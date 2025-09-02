package usecase

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
	"github.com/russo2642/renti_kz/pkg/logger"
	"github.com/russo2642/renti_kz/pkg/storage/s3"
)

type ApartmentUseCase struct {
	apartmentRepo   domain.ApartmentRepository
	userRepo        domain.UserRepository
	ownerRepo       domain.PropertyOwnerRepository
	bookingUseCase  domain.BookingUseCase
	bookingRepo     domain.BookingRepository
	contractUseCase domain.ContractUseCase
	s3Storage       *s3.Storage
}

func NewApartmentUseCase(
	apartmentRepo domain.ApartmentRepository,
	userRepo domain.UserRepository,
	ownerRepo domain.PropertyOwnerRepository,
	bookingUseCase domain.BookingUseCase,
	bookingRepo domain.BookingRepository,
	contractUseCase domain.ContractUseCase,
	s3Storage *s3.Storage,
) *ApartmentUseCase {
	return &ApartmentUseCase{
		apartmentRepo:   apartmentRepo,
		userRepo:        userRepo,
		ownerRepo:       ownerRepo,
		bookingUseCase:  bookingUseCase,
		bookingRepo:     bookingRepo,
		contractUseCase: contractUseCase,
		s3Storage:       s3Storage,
	}
}

func (uc *ApartmentUseCase) Create(apartment *domain.Apartment) error {
	owner, err := uc.ownerRepo.GetByID(apartment.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to get property owner: %w", err)
	}
	if owner == nil {
		return fmt.Errorf("property owner with id %d not found", apartment.OwnerID)
	}

	if !apartment.RentalTypeHourly && !apartment.RentalTypeDaily {
		return fmt.Errorf("должен быть выбран хотя бы один тип аренды (почасовая или посуточная)")
	}

	apartment.Status = domain.AptStatusPending
	apartment.IsAgreementAccepted = false

	if err := uc.apartmentRepo.Create(apartment); err != nil {
		return fmt.Errorf("failed to create apartment: %w", err)
	}

	if uc.contractUseCase != nil {
		contract, err := uc.contractUseCase.CreateApartmentContract(apartment.ID)
		if err != nil {
			logger.Warn("failed to create apartment contract",
				slog.Int("apartment_id", apartment.ID),
				slog.String("error", err.Error()))
		} else {
			apartment.ContractID = &contract.ID
			if updateErr := uc.apartmentRepo.Update(apartment); updateErr != nil {
				logger.Warn("failed to link apartment to contract",
					slog.Int("apartment_id", apartment.ID),
					slog.Int("contract_id", contract.ID),
					slog.String("error", updateErr.Error()))
			}
			logger.Info("apartment contract created successfully", slog.Int("apartment_id", apartment.ID))
		}
	}

	return nil
}

func (uc *ApartmentUseCase) GetByID(id int) (*domain.Apartment, error) {
	apartment, err := uc.apartmentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}

	return apartment, nil
}

func (uc *ApartmentUseCase) GetByIDWithUserContext(id int, userID *int) (*domain.Apartment, error) {
	apartment, err := uc.apartmentRepo.GetByIDWithUserContext(id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}

	return apartment, nil
}

func (uc *ApartmentUseCase) GetByOwnerID(ownerID int) ([]*domain.Apartment, error) {
	owner, err := uc.ownerRepo.GetByID(ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property owner: %w", err)
	}
	if owner == nil {
		return nil, fmt.Errorf("property owner with id %d not found", ownerID)
	}

	apartments, err := uc.apartmentRepo.GetByOwnerID(ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments by owner: %w", err)
	}

	return apartments, nil
}

func (uc *ApartmentUseCase) GetAll(filters map[string]interface{}, page, pageSize int) ([]*domain.Apartment, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	apartments, total, err := uc.apartmentRepo.GetAll(filters, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get apartments: %w", err)
	}

	return apartments, total, nil
}

func (uc *ApartmentUseCase) GetAllWithUserContext(filters map[string]interface{}, page, pageSize int, userID *int) ([]*domain.Apartment, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	apartments, total, err := uc.apartmentRepo.GetAllWithUserContext(filters, page, pageSize, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get apartments: %w", err)
	}

	return apartments, total, nil
}

