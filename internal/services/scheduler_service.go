package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"sync"
	"sync/atomic"

	"github.com/redis/go-redis/v9"
	"github.com/russo2642/renti_kz/internal/config"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type SchedulerService struct {
	redisClient         *redis.Client
	db                  *sql.DB
	bookingRepo         domain.BookingRepository
	apartmentRepo       domain.ApartmentRepository
	availabilityService domain.ApartmentAvailabilityService
	notificationUseCase domain.NotificationUseCase
	lockUseCase         domain.LockUseCase
	renterRepo          domain.RenterRepository
	propertyOwnerRepo   domain.PropertyOwnerRepository
	userUseCase         domain.UserUseCase
	chatUseCase         domain.ChatUseCase
	chatRoomRepo        domain.ChatRoomRepository
	paymentRepo         domain.PaymentRepository
	paymentUseCase      domain.PaymentUseCase
	config              config.RedisConfig
	isRunning           bool
	stopChan            chan struct{}

	workerPool chan struct{}
	metrics    *SchedulerMetrics
}

type SchedulerMetrics struct {
	TasksProcessed       int64
	TasksSkipped         int64
	ProcessingTimeMs     int64
	LastProcessingTime   time.Time
	ActiveWorkers        int32
	TotalErrors          int64
	DatabaseQueriesCount int64
}

type ScheduledTask struct {
	Type        string                 `json:"type"` // "activate_booking", "complete_booking", "send_reminder"
	BookingID   int                    `json:"booking_id"`
	Data        map[string]interface{} `json:"data"`
	ScheduledAt time.Time              `json:"scheduled_at"`
}

const (
	TaskActivateBooking   = "activate_booking"
	TaskCompleteBooking   = "complete_booking"
	TaskSendReminder      = "send_reminder"
	TaskOpenChat          = "open_chat"
	TaskCloseChat         = "close_chat"
	TaskCleanupBookings   = "cleanup_expired_bookings"
	TaskCleanupExtensions = "cleanup_expired_extensions"

	SchedulerLockKey     = "scheduler:lock"
	SchedulerInstanceKey = "scheduler:instance"
	TaskQueueKey         = "scheduler:tasks"
	ProcessedTasksKey    = "scheduler:processed"
)

func NewSchedulerService(
	redisConfig config.RedisConfig,
	db *sql.DB,
	bookingRepo domain.BookingRepository,
	apartmentRepo domain.ApartmentRepository,
	availabilityService domain.ApartmentAvailabilityService,
	notificationUseCase domain.NotificationUseCase,
	lockUseCase domain.LockUseCase,
	renterRepo domain.RenterRepository,
	propertyOwnerRepo domain.PropertyOwnerRepository,
	userUseCase domain.UserUseCase,
	chatUseCase domain.ChatUseCase,
	chatRoomRepo domain.ChatRoomRepository,
	paymentRepo domain.PaymentRepository,
	paymentUseCase domain.PaymentUseCase,
) *SchedulerService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr(),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	return &SchedulerService{
		redisClient:         rdb,
		db:                  db,
		bookingRepo:         bookingRepo,
		apartmentRepo:       apartmentRepo,
		availabilityService: availabilityService,
		notificationUseCase: notificationUseCase,
		lockUseCase:         lockUseCase,
		renterRepo:          renterRepo,
		propertyOwnerRepo:   propertyOwnerRepo,
		userUseCase:         userUseCase,
		chatUseCase:         chatUseCase,
		chatRoomRepo:        chatRoomRepo,
		paymentRepo:         paymentRepo,
		paymentUseCase:      paymentUseCase,
		config:              redisConfig,
		stopChan:            make(chan struct{}),
		workerPool:          make(chan struct{}, 50),
		metrics:             &SchedulerMetrics{},
	}
}

func (s *SchedulerService) StartScheduler() {
	if s.isRunning {
		log.Println("⚠️ Scheduler уже запущен")
		return
	}

	s.isRunning = true
	instanceID := fmt.Sprintf("scheduler_%d", time.Now().Unix())

	log.Printf("🚀 Запускаем Redis Scheduler (instance: %s)", instanceID)

	ctx := context.Background()
	s.redisClient.Set(ctx, SchedulerInstanceKey, instanceID, time.Minute*5)

	ticker := time.NewTicker(30 * time.Second)
	selfCheckTicker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	defer selfCheckTicker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processScheduledTasks(ctx)
			s.scheduleNewTasks(ctx)
		case <-selfCheckTicker.C:
			s.performSelfCheck(ctx)
		case <-s.stopChan:
			log.Println("🛑 Остановка Redis Scheduler")
			s.isRunning = false
			return
		}
	}
}

func (s *SchedulerService) StopScheduler() {
	if s.isRunning {
		close(s.stopChan)
	}
}

func (s *SchedulerService) processScheduledTasks(ctx context.Context) {
	if !s.acquireLock(ctx) {
		return
	}
	defer s.releaseLock(ctx)

	startTime := time.Now()
	now := utils.GetCurrentTimeUTC()

	tasks, err := s.getTasksToProcess(ctx, now)
	if err != nil {
		log.Printf("❌ Ошибка получения задач: %v", err)
		s.metrics.TotalErrors++
		return
	}

	if len(tasks) > 0 {
		log.Printf("📋 Обрабатываем %d задач на %s UTC", len(tasks), now.Format("15:04:05"))
		s.processTasksConcurrently(ctx, tasks)
	}

	processingTime := time.Since(startTime)
	s.metrics.ProcessingTimeMs = processingTime.Milliseconds()
	s.metrics.LastProcessingTime = time.Now()
}

func (s *SchedulerService) processTasksConcurrently(ctx context.Context, tasks []ScheduledTask) {
	if len(tasks) == 0 {
		return
	}

	if len(tasks) <= 3 {
		for _, task := range tasks {
			s.executeTaskWithMetrics(ctx, task)
		}
		return
	}

	var wg sync.WaitGroup
	taskChan := make(chan ScheduledTask, len(tasks))

	workerCount := minInt(len(tasks), cap(s.workerPool))
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				s.executeTaskWithMetrics(ctx, task)
			}
		}()
	}

	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	wg.Wait()

	log.Printf("✅ Все задачи обработаны параллельно (%d воркеров)", workerCount)
}

