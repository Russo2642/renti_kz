package usecase

import (
	"fmt"
	"log"

	"github.com/russo2642/renti_kz/internal/domain"
)

type cleanerUseCase struct {
	cleanerRepo   domain.CleanerRepository
	userUseCase   domain.UserUseCase
	apartmentRepo domain.ApartmentRepository
}

func NewCleanerUseCase(
	cleanerRepo domain.CleanerRepository,
	userUseCase domain.UserUseCase,
	apartmentRepo domain.ApartmentRepository,
	logger interface{},
) domain.CleanerUseCase {
	return &cleanerUseCase{
		cleanerRepo:   cleanerRepo,
		userUseCase:   userUseCase,
		apartmentRepo: apartmentRepo,
	}
}

func (uc *cleanerUseCase) CreateCleaner(request *domain.CreateCleanerRequest, adminID int) (*domain.Cleaner, error) {
	user, err := uc.userUseCase.GetByID(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("пользователь с ID %d не найден", request.UserID)
	}

	exists, err := uc.cleanerRepo.IsUserCleaner(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существования уборщицы: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("пользователь уже является уборщицей")
	}

	cleaner := &domain.Cleaner{
		UserID:   request.UserID,
		IsActive: true,
		Schedule: request.Schedule,
	}

	err = uc.cleanerRepo.Create(cleaner)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания уборщицы: %w", err)
	}

	err = uc.userUseCase.UpdateUserRole(request.UserID, domain.RoleCleaner, adminID)
	if err != nil {
		log.Printf("Ошибка обновления роли пользователя %d на cleaner: %v", request.UserID, err)
	}

	for _, apartmentID := range request.ApartmentIDs {
		err = uc.cleanerRepo.AssignToApartment(cleaner.ID, apartmentID)
		if err != nil {
			log.Printf("Ошибка назначения квартиры %d уборщице %d: %v", apartmentID, cleaner.ID, err)
		}
	}

	return uc.cleanerRepo.GetByID(cleaner.ID)
}

func (uc *cleanerUseCase) GetCleanerByID(id int) (*domain.Cleaner, error) {
	cleaner, err := uc.cleanerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения уборщицы: %w", err)
	}

	if cleaner == nil {
		return nil, fmt.Errorf("уборщица с ID %d не найдена", id)
	}

	return cleaner, nil
}

func (uc *cleanerUseCase) GetCleanersByApartment(apartmentID int) ([]*domain.Cleaner, error) {
	return uc.cleanerRepo.GetByApartmentIDActive(apartmentID)
}

func (uc *cleanerUseCase) GetAllCleaners(filters map[string]interface{}, page, pageSize int) ([]*domain.Cleaner, int, error) {
	return uc.cleanerRepo.GetAll(filters, page, pageSize)
}

func (uc *cleanerUseCase) UpdateCleaner(id int, request *domain.UpdateCleanerRequest, adminID int) error {
	cleaner, err := uc.cleanerRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return fmt.Errorf("уборщица с ID %d не найдена", id)
	}

	err = uc.cleanerRepo.UpdatePartial(id, request)
	if err != nil {
		return fmt.Errorf("ошибка обновления уборщицы: %w", err)
	}

	return nil
}

func (uc *cleanerUseCase) DeleteCleaner(id int, adminID int) error {
	cleaner, err := uc.cleanerRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return fmt.Errorf("уборщица с ID %d не найдена", id)
	}

	err = uc.cleanerRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("ошибка удаления уборщицы: %w", err)
	}

	return nil
}

func (uc *cleanerUseCase) GetCleanersByOwner(ownerID int) ([]*domain.Cleaner, error) {
	return uc.cleanerRepo.GetCleanersByOwner(ownerID)
}

func (uc *cleanerUseCase) IsUserCleaner(userID int) (bool, error) {
	return uc.cleanerRepo.IsUserCleaner(userID)
}

func (uc *cleanerUseCase) AssignCleanerToApartment(cleanerID, apartmentID, adminID int) error {
	cleaner, err := uc.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return fmt.Errorf("уборщица с ID %d не найдена", cleanerID)
	}

	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return fmt.Errorf("квартира с ID %d не найдена", apartmentID)
	}

	err = uc.cleanerRepo.AssignToApartment(cleanerID, apartmentID)
	if err != nil {
		return fmt.Errorf("ошибка назначения уборщицы на квартиру: %w", err)
	}

	return nil
}

func (uc *cleanerUseCase) RemoveCleanerFromApartment(cleanerID, apartmentID, adminID int) error {
	err := uc.cleanerRepo.RemoveFromApartment(cleanerID, apartmentID)
	if err != nil {
		return fmt.Errorf("ошибка удаления уборщицы с квартиры: %w", err)
	}

	return nil
}

func (uc *cleanerUseCase) GetCleanerByUserID(userID int) (*domain.Cleaner, error) {
	return uc.cleanerRepo.GetByUserID(userID)
}

func (uc *cleanerUseCase) GetCleanerApartments(userID int) ([]*domain.Apartment, error) {
	cleaner, err := uc.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return nil, fmt.Errorf("пользователь не является уборщицей")
	}

	cleanerApartments, err := uc.cleanerRepo.GetCleanerApartments(cleaner.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения квартир уборщицы: %w", err)
	}

	var apartments []*domain.Apartment
	for _, ca := range cleanerApartments {
		apartment, err := uc.apartmentRepo.GetByID(ca.ApartmentID)
		if err != nil {
			log.Printf("Ошибка получения квартиры %d: %v", ca.ApartmentID, err)
			continue
		}
		if apartment != nil {
			apartments = append(apartments, apartment)
		}
	}

	return apartments, nil
}