func (uc *ApartmentUseCase) Update(apartment *domain.Apartment) error {
	existingApartment, err := uc.apartmentRepo.GetByID(apartment.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing apartment: %w", err)
	}
	if existingApartment == nil {
		return fmt.Errorf("apartment with id %d not found", apartment.ID)
	}

	if existingApartment.OwnerID != apartment.OwnerID {
		return fmt.Errorf("changing apartment owner is not allowed")
	}

	if existingApartment.Status == domain.AptStatusRejected {
		return fmt.Errorf("cannot update rejected apartment")
	}

	if !apartment.RentalTypeHourly && !apartment.RentalTypeDaily {
		return fmt.Errorf("должен быть выбран хотя бы один тип аренды (почасовая или посуточная)")
	}

	apartment.Status = domain.AptStatusPending
	if apartment.ModeratorComment == "" {
		apartment.ModeratorComment = existingApartment.ModeratorComment
	}

	if err := uc.apartmentRepo.Update(apartment); err != nil {
		return fmt.Errorf("failed to update apartment: %w", err)
	}

	return nil
}

func (uc *ApartmentUseCase) Delete(id int) error {
	existingApartment, err := uc.apartmentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get existing apartment: %w", err)
	}
	if existingApartment == nil {
		return fmt.Errorf("apartment with id %d not found", id)
	}

	photos, err := uc.apartmentRepo.GetPhotosByApartmentID(id)
	if err != nil {
		return fmt.Errorf("failed to get apartment photos: %w", err)
	}

	if err := uc.apartmentRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete apartment: %w", err)
	}

	for _, photo := range photos {
		objectKey := uc.s3Storage.ExtractObjectKey(photo.URL)
		if err := uc.s3Storage.DeleteFile(objectKey); err != nil {
			fmt.Printf("failed to delete photo from S3: %v\n", err)
		}
	}

	return nil
}

func (uc *ApartmentUseCase) AddPhotos(apartmentID int, filesData [][]byte) ([]string, error) {
	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}
	if apartment == nil {
		return nil, fmt.Errorf("apartment with id %d not found", apartmentID)
	}

	owner, err := uc.ownerRepo.GetByID(apartment.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}

	user, err := uc.userRepo.GetByID(owner.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	existingPhotos, err := uc.apartmentRepo.GetPhotosByApartmentID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment photos: %w", err)
	}

	if len(existingPhotos)+len(filesData) > 20 {
		return nil, fmt.Errorf("total number of photos would exceed maximum (20): current %d, trying to add %d",
			len(existingPhotos), len(filesData))
	}

	startOrder := 0
	if len(existingPhotos) > 0 {

		for _, photo := range existingPhotos {
			if photo.Order > startOrder {
				startOrder = photo.Order
			}
		}

		startOrder++
	}

	uploadedURLs := make([]string, 0, len(filesData))

	uploadedObjects := make([]string, 0, len(filesData))

	date := time.Now().Format("2006-01-02")

	for i, fileData := range filesData {

		objectKey := fmt.Sprintf("%s/ad/%s/%d_%d.jpg", user.Phone, date, apartmentID, startOrder+i)

		url, err := uc.s3Storage.UploadFile(objectKey, fileData)
		if err != nil {

			for _, objKey := range uploadedObjects {
				uc.s3Storage.DeleteFile(objKey)
			}
			return nil, fmt.Errorf("failed to upload photo %d: %w", i+1, err)
		}

		uploadedObjects = append(uploadedObjects, objectKey)

		photo := &domain.ApartmentPhoto{
			ApartmentID: apartmentID,
			URL:         url,
			Order:       startOrder + i,
		}

		if err := uc.apartmentRepo.AddPhoto(photo); err != nil {

			for _, objKey := range uploadedObjects {
				uc.s3Storage.DeleteFile(objKey)
			}
			return nil, fmt.Errorf("failed to save photo %d info: %w", i+1, err)
		}

		uploadedURLs = append(uploadedURLs, url)
	}

	return uploadedURLs, nil
}

