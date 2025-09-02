package services

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"context"

	"github.com/redis/go-redis/v9"
	"github.com/russo2642/renti_kz/internal/domain"
)

type contractService struct {
	contractRepo         domain.ContractRepository
	contractTemplateRepo domain.ContractTemplateRepository
	bookingRepo          domain.BookingRepository
	apartmentRepo        domain.ApartmentRepository
	userRepo             domain.UserRepository
	renterRepo           domain.RenterRepository
	propertyOwnerRepo    domain.PropertyOwnerRepository
	redisClient          *redis.Client
	templatesPath        string
}

type ContractServiceConfig struct {
	ContractRepo         domain.ContractRepository
	ContractTemplateRepo domain.ContractTemplateRepository
	BookingRepo          domain.BookingRepository
	ApartmentRepo        domain.ApartmentRepository
	UserRepo             domain.UserRepository
	RenterRepo           domain.RenterRepository
	PropertyOwnerRepo    domain.PropertyOwnerRepository
	RedisClient          *redis.Client
	TemplatesPath        string
}

func NewContractService(config ContractServiceConfig) domain.ContractService {
	return &contractService{
		contractRepo:         config.ContractRepo,
		contractTemplateRepo: config.ContractTemplateRepo,
		bookingRepo:          config.BookingRepo,
		apartmentRepo:        config.ApartmentRepo,
		userRepo:             config.UserRepo,
		renterRepo:           config.RenterRepo,
		propertyOwnerRepo:    config.PropertyOwnerRepo,
		redisClient:          config.RedisClient,
		templatesPath:        config.TemplatesPath,
	}
}

const (
	contractCacheKeyPrefix = "contract_html:"
	cacheExpiration        = 7 * 24 * time.Hour
)

func (s *contractService) GetOrGenerateContractHTML(contractID int) (string, error) {
	cacheKey := fmt.Sprintf("%s%d", contractCacheKeyPrefix, contractID)
	ctx := context.Background()

	cachedHTML, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil && cachedHTML != "" {
		return cachedHTML, nil
	}

	html, err := s.GenerateContractHTML(contractID)
	if err != nil {
		return "", err
	}

	err = s.redisClient.Set(ctx, cacheKey, html, cacheExpiration).Err()
	if err != nil {
		fmt.Printf("Warning: failed to cache contract HTML: %v\n", err)
	}

	return html, nil
}

func (s *contractService) GenerateContractHTML(contractID int) (string, error) {
	contract, err := s.contractRepo.GetByID(contractID)
	if err != nil {
		return "", fmt.Errorf("договор не найден: %w", err)
	}

	var templateData *domain.ContractTemplateData

	switch contract.Type {
	case domain.ContractTypeRental:
		templateData, err = s.prepareRentalTemplateData(contract)
	case domain.ContractTypeApartment:
		templateData, err = s.prepareApartmentTemplateData(contract)
	default:
		return "", fmt.Errorf("неподдерживаемый тип договора: %s", contract.Type)
	}

	if err != nil {
		return "", fmt.Errorf("ошибка подготовки данных: %w", err)
	}

	html, err := s.RenderTemplate(contract.Type, templateData)
	if err != nil {
		return "", fmt.Errorf("ошибка рендеринга шаблона: %w", err)
	}

	return html, nil
}

func (s *contractService) prepareRentalTemplateData(contract *domain.Contract) (*domain.ContractTemplateData, error) {
	if contract.BookingID == nil {
		return nil, fmt.Errorf("у rental договора должно быть booking_id")
	}

	snapshot, err := contract.GetSnapshotData()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения снапшота: %w", err)
	}

	if snapshot == nil {
		return nil, fmt.Errorf("снапшот данных не найден")
	}

	contactInfo, err := s.getCurrentContactInfo(*contract.BookingID, contract.ApartmentID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения контактов: %w", err)
	}

	return &domain.ContractTemplateData{
		RentalContractSnapshot: snapshot,
		ContractContactInfo:    contactInfo,
		ContractDate:           snapshot.ContractDate,
		TemplateVersion:        contract.TemplateVersion,
	}, nil
}

