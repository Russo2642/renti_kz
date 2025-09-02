package usecase

import (
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/pkg/logger"
)

type paymentUseCase struct {
	freedomPayService domain.FreedomPayService
	paymentRepo       domain.PaymentRepository
	paymentLogRepo    domain.PaymentLogRepository
}

func NewPaymentUseCase(
	freedomPayService domain.FreedomPayService,
	paymentRepo domain.PaymentRepository,
	paymentLogRepo domain.PaymentLogRepository,
) domain.PaymentUseCase {
	return &paymentUseCase{
		freedomPayService: freedomPayService,
		paymentRepo:       paymentRepo,
		paymentLogRepo:    paymentLogRepo,
	}
}

func (uc *paymentUseCase) CheckPaymentStatus(paymentID string) (*domain.PaymentStatusResponse, error) {
	startTime := time.Now()

	fpResponse, err := uc.freedomPayService.GetPaymentStatus(paymentID)
	processingDuration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		logger.Error("failed to get payment status from FreedomPay",
			slog.String("payment_id", paymentID),
			slog.String("error", err.Error()),
			slog.Int("duration_ms", processingDuration))
		return nil, fmt.Errorf("ошибка получения статуса платежа: %w", err)
	}

	response := &domain.PaymentStatusResponse{
		PaymentID: paymentID,
	}

	if fpResponse.Status == "ok" {
		response.Exists = true
		response.Amount = fpResponse.Amount
		response.Currency = fpResponse.Currency
		response.Status = fpResponse.PaymentStatus
		response.PaymentMethod = fpResponse.PaymentMethod
		response.CreateDate = fpResponse.CreateDate

		if fpResponse.PaymentID == "0" {
			response.Exists = false
			response.ErrorMessage = "Платеж не найден в системе FreedomPay"
		}
	} else {
		response.Exists = false
		if fpResponse.ErrorDescription != "" {
			response.ErrorMessage = fpResponse.ErrorDescription
		} else {
			response.ErrorMessage = "Неизвестная ошибка при проверке статуса платежа"
		}
	}

	logger.Info("payment status checked",
		slog.String("payment_id", paymentID),
		slog.String("status", response.Status),
		slog.Bool("exists", response.Exists),
		slog.Int("duration_ms", processingDuration))

	return response, nil
}

func (uc *paymentUseCase) CheckPaymentStatusWithBooking(paymentID string, bookingID int64) (*domain.PaymentStatusResponse, error) {
	startTime := time.Now()

	fpResponse, err := uc.freedomPayService.GetPaymentStatus(paymentID)
	processingDuration := int(time.Since(startTime).Milliseconds())

	logEntry := &domain.PaymentLog{
		BookingID:          bookingID,
		FPPaymentID:        paymentID,
		Action:             domain.PaymentLogActionCheckStatus,
		FPResponse:         fpResponse,
		ProcessingDuration: &processingDuration,
		Source:             domain.PaymentLogSourceAPI,
		Success:            err == nil && fpResponse != nil,
	}

	if err != nil {
		errorMessage := err.Error()
		logEntry.ErrorMessage = &errorMessage
		logger.Error("failed to get payment status from FreedomPay",
			slog.String("payment_id", paymentID),
			slog.Int64("booking_id", bookingID),
			slog.String("error", err.Error()),
			slog.Int("duration_ms", processingDuration))
	}

	if logErr := uc.paymentLogRepo.Create(logEntry); logErr != nil {
		logger.Warn("failed to save payment log",
			slog.String("payment_id", paymentID),
			slog.Int64("booking_id", bookingID),
			slog.String("error", logErr.Error()))
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка получения статуса платежа: %w", err)
	}

	response := &domain.PaymentStatusResponse{
		PaymentID: paymentID,
	}

	if fpResponse.Status == "ok" {
		response.Exists = true
		response.Amount = fpResponse.Amount
		response.Currency = fpResponse.Currency
		response.Status = fpResponse.PaymentStatus
		response.PaymentMethod = fpResponse.PaymentMethod
		response.CreateDate = fpResponse.CreateDate

		if fpResponse.PaymentID == "0" {
			response.Exists = false
			response.ErrorMessage = "Платеж не найден в системе FreedomPay"
		}
	} else {
		response.Exists = false
		if fpResponse.ErrorDescription != "" {
			response.ErrorMessage = fpResponse.ErrorDescription
		} else {
			response.ErrorMessage = "Неизвестная ошибка при проверке статуса платежа"
		}
	}

	logger.Info("payment status checked with booking",
		slog.String("payment_id", paymentID),
		slog.Int64("booking_id", bookingID),
		slog.String("status", response.Status),
		slog.Bool("exists", response.Exists),
		slog.Int("duration_ms", processingDuration))

	return response, nil
}