func (uc *ApartmentUseCase) AddPhotosParallel(apartmentID int, filesData [][]byte) ([]string, error) {
	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return nil, err
	}

	if apartment == nil {
		return nil, fmt.Errorf("квартира с ID %d не найдена", apartmentID)
	}

	existingPhotos, err := uc.apartmentRepo.GetPhotosByApartmentID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке существующих фотографий: %w", err)
	}

	if len(existingPhotos)+len(filesData) > 20 {
		return nil, fmt.Errorf("превышено максимальное количество фотографий (20). Текущее: %d, добавляется: %d",
			len(existingPhotos), len(filesData))
	}

	urls, err := uc.s3Storage.UploadApartmentPhotosParallel(apartmentID, filesData)
	if err != nil {
		return nil, fmt.Errorf("ошибка при загрузке фотографий: %w", err)
	}

	var savedURLs []string
	for i, url := range urls {
		photo := &domain.ApartmentPhoto{
			ApartmentID: apartmentID,
			URL:         url,
			Order:       len(existingPhotos) + i + 1,
		}

		if err := uc.apartmentRepo.AddPhoto(photo); err != nil {
			objectKey := uc.s3Storage.ExtractObjectKey(url)
			_ = uc.s3Storage.DeleteFile(objectKey)
			return nil, fmt.Errorf("ошибка при сохранении фотографии %d: %w", i+1, err)
		}

		savedURLs = append(savedURLs, url)
	}

	return savedURLs, nil
}

func (uc *ApartmentUseCase) GetPhotosByApartmentID(apartmentID int) ([]*domain.ApartmentPhoto, error) {
	return uc.apartmentRepo.GetPhotosByApartmentID(apartmentID)
}

func (uc *ApartmentUseCase) DeletePhoto(id int) error {
	photo, err := uc.apartmentRepo.GetPhotoByID(id)
	if err != nil {
		return fmt.Errorf("failed to get photo: %w", err)
	}

	if photo == nil {
		return fmt.Errorf("photo with id %d not found", id)
	}

	if err := uc.apartmentRepo.DeletePhoto(id); err != nil {
		return fmt.Errorf("failed to delete photo from database: %w", err)
	}

	objectKey := uc.s3Storage.ExtractObjectKey(photo.URL)
	if err := uc.s3Storage.DeleteFile(objectKey); err != nil {
		fmt.Printf("failed to delete photo from S3: %v\n", err)
	}

	return nil
}

func (uc *ApartmentUseCase) AddDocuments(apartmentID int, filesData [][]byte) ([]string, error) {
	return uc.AddDocumentsWithType(apartmentID, filesData, domain.DocumentTypeOwner)
}

func (uc *ApartmentUseCase) AddDocumentsWithType(apartmentID int, filesData [][]byte, documentType string) ([]string, error) {
	var urls []string

	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}
	if apartment == nil {
		return nil, fmt.Errorf("apartment with id %d not found", apartmentID)
	}

	owner, err := uc.ownerRepo.GetByID(apartment.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment owner: %w", err)
	}

	user, err := uc.userRepo.GetByID(owner.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner user: %w", err)
	}

	for _, fileData := range filesData {

		fileName := fmt.Sprintf("document_%d_%d.jpg", apartmentID, time.Now().Unix())

		currentDate := time.Now().Format("2006-01-02")
		objectKey := fmt.Sprintf("%s/apartment_docs/%d/%s/%s", user.Phone, apartmentID, currentDate, fileName)

		url, err := uc.s3Storage.UploadFile(objectKey, fileData)
		if err != nil {
			return nil, fmt.Errorf("failed to upload document to S3: %w", err)
		}

		document := &domain.ApartmentDocument{
			ApartmentID: apartmentID,
			URL:         url,
			Type:        documentType,
		}

		if err := uc.apartmentRepo.AddDocument(document); err != nil {
			return nil, fmt.Errorf("failed to save document to database: %w", err)
		}

		urls = append(urls, url)
	}

	return urls, nil
}

func (uc *ApartmentUseCase) GetDocumentsByApartmentID(apartmentID int) ([]*domain.ApartmentDocument, error) {
	return uc.apartmentRepo.GetDocumentsByApartmentID(apartmentID)
}

func (uc *ApartmentUseCase) GetDocumentsByApartmentIDAndType(apartmentID int, documentType string) ([]*domain.ApartmentDocument, error) {
	return uc.apartmentRepo.GetDocumentsByApartmentIDAndType(apartmentID, documentType)
}

func (uc *ApartmentUseCase) GetDocumentByID(id int) (*domain.ApartmentDocument, error) {
	return uc.apartmentRepo.GetDocumentByID(id)
}