func (s *contractService) prepareApartmentTemplateData(contract *domain.Contract) (*domain.ContractTemplateData, error) {
	snapshot, err := contract.GetApartmentSnapshotData()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения снапшота apartment contract: %w", err)
	}

	if snapshot == nil {
		return nil, fmt.Errorf("snapshot данные отсутствуют в apartment contract")
	}

	templateData := &domain.ContractTemplateData{
		ApartmentContractSnapshot: snapshot,
		ContractDate:              snapshot.ContractDate,
		TemplateVersion:           contract.TemplateVersion,
	}

	return templateData, nil
}

func (s *contractService) getCurrentContactInfo(bookingID, apartmentID int) (*domain.ContractContactInfo, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("бронирование не найдено: %w", err)
	}

	renter, err := s.renterRepo.GetByID(booking.RenterID)
	if err != nil {
		return nil, fmt.Errorf("арендатор не найден: %w", err)
	}

	renterUser, err := s.userRepo.GetByID(renter.UserID)
	if err != nil {
		return nil, fmt.Errorf("пользователь арендатора не найден: %w", err)
	}

	apartment, err := s.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("квартира не найдена: %w", err)
	}

	owner, err := s.propertyOwnerRepo.GetByID(apartment.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("арендодатель не найден: %w", err)
	}

	ownerUser, err := s.userRepo.GetByID(owner.UserID)
	if err != nil {
		return nil, fmt.Errorf("пользователь арендодателя не найден: %w", err)
	}

	return &domain.ContractContactInfo{
		RenterPhone: renterUser.Phone,
		RenterEmail: renterUser.Email,
		OwnerPhone:  ownerUser.Phone,
		OwnerEmail:  ownerUser.Email,
	}, nil
}

func (s *contractService) RenderTemplate(contractType domain.ContractType, data *domain.ContractTemplateData) (string, error) {
	var templateFileName string

	switch contractType {
	case domain.ContractTypeRental:
		templateFileName = "rental_contract.html"
	case domain.ContractTypeApartment:
		templateFileName = "apartment_contract.html"
	default:
		return "", fmt.Errorf("неподдерживаемый тип договора: %s", contractType)
	}

	templatePath := filepath.Join(s.templatesPath, templateFileName)

	tmpl := template.New(templateFileName).Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
		"replace": func(old, new, str string) string {
			return strings.ReplaceAll(str, old, new)
		},
	})

	tmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга шаблона %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения шаблона: %w", err)
	}

	return buf.String(), nil
}

func (s *contractService) minifyHTML(html string) string {
	lines := strings.Split(html, "\n")
	var cleanLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	return strings.Join(cleanLines, "\n")
}

func (s *contractService) InvalidateContractCache(contractID int) error {
	cacheKey := fmt.Sprintf("%s%d", contractCacheKeyPrefix, contractID)
	ctx := context.Background()

	err := s.redisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		return fmt.Errorf("ошибка удаления из кэша: %w", err)
	}

	return nil
}

func (s *contractService) WarmupContractCache(contractID int) error {
	_, err := s.GetOrGenerateContractHTML(contractID)
	return err
}

func (s *contractService) GetCacheStats() (map[string]interface{}, error) {
	ctx := context.Background()

	keys, err := s.redisClient.Keys(ctx, contractCacheKeyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"cached_contracts": len(keys),
		"cache_prefix":     contractCacheKeyPrefix,
		"cache_ttl":        cacheExpiration.String(),
	}

	return stats, nil
}

func (s *contractService) ClearAllContractCache() error {
	ctx := context.Background()

	keys, err := s.redisClient.Keys(ctx, contractCacheKeyPrefix+"*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		err = s.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			return err
		}
	}

	return nil
}