func (uc *cleanerUseCase) GetApartmentsForCleaning(userID int) ([]*domain.ApartmentForCleaning, error) {
	cleaner, err := uc.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return nil, fmt.Errorf("пользователь не является уборщицей")
	}

	if !cleaner.IsActive {
		return nil, fmt.Errorf("уборщица неактивна")
	}

	return uc.cleanerRepo.GetApartmentsForCleaning(cleaner.ID)
}

func (uc *cleanerUseCase) GetCleanerStats(userID int) (map[string]interface{}, error) {
	cleaner, err := uc.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return nil, fmt.Errorf("пользователь не является уборщицей")
	}

	apartmentsForCleaning, err := uc.cleanerRepo.GetApartmentsForCleaning(cleaner.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения квартир для уборки: %w", err)
	}

	cleanerApartments, err := uc.cleanerRepo.GetCleanerApartments(cleaner.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения назначенных квартир: %w", err)
	}

	stats := map[string]interface{}{
		"total_apartments":         len(cleanerApartments),
		"apartments_need_cleaning": len(apartmentsForCleaning),
		"apartments_clean":         len(cleanerApartments) - len(apartmentsForCleaning),
		"is_active":                cleaner.IsActive,
	}

	return stats, nil
}

func (uc *cleanerUseCase) UpdateCleanerSchedule(userID int, schedule *domain.CleanerSchedule) error {
	cleaner, err := uc.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return fmt.Errorf("пользователь не является уборщицей")
	}

	// Для PUT запроса полностью заменяем расписание
	cleaner.Schedule = schedule
	err = uc.cleanerRepo.Update(cleaner)
	if err != nil {
		return fmt.Errorf("ошибка обновления расписания: %w", err)
	}

	return nil
}

func (uc *cleanerUseCase) UpdateCleanerSchedulePatch(userID int, schedulePatch *domain.CleanerSchedulePatch) error {
	cleaner, err := uc.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return fmt.Errorf("пользователь не является уборщицей")
	}

	err = uc.cleanerRepo.UpdateSchedulePatch(cleaner.ID, schedulePatch)
	if err != nil {
		return fmt.Errorf("ошибка частичного обновления расписания: %w", err)
	}

	return nil
}

func (uc *cleanerUseCase) StartCleaning(userID int, request *domain.StartCleaningRequest) error {
	cleaner, err := uc.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return fmt.Errorf("пользователь не является уборщицей")
	}

	if !cleaner.IsActive {
		return fmt.Errorf("уборщица неактивна")
	}

	cleanerApartments, err := uc.cleanerRepo.GetCleanerApartments(cleaner.ID)
	if err != nil {
		return fmt.Errorf("ошибка получения квартир уборщицы: %w", err)
	}

	var hasAccess bool
	for _, ca := range cleanerApartments {
		if ca.ApartmentID == request.ApartmentID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return fmt.Errorf("квартира не назначена данной уборщице")
	}

	apartmentsForCleaning, err := uc.cleanerRepo.GetApartmentsForCleaning(cleaner.ID)
	if err != nil {
		return fmt.Errorf("ошибка получения квартир для уборки: %w", err)
	}

	var needsCleaning bool
	for _, apt := range apartmentsForCleaning {
		if apt.ID == request.ApartmentID {
			needsCleaning = true
			break
		}
	}

	if !needsCleaning {
		return fmt.Errorf("квартира не нуждается в уборке")
	}

	// TODO: Здесь можно добавить логирование начала уборки
	// Например, создать запись в таблице cleaning_logs
	log.Printf("Уборщица %d начала уборку квартиры %d. Заметки: %s", userID, request.ApartmentID, request.Notes)

	return nil
}

func (uc *cleanerUseCase) CompleteCleaning(userID int, request *domain.CompleteCleaningRequest) error {
	cleaner, err := uc.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("уборщица не найдена: %w", err)
	}

	if cleaner == nil {
		return fmt.Errorf("пользователь не является уборщицей")
	}

	if !cleaner.IsActive {
		return fmt.Errorf("уборщица неактивна")
	}

	cleanerApartments, err := uc.cleanerRepo.GetCleanerApartments(cleaner.ID)
	if err != nil {
		return fmt.Errorf("ошибка получения квартир уборщицы: %w", err)
	}

	var hasAccess bool
	for _, ca := range cleanerApartments {
		if ca.ApartmentID == request.ApartmentID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return fmt.Errorf("квартира не назначена данной уборщице")
	}

	err = uc.apartmentRepo.UpdateIsFree(request.ApartmentID, true)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса квартиры: %w", err)
	}

	// TODO: Здесь можно добавить:
	// 1. Сохранение фотографий после уборки
	// 2. Логирование завершения уборки
	// 3. Отправка уведомлений владельцу
	// 4. Создание записи в таблице cleaning_logs

	log.Printf("Уборщица %d завершила уборку квартиры %d. Заметки: %s", userID, request.ApartmentID, request.Notes)

	return nil
}

func (uc *cleanerUseCase) GetCleaningHistory(userID int, filters map[string]interface{}, page, pageSize int) ([]map[string]interface{}, int, error) {
	// TODO: Реализовать получение истории уборки
	// Пока возвращаем пустой результат
	return []map[string]interface{}{}, 0, nil
}

func (uc *cleanerUseCase) GetApartmentsNeedingCleaning() ([]*domain.ApartmentForCleaning, error) {
	apartments, err := uc.cleanerRepo.GetApartmentsNeedingCleaning()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения квартир, нуждающихся в уборке: %w", err)
	}

	return apartments, nil
}