func (uc *ApartmentUseCase) DeleteDocument(id int) error {

	document, err := uc.apartmentRepo.GetDocumentByID(id)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	if document == nil {
		return fmt.Errorf("document with id %d not found", id)
	}

	if err := uc.apartmentRepo.DeleteDocument(id); err != nil {
		return fmt.Errorf("failed to delete document from database: %w", err)
	}

	objectKey := uc.s3Storage.ExtractObjectKey(document.URL)
	if err := uc.s3Storage.DeleteFile(objectKey); err != nil {
		fmt.Printf("failed to delete document from S3: %v\n", err)
	}

	return nil
}

func (uc *ApartmentUseCase) AddLocation(location *domain.ApartmentLocation) error {
	apartment, err := uc.apartmentRepo.GetByID(location.ApartmentID)
	if err != nil {
		return fmt.Errorf("failed to get apartment: %w", err)
	}
	if apartment == nil {
		return fmt.Errorf("apartment with id %d not found", location.ApartmentID)
	}

	existingLocation, err := uc.apartmentRepo.GetLocationByApartmentID(location.ApartmentID)
	if err != nil {
		return fmt.Errorf("failed to check existing location: %w", err)
	}

	if existingLocation != nil {
		location.ID = existingLocation.ID
		return uc.UpdateLocation(location)
	}

	if err := uc.apartmentRepo.AddLocation(location); err != nil {
		return fmt.Errorf("failed to add location: %w", err)
	}

	return nil
}

func (uc *ApartmentUseCase) UpdateLocation(location *domain.ApartmentLocation) error {
	if err := uc.apartmentRepo.UpdateLocation(location); err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}

	return nil
}

func (uc *ApartmentUseCase) GetLocationByApartmentID(apartmentID int) (*domain.ApartmentLocation, error) {
	return uc.apartmentRepo.GetLocationByApartmentID(apartmentID)
}

func (uc *ApartmentUseCase) GetAllConditions() ([]*domain.ApartmentCondition, error) {
	return uc.apartmentRepo.GetAllConditions()
}

func (uc *ApartmentUseCase) GetConditionByID(id int) (*domain.ApartmentCondition, error) {
	return uc.apartmentRepo.GetConditionByID(id)
}

func (uc *ApartmentUseCase) UpdateStatus(apartmentID int, status domain.ApartmentStatus, comment string) error {
	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return fmt.Errorf("failed to get apartment: %w", err)
	}
	if apartment == nil {
		return fmt.Errorf("apartment with id %d not found", apartmentID)
	}

	apartment.Status = status
	if comment != "" {
		apartment.ModeratorComment = comment
	}

	if err := uc.apartmentRepo.Update(apartment); err != nil {
		return fmt.Errorf("failed to update apartment status: %w", err)
	}

	return nil
}

func (uc *ApartmentUseCase) UpdateApartmentType(apartmentID int, apartmentTypeID int) error {
	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return fmt.Errorf("failed to get apartment: %w", err)
	}
	if apartment == nil {
		return fmt.Errorf("apartment with id %d not found", apartmentID)
	}

	if apartmentTypeID == 0 {
		apartment.ApartmentTypeID = nil
	} else {
		apartment.ApartmentTypeID = &apartmentTypeID
	}

	if err := uc.apartmentRepo.Update(apartment); err != nil {
		return fmt.Errorf("failed to update apartment type: %w", err)
	}

	return nil
}

func (uc *ApartmentUseCase) GetAllHouseRules() ([]*domain.HouseRules, error) {
	return uc.apartmentRepo.GetAllHouseRules()
}

func (uc *ApartmentUseCase) GetHouseRulesByID(id int) (*domain.HouseRules, error) {
	return uc.apartmentRepo.GetHouseRulesByID(id)
}

func (uc *ApartmentUseCase) GetAllPopularAmenities() ([]*domain.PopularAmenities, error) {
	return uc.apartmentRepo.GetAllPopularAmenities()
}

func (uc *ApartmentUseCase) GetPopularAmenitiesByID(id int) (*domain.PopularAmenities, error) {
	return uc.apartmentRepo.GetPopularAmenitiesByID(id)
}

func (uc *ApartmentUseCase) AddHouseRulesToApartment(apartmentID int, houseRuleIDs []int) error {
	return uc.apartmentRepo.AddHouseRulesToApartment(apartmentID, houseRuleIDs)
}

