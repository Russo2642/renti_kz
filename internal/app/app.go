package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/russo2642/renti_kz/internal/config"
	httpDelivery "github.com/russo2642/renti_kz/internal/delivery/http"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/repository/postgres"
	redisRepo "github.com/russo2642/renti_kz/internal/repository/redis"
	"github.com/russo2642/renti_kz/internal/services"
	"github.com/russo2642/renti_kz/internal/usecase"
	"github.com/russo2642/renti_kz/pkg/migrator"
)

type App struct {
	db         *sql.DB
	httpServer *http.Server
	wsService  *services.ChatWebSocketService
}

func (a *App) Cleanup() {
	if a.db != nil {
		a.db.Close()
	}
}

func (a *App) GetHTTPServer() *http.Server {
	return a.httpServer
}

func InitApp(cfg *config.Config) (*App, error) {
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	services.InitHTTPClientPool()

	if err := migrator.RunMigrations(cfg.Database, cfg.Migration); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	db, err := initDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	s3Storage, err := initS3Storage(cfg.S3)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 storage: %w", err)
	}

	tokenManager := initTokenManager(cfg.JWT)

	roleRepo := postgres.NewRoleRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	userRepo := postgres.NewUserRepository(db, roleRepo, locationRepo)
	propertyOwnerRepo := postgres.NewPropertyOwnerRepository(db)
	renterRepo := postgres.NewRenterRepository(db)
	apartmentRepo := postgres.NewApartmentRepository(db)
	bookingRepo := postgres.NewBookingRepository(db)
	favoriteRepo := postgres.NewFavoriteRepository(db)
	lockRepo := postgres.NewLockRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)

	conciergeRepo := postgres.NewConciergeRepository(db)
	cleanerRepo := postgres.NewCleanerRepository(db)
	chatRoomRepo := postgres.NewChatRoomRepository(db)
	chatParticipantRepo := postgres.NewChatParticipantRepository(db)
	chatMessageRepo := postgres.NewChatMessageRepository(db)
	cancellationRuleRepo := postgres.NewCancellationRuleRepository(db)
	contractRepo := postgres.NewContractRepository(db)
	settingsRepo := postgres.NewPlatformSettingsRepository(db)
	paymentRepo := postgres.NewPaymentRepository(db)
	paymentLogRepo := postgres.NewPaymentLogRepository(db)
	apartmentTypeRepo := postgres.NewApartmentTypeRepository(db)

	tuyaConfig := services.TuyaConfig{
		ClientID:     cfg.Tuya.ClientID,
		ClientSecret: cfg.Tuya.ClientSecret,
		APIBase:      cfg.Tuya.APIBase,
		TimeZone:     cfg.Tuya.TimeZone,
	}
	tuyaService := services.NewTuyaLockService(tuyaConfig)

	otpService := services.NewOTPService(&cfg.OTP)
	otpRepo, err := redisRepo.NewOTPRepository(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OTP repository: %w", err)
	}

	pushService := services.NewPushNotificationService(notificationRepo)

	queueService, err := services.NewRedisQueueService(cfg.Redis, cfg.Notification)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize queue service: %w", err)
	}

	userCacheService, err := services.NewUserCacheService(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize user cache service: %w", err)
	}

	responseCacheService, err := services.NewResponseCacheService(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize response cache service: %w", err)
	}

	userUseCase := usecase.NewUserUseCase(userRepo, roleRepo, cfg.App.PasswordSalt)
	propertyOwnerUseCase := usecase.NewPropertyOwnerUseCase(propertyOwnerRepo, userRepo, roleRepo, s3Storage, cfg.App.PasswordSalt)
	renterUseCase := usecase.NewRenterUseCase(renterRepo, userRepo, roleRepo, s3Storage, cfg.App.PasswordSalt)
	authUseCase := usecase.NewAuthUseCase(userRepo, tokenManager, cfg.App.PasswordSalt, userCacheService)
	otpUseCase := usecase.NewOTPUseCase(otpService, otpRepo, userRepo, tokenManager)
	locationUseCase := usecase.NewLocationUseCase(locationRepo)

	lockAutoUpdateService := services.NewLockAutoUpdateService(
		tuyaService,
		cfg.Tuya.APIBase,
		cfg.Tuya.ClientID,
		cfg.Tuya.ClientSecret,
	)

	lockUseCase := usecase.NewLockUseCase(lockRepo, apartmentRepo, bookingRepo, propertyOwnerRepo, renterRepo, userUseCase, tuyaService, lockAutoUpdateService)

	freedomPayService := services.NewFreedomPayService(
		cfg.FreedomPay.MerchantID,
		cfg.FreedomPay.SecretKey,
		cfg.FreedomPay.APIBase,
	)

	lockAutoUpdateService.SetLockUseCase(lockUseCase)
	favoriteUseCase := usecase.NewFavoriteUseCase(favoriteRepo, apartmentRepo, userRepo, propertyOwnerRepo)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, pushService, queueService)
	lockUseCase.SetNotificationUseCase(notificationUseCase)

	conciergeUseCase := usecase.NewConciergeUseCase(conciergeRepo, userRepo, apartmentRepo, roleRepo, bookingRepo, chatRoomRepo)
	cleanerUseCase := usecase.NewCleanerUseCase(cleanerRepo, userUseCase, apartmentRepo, nil)
	cancellationRuleUseCase := usecase.NewCancellationRuleUseCase(cancellationRuleRepo)

	contractTemplateRepo := postgres.NewContractTemplateRepository(db)

	redisConn := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolTimeout:  cfg.Redis.PoolTimeout,
	})

	contractService := services.NewContractService(services.ContractServiceConfig{
		ContractRepo:         contractRepo,
		ContractTemplateRepo: contractTemplateRepo,
		BookingRepo:          bookingRepo,
		ApartmentRepo:        apartmentRepo,
		UserRepo:             userRepo,
		RenterRepo:           renterRepo,
		PropertyOwnerRepo:    propertyOwnerRepo,
		RedisClient:          redisConn,
		TemplatesPath:        "internal/templates",
	})

	contractUseCase := usecase.NewContractUseCase(contractRepo, contractService, bookingRepo, apartmentRepo, userRepo, renterRepo, propertyOwnerRepo)
	settingsUseCase := usecase.NewPlatformSettingsUseCase(settingsRepo)

	wsService := services.NewChatWebSocketService(nil, userUseCase)

	chatUseCase := usecase.NewChatUseCase(chatRoomRepo, chatMessageRepo, chatParticipantRepo, bookingRepo, conciergeRepo, renterRepo, wsService, notificationRepo)

	wsService.SetChatUseCase(chatUseCase)

	paymentUseCase := usecase.NewPaymentUseCase(freedomPayService, paymentRepo, paymentLogRepo)

	availabilityService := services.NewApartmentAvailabilityService(db, apartmentRepo)

	redisScheduler := services.NewSchedulerService(
		cfg.Redis,
		db,
		bookingRepo,
		apartmentRepo,
		availabilityService,
		notificationUseCase,
		lockUseCase,
		renterRepo,
		propertyOwnerRepo,
		userUseCase,
		chatUseCase,
		chatRoomRepo,
		paymentRepo,
		paymentUseCase,
	)

	bookingUseCase := usecase.NewBookingUseCase(bookingRepo, apartmentRepo, renterRepo, propertyOwnerRepo, lockUseCase, userUseCase, notificationUseCase, redisScheduler, chatUseCase, chatRoomRepo, conciergeRepo, contractUseCase, settingsUseCase, paymentUseCase, paymentRepo, paymentLogRepo, availabilityService)

	apartmentTypeUseCase := usecase.NewApartmentTypeUseCase(apartmentTypeRepo, userUseCase)
	apartmentUseCase := usecase.NewApartmentUseCase(apartmentRepo, userRepo, propertyOwnerRepo, bookingUseCase, bookingRepo, contractUseCase, s3Storage)
	apartmentUseCase.SetNotificationUseCase(notificationUseCase)

	go redisScheduler.StartScheduler()

	notificationUseCase.StartNotificationConsumer()

	services.StartPerformanceMonitoring()

	middleware := httpDelivery.NewMiddleware(tokenManager, authUseCase, userCacheService)
	authHandler := httpDelivery.NewAuthHandler(authUseCase, userUseCase, otpUseCase, tokenManager, renterRepo, propertyOwnerRepo, renterUseCase)
	userHandler := httpDelivery.NewUserHandler(userUseCase, propertyOwnerUseCase, renterUseCase, renterRepo, apartmentUseCase, bookingUseCase, otpUseCase, responseCacheService)

	apartmentHandler := httpDelivery.NewApartmentHandler(
		apartmentUseCase,
		userUseCase,
		propertyOwnerUseCase,
		locationUseCase,
		notificationUseCase,
		settingsUseCase,
		bookingUseCase,
		lockUseCase,
		middleware,
		userRepo,
		propertyOwnerRepo,
		roleRepo,
		responseCacheService,
	)
	dictionaryHandler := httpDelivery.NewDictionaryHandler(apartmentUseCase)
	bookingHandler := httpDelivery.NewBookingHandler(bookingUseCase, userUseCase, lockUseCase, responseCacheService)
	paymentHandler := httpDelivery.NewPaymentHandler(paymentUseCase)
	favoriteHandler := httpDelivery.NewFavoriteHandler(favoriteUseCase)
	lockHandler := httpDelivery.NewLockHandler(lockUseCase, userUseCase, bookingUseCase, notificationUseCase, apartmentRepo)
	notificationHandler := httpDelivery.NewNotificationHandler(notificationUseCase)
	schedulerHandler := httpDelivery.NewSchedulerHandler(redisScheduler)
	systemHandler := httpDelivery.NewSystemHandler(userCacheService)

	tuyaWebhookHandler := httpDelivery.NewTuyaWebhookHandler(lockUseCase)

	conciergeHandler := httpDelivery.NewConciergeHandler(conciergeUseCase)
	conciergeInterfaceHandler := httpDelivery.NewConciergeInterfaceHandler(conciergeUseCase, apartmentUseCase, bookingUseCase, chatUseCase)
	cleanerHandler := httpDelivery.NewCleanerHandler(cleanerUseCase)
	chatHandler := httpDelivery.NewChatHandler(chatUseCase, wsService)
	cancellationRuleHandler := httpDelivery.NewCancellationRuleHandler(cancellationRuleUseCase)
	contractHandler := httpDelivery.NewContractHandler(contractUseCase, userUseCase)
	settingsHandler := httpDelivery.NewPlatformSettingsHandler(settingsUseCase, middleware)
	apartmentTypeHandler := httpDelivery.NewApartmentTypeHandler(apartmentTypeUseCase)

	router := initRouter(
		authHandler,
		userHandler,
		apartmentHandler,
		dictionaryHandler,
		bookingHandler,
		paymentHandler,
		favoriteHandler,
		lockHandler,
		notificationHandler,
		schedulerHandler,
		systemHandler,
		conciergeHandler,
		conciergeInterfaceHandler,
		cleanerHandler,
		chatHandler,
		cancellationRuleHandler,
		contractHandler,
		settingsHandler,
		apartmentTypeHandler,
		middleware,
		locationUseCase,
		tuyaWebhookHandler,
		responseCacheService,
	)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Server.HttpPort),
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
	}

	if cfg.Server.EnableKeepAlive {
		httpServer.SetKeepAlivesEnabled(true)
	}

	return &App{
		db:         db,
		httpServer: httpServer,
		wsService:  wsService,
	}, nil
}