func (uc *paymentUseCase) CheckPaymentStatusByOrderID(orderID string, bookingID int64) (*domain.PaymentStatusResponse, error) {
	startTime := time.Now()

	fpResponse, err := uc.freedomPayService.GetPaymentStatusByOrderID(orderID)
	processingDuration := int(time.Since(startTime).Milliseconds())

	logEntry := &domain.PaymentLog{
		BookingID:          bookingID,
		FPPaymentID:        orderID,
		Action:             domain.PaymentLogActionCheckStatus,
		FPResponse:         fpResponse,
		ProcessingDuration: &processingDuration,
		Source:             domain.PaymentLogSourceAPI,
		Success:            err == nil && fpResponse != nil,
	}

	if err != nil {
		errorMessage := err.Error()
		logEntry.ErrorMessage = &errorMessage
		logger.Error("failed to get payment status from FreedomPay by order ID",
			slog.String("order_id", orderID),
			slog.Int64("booking_id", bookingID),
			slog.String("error", err.Error()),
			slog.Int("duration_ms", processingDuration))
	}

	if logErr := uc.paymentLogRepo.Create(logEntry); logErr != nil {
		logger.Warn("failed to save payment log",
			slog.String("order_id", orderID),
			slog.Int64("booking_id", bookingID),
			slog.String("error", logErr.Error()))
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка получения статуса платежа по order_id: %w", err)
	}

	response := &domain.PaymentStatusResponse{
		PaymentID: fpResponse.PaymentID,
	}

	if fpResponse.Status == "ok" {
		response.Exists = true
		response.Amount = fpResponse.Amount
		response.Currency = fpResponse.Currency
		response.Status = fpResponse.PaymentStatus
		response.PaymentMethod = fpResponse.PaymentMethod
		response.CreateDate = fpResponse.CreateDate

		if fpResponse.PaymentID == "0" {
			response.Exists = false
			response.ErrorMessage = "Платеж не найден в системе FreedomPay"
		}
	} else {
		response.Exists = false
		if fpResponse.ErrorDescription != "" {
			response.ErrorMessage = fpResponse.ErrorDescription
		} else {
			response.ErrorMessage = "Неизвестная ошибка при проверке статуса платежа"
		}
	}

	logger.Info("payment status checked by order ID",
		slog.String("order_id", orderID),
		slog.Int64("booking_id", bookingID),
		slog.String("payment_id", response.PaymentID),
		slog.String("status", response.Status),
		slog.Bool("exists", response.Exists),
		slog.Int("duration_ms", processingDuration))

	return response, nil
}

func (uc *paymentUseCase) RefundPayment(paymentID string, refundAmount *int) (*domain.RefundResponse, error) {
	startTime := time.Now()

	fpResponse, err := uc.freedomPayService.RefundPayment(paymentID, refundAmount)
	processingDuration := int(time.Since(startTime).Milliseconds())

	logEntry := &domain.PaymentLog{
		FPPaymentID:        paymentID,
		Action:             domain.PaymentLogActionRefundPayment,
		ProcessingDuration: &processingDuration,
		Source:             domain.PaymentLogSourceAPI,
		Success:            err == nil && fpResponse != nil && fpResponse.Status == "ok",
	}

	if err != nil {
		errorMessage := err.Error()
		logEntry.ErrorMessage = &errorMessage
		logger.Error("failed to refund payment via FreedomPay",
			slog.String("payment_id", paymentID),
			slog.String("error", err.Error()),
			slog.Int("duration_ms", processingDuration))
	}

	if logErr := uc.paymentLogRepo.Create(logEntry); logErr != nil {
		logger.Warn("failed to save refund payment log",
			slog.String("payment_id", paymentID),
			slog.String("error", logErr.Error()))
	}

	if err != nil {
		return &domain.RefundResponse{
			Success:   false,
			PaymentID: paymentID,
			Message:   fmt.Sprintf("Ошибка возврата платежа: %s", err.Error()),
		}, err
	}

	if fpResponse.Status != "ok" {
		errorMsg := fmt.Sprintf("Возврат отклонен FreedomPay: %s - %s", fpResponse.ErrorCode, fpResponse.ErrorDescription)
		return &domain.RefundResponse{
			Success:   false,
			PaymentID: paymentID,
			Message:   errorMsg,
		}, errors.New(errorMsg)
	}

	logger.Info("payment refunded successfully",
		slog.String("payment_id", paymentID),
		slog.Int("duration_ms", processingDuration))

	return &domain.RefundResponse{
		Success:   true,
		PaymentID: paymentID,
		Message:   "Платеж успешно возвращен",
	}, nil
}