func (s *SchedulerService) executeTaskWithMetrics(ctx context.Context, task ScheduledTask) {
	atomic.AddInt32(&s.metrics.ActiveWorkers, 1)
	defer atomic.AddInt32(&s.metrics.ActiveWorkers, -1)

	startTime := time.Now()
	s.executeTask(ctx, task)

	atomic.AddInt64(&s.metrics.TasksProcessed, 1)
	log.Printf("⚡ Задача %s-%d выполнена за %v", task.Type, task.BookingID, time.Since(startTime))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *SchedulerService) scheduleNewTasks(ctx context.Context) {
	startTime := time.Now()

	processedTasks, err := s.redisClient.SMembers(ctx, ProcessedTasksKey).Result()
	if err != nil {
		log.Printf("❌ Ошибка получения processed tasks: %v", err)
		s.metrics.TotalErrors++
		return
	}

	processedSet := make(map[string]bool, len(processedTasks))
	for _, task := range processedTasks {
		processedSet[task] = true
	}

	const maxBookingsPerBatch = 500

	approvedBookings, err := s.getBookingsBatch([]domain.BookingStatus{domain.BookingStatusApproved}, maxBookingsPerBatch)
	if err != nil {
		log.Printf("❌ Ошибка получения одобренных бронирований: %v", err)
		s.metrics.TotalErrors++
		return
	}
	s.metrics.DatabaseQueriesCount++

	activeBookings, err := s.getBookingsBatch([]domain.BookingStatus{domain.BookingStatusActive}, maxBookingsPerBatch)
	if err != nil {
		log.Printf("❌ Ошибка получения активных бронирований: %v", err)
		s.metrics.TotalErrors++
		return
	}
	s.metrics.DatabaseQueriesCount++

	s.scheduleTasksBatch(ctx, approvedBookings, processedSet, "approved")

	s.scheduleTasksBatch(ctx, activeBookings, processedSet, "active")

	s.scheduleCleanupTasks(ctx, processedSet)

	log.Printf("📊 Планирование задач завершено за %v (approved: %d, active: %d)",
		time.Since(startTime), len(approvedBookings), len(activeBookings))
}

func (s *SchedulerService) getBookingsBatch(statuses []domain.BookingStatus, limit int) ([]*domain.Booking, error) {
	// TODO: Добавить поддержку LIMIT в BookingRepository.GetByStatus для оптимизации
	bookings, err := s.bookingRepo.GetByStatus(statuses)
	if err != nil {
		return nil, err
	}

	if len(bookings) > limit {
		log.Printf("⚠️ Слишком много бронирований (%d), ограничиваем до %d", len(bookings), limit)
		bookings = bookings[:limit]
	}

	return bookings, nil
}

func (s *SchedulerService) scheduleTasksBatch(ctx context.Context, bookings []*domain.Booking, processedSet map[string]bool, bookingType string) {
	if len(bookings) == 0 {
		return
	}

	tasksScheduled := 0
	tasksSkipped := 0

	for _, booking := range bookings {
		if bookingType == "approved" {
			if s.scheduleActivationTaskOptimized(ctx, booking, processedSet) {
				tasksScheduled++
			} else {
				tasksSkipped++
			}

			if s.scheduleReminderTasksOptimized(ctx, booking, processedSet) {
				tasksScheduled++
			}

			if s.scheduleChatTasksOptimized(ctx, booking, processedSet) {
				tasksScheduled++
			}
		} else if bookingType == "active" {
			if s.scheduleCompletionTaskOptimized(ctx, booking, processedSet) {
				tasksScheduled++
			} else {
				tasksSkipped++
			}
		}
	}

	atomic.AddInt64(&s.metrics.TasksSkipped, int64(tasksSkipped))
	log.Printf("📋 %s booking tasks: запланировано %d, пропущено %d", bookingType, tasksScheduled, tasksSkipped)
}

func (s *SchedulerService) scheduleActivationTask(ctx context.Context, booking *domain.Booking) {
	taskKey := fmt.Sprintf("activate_%d", booking.ID)

	exists, _ := s.redisClient.SIsMember(ctx, ProcessedTasksKey, taskKey).Result()
	if exists {
		return
	}

	task := ScheduledTask{
		Type:        TaskActivateBooking,
		BookingID:   booking.ID,
		ScheduledAt: booking.StartDate,
		Data: map[string]interface{}{
			"apartment_id": booking.ApartmentID,
			"renter_id":    booking.RenterID,
		},
	}

	s.scheduleTask(ctx, task, booking.StartDate)
}

func (s *SchedulerService) scheduleCompletionTask(ctx context.Context, booking *domain.Booking) {
	taskKey := fmt.Sprintf("complete_%d", booking.ID)

	exists, _ := s.redisClient.SIsMember(ctx, ProcessedTasksKey, taskKey).Result()
	if exists {
		return
	}

	endDate := booking.EndDate
	if booking.ExtensionEndDate != nil && !booking.ExtensionRequested {
		endDate = *booking.ExtensionEndDate
	}

	task := ScheduledTask{
		Type:        TaskCompleteBooking,
		BookingID:   booking.ID,
		ScheduledAt: endDate,
		Data: map[string]interface{}{
			"apartment_id": booking.ApartmentID,
			"end_date":     endDate.Format(time.RFC3339),
		},
	}

	s.scheduleTask(ctx, task, endDate)
}

