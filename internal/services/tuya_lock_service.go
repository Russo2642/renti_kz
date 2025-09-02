package services

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	EmptySHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

type TuyaConfig struct {
	ClientID     string
	ClientSecret string
	APIBase      string
	TimeZone     string
}

type TuyaLockService struct {
	config TuyaConfig
}

type TuyaTokenResponse struct {
	Success bool `json:"success"`
	Result  struct {
		AccessToken string `json:"access_token"`
		ExpireIn    int    `json:"expire_in"`
	} `json:"result"`
}

type TuyaTicketResponse struct {
	Success bool `json:"success"`
	Result  struct {
		TicketID  string `json:"ticket_id"`
		TicketKey string `json:"ticket_key"`
	} `json:"result"`
}

type TuyaPasswordResponse struct {
	Success bool `json:"success"`
	Result  struct {
		ID int64 `json:"id"`
	} `json:"result"`
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type ScheduleItem struct {
	EffectiveTime int `json:"effective_time"`
	InvalidTime   int `json:"invalid_time"`
	WorkingDay    int `json:"working_day"`
}

type TempPasswordRequest struct {
	Name          string         `json:"name"`
	Password      string         `json:"password"`
	PasswordType  string         `json:"password_type"`
	TicketID      string         `json:"ticket_id"`
	EffectiveTime int64          `json:"effective_time"`
	InvalidTime   int64          `json:"invalid_time"`
	TimeZone      string         `json:"time_zone"`
	ScheduleList  []ScheduleItem `json:"schedule_list"`
}

func NewTuyaLockService(config TuyaConfig) *TuyaLockService {
	return &TuyaLockService{
		config: config,
	}
}

func (t *TuyaLockService) hmacSHA256Hex(key, message string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func (t *TuyaLockService) sha256Hex(data []byte) string {
	if data == nil {
		return EmptySHA256
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func (t *TuyaLockService) stringToSign(method, path string, query map[string]string, body []byte) string {
	var queryString string
	if len(query) > 0 {
		var params []string
		for k, v := range query {
			params = append(params, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
		}
		sort.Strings(params)
		queryString = "?" + strings.Join(params, "&")
	}

	fullURL := path + queryString
	bodyHash := t.sha256Hex(body)

	return fmt.Sprintf("%s\n%s\n\n%s", strings.ToUpper(method), bodyHash, fullURL)
}

func (t *TuyaLockService) signToken(path string, query map[string]string) (string, string) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	stringToSign := t.stringToSign("GET", path, query, nil)
	sign := t.hmacSHA256Hex(t.config.ClientSecret, t.config.ClientID+timestamp+stringToSign)
	return sign, timestamp
}

func (t *TuyaLockService) signBusiness(path string, body []byte, token string) (string, string) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	stringToSign := t.stringToSign("POST", path, nil, body)
	sign := t.hmacSHA256Hex(t.config.ClientSecret, t.config.ClientID+token+timestamp+stringToSign)
	return sign, timestamp
}

func (t *TuyaLockService) signDelete(path string, token string) (string, string) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	stringToSign := t.stringToSign("DELETE", path, nil, nil)
	sign := t.hmacSHA256Hex(t.config.ClientSecret, t.config.ClientID+token+timestamp+stringToSign)
	return sign, timestamp
}

func (t *TuyaLockService) GetAccessToken() (string, error) {
	path := "/v1.0/token"
	query := map[string]string{"grant_type": "1"}

	sign, timestamp := t.signToken(path, query)

	req, err := http.NewRequest("GET", t.config.APIBase+path, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("client_id", t.config.ClientID)
	req.Header.Set("sign_method", "HMAC-SHA256")
	req.Header.Set("t", timestamp)
	req.Header.Set("sign", sign)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var tokenResp TuyaTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	if !tokenResp.Success {
		return "", fmt.Errorf("ошибка получения токена: %s", string(body))
	}

	return tokenResp.Result.AccessToken, nil
}

func (t *TuyaLockService) GetPasswordTicket(token, deviceID string) (string, string, error) {
	path := fmt.Sprintf("/v1.0/devices/%s/door-lock/password-ticket", deviceID)
	body := []byte("{}")

	sign, timestamp := t.signBusiness(path, body, token)

	req, err := http.NewRequest("POST", t.config.APIBase+path, bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("client_id", t.config.ClientID)
	req.Header.Set("sign_method", "HMAC-SHA256")
	req.Header.Set("t", timestamp)
	req.Header.Set("sign", sign)
	req.Header.Set("access_token", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var ticketResp TuyaTicketResponse
	if err := json.Unmarshal(respBody, &ticketResp); err != nil {
		return "", "", fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	if !ticketResp.Success {
		return "", "", fmt.Errorf("ошибка получения ticket: %s", string(respBody))
	}

	return ticketResp.Result.TicketID, ticketResp.Result.TicketKey, nil
}

func (t *TuyaLockService) DecryptTicketKey(ticketKeyHex string) ([]byte, error) {
	key := []byte(t.config.ClientSecret)
	if len(key) != 32 {
		return nil, fmt.Errorf("неверная длина ключа: ожидается 32 байта, получено %d", len(key))
	}

	encrypted, err := hex.DecodeString(ticketKeyHex)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования hex: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания AES cipher: %w", err)
	}

	if len(encrypted)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("зашифрованные данные не кратны размеру блока")
	}

	decrypted := make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], encrypted[i:i+aes.BlockSize])
	}

	originalKey := t.removePKCS7Padding(decrypted)
	if len(originalKey) < 16 {
		return nil, fmt.Errorf("недостаточно данных после удаления padding")
	}

	return originalKey[:16], nil
}

func (t *TuyaLockService) removePKCS7Padding(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	paddingLen := int(data[len(data)-1])
	if paddingLen > len(data) || paddingLen > aes.BlockSize {
		return data
	}

	for i := len(data) - paddingLen; i < len(data); i++ {
		if data[i] != byte(paddingLen) {
			return data
		}
	}

	return data[:len(data)-paddingLen]
}

func (t *TuyaLockService) EncryptPassword(rawPassword string, originalKey []byte) (string, error) {
	if len(originalKey) != 16 {
		return "", fmt.Errorf("неверная длина ключа: ожидается 16 байт, получено %d", len(originalKey))
	}

	block, err := aes.NewCipher(originalKey)
	if err != nil {
		return "", fmt.Errorf("ошибка создания AES cipher: %w", err)
	}

	passwordBytes := []byte(rawPassword)
	paddedPassword := t.addPKCS7Padding(passwordBytes, aes.BlockSize)

	encrypted := make([]byte, len(paddedPassword))
	for i := 0; i < len(paddedPassword); i += aes.BlockSize {
		block.Encrypt(encrypted[i:i+aes.BlockSize], paddedPassword[i:i+aes.BlockSize])
	}

	return strings.ToUpper(hex.EncodeToString(encrypted)), nil
}

func (t *TuyaLockService) addPKCS7Padding(data []byte, blockSize int) []byte {
	paddingLen := blockSize - (len(data) % blockSize)
	padding := bytes.Repeat([]byte{byte(paddingLen)}, paddingLen)
	return append(data, padding...)
}

func (t *TuyaLockService) CreateTempPasswordWithTimes(token, deviceID, name, encryptedPassword, ticketID string, validFrom, validUntil time.Time) (*TuyaPasswordResponse, error) {
	kzZone := time.FixedZone("Asia/Almaty", 5*60*60)

	validFromLocal := validFrom.In(kzZone)
	validUntilLocal := validUntil.In(kzZone)

	nowLocal := time.Now().In(kzZone)
	if validFromLocal.Before(nowLocal.Add(-1 * time.Minute)) {
		validFromLocal = nowLocal
	}

	request := TempPasswordRequest{
		Name:          name,
		Password:      encryptedPassword,
		PasswordType:  "ticket",
		TicketID:      ticketID,
		EffectiveTime: validFromLocal.Unix(),
		InvalidTime:   validUntilLocal.Unix(),
		TimeZone:      t.config.TimeZone,
		ScheduleList: []ScheduleItem{
			{
				EffectiveTime: 0,
				InvalidTime:   1440,
				WorkingDay:    127,
			},
		},
	}

	return t.createTempPasswordRequest(token, deviceID, request)
}

func (t *TuyaLockService) createTempPasswordRequest(token, deviceID string, request TempPasswordRequest) (*TuyaPasswordResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации JSON: %w", err)
	}

	path := fmt.Sprintf("/v1.0/devices/%s/door-lock/temp-password", deviceID)
	sign, timestamp := t.signBusiness(path, body, token)

	req, err := http.NewRequest("POST", t.config.APIBase+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("client_id", t.config.ClientID)
	req.Header.Set("sign_method", "HMAC-SHA256")
	req.Header.Set("t", timestamp)
	req.Header.Set("sign", sign)
	req.Header.Set("access_token", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var passwordResp TuyaPasswordResponse
	if err := json.Unmarshal(respBody, &passwordResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	return &passwordResp, nil
}

func (t *TuyaLockService) GenerateTemporaryPasswordWithTimes(deviceID, name, rawPassword string, validFrom, validUntil time.Time) (string, int64, error) {
	token, err := t.GetAccessToken()
	if err != nil {
		return "", 0, fmt.Errorf("ошибка получения токена: %w", err)
	}

	ticketID, ticketKey, err := t.GetPasswordTicket(token, deviceID)
	if err != nil {
		return "", 0, fmt.Errorf("ошибка получения ticket: %w", err)
	}

	originalKey, err := t.DecryptTicketKey(ticketKey)
	if err != nil {
		return "", 0, fmt.Errorf("ошибка расшифровки ticket key: %w", err)
	}

	encryptedPassword, err := t.EncryptPassword(rawPassword, originalKey)
	if err != nil {
		return "", 0, fmt.Errorf("ошибка шифрования пароля: %w", err)
	}

	resp, err := t.CreateTempPasswordWithTimes(token, deviceID, name, encryptedPassword, ticketID, validFrom, validUntil)
	if err != nil {
		return "", 0, fmt.Errorf("ошибка создания пароля: %w", err)
	}

	if !resp.Success {
		return "", 0, fmt.Errorf("не удалось создать пароль: код %d, сообщение: %s", resp.Code, resp.Msg)
	}

	return rawPassword, resp.Result.ID, nil
}

func (t *TuyaLockService) DeleteTempPassword(deviceID string, passwordID int64) error {
	token, err := t.GetAccessToken()
	if err != nil {
		return fmt.Errorf("ошибка получения токена: %w", err)
	}

	path := fmt.Sprintf("/v1.0/devices/%s/door-lock/temp-passwords/%d", deviceID, passwordID)
	sign, timestamp := t.signDelete(path, token)

	req, err := http.NewRequest("DELETE", t.config.APIBase+path, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("client_id", t.config.ClientID)
	req.Header.Set("sign_method", "HMAC-SHA256")
	req.Header.Set("t", timestamp)
	req.Header.Set("sign", sign)
	req.Header.Set("access_token", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var tuyaResp struct {
		Success bool   `json:"success"`
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
	}

	if err := json.Unmarshal(respBody, &tuyaResp); err == nil {
		if !tuyaResp.Success {
			if tuyaResp.Code == 1004 {
				return fmt.Errorf("ошибка подписи Tuya API: %s", tuyaResp.Msg)
			}
			return fmt.Errorf("ошибка Tuya API: код %d, сообщение: %s", tuyaResp.Code, tuyaResp.Msg)
		}
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка удаления пароля: Status %d, Body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
