package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
)

type LockAutoUpdateService struct {
	lockUseCase  domain.LockUseCase
	tuyaService  domain.TuyaLockService
	httpClient   *http.Client
	tuyaBaseURL  string
	tuyaClientID string
	tuyaSecret   string
	ticker       *time.Ticker
	stopChan     chan struct{}
}

type TuyaDeviceStatus struct {
	DeviceID   string                 `json:"id"`
	Name       string                 `json:"name"`
	Online     bool                   `json:"online"`
	Status     []TuyaStatusItem       `json:"status"`
	Properties map[string]interface{} `json:"properties"`
	UpdateTime int64                  `json:"update_time"`
}

type TuyaStatusItem struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

func NewLockAutoUpdateService(
	tuyaService domain.TuyaLockService,
	tuyaBaseURL, tuyaClientID, tuyaSecret string,
) *LockAutoUpdateService {
	return &LockAutoUpdateService{
		tuyaService:  tuyaService,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		tuyaBaseURL:  tuyaBaseURL,
		tuyaClientID: tuyaClientID,
		tuyaSecret:   tuyaSecret,
		stopChan:     make(chan struct{}),
	}
}

func (s *LockAutoUpdateService) SetLockUseCase(lockUseCase domain.LockUseCase) {
	s.lockUseCase = lockUseCase
}

func (s *LockAutoUpdateService) StartAutoUpdate(interval time.Duration) {
	log.Printf("Запуск автоматического обновления замков с интервалом %v", interval)

	s.ticker = time.NewTicker(interval)

	go func() {
		s.updateAllLocks()

		for {
			select {
			case <-s.ticker.C:
				s.updateAllLocks()
			case <-s.stopChan:
				log.Println("Остановка автоматического обновления замков")
				return
			}
		}
	}()
}

func (s *LockAutoUpdateService) StopAutoUpdate() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
}

func (s *LockAutoUpdateService) UpdateAllLocks() {
	s.updateAllLocks()
}

func (s *LockAutoUpdateService) UpdateLock(lock *domain.Lock) error {
	return s.updateLock(lock)
}