func (s *SchedulerService) scheduleReminderTasks(ctx context.Context, booking *domain.Booking) {
	now := utils.GetCurrentTimeUTC()

	reminderTime := booking.StartDate.Add(-time.Hour)
	if reminderTime.After(now) {
		task := ScheduledTask{
			Type:        TaskSendReminder,
			BookingID:   booking.ID,
			ScheduledAt: reminderTime,
			Data: map[string]interface{}{
				"reminder_type": "starting_soon",
				"renter_id":     booking.RenterID,
			},
		}
		s.scheduleTask(ctx, task, reminderTime)
	}

	duration := booking.EndDate.Sub(booking.StartDate)

	if duration.Hours() >= 24 {
		endReminderTime := booking.EndDate.Add(-2 * time.Hour)
		if endReminderTime.After(now) {
			task := ScheduledTask{
				Type:        TaskSendReminder,
				BookingID:   booking.ID,
				ScheduledAt: endReminderTime,
				Data: map[string]interface{}{
					"reminder_type": "ending_soon",
					"renter_id":     booking.RenterID,
				},
			}
			s.scheduleTask(ctx, task, endReminderTime)
		}
	} else if duration.Hours() >= 10 {
		endReminderTime := booking.EndDate.Add(-time.Hour)
		if endReminderTime.After(now) {
			task := ScheduledTask{
				Type:        TaskSendReminder,
				BookingID:   booking.ID,
				ScheduledAt: endReminderTime,
				Data: map[string]interface{}{
					"reminder_type": "ending_soon",
					"renter_id":     booking.RenterID,
				},
			}
			s.scheduleTask(ctx, task, endReminderTime)
		}
	} else {
		endReminderTime := booking.EndDate.Add(-30 * time.Minute)
		if endReminderTime.After(now) {
			task := ScheduledTask{
				Type:        TaskSendReminder,
				BookingID:   booking.ID,
				ScheduledAt: endReminderTime,
				Data: map[string]interface{}{
					"reminder_type": "ending_soon",
					"renter_id":     booking.RenterID,
				},
			}
			s.scheduleTask(ctx, task, endReminderTime)
		}
	}
}

func (s *SchedulerService) scheduleChatTasks(ctx context.Context, booking *domain.Booking) {
	now := utils.GetCurrentTimeUTC()

	chatOpenTime := booking.StartDate.Add(-15 * time.Minute)
	if chatOpenTime.After(now) {
		openChatTaskKey := fmt.Sprintf("open_chat_%d", booking.ID)
		exists, _ := s.redisClient.SIsMember(ctx, ProcessedTasksKey, openChatTaskKey).Result()
		if !exists {
			openTask := ScheduledTask{
				Type:        TaskOpenChat,
				BookingID:   booking.ID,
				ScheduledAt: chatOpenTime,
				Data: map[string]interface{}{
					"apartment_id": booking.ApartmentID,
					"renter_id":    booking.RenterID,
				},
			}
			s.scheduleTask(ctx, openTask, chatOpenTime)
		}
	}

	endDate := booking.EndDate
	if booking.ExtensionEndDate != nil && !booking.ExtensionRequested {
		endDate = *booking.ExtensionEndDate
	}

	chatCloseTime := endDate.Add(24 * time.Hour)
	closeChatTaskKey := fmt.Sprintf("close_chat_%d", booking.ID)
	exists, _ := s.redisClient.SIsMember(ctx, ProcessedTasksKey, closeChatTaskKey).Result()
	if !exists {
		closeTask := ScheduledTask{
			Type:        TaskCloseChat,
			BookingID:   booking.ID,
			ScheduledAt: chatCloseTime,
			Data: map[string]interface{}{
				"apartment_id": booking.ApartmentID,
				"renter_id":    booking.RenterID,
			},
		}
		s.scheduleTask(ctx, closeTask, chatCloseTime)
	}
}

func (s *SchedulerService) scheduleCleanupTasks(ctx context.Context, processedSet map[string]bool) {
	now := time.Now()

	if now.Hour()%2 == 0 && now.Minute() < 5 {
		cleanupBookingsKey := fmt.Sprintf("cleanup_bookings_%s", now.Format("2006010215"))
		if !processedSet[cleanupBookingsKey] {
			cleanupBookingsTask := ScheduledTask{
				Type:        TaskCleanupBookings,
				BookingID:   0,
				ScheduledAt: now,
				Data: map[string]interface{}{
					"batch_size": 1000,
				},
			}
			s.scheduleTask(ctx, cleanupBookingsTask, now)
			log.Printf("🧹 Запланирована очистка просроченных бронирований на %s", now.Format("15:04:05"))
		}

		cleanupExtensionsKey := fmt.Sprintf("cleanup_extensions_%s", now.Format("2006010215"))
		if !processedSet[cleanupExtensionsKey] {
			cleanupExtensionsTask := ScheduledTask{
				Type:        TaskCleanupExtensions,
				BookingID:   0,
				ScheduledAt: now,
				Data: map[string]interface{}{
					"batch_size": 1000,
				},
			}
			s.scheduleTask(ctx, cleanupExtensionsTask, now)
			log.Printf("🧹 Запланирована очистка просроченных продлений на %s", now.Format("15:04:05"))
		}
	}
}

func (s *SchedulerService) scheduleTask(ctx context.Context, task ScheduledTask, executeAt time.Time) {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		log.Printf("❌ Ошибка сериализации задачи: %v", err)
		return
	}

	score := float64(executeAt.Unix())

	err = s.redisClient.ZAdd(ctx, TaskQueueKey, redis.Z{
		Score:  score,
		Member: string(taskJSON),
	}).Err()

	if err != nil {
		log.Printf("❌ Ошибка добавления задачи в Redis: %v", err)
		return
	}

	log.Printf("📅 Запланирована задача %s для бронирования %d на %s UTC",
		task.Type, task.BookingID, executeAt.Format("2006-01-02 15:04:05"))
}

func (s *SchedulerService) getTasksToProcess(ctx context.Context, now time.Time) ([]ScheduledTask, error) {
	results, err := s.redisClient.ZRangeByScore(ctx, TaskQueueKey, &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(now.Unix(), 10),
	}).Result()

	if err != nil {
		return nil, err
	}

	var tasks []ScheduledTask
	for _, result := range results {
		var task ScheduledTask
		if err := json.Unmarshal([]byte(result), &task); err != nil {
			log.Printf("❌ Ошибка десериализации задачи: %v", err)
			continue
		}
		tasks = append(tasks, task)
	}

	if len(results) > 0 {
		s.redisClient.ZRem(ctx, TaskQueueKey, results)
	}

	return tasks, nil
}

