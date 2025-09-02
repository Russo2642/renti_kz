package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/russo2642/renti_kz/internal/config"
	"github.com/russo2642/renti_kz/internal/domain"
)

type OTPService struct {
	config *config.OTPConfig
	client *http.Client
}

func NewOTPService(cfg *config.OTPConfig) *OTPService {
	return &OTPService{
		config: cfg,
		client: GetFastClient(),
	}
}

type otpRequestPayload struct {
	Phone string `json:"phone"`
	From  string `json:"from"`
	Type  string `json:"type"`
	Msg   string `json:"msg"`
}

type otpVerifyPayload struct {
	ID  string `json:"id"`
	Pin string `json:"pin"`
}

type otpStatusPayload struct {
	ID string `json:"id"`
}

func (s *OTPService) RequestOTP(phone string) (*domain.OTPRequestResponse, error) {
	url := fmt.Sprintf("%s/request", s.config.APIBase)

	payload := otpRequestPayload{
		Phone: phone,
		From:  s.config.From,
		Type:  "sms",
		Msg:   s.config.Template,
	}

	response, err := s.makeRequest("POST", url, payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса OTP: %w", err)
	}

	var otpResponse domain.OTPRequestResponse
	if err := json.Unmarshal(response, &otpResponse); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа OTP: %w", err)
	}

	return &otpResponse, nil
}

func (s *OTPService) VerifyOTP(id, code string) (*domain.OTPVerifyResponse, error) {
	url := fmt.Sprintf("%s/verify", s.config.APIBase)

	payload := otpVerifyPayload{
		ID:  id,
		Pin: code,
	}

	response, err := s.makeRequest("POST", url, payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки OTP: %w", err)
	}

	var otpResponse domain.OTPVerifyResponse
	if err := json.Unmarshal(response, &otpResponse); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа проверки OTP: %w", err)
	}

	return &otpResponse, nil
}

func (s *OTPService) CheckStatus(id string) (*domain.OTPStatusResponse, error) {
	url := fmt.Sprintf("%s/status", s.config.APIBase)

	payload := otpStatusPayload{
		ID: id,
	}

	response, err := s.makeRequest("POST", url, payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки статуса OTP: %w", err)
	}

	var otpResponse domain.OTPStatusResponse
	if err := json.Unmarshal(response, &otpResponse); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа статуса OTP: %w", err)
	}

	return &otpResponse, nil
}

func (s *OTPService) makeRequest(method, url string, payload interface{}) ([]byte, error) {
	var body []byte
	var err error

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("ошибка маршалинга payload: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Token", s.config.Token)
	req.Header.Set("Cookie", "PHPSESSID=p8u9ao3dqd9pktg42bb6rq4ev3")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP ошибка: %d", resp.StatusCode)
	}

	responseBody := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			responseBody = append(responseBody, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}

	return responseBody, nil
}