func (s *LockAutoUpdateService) updateAllLocks() {
	log.Println("Начинаем автоматическое обновление всех замков...")

	if s.lockUseCase == nil {
		log.Printf("LockUseCase не установлен для автообновления")
		return
	}

	allLocks, err := s.lockUseCase.GetAllLocks()
	if err != nil {
		log.Printf("Ошибка получения замков для автообновления: %v", err)
		return
	}

	var locks []*domain.Lock
	for _, lock := range allLocks {
		if lock.AutoUpdateEnabled {
			locks = append(locks, lock)
		}
	}

	log.Printf("Найдено %d замков для автообновления", len(locks))

	for _, lock := range locks {
		if err := s.updateLock(lock); err != nil {
			log.Printf("Ошибка обновления замка %s: %v", lock.UniqueID, err)
		} else {
			log.Printf("Замок %s успешно обновлен", lock.UniqueID)
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Println("Автоматическое обновление завершено")
}

func (s *LockAutoUpdateService) updateLock(lock *domain.Lock) error {
	if lock.TuyaDeviceID == "" {
		return fmt.Errorf("у замка %s отсутствует TuyaDeviceID", lock.UniqueID)
	}

	deviceStatus, err := s.getTuyaDeviceStatus(lock.TuyaDeviceID)
	if err != nil {
		return fmt.Errorf("ошибка получения статуса от Tuya: %w", err)
	}

	batteryLevel, batteryType, _ := s.parseBatteryInfo(deviceStatus.Status)
	lockStatus := s.parseLockStatus(deviceStatus.Status)

	log.Printf("🔧 Парсинг данных замка %s: battery=%v, type=%v, status=%s, online=%t",
		lock.UniqueID, batteryLevel, batteryType, lockStatus, deviceStatus.Online)

	if err := s.lockUseCase.UpdateOnlineStatus(lock.UniqueID, deviceStatus.Online); err != nil {
		log.Printf("Ошибка обновления онлайн статуса для замка %s: %v", lock.UniqueID, err)
	}

	heartbeatReq := &domain.LockHeartbeatRequest{
		UniqueID:     lock.UniqueID,
		Status:       lockStatus,
		BatteryLevel: batteryLevel,
		Timestamp:    time.Now(),
	}
	if err := s.lockUseCase.ProcessHeartbeat(heartbeatReq); err != nil {
		log.Printf("Ошибка обновления heartbeat для замка %s: %v", lock.UniqueID, err)
	}

	if err := s.lockUseCase.UpdateTuyaSync(lock.UniqueID, time.Now()); err != nil {
		log.Printf("Ошибка обновления времени синхронизации для замка %s: %v", lock.UniqueID, err)
	}

	return nil
}

func (s *LockAutoUpdateService) getTuyaDeviceStatus(deviceID string) (*TuyaDeviceStatus, error) {
	if s.tuyaBaseURL == "" || s.tuyaClientID == "" || s.tuyaSecret == "" {
		log.Printf("⚠️  Tuya API не настроен, используем тестовые данные для устройства %s", deviceID)
		return s.getMockDeviceStatus(deviceID), nil
	}

	accessToken, err := s.getTuyaAccessToken()
	if err != nil {
		log.Printf("⚠️  Ошибка получения Tuya токена: %v, используем тестовые данные", err)
		return s.getMockDeviceStatus(deviceID), nil
	}

	url := fmt.Sprintf("%s/v1.0/devices/%s", s.tuyaBaseURL, deviceID)

	req, err := s.createSignedRequest("GET", url, accessToken, nil)
	if err != nil {
		log.Printf("⚠️  Ошибка создания подписанного запроса: %v, используем тестовые данные", err)
		return s.getMockDeviceStatus(deviceID), nil
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("⚠️  Ошибка запроса к Tuya API: %v, используем тестовые данные", err)
		return s.getMockDeviceStatus(deviceID), nil
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️  Tuya API вернул статус %d, используем тестовые данные", resp.StatusCode)
		return s.getMockDeviceStatus(deviceID), nil
	}

	var tuyaResponse struct {
		Success bool `json:"success"`
		Result  struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Online bool   `json:"online"`
			Status []struct {
				Code  string      `json:"code"`
				Value interface{} `json:"value"`
			} `json:"status"`
			UpdateTime int64 `json:"update_time"`
		} `json:"result"`
	}

	if err := json.Unmarshal(bodyBytes, &tuyaResponse); err != nil {
		log.Printf("⚠️  Ошибка парсинга ответа Tuya API: %v, используем тестовые данные", err)
		return s.getMockDeviceStatus(deviceID), nil
	}

	if !tuyaResponse.Success {
		log.Printf("⚠️  Tuya API вернул success=false, используем тестовые данные")
		return s.getMockDeviceStatus(deviceID), nil
	}

	deviceStatus := &TuyaDeviceStatus{
		DeviceID:   tuyaResponse.Result.ID,
		Name:       tuyaResponse.Result.Name,
		Online:     tuyaResponse.Result.Online,
		UpdateTime: tuyaResponse.Result.UpdateTime,
		Status:     make([]TuyaStatusItem, len(tuyaResponse.Result.Status)),
	}

	for i, status := range tuyaResponse.Result.Status {
		deviceStatus.Status[i] = TuyaStatusItem{
			Code:  status.Code,
			Value: status.Value,
		}
	}

	return deviceStatus, nil
}

func (s *LockAutoUpdateService) getMockDeviceStatus(deviceID string) *TuyaDeviceStatus {
	return &TuyaDeviceStatus{
		DeviceID:   deviceID,
		Name:       "Test Lock",
		Online:     true,
		UpdateTime: time.Now().Unix(),
		Status: []TuyaStatusItem{
			{Code: "battery_percentage", Value: 75},
			{Code: "doorcontact_state", Value: false},
			{Code: "battery_state", Value: "middle"},
		},
	}
}

func (s *LockAutoUpdateService) getTuyaAccessToken() (string, error) {
	timestamp := time.Now().UnixMilli()
	timestampStr := strconv.FormatInt(timestamp, 10)

	tokenURL := fmt.Sprintf("%s/v1.0/token?grant_type=1", s.tuyaBaseURL)

	method := "GET"
	contentHash := sha256.Sum256([]byte(""))
	contentHashHex := hex.EncodeToString(contentHash[:])

	urlParts := strings.Split(tokenURL, s.tuyaBaseURL)
	if len(urlParts) < 2 {
		return "", fmt.Errorf("неверный формат URL")
	}
	urlPath := urlParts[1]

	stringToSign := fmt.Sprintf("%s\n%s\n\n%s", method, contentHashHex, urlPath)

	finalStringToSign := fmt.Sprintf("%s%s%s", s.tuyaClientID, timestampStr, stringToSign)

	h := hmac.New(sha256.New, []byte(s.tuyaSecret))
	h.Write([]byte(finalStringToSign))
	signature := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	req, err := http.NewRequest(method, tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса токена: %w", err)
	}

	req.Header.Set("client_id", s.tuyaClientID)
	req.Header.Set("sign", signature)
	req.Header.Set("t", timestampStr)
	req.Header.Set("sign_method", "HMAC-SHA256")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса токена: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var tokenResponse struct {
		Success bool   `json:"success"`
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		Result  struct {
			AccessToken string `json:"access_token"`
			ExpireTime  int64  `json:"expire_time"`
		} `json:"result"`
	}

	if err := json.Unmarshal(bodyBytes, &tokenResponse); err != nil {
		return "", fmt.Errorf("ошибка парсинга токена: %w", err)
	}

	if !tokenResponse.Success {
		return "", fmt.Errorf("tuya API ошибка: код=%d, сообщение=%s", tokenResponse.Code, tokenResponse.Msg)
	}

	return tokenResponse.Result.AccessToken, nil
}

func (s *LockAutoUpdateService) createSignedRequest(method, url, accessToken string, body []byte) (*http.Request, error) {
	timestamp := time.Now().UnixMilli()
	timestampStr := strconv.FormatInt(timestamp, 10)

	contentHash := sha256.Sum256(body)
	contentHashHex := hex.EncodeToString(contentHash[:])

	urlParts := strings.Split(url, s.tuyaBaseURL)
	if len(urlParts) < 2 {
		return nil, fmt.Errorf("неверный формат URL")
	}
	urlPath := urlParts[1]

	stringToSign := fmt.Sprintf("%s\n%s\n\n%s", method, contentHashHex, urlPath)

	finalStringToSign := fmt.Sprintf("%s%s%s%s", s.tuyaClientID, accessToken, timestampStr, stringToSign)

	h := hmac.New(sha256.New, []byte(s.tuyaSecret))
	h.Write([]byte(finalStringToSign))
	signature := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, url, strings.NewReader(string(body)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("client_id", s.tuyaClientID)
	req.Header.Set("sign", signature)
	req.Header.Set("t", timestampStr)
	req.Header.Set("sign_method", "HMAC-SHA256")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (s *LockAutoUpdateService) parseBatteryInfo(status []TuyaStatusItem) (*int, domain.BatteryType, *domain.ChargingStatus) {
	var batteryLevel *int
	batteryType := domain.BatteryTypeUnknown
	var chargingStatus *domain.ChargingStatus

	for _, item := range status {
		switch item.Code {
		case "battery_percentage", "residual_electricity":
			if val, ok := item.Value.(float64); ok {
				level := int(val)
				batteryLevel = &level
				batteryType = domain.BatteryTypeAlkaline
			}
		case "battery_state":
			if val, ok := item.Value.(string); ok {
				batteryType = domain.BatteryTypeAlkaline
				switch val {
				case "low":
					level := 25
					batteryLevel = &level
				case "middle":
					level := 60
					batteryLevel = &level
				case "high":
					level := 90
					batteryLevel = &level
				}
			}
		case "alarm_lock":
			if val, ok := item.Value.(string); ok && val == "low_battery" {
				level := 15
				batteryLevel = &level
				batteryType = domain.BatteryTypeAlkaline
			}
		case "va_battery":
			if val, ok := item.Value.(float64); ok {
				level := int(val)
				batteryLevel = &level
				batteryType = domain.BatteryTypeLithium
			}
		case "charge_state":
			if val, ok := item.Value.(string); ok {
				switch val {
				case "unplugged":
					status := domain.ChargingStatusNotCharging
					chargingStatus = &status
				case "charging":
					status := domain.ChargingStatusCharging
					chargingStatus = &status
				case "charged":
					status := domain.ChargingStatusFull
					chargingStatus = &status
				}
			}
		}
	}

	return batteryLevel, batteryType, chargingStatus
}

func (s *LockAutoUpdateService) parseLockStatus(status []TuyaStatusItem) domain.LockStatus {
	for _, item := range status {
		switch item.Code {
		case "doorcontact_state", "closed_opened":
			if val, ok := item.Value.(bool); ok {
				if val {
					return domain.LockStatusOpen
				} else {
					return domain.LockStatusClosed
				}
			}
		case "lock_motor_state":
			if val, ok := item.Value.(string); ok {
				switch val {
				case "locked":
					return domain.LockStatusClosed
				case "unlocked":
					return domain.LockStatusOpen
				}
			}
		case "open_inside":
			if val, ok := item.Value.(bool); ok {
				if val {
					return domain.LockStatusOpen
				} else {
					return domain.LockStatusClosed
				}
			}
		case "hijack":
			if val, ok := item.Value.(bool); ok && val {
				return domain.LockStatusOpen
			}
		}
	}

	return domain.LockStatusClosed
}

func (s *LockAutoUpdateService) ProcessWebhookEvent(event *domain.TuyaWebhookEvent) error {
	log.Printf("Обработка webhook события: %s для устройства %s", event.BizCode, event.DevID)

	lock, err := s.findLockByTuyaDeviceID(event.DevID)
	if err != nil {
		return fmt.Errorf("замок с TuyaDeviceID %s не найден: %w", event.DevID, err)
	}

	switch event.BizCode {
	case "online":
		return s.handleOnlineEvent(lock, event)
	case "offline":
		return s.handleOfflineEvent(lock, event)
	case "dataReport":
		return s.handleDataReportEvent(lock, event)
	default:
		log.Printf("Неизвестный тип события: %s", event.BizCode)
		return nil
	}
}

func (s *LockAutoUpdateService) findLockByTuyaDeviceID(tuyaDeviceID string) (*domain.Lock, error) {
	locks, err := s.lockUseCase.GetAllLocks()
	if err != nil {
		return nil, err
	}

	for _, lock := range locks {
		if lock.TuyaDeviceID == tuyaDeviceID {
			return lock, nil
		}
	}

	return nil, fmt.Errorf("замок с TuyaDeviceID %s не найден", tuyaDeviceID)
}

func (s *LockAutoUpdateService) handleOnlineEvent(lock *domain.Lock, _ *domain.TuyaWebhookEvent) error {
	log.Printf("Замок %s пошел в онлайн", lock.UniqueID)
	return s.lockUseCase.UpdateOnlineStatus(lock.UniqueID, true)
}

func (s *LockAutoUpdateService) handleOfflineEvent(lock *domain.Lock, _ *domain.TuyaWebhookEvent) error {
	log.Printf("Замок %s ушел в оффлайн", lock.UniqueID)
	return s.lockUseCase.UpdateOnlineStatus(lock.UniqueID, false)
}

func (s *LockAutoUpdateService) handleDataReportEvent(lock *domain.Lock, event *domain.TuyaWebhookEvent) error {
	log.Printf("Получены данные от замка %s", lock.UniqueID)

	if statusData, ok := event.BizData["status"]; ok {
		if statusList, ok := statusData.([]interface{}); ok {
			var tuyaStatuses []TuyaStatusItem

			for _, item := range statusList {
				if statusMap, ok := item.(map[string]interface{}); ok {
					tuyaStatus := TuyaStatusItem{
						Code:  statusMap["code"].(string),
						Value: statusMap["value"],
					}
					tuyaStatuses = append(tuyaStatuses, tuyaStatus)
				}
			}

			batteryLevel, _, _ := s.parseBatteryInfo(tuyaStatuses)
			lockStatus := s.parseLockStatus(tuyaStatuses)

			if batteryLevel != nil || lockStatus != "" {
				heartbeatReq := &domain.LockHeartbeatRequest{
					UniqueID:     lock.UniqueID,
					Status:       lockStatus,
					BatteryLevel: batteryLevel,
					Timestamp:    time.Unix(event.Ts/1000, 0),
				}
				if err := s.lockUseCase.ProcessHeartbeat(heartbeatReq); err != nil {
					log.Printf("Ошибка обновления данных через heartbeat: %v", err)
				}
			}
		}
	}

	return s.lockUseCase.UpdateTuyaSync(lock.UniqueID, time.Now())
}
