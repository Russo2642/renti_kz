package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
)

type PaymentHandler struct {
	paymentUseCase domain.PaymentUseCase
}

func NewPaymentHandler(paymentUseCase domain.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{
		paymentUseCase: paymentUseCase,
	}
}

func (h *PaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		payments.POST("/check-status", h.CheckPaymentStatus)
		payments.POST("/check-status-by-order", h.CheckPaymentStatusByOrder)
	}
}

// @Summary Проверка статуса платежа
// @Description Проверяет статус платежа в системе FreedomPay по payment_id
// @Tags payments
// @Accept json
// @Produce json
// @Param request body domain.PaymentStatusRequest true "Payment ID для проверки"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /payments/check-status [post]
func (h *PaymentHandler) CheckPaymentStatus(c *gin.Context) {
	var request domain.PaymentStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	response, err := h.paymentUseCase.CheckPaymentStatus(request.PaymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статус платежа получен", response))
}

// @Summary Проверка статуса платежа по Order ID
// @Description Проверяет статус платежа в системе FreedomPay по order_id
// @Tags payments
// @Accept json
// @Produce json
// @Param request body domain.PaymentOrderStatusRequest true "Order ID и Booking ID для проверки"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /payments/check-status-by-order [post]
func (h *PaymentHandler) CheckPaymentStatusByOrder(c *gin.Context) {
	var request domain.PaymentOrderStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	response, err := h.paymentUseCase.CheckPaymentStatusByOrderID(request.OrderID, request.BookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статус платежа получен", response))
}