func (s *SchedulerService) executeTask(ctx context.Context, task ScheduledTask) {
	var taskKey string

	if task.Type == TaskCleanupBookings || task.Type == TaskCleanupExtensions {
		taskKey = fmt.Sprintf("%s_%s", task.Type, task.ScheduledAt.Format("2006010215"))
	} else {
		taskKey = fmt.Sprintf("%s_%d", task.Type, task.BookingID)
	}

	exists, _ := s.redisClient.SIsMember(ctx, ProcessedTasksKey, taskKey).Result()
	if exists {
		log.Printf("⏭️ Задача %s уже выполнена", taskKey)
		return
	}

	if task.Type == TaskCleanupBookings || task.Type == TaskCleanupExtensions {
		log.Printf("⚡ Выполняем задачу очистки: %s", task.Type)
	} else {
		log.Printf("⚡ Выполняем задачу: %s для бронирования %d", task.Type, task.BookingID)
	}

	switch task.Type {
	case TaskActivateBooking:
		s.executeActivateBooking(ctx, task)
	case TaskCompleteBooking:
		s.executeCompleteBooking(ctx, task)
	case TaskSendReminder:
		s.executeSendReminder(ctx, task)
	case TaskOpenChat:
		s.executeOpenChat(ctx, task)
	case TaskCloseChat:
		s.executeCloseChat(ctx, task)
	case TaskCleanupBookings:
		s.executeCleanupBookings(ctx, task)
	case TaskCleanupExtensions:
		s.executeCleanupExtensions(ctx, task)
	default:
		log.Printf("⚠️ Неизвестный тип задачи: %s", task.Type)
		return
	}

	s.redisClient.SAdd(ctx, ProcessedTasksKey, taskKey)
	s.redisClient.Expire(ctx, ProcessedTasksKey, time.Hour*48)
}

func (s *SchedulerService) executeActivateBooking(_ context.Context, task ScheduledTask) {
	booking, err := s.bookingRepo.GetByID(task.BookingID)
	if err != nil {
		log.Printf("❌ Ошибка получения бронирования %d: %v", task.BookingID, err)
		return
	}

	if booking.Status != domain.BookingStatusApproved {
		log.Printf("⚠️ Бронирование %d не в статусе approved (текущий: %s)", task.BookingID, booking.Status)
		return
	}

	log.Printf("🚀 Активируем бронирование %d", task.BookingID)

	booking.Status = domain.BookingStatusActive
	err = s.bookingRepo.Update(booking)
	if err != nil {
		log.Printf("❌ Ошибка обновления бронирования %d: %v", task.BookingID, err)
		return
	}

	if s.availabilityService != nil {
		if err := s.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			log.Printf("❌ Ошибка пересчета доступности квартиры %d: %v", booking.ApartmentID, err)
		}
	} else {
		err = s.apartmentRepo.UpdateIsFree(booking.ApartmentID, false)
		if err != nil {
			log.Printf("❌ Ошибка обновления статуса квартиры %d: %v", booking.ApartmentID, err)
		}
	}

	apartment, _ := s.apartmentRepo.GetByID(booking.ApartmentID)
	apartmentTitle := "квартира"
	if apartment != nil {
		apartmentTitle = fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)
	}

	renter, renterErr := s.renterRepo.GetByID(booking.RenterID)
	if renterErr == nil && renter != nil {
		s.notificationUseCase.NotifyRenterBookingStarted(renter.UserID, booking.ID, apartmentTitle)
	}

	if apartment != nil {
		renter, _ := s.renterRepo.GetByID(booking.RenterID)
		renterName := "арендатор"

		if renter != nil {
			renterUser, userErr := s.userUseCase.GetByID(renter.UserID)
			if userErr == nil && renterUser != nil {
				renterName = fmt.Sprintf("%s %s", renterUser.FirstName, renterUser.LastName)
			}
		}

		propertyOwner, _ := s.propertyOwnerRepo.GetByID(apartment.OwnerID)
		if propertyOwner != nil {
			s.notificationUseCase.NotifyBookingStarted(propertyOwner.UserID, booking.ID, apartmentTitle, renterName)
		}
	}

	if s.chatUseCase != nil && renter != nil {
		chatRoom, err := s.chatRoomRepo.GetByBookingID(booking.ID)
		if err == nil && chatRoom != nil && chatRoom.Status == domain.ChatRoomStatusPending {
			err = s.chatUseCase.ActivateChat(chatRoom.ID, renter.UserID)
			if err != nil {
				log.Printf("❌ Ошибка активации чата для бронирования %d: %v", task.BookingID, err)
			} else {
				log.Printf("💬 Чат для бронирования %d успешно активирован", task.BookingID)
			}
		}
	}

	log.Printf("✅ Бронирование %d успешно активировано", task.BookingID)
}

