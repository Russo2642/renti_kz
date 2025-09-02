package services

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/russo2642/renti_kz/internal/domain"
)

type FreedomPayService struct {
	merchantID string
	secretKey  string
	apiURL     string
	client     *http.Client
}

func NewFreedomPayService(merchantID, secretKey, apiURL string) domain.FreedomPayService {
	return &FreedomPayService{
		merchantID: merchantID,
		secretKey:  secretKey,
		apiURL:     apiURL,
		client:     GetFastClient(),
	}
}

func (s *FreedomPayService) GetPaymentStatus(paymentID string) (*domain.FreedomPayStatusResponse, error) {
	salt := uuid.New().String()

	data := map[string]string{
		"pg_merchant_id": s.merchantID,
		"pg_payment_id":  paymentID,
		"pg_salt":        salt,
	}

	signature := s.generateSignature(data, "get_status3.php")
	data["pg_sig"] = signature

	return s.makeRequest("get_status3.php", data)
}

func (s *FreedomPayService) GetPaymentStatusByOrderID(orderID string) (*domain.FreedomPayStatusResponse, error) {
	salt := uuid.New().String()

	data := map[string]string{
		"pg_merchant_id": s.merchantID,
		"pg_order_id":    orderID,
		"pg_salt":        salt,
	}

	signature := s.generateSignatureByOrderID(data, "get_status3.php")
	data["pg_sig"] = signature

	return s.makeRequest("get_status3.php", data)
}

func (s *FreedomPayService) RefundPayment(paymentID string, refundAmount *int) (*domain.FreedomPayRefundResponse, error) {
	salt := uuid.New().String()

	data := map[string]string{
		"pg_merchant_id": s.merchantID,
		"pg_payment_id":  paymentID,
		"pg_salt":        salt,
	}

	if refundAmount != nil && *refundAmount > 0 {
		data["pg_refund_amount"] = fmt.Sprintf("%d", *refundAmount)
	}

	signature := s.generateRefundSignature(data)
	data["pg_sig"] = signature

	return s.makeRefundRequest(data)
}

func (s *FreedomPayService) generateSignature(data map[string]string, endpoint string) string {
	// Согласно документации FreedomPay для get_status3.php
	// Формат: 'get_status3.php;pg_merchant_id;pg_order_id;pg_payment_id;pg_salt;secret_key'
	// Если передан только pg_payment_id: 'get_status3.php;pg_merchant_id;pg_payment_id;pg_salt;secret_key'

	signatureString := fmt.Sprintf("%s;%s;%s;%s;%s",
		endpoint,
		s.merchantID,
		data["pg_payment_id"],
		data["pg_salt"],
		s.secretKey,
	)

	hash := md5.Sum([]byte(signatureString))
	return fmt.Sprintf("%x", hash)
}

func (s *FreedomPayService) generateSignatureByOrderID(data map[string]string, endpoint string) string {
	// Согласно документации FreedomPay для get_status3.php с pg_order_id
	// Формат: 'get_status3.php;pg_merchant_id;pg_order_id;pg_salt;secret_key'

	signatureString := fmt.Sprintf("%s;%s;%s;%s;%s",
		endpoint,
		s.merchantID,
		data["pg_order_id"],
		data["pg_salt"],
		s.secretKey,
	)

	hash := md5.Sum([]byte(signatureString))
	return fmt.Sprintf("%x", hash)
}

func (s *FreedomPayService) makeRequest(endpoint string, data map[string]string) (*domain.FreedomPayStatusResponse, error) {
	requestURL := fmt.Sprintf("%s/%s", s.apiURL, endpoint)

	formData := url.Values{}
	for key, value := range data {
		formData.Set(key, value)
	}

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var response domain.FreedomPayStatusResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга XML ответа: %w", err)
	}

	return &response, nil
}

func (s *FreedomPayService) generateRefundSignature(data map[string]string) string {
	// Согласно документации FreedomPay для revoke.php
	// Формат: 'revoke.php;pg_merchant_id;pg_payment_id;pg_refund_amount;pg_salt;secret_key'
	// Если pg_refund_amount не передан, то: 'revoke.php;pg_merchant_id;pg_payment_id;pg_salt;secret_key'

	var signatureString string

	if refundAmount, exists := data["pg_refund_amount"]; exists && refundAmount != "" {
		signatureString = fmt.Sprintf("revoke.php;%s;%s;%s;%s;%s",
			s.merchantID,
			data["pg_payment_id"],
			refundAmount,
			data["pg_salt"],
			s.secretKey,
		)
	} else {
		signatureString = fmt.Sprintf("revoke.php;%s;%s;%s;%s",
			s.merchantID,
			data["pg_payment_id"],
			data["pg_salt"],
			s.secretKey,
		)
	}

	hash := md5.Sum([]byte(signatureString))
	return fmt.Sprintf("%x", hash)
}

func (s *FreedomPayService) makeRefundRequest(data map[string]string) (*domain.FreedomPayRefundResponse, error) {
	requestURL := fmt.Sprintf("%s/revoke.php", s.apiURL)

	formData := url.Values{}
	for key, value := range data {
		formData.Set(key, value)
	}

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса возврата: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса возврата: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа возврата: %w", err)
	}

	var response domain.FreedomPayRefundResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга XML ответа возврата: %w", err)
	}

	return &response, nil
}