func initRouter(
	authHandler *httpDelivery.AuthHandler,
	userHandler *httpDelivery.UserHandler,
	apartmentHandler *httpDelivery.ApartmentHandler,
	dictionaryHandler *httpDelivery.DictionaryHandler,
	bookingHandler *httpDelivery.BookingHandler,
	paymentHandler *httpDelivery.PaymentHandler,
	favoriteHandler *httpDelivery.FavoriteHandler,
	lockHandler *httpDelivery.LockHandler,
	notificationHandler *httpDelivery.NotificationHandler,
	schedulerHandler *httpDelivery.SchedulerHandler,
	systemHandler *httpDelivery.SystemHandler,
	conciergeHandler *httpDelivery.ConciergeHandler,
	conciergeInterfaceHandler *httpDelivery.ConciergeInterfaceHandler,
	cleanerHandler *httpDelivery.CleanerHandler,
	chatHandler *httpDelivery.ChatHandler,
	cancellationRuleHandler *httpDelivery.CancellationRuleHandler,
	contractHandler *httpDelivery.ContractHandler,
	settingsHandler *httpDelivery.PlatformSettingsHandler,
	apartmentTypeHandler *httpDelivery.ApartmentTypeHandler,
	middleware *httpDelivery.Middleware,
	locationUseCase domain.LocationUseCase,
	tuyaWebhookHandler *httpDelivery.TuyaWebhookHandler,
	responseCacheService *services.ResponseCacheService,
) *gin.Engine {
	router := gin.Default()

	router.Use(services.PerformanceMiddleware())
	router.Use(middleware.CORS())
	router.Use(httpDelivery.ErrorMiddleware())

	router.Static("/uploads", "./uploads")

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/metrics", services.MetricsHandler())

	api := router.Group("/api")

	authHandler.RegisterRoutes(api)

	apartments := api.Group("/apartments")
	{
		apartments.GET("", middleware.OptionalAuthMiddleware(), apartmentHandler.GetAll)
		apartments.GET("/search/geo", httpDelivery.CacheMiddlewareWithTTL(responseCacheService, 5*time.Minute), apartmentHandler.GetByCoordinates)
		apartments.GET("/:id", middleware.OptionalAuthMiddleware(), apartmentHandler.GetByID)
		apartments.GET("/:id/photos", httpDelivery.LongCacheMiddleware(responseCacheService), apartmentHandler.GetPhotosByApartmentID)
		apartments.GET("/:id/location", httpDelivery.LongCacheMiddleware(responseCacheService), apartmentHandler.GetLocationByApartmentID)
		apartments.GET("/:id/available-durations", apartmentHandler.GetAvailableDurations)
		apartments.GET("/:id/can-book-now", apartmentHandler.CanBookNow)
		apartments.GET("/:id/calculate-price", apartmentHandler.CalculatePrice)
		apartments.GET("/:id/booked-dates", apartmentHandler.GetBookedDates)
		apartments.GET("/:id/availability", apartmentHandler.CheckApartmentAvailability)
		apartments.GET("/:id/available-slots", apartmentHandler.GetAvailableTimeSlots)

		authorized := apartments.Group("/", middleware.AuthMiddleware())
		{
			authorized.POST("", apartmentHandler.Create)
			authorized.PUT("/:id", apartmentHandler.Update)
			authorized.DELETE("/:id", apartmentHandler.Delete)

			authorized.POST("/:id/photos", apartmentHandler.AddPhotos)
			authorized.DELETE("/photos/:photoId", apartmentHandler.DeletePhoto)

			authorized.POST("/:id/location", apartmentHandler.AddLocation)
			authorized.PUT("/:id/location", apartmentHandler.UpdateLocation)

			authorized.GET("/:id/documents", apartmentHandler.GetDocumentsByApartmentID)
			authorized.POST("/:id/documents", apartmentHandler.AddDocuments)
			authorized.DELETE("/documents/:documentId", apartmentHandler.DeleteDocument)

			authorized.POST("/:id/confirm-agreement", apartmentHandler.ConfirmApartmentAgreement)

			authorized.GET("/owner/statistics", httpDelivery.CacheMiddlewareWithTTL(responseCacheService, 2*time.Minute), apartmentHandler.GetOwnerStatistics)

			adminModerator := authorized.Group("/", middleware.RoleMiddleware(domain.RoleAdmin, domain.RoleModerator))
			{
				adminModerator.GET("/dashboard", apartmentHandler.GetDashboardStats)
			}
		}
	}

	dictionariesGroup := api.Group("/dictionaries")
	dictionariesGroup.Use(httpDelivery.LongCacheMiddleware(responseCacheService))
	dictionaryHandler.RegisterRoutes(dictionariesGroup)

	locationsGroup := api.Group("/locations")
	locationsGroup.Use(httpDelivery.LongCacheMiddleware(responseCacheService))
	httpDelivery.NewLocationHandler(locationsGroup, locationUseCase)

	cancellationRuleHandler.RegisterRoutes(api)

	settingsHandler.RegisterRoutes(api)
	apartmentTypeHandler.RegisterRoutes(api)

	tuyaWebhookHandler.RegisterRoutes(api)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		userHandler.RegisterRoutes(protected)

		adminRoutes := protected.Group("/admin")
		adminRoutes.Use(middleware.RoleMiddleware(domain.RoleAdmin))
		{
			adminRoutes.GET("/apartments", httpDelivery.CacheMiddlewareWithTTL(responseCacheService, 3*time.Minute), apartmentHandler.AdminGetAllApartments)
			adminRoutes.GET("/apartments/:id", httpDelivery.CacheMiddlewareWithTTL(responseCacheService, 5*time.Minute), apartmentHandler.AdminGetApartmentByID)
			adminRoutes.PUT("/apartments/:id/status", apartmentHandler.UpdateStatus)
			adminRoutes.PUT("/apartments/:id/apartment-type", apartmentHandler.UpdateApartmentType)
			adminRoutes.PUT("/apartments/:id/counters", apartmentHandler.AdminUpdateCounters)
			adminRoutes.POST("/apartments/:id/counters/reset", apartmentHandler.AdminResetCounters)
			adminRoutes.DELETE("/apartments/:id", apartmentHandler.AdminDeleteApartment)
			adminRoutes.GET("/apartments/statistics", httpDelivery.CacheMiddlewareWithTTL(responseCacheService, 10*time.Minute), apartmentHandler.AdminGetApartmentStatistics)
			adminRoutes.GET("/apartments/:id/bookings-history", apartmentHandler.AdminGetApartmentBookingsHistory)
			adminRoutes.GET("/dashboard/statistics", httpDelivery.CacheMiddlewareWithTTL(responseCacheService, 5*time.Minute), apartmentHandler.AdminGetFullDashboardStats)

			bookingHandler.RegisterAdminRoutes(adminRoutes)
			lockHandler.RegisterAdminRoutes(adminRoutes)
			schedulerHandler.RegisterRoutes(adminRoutes)
			conciergeHandler.RegisterRoutes(adminRoutes)
			cleanerHandler.RegisterAdminRoutes(adminRoutes)
			systemHandler.RegisterAdminRoutes(adminRoutes)
			apartmentTypeHandler.RegisterAdminRoutes(adminRoutes)
		}

		conciergeRoutes := protected.Group("/concierge")
		conciergeRoutes.Use(middleware.RoleMiddleware(domain.RoleConcierge))
		{
			conciergeInterfaceHandler.RegisterRoutes(conciergeRoutes)
		}

		cleanerRoutes := protected.Group("/cleaner")
		cleanerRoutes.Use(middleware.RoleMiddleware(domain.RoleCleaner))
		{
			cleanerHandler.RegisterCleanerRoutes(cleanerRoutes)
		}

		bookingHandler.RegisterRoutes(protected)

		paymentHandler.RegisterRoutes(protected)

		favoriteHandler.RegisterRoutes(protected)

		lockHandler.RegisterRoutes(protected, middleware)

		tuyaWebhookHandler.RegisterProtectedRoutes(protected)

		notificationHandler.RegisterRoutes(protected)

		chatHandler.RegisterRoutes(protected)

		contractHandler.RegisterRoutes(protected)
	}

	return router
}