func (s *SchedulerService) executeCompleteBooking(_ context.Context, task ScheduledTask) {
	booking, err := s.bookingRepo.GetByID(task.BookingID)
	if err != nil {
		log.Printf("❌ Ошибка получения бронирования %d: %v", task.BookingID, err)
		return
	}

	if booking.Status != domain.BookingStatusActive {
		log.Printf("⚠️ Бронирование %d не активно (текущий статус: %s)", task.BookingID, booking.Status)
		return
	}

	// ОПТИМИЗАЦИЯ: проверяем extensions только если есть активный запрос на продление
	if booking.ExtensionRequested {
		extensions, err := s.bookingRepo.GetExtensionsByBookingID(booking.ID)
		if err == nil {
			now := time.Now()
			gracePeriodExpired := false

			for _, ext := range extensions {
				if ext.Status == domain.BookingStatusPending {
					if now.Sub(ext.RequestedAt) > 30*time.Minute {
						log.Printf("⏰ Grace period истёк для продления %d, возвращаем деньги", ext.ID)
						gracePeriodExpired = true

						if err := s.refundExtensionPayment(ext, booking); err != nil {
							log.Printf("❌ Ошибка возврата за продление %d: %v", ext.ID, err)
						} else {
							log.Printf("💸 Возврат выполнен для продления %d", ext.ID)
						}
					} else {
						log.Printf("⏸️ Откладываем завершение бронирования %d - есть pending продление %d", booking.ID, ext.ID)

						remainingTime := 30*time.Minute - now.Sub(ext.RequestedAt)
						gracePeriod := now.Add(remainingTime)
						s.RescheduleCompletionTask(booking.ID, gracePeriod)
						return
					}
				}
			}

			if gracePeriodExpired {
				booking, err = s.bookingRepo.GetByID(task.BookingID)
				if err != nil {
					log.Printf("❌ Ошибка перезагрузки бронирования %d: %v", task.BookingID, err)
					return
				}
			}
		}
	}

	log.Printf("🏁 Завершаем бронирование %d", task.BookingID)

	booking.Status = domain.BookingStatusCompleted
	booking.DoorStatus = domain.DoorStatusClosed

	err = s.bookingRepo.Update(booking)
	if err != nil {
		log.Printf("❌ Ошибка завершения бронирования %d: %v", task.BookingID, err)
		return
	}

	if s.availabilityService != nil {
		if err := s.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			log.Printf("❌ Ошибка пересчета доступности при освобождении квартиры %d: %v", booking.ApartmentID, err)
		}
	} else {
		err = s.apartmentRepo.UpdateIsFree(booking.ApartmentID, true)
		if err != nil {
			log.Printf("❌ Ошибка освобождения квартиры %d: %v", booking.ApartmentID, err)
		}
	}

	if s.lockUseCase != nil {
		err = s.lockUseCase.DeactivatePasswordForBooking(booking.ID)
		if err != nil {
			log.Printf("❌ Ошибка деактивации временного пароля: %v", err)
		}
	}

	log.Printf("✅ Бронирование %d успешно завершено", task.BookingID)
}

func (s *SchedulerService) executeSendReminder(_ context.Context, task ScheduledTask) {
	reminderType, ok := task.Data["reminder_type"].(string)
	if !ok {
		log.Printf("❌ Неверный тип напоминания в задаче")
		return
	}

	booking, err := s.bookingRepo.GetByID(task.BookingID)
	if err != nil {
		log.Printf("❌ Ошибка получения бронирования %d: %v", task.BookingID, err)
		return
	}

	apartment, _ := s.apartmentRepo.GetByID(booking.ApartmentID)
	apartmentTitle := "квартира"
	if apartment != nil {
		apartmentTitle = fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)
	}

	renter, renterErr := s.renterRepo.GetByID(booking.RenterID)
	if renterErr != nil || renter == nil {
		log.Printf("❌ Ошибка получения арендатора для уведомления: %v", renterErr)
		return
	}

	switch reminderType {
	case "starting_soon":
		duration := time.Until(booking.StartDate)
		s.notificationUseCase.NotifyBookingStartingSoon(renter.UserID, booking.ID, apartmentTitle, duration)
		log.Printf("📢 Отправлено напоминание о начале бронирования %d", task.BookingID)

	case "ending_soon":
		s.notificationUseCase.NotifyBookingEnding(renter.UserID, booking.ID, apartmentTitle)
		log.Printf("📢 Отправлено напоминание об окончании бронирования %d", task.BookingID)
	}
}

func (s *SchedulerService) executeOpenChat(_ context.Context, task ScheduledTask) {
	if s.chatUseCase == nil {
		log.Printf("⚠️ ChatUseCase не инициализирован")
		return
	}

	log.Printf("💬 Открываем чат для бронирования %d", task.BookingID)

	err := s.chatUseCase.OpenScheduledChats()
	if err != nil {
		log.Printf("❌ Ошибка открытия чатов: %v", err)
		return
	}

	log.Printf("✅ Чаты успешно открыты для готовых бронирований")
}

func (s *SchedulerService) executeCloseChat(_ context.Context, task ScheduledTask) {
	if s.chatUseCase == nil {
		log.Printf("⚠️ ChatUseCase не инициализирован")
		return
	}

	log.Printf("💬 Закрываем просроченные чаты для бронирования %d", task.BookingID)

	err := s.chatUseCase.CloseExpiredChats()
	if err != nil {
		log.Printf("❌ Ошибка закрытия чатов: %v", err)
		return
	}

	log.Printf("✅ Просроченные чаты успешно закрыты")
}

func (s *SchedulerService) acquireLock(ctx context.Context) bool {
	result, err := s.redisClient.SetNX(ctx, SchedulerLockKey, "locked", time.Minute).Result()
	if err != nil {
		log.Printf("❌ Ошибка получения блокировки: %v", err)
		return false
	}
	return result
}

func (s *SchedulerService) releaseLock(ctx context.Context) {
	s.redisClient.Del(ctx, SchedulerLockKey)
}

func (s *SchedulerService) RemoveScheduledTasksForBooking(bookingID int) error {
	ctx := context.Background()

	log.Printf("🗑️ Удаляем запланированные задачи для бронирования %d", bookingID)

	pipe := s.redisClient.Pipeline()

	results, err := s.redisClient.ZRange(ctx, TaskQueueKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("ошибка получения задач из очереди: %w", err)
	}

	var tasksToRemove []string
	var processedKeys []string
	removedCount := 0

	for _, result := range results {
		var task ScheduledTask
		if err := json.Unmarshal([]byte(result), &task); err != nil {
			continue
		}

		if task.BookingID == bookingID {
			tasksToRemove = append(tasksToRemove, result)

			taskKey := fmt.Sprintf("%s_%d", task.Type, bookingID)
			processedKeys = append(processedKeys, taskKey)

			log.Printf("🗑️ Помечаем для удаления задачу %s для бронирования %d", task.Type, bookingID)
			removedCount++
		}
	}

	if len(tasksToRemove) > 0 {
		pipe.ZRem(ctx, TaskQueueKey, tasksToRemove)

		if len(processedKeys) > 0 {
			pipe.SAdd(ctx, ProcessedTasksKey, processedKeys)
			pipe.Expire(ctx, ProcessedTasksKey, time.Hour*48)
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("ошибка выполнения batch операций: %w", err)
		}
	}

	log.Printf("✅ Удалено %d задач для бронирования %d", removedCount, bookingID)
	atomic.AddInt64(&s.metrics.TasksProcessed, int64(removedCount))
	return nil
}