func (uc *ApartmentUseCase) AddAmenitiesToApartment(apartmentID int, amenityIDs []int) error {
	return uc.apartmentRepo.AddAmenitiesToApartment(apartmentID, amenityIDs)
}

func (uc *ApartmentUseCase) GetHouseRulesByApartmentID(apartmentID int) ([]*domain.HouseRules, error) {
	return uc.apartmentRepo.GetHouseRulesByApartmentID(apartmentID)
}

func (uc *ApartmentUseCase) GetAmenitiesByApartmentID(apartmentID int) ([]*domain.PopularAmenities, error) {
	return uc.apartmentRepo.GetAmenitiesByApartmentID(apartmentID)
}

func (uc *ApartmentUseCase) GetBookedDates(apartmentID int, daysAhead int) ([]string, error) {
	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}
	if apartment == nil {
		return nil, fmt.Errorf("apartment with id %d not found", apartmentID)
	}

	startDate := time.Now().UTC()
	endDate := startDate.AddDate(0, 0, daysAhead)

	var bookedDates []string

	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		targetDate := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
		hasAvailableSlots := false

		startHour := 0
		endHour := 24

		if apartment.RentalTypeHourly {
			startHour = 10
			endHour = 22
		}

		for hour := startHour; hour < endHour; hour++ {

			durations := []int{}

			if apartment.RentalTypeDaily {
				durations = append(durations, 24)
			}

			if apartment.RentalTypeHourly {

				possibleDurations := []int{utils.RentalDuration3Hours, utils.RentalDuration6Hours, utils.RentalDuration12Hours}
				for _, duration := range possibleDurations {
					endHour := hour + duration
					if endHour <= 22 {
						durations = append(durations, duration)
					}
				}
			}

			for _, duration := range durations {
				startTimeLocal := time.Date(
					targetDate.Year(), targetDate.Month(), targetDate.Day(),
					hour, 0, 0, 0, utils.KazakhstanTZ,
				)

				startTimeUTC := startTimeLocal.UTC()
				endTimeUTC := startTimeUTC.Add(time.Duration(duration) * time.Hour)

				isAvailable, err := uc.bookingRepo.CheckApartmentAvailability(
					apartmentID,
					startTimeUTC,
					endTimeUTC,
					nil,
				)

				if err == nil && isAvailable {
					hasAvailableSlots = true
					break
				}
			}

			if hasAvailableSlots {
				break
			}
		}

		if !hasAvailableSlots {
			bookedDates = append(bookedDates, targetDate.Format("2006-01-02"))
		}
	}

	return bookedDates, nil
}

func (uc *ApartmentUseCase) GetByCoordinates(minLat, maxLat, minLng, maxLng float64) ([]*domain.ApartmentCoordinates, error) {
	if minLat < -90 || minLat > 90 || maxLat < -90 || maxLat > 90 {
		return nil, fmt.Errorf("широта должна быть между -90 и 90")
	}
	if minLng < -180 || minLng > 180 || maxLng < -180 || maxLng > 180 {
		return nil, fmt.Errorf("долгота должна быть между -180 и 180")
	}
	if minLat > maxLat {
		return nil, fmt.Errorf("минимальная широта должна быть меньше максимальной")
	}
	if minLng > maxLng {
		return nil, fmt.Errorf("минимальная долгота должна быть меньше максимальной")
	}

	return uc.apartmentRepo.GetByCoordinates(minLat, maxLat, minLng, maxLng)
}

func (uc *ApartmentUseCase) GetByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng float64, filters map[string]interface{}) ([]*domain.ApartmentCoordinates, error) {
	if minLat < -90 || minLat > 90 || maxLat < -90 || maxLat > 90 {
		return nil, fmt.Errorf("широта должна быть между -90 и 90")
	}
	if minLng < -180 || minLng > 180 || maxLng < -180 || maxLng > 180 {
		return nil, fmt.Errorf("долгота должна быть между -180 и 180")
	}
	if minLat > maxLat {
		return nil, fmt.Errorf("минимальная широта должна быть меньше максимальной")
	}
	if minLng > maxLng {
		return nil, fmt.Errorf("минимальная долгота должна быть меньше максимальной")
	}

	if _, hasStatus := filters["status"]; !hasStatus {
		filters["status"] = domain.AptStatusApproved
	}

	return uc.apartmentRepo.GetByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng, filters)
}