func (s *SchedulerService) RescheduleCompletionTask(bookingID int, newEndDate time.Time) error {
	ctx := context.Background()

	log.Printf("🔄 Переносим задачу завершения для бронирования %d на %s", bookingID, newEndDate.Format("2006-01-02 15:04:05"))

	err := s.RemoveScheduledTasksForBooking(bookingID)
	if err != nil {
		log.Printf("⚠️ Ошибка удаления старых задач: %v", err)
	}

	task := ScheduledTask{
		Type:        TaskCompleteBooking,
		BookingID:   bookingID,
		ScheduledAt: newEndDate,
		Data: map[string]interface{}{
			"booking_id": bookingID,
			"end_date":   newEndDate.Format(time.RFC3339),
		},
	}

	s.scheduleTask(ctx, task, newEndDate)

	log.Printf("✅ Задача завершения перенесена на %s", newEndDate.Format("2006-01-02 15:04:05"))
	return nil
}

func (s *SchedulerService) GetSchedulerStats(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	queueSize, _ := s.redisClient.ZCard(ctx, TaskQueueKey).Result()
	stats["queue_size"] = queueSize

	processedSize, _ := s.redisClient.SCard(ctx, ProcessedTasksKey).Result()
	stats["processed_count"] = processedSize

	instance, _ := s.redisClient.Get(ctx, SchedulerInstanceKey).Result()
	stats["instance"] = instance
	stats["is_running"] = s.isRunning

	stats["tasks_processed"] = atomic.LoadInt64(&s.metrics.TasksProcessed)
	stats["tasks_skipped"] = atomic.LoadInt64(&s.metrics.TasksSkipped)
	stats["processing_time_ms"] = s.metrics.ProcessingTimeMs
	stats["last_processing_time"] = s.metrics.LastProcessingTime
	stats["active_workers"] = atomic.LoadInt32(&s.metrics.ActiveWorkers)
	stats["total_errors"] = atomic.LoadInt64(&s.metrics.TotalErrors)
	stats["database_queries_count"] = atomic.LoadInt64(&s.metrics.DatabaseQueriesCount)
	stats["worker_pool_capacity"] = cap(s.workerPool)

	return stats
}

func (s *SchedulerService) GetMetrics() *SchedulerMetrics {
	return &SchedulerMetrics{
		TasksProcessed:       atomic.LoadInt64(&s.metrics.TasksProcessed),
		TasksSkipped:         atomic.LoadInt64(&s.metrics.TasksSkipped),
		ProcessingTimeMs:     s.metrics.ProcessingTimeMs,
		LastProcessingTime:   s.metrics.LastProcessingTime,
		ActiveWorkers:        atomic.LoadInt32(&s.metrics.ActiveWorkers),
		TotalErrors:          atomic.LoadInt64(&s.metrics.TotalErrors),
		DatabaseQueriesCount: atomic.LoadInt64(&s.metrics.DatabaseQueriesCount),
	}
}

func (s *SchedulerService) scheduleActivationTaskOptimized(ctx context.Context, booking *domain.Booking, processedSet map[string]bool) bool {
	taskKey := fmt.Sprintf("activate_%d", booking.ID)

	if processedSet[taskKey] {
		return false
	}

	task := ScheduledTask{
		Type:        TaskActivateBooking,
		BookingID:   booking.ID,
		ScheduledAt: booking.StartDate,
		Data: map[string]interface{}{
			"apartment_id": booking.ApartmentID,
			"renter_id":    booking.RenterID,
		},
	}

	s.scheduleTask(ctx, task, booking.StartDate)
	return true
}

func (s *SchedulerService) scheduleCompletionTaskOptimized(ctx context.Context, booking *domain.Booking, processedSet map[string]bool) bool {
	taskKey := fmt.Sprintf("complete_%d", booking.ID)

	if processedSet[taskKey] {
		return false
	}

	endDate := booking.EndDate
	if booking.ExtensionEndDate != nil && !booking.ExtensionRequested {
		endDate = *booking.ExtensionEndDate
	}

	task := ScheduledTask{
		Type:        TaskCompleteBooking,
		BookingID:   booking.ID,
		ScheduledAt: endDate,
		Data: map[string]interface{}{
			"apartment_id": booking.ApartmentID,
			"end_date":     endDate.Format(time.RFC3339),
		},
	}

	s.scheduleTask(ctx, task, endDate)
	return true
}

func (s *SchedulerService) scheduleReminderTasksOptimized(ctx context.Context, booking *domain.Booking, processedSet map[string]bool) bool {
	now := utils.GetCurrentTimeUTC()
	tasksScheduled := 0

	reminderTime := booking.StartDate.Add(-time.Hour)
	if reminderTime.After(now) {
		startReminderKey := fmt.Sprintf("reminder_start_%d", booking.ID)
		if !processedSet[startReminderKey] {
			task := ScheduledTask{
				Type:        TaskSendReminder,
				BookingID:   booking.ID,
				ScheduledAt: reminderTime,
				Data: map[string]interface{}{
					"reminder_type": "starting_soon",
					"renter_id":     booking.RenterID,
				},
			}
			s.scheduleTask(ctx, task, reminderTime)
			tasksScheduled++
		}
	}

	duration := booking.EndDate.Sub(booking.StartDate)
	var endReminderTime time.Time

	if duration.Hours() >= 24 {
		endReminderTime = booking.EndDate.Add(-2 * time.Hour)
	} else if duration.Hours() >= 10 {
		endReminderTime = booking.EndDate.Add(-time.Hour)
	} else {
		endReminderTime = booking.EndDate.Add(-30 * time.Minute)
	}

	if endReminderTime.After(now) {
		endReminderKey := fmt.Sprintf("reminder_end_%d", booking.ID)
		if !processedSet[endReminderKey] {
			task := ScheduledTask{
				Type:        TaskSendReminder,
				BookingID:   booking.ID,
				ScheduledAt: endReminderTime,
				Data: map[string]interface{}{
					"reminder_type": "ending_soon",
					"renter_id":     booking.RenterID,
				},
			}
			s.scheduleTask(ctx, task, endReminderTime)
			tasksScheduled++
		}
	}

	return tasksScheduled > 0
}

func (s *SchedulerService) scheduleChatTasksOptimized(ctx context.Context, booking *domain.Booking, processedSet map[string]bool) bool {
	now := utils.GetCurrentTimeUTC()
	tasksScheduled := 0

	chatOpenTime := booking.StartDate.Add(-15 * time.Minute)
	if chatOpenTime.After(now) {
		openChatTaskKey := fmt.Sprintf("open_chat_%d", booking.ID)
		if !processedSet[openChatTaskKey] {
			openTask := ScheduledTask{
				Type:        TaskOpenChat,
				BookingID:   booking.ID,
				ScheduledAt: chatOpenTime,
				Data: map[string]interface{}{
					"apartment_id": booking.ApartmentID,
					"renter_id":    booking.RenterID,
				},
			}
			s.scheduleTask(ctx, openTask, chatOpenTime)
			tasksScheduled++
		}
	}

	endDate := booking.EndDate
	if booking.ExtensionEndDate != nil && !booking.ExtensionRequested {
		endDate = *booking.ExtensionEndDate
	}

	chatCloseTime := endDate.Add(24 * time.Hour)
	closeChatTaskKey := fmt.Sprintf("close_chat_%d", booking.ID)
	if !processedSet[closeChatTaskKey] {
		closeTask := ScheduledTask{
			Type:        TaskCloseChat,
			BookingID:   booking.ID,
			ScheduledAt: chatCloseTime,
			Data: map[string]interface{}{
				"apartment_id": booking.ApartmentID,
				"renter_id":    booking.RenterID,
			},
		}
		s.scheduleTask(ctx, closeTask, chatCloseTime)
		tasksScheduled++
	}

	return tasksScheduled > 0
}

func (s *SchedulerService) refundExtensionPayment(extension *domain.BookingExtension, booking *domain.Booking) error {
	extension.Status = domain.BookingStatusRejected

	err := s.bookingRepo.UpdateExtension(extension)
	if err != nil {
		return fmt.Errorf("ошибка обновления продления: %w", err)
	}

	booking.ExtensionRequested = false
	booking.ExtensionEndDate = nil
	booking.ExtensionDuration = 0
	booking.ExtensionPrice = 0

	err = s.bookingRepo.Update(booking)
	if err != nil {
		return fmt.Errorf("ошибка обновления бронирования: %w", err)
	}

	if extension.PaymentID != nil {
		paymentRecord, err := s.paymentRepo.GetByID(*extension.PaymentID)
		if err != nil {
			log.Printf("❌ Ошибка получения платежа ID: %d для возврата: %v", *extension.PaymentID, err)
		} else {
			log.Printf("💳 Выполняем возврат платежа %s для продления %d", paymentRecord.PaymentID, extension.ID)

			refundResponse, refundErr := s.paymentUseCase.RefundPayment(paymentRecord.PaymentID, nil)
			if refundErr != nil {
				log.Printf("❌ Ошибка возврата платежа %s: %v", paymentRecord.PaymentID, refundErr)
			} else if refundResponse.Success {
				log.Printf("💸 Возврат успешно выполнен для платежа %s", paymentRecord.PaymentID)
			} else {
				log.Printf("⚠️ Возврат не удался для платежа %s", paymentRecord.PaymentID)
			}
		}
	}

	if s.notificationUseCase != nil {
		apartment, err := s.apartmentRepo.GetByID(booking.ApartmentID)
		if err == nil && apartment != nil {
			apartmentTitle := fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)

			renter, err := s.renterRepo.GetByID(booking.RenterID)
			if err == nil && renter != nil {
				notifyErr := s.notificationUseCase.NotifyExtensionTimeoutRefund(
					renter.UserID,
					booking.ID,
					apartmentTitle,
					extension.Duration,
				)
				if notifyErr != nil {
					log.Printf("⚠️ Ошибка отправки уведомления о возврате: %v", notifyErr)
				}
			}
		}
	}

	log.Printf("💸 Продление %d отклонено по таймауту, выполнен возврат", extension.ID)

	return nil
}

func (s *SchedulerService) executeCleanupBookings(_ context.Context, task ScheduledTask) {
	batchSizeInterface, exists := task.Data["batch_size"]
	if !exists {
		log.Printf("❌ Batch size не указан для очистки бронирований")
		return
	}

	batchSize, ok := batchSizeInterface.(float64)
	if !ok {
		log.Printf("❌ Некорректный batch size для очистки бронирований")
		return
	}

	log.Printf("🧹 Начинаем очистку просроченных бронирований (batch size: %d)...", int(batchSize))

	deletedCount, err := s.bookingRepo.CleanupExpiredBookings(int(batchSize))
	if err != nil {
		log.Printf("❌ Ошибка очистки просроченных бронирований: %v", err)
		return
	}

	if deletedCount > 0 {
		log.Printf("✅ Очистка бронирований завершена: удалено %d просроченных записей", deletedCount)
	} else {
		log.Printf("✨ Очистка бронирований завершена: просроченных записей не найдено")
	}
}