func (uc *ApartmentUseCase) GetFullApartmentsByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng float64, filters map[string]interface{}) ([]*domain.Apartment, error) {
	if minLat < -90 || minLat > 90 || maxLat < -90 || maxLat > 90 {
		return nil, fmt.Errorf("широта должна быть между -90 и 90")
	}
	if minLng < -180 || minLng > 180 || maxLng < -180 || maxLng > 180 {
		return nil, fmt.Errorf("долгота должна быть между -180 и 180")
	}
	if minLat > maxLat {
		return nil, fmt.Errorf("минимальная широта должна быть меньше максимальной")
	}
	if minLng > maxLng {
		return nil, fmt.Errorf("минимальная долгота должна быть меньше максимальной")
	}

	if _, hasStatus := filters["status"]; !hasStatus {
		filters["status"] = domain.AptStatusApproved
	}

	return uc.apartmentRepo.GetFullApartmentsByCoordinatesWithFilters(minLat, maxLat, minLng, maxLng, filters)
}

func (uc *ApartmentUseCase) GetStatusStatistics() (map[string]int, error) {
	return uc.apartmentRepo.GetStatusStatistics()
}

func (uc *ApartmentUseCase) GetCityStatistics() (map[string]int, error) {
	return uc.apartmentRepo.GetCityStatistics()
}

func (uc *ApartmentUseCase) GetDistrictStatistics() (map[string]int, error) {
	return uc.apartmentRepo.GetDistrictStatistics()
}

func (uc *ApartmentUseCase) ConfirmApartmentAgreement(apartmentID, userID int, request *domain.ConfirmApartmentAgreementRequest) (*domain.Apartment, error) {
	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("квартира не найдена: %w", err)
	}

	owner, err := uc.ownerRepo.GetByUserID(userID)
	if err != nil || owner == nil {
		return nil, fmt.Errorf("пользователь не является владельцем недвижимости")
	}

	if apartment.OwnerID != owner.ID {
		return nil, fmt.Errorf("нет прав для принятия договора этой квартиры")
	}

	if apartment.IsAgreementAccepted {
		return nil, fmt.Errorf("договор уже принят")
	}

	if !request.IsAgreementAccepted {
		return nil, fmt.Errorf("необходимо принять условия договора публикации")
	}

	now := time.Now()
	apartment.IsAgreementAccepted = true
	apartment.AgreementAcceptedAt = &now

	if uc.contractUseCase != nil {
		contract, contractErr := uc.contractUseCase.CreateApartmentContract(apartmentID)
		if contractErr != nil {
			logger.Warn("failed to create contract for apartment",
				slog.Int("apartment_id", apartmentID),
				slog.String("error", contractErr.Error()))
		} else {
			apartment.ContractID = &contract.ID
			logger.Info("contract created and linked to apartment",
				slog.Int("contract_id", contract.ID),
				slog.Int("apartment_id", apartmentID))
		}
	}

	err = uc.apartmentRepo.Update(apartment)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления квартиры: %w", err)
	}

	return apartment, nil
}

func (uc *ApartmentUseCase) IncrementViewCount(apartmentID int) error {
	return uc.apartmentRepo.IncrementViewCount(apartmentID)
}

func (uc *ApartmentUseCase) IncrementBookingCount(apartmentID int) error {
	return uc.apartmentRepo.IncrementBookingCount(apartmentID)
}

func (uc *ApartmentUseCase) AdminUpdateViewCount(apartmentID int, viewCount int) error {
	if viewCount < 0 {
		return fmt.Errorf("количество просмотров не может быть отрицательным")
	}
	return uc.apartmentRepo.AdminUpdateViewCount(apartmentID, viewCount)
}

func (uc *ApartmentUseCase) AdminUpdateBookingCount(apartmentID int, bookingCount int) error {
	if bookingCount < 0 {
		return fmt.Errorf("количество бронирований не может быть отрицательным")
	}
	return uc.apartmentRepo.AdminUpdateBookingCount(apartmentID, bookingCount)
}

func (uc *ApartmentUseCase) AdminResetCounters(apartmentID int) error {
	return uc.apartmentRepo.AdminResetCounters(apartmentID)
}