func (s *SchedulerService) executeCleanupExtensions(_ context.Context, task ScheduledTask) {
	batchSizeInterface, exists := task.Data["batch_size"]
	if !exists {
		log.Printf("❌ Batch size не указан для очистки продлений")
		return
	}

	batchSize, ok := batchSizeInterface.(float64)
	if !ok {
		log.Printf("❌ Некорректный batch size для очистки продлений")
		return
	}

	log.Printf("🧹 Начинаем очистку просроченных продлений (batch size: %d)...", int(batchSize))

	deletedCount, err := s.bookingRepo.CleanupExpiredExtensions(int(batchSize))
	if err != nil {
		log.Printf("❌ Ошибка очистки просроченных продлений: %v", err)
		return
	}

	if deletedCount > 0 {
		log.Printf("✅ Очистка продлений завершена: удалено %d просроченных записей", deletedCount)
	} else {
		log.Printf("✨ Очистка продлений завершена: просроченных записей не найдено")
	}
}

func (s *SchedulerService) performSelfCheck(ctx context.Context) {
	log.Printf("🔍 Начинаем самодиагностику scheduler...")

	start := time.Now()
	issues := 0

	stuckApartments, err := s.findStuckApartments()
	if err != nil {
		log.Printf("❌ Ошибка проверки застрявших квартир: %v", err)
		issues++
	} else if len(stuckApartments) > 0 {
		log.Printf("⚠️ Найдено %d застрявших квартир, исправляем...", len(stuckApartments))
		if s.availabilityService != nil {
			if err := s.availabilityService.RecalculateMultipleApartments(stuckApartments); err != nil {
				log.Printf("❌ Ошибка исправления застрявших квартир: %v", err)
				issues++
			} else {
				log.Printf("✅ Исправлено %d застрявших квартир", len(stuckApartments))
			}
		}
	}

	missedCompletions, err := s.findMissedCompletionTasks()
	if err != nil {
		log.Printf("❌ Ошибка проверки пропущенных завершений: %v", err)
		issues++
	} else if len(missedCompletions) > 0 {
		log.Printf("⚠️ Найдено %d пропущенных завершений, исправляем...", len(missedCompletions))
		for _, bookingID := range missedCompletions {
			s.executeCompleteBookingFallback(bookingID)
		}
		log.Printf("✅ Обработано %d пропущенных завершений", len(missedCompletions))
	}

	freeActiveApartments, err := s.findActiveBookingsWithFreeApartments()
	if err != nil {
		log.Printf("❌ Ошибка проверки активных бронирований: %v", err)
		issues++
	} else if len(freeActiveApartments) > 0 {
		log.Printf("⚠️ Найдено %d активных бронирований с is_free=true, исправляем...", len(freeActiveApartments))
		if s.availabilityService != nil {
			if err := s.availabilityService.RecalculateMultipleApartments(freeActiveApartments); err != nil {
				log.Printf("❌ Ошибка исправления активных бронирований: %v", err)
				issues++
			} else {
				log.Printf("✅ Исправлено %d активных бронирований", len(freeActiveApartments))
			}
		}
	}

	duration := time.Since(start)
	if issues == 0 {
		log.Printf("✅ Самодиагностика завершена успешно за %v", duration)
	} else {
		log.Printf("⚠️ Самодиагностика завершена с %d проблемами за %v", issues, duration)
	}
}

func (s *SchedulerService) findStuckApartments() ([]int, error) {
	query := `
		SELECT DISTINCT a.id 
		FROM apartments a
		WHERE a.is_free = false
		AND NOT EXISTS (
			SELECT 1 FROM bookings b 
			WHERE b.apartment_id = a.id 
			AND b.status IN ('active', 'approved', 'pending', 'awaiting_payment')
			AND (
				(b.status = 'active' AND b.start_date <= NOW() AND b.end_date > NOW()) OR
				(b.status IN ('approved', 'pending', 'awaiting_payment') AND b.start_date <= NOW() + INTERVAL '2 hours')
			)
		)
		AND NOT EXISTS (
			SELECT 1 FROM bookings b2
			WHERE b2.apartment_id = a.id 
			AND b2.status = 'created'
			AND b2.created_at > NOW() - INTERVAL '30 minutes'
			AND b2.start_date <= NOW() + INTERVAL '2 hours'
		)`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apartmentIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		apartmentIDs = append(apartmentIDs, id)
	}

	return apartmentIDs, nil
}

func (s *SchedulerService) findMissedCompletionTasks() ([]int, error) {
	query := `
		SELECT id 
		FROM bookings 
		WHERE status = 'active' 
		AND end_date < NOW() - INTERVAL '10 minutes'
		ORDER BY end_date ASC
		LIMIT 50`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookingIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		bookingIDs = append(bookingIDs, id)
	}

	return bookingIDs, nil
}

func (s *SchedulerService) findActiveBookingsWithFreeApartments() ([]int, error) {
	query := `
		SELECT DISTINCT b.apartment_id 
		FROM bookings b
		JOIN apartments a ON b.apartment_id = a.id
		WHERE b.status = 'active'
		AND b.start_date <= NOW()
		AND b.end_date > NOW()
		AND a.is_free = true`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apartmentIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		apartmentIDs = append(apartmentIDs, id)
	}

	return apartmentIDs, nil
}

func (s *SchedulerService) executeCompleteBookingFallback(bookingID int) {
	log.Printf("🔧 Аварийное завершение бронирования %d", bookingID)

	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		log.Printf("❌ Ошибка получения бронирования %d: %v", bookingID, err)
		return
	}

	if booking.Status != domain.BookingStatusActive {
		log.Printf("⚠️ Бронирование %d не активно (статус: %s)", bookingID, booking.Status)
		return
	}

	booking.Status = domain.BookingStatusCompleted
	booking.DoorStatus = domain.DoorStatusClosed

	err = s.bookingRepo.Update(booking)
	if err != nil {
		log.Printf("❌ Ошибка завершения бронирования %d: %v", bookingID, err)
		return
	}

	if s.availabilityService != nil {
		if err := s.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			log.Printf("❌ Ошибка пересчета доступности квартиры %d: %v", booking.ApartmentID, err)
		}
	}

	if s.lockUseCase != nil {
		if err := s.lockUseCase.DeactivatePasswordForBooking(bookingID); err != nil {
			log.Printf("❌ Ошибка деактивации замка для бронирования %d: %v", bookingID, err)
		}
	}

	log.Printf("✅ Аварийное завершение бронирования %d выполнено", bookingID)
}
