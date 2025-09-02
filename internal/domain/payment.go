package domain

import "time"

type PaymentStatusRequest struct {
	PaymentID string `json:"payment_id" binding:"required"`
}

type PaymentOrderStatusRequest struct {
	OrderID   string `json:"order_id" binding:"required"`
	BookingID int64  `json:"booking_id" binding:"required"`
}

type ProcessPaymentRequest struct {
	PaymentID string `json:"payment_id,omitempty"`
	OrderID   string `json:"order_id,omitempty"`
}

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusSuccess    PaymentStatus = "success"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusExpired    PaymentStatus = "expired"
	PaymentStatusCanceled   PaymentStatus = "canceled"
)

type PaymentLogAction string

const (
	PaymentLogActionCreatePayment       PaymentLogAction = "create_payment"
	PaymentLogActionCheckStatus         PaymentLogAction = "check_status"
	PaymentLogActionProcessPayment      PaymentLogAction = "process_payment"
	PaymentLogActionWebhookNotification PaymentLogAction = "webhook_notification"
	PaymentLogActionRefundPayment       PaymentLogAction = "refund_payment"
)

type PaymentLogSource string

const (
	PaymentLogSourceAPI       PaymentLogSource = "api"
	PaymentLogSourceWebhook   PaymentLogSource = "webhook"
	PaymentLogSourceManual    PaymentLogSource = "manual"
	PaymentLogSourceScheduler PaymentLogSource = "scheduler"
)

type Payment struct {
	ID                 int64                     `json:"id"`
	BookingID          int64                     `json:"booking_id"`
	PaymentID          string                    `json:"payment_id"`
	Amount             int                       `json:"amount"`
	Currency           string                    `json:"currency"`
	Status             PaymentStatus             `json:"status"`
	PaymentMethod      *string                   `json:"payment_method,omitempty"`
	ProviderStatus     *string                   `json:"provider_status,omitempty"`
	ProviderResponse   *FreedomPayStatusResponse `json:"provider_response,omitempty"`
	FinalBookingStatus *string                   `json:"final_booking_status,omitempty"`
	ProcessedAt        *time.Time                `json:"processed_at,omitempty"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
}

type PaymentLog struct {
	ID                 int64                     `json:"id"`
	PaymentID          *int64                    `json:"payment_id,omitempty"`
	BookingID          int64                     `json:"booking_id"`
	FPPaymentID        string                    `json:"fp_payment_id"`
	Action             PaymentLogAction          `json:"action"`
	OldStatus          *string                   `json:"old_status,omitempty"`
	NewStatus          *string                   `json:"new_status,omitempty"`
	FPResponse         *FreedomPayStatusResponse `json:"fp_response,omitempty"`
	ProcessingDuration *int                      `json:"processing_duration,omitempty"`
	UserID             *int64                    `json:"user_id,omitempty"`
	Source             PaymentLogSource          `json:"source"`
	Success            bool                      `json:"success"`
	ErrorMessage       *string                   `json:"error_message,omitempty"`
	IPAddress          *string                   `json:"ip_address,omitempty"`
	UserAgent          *string                   `json:"user_agent,omitempty"`
	CreatedAt          time.Time                 `json:"created_at"`
}

type FreedomPayStatusResponse struct {
	Status           string `xml:"pg_status" json:"pg_status"`
	PaymentID        string `xml:"pg_payment_id" json:"pg_payment_id"`
	CanReject        string `xml:"pg_can_reject" json:"pg_can_reject"`
	PaymentMethod    string `xml:"pg_payment_method" json:"pg_payment_method"`
	Amount           string `xml:"pg_amount" json:"pg_amount"`
	Currency         string `xml:"pg_currency" json:"pg_currency"`
	PaymentStatus    string `xml:"pg_payment_status" json:"pg_payment_status"`
	ClearingAmount   string `xml:"pg_clearing_amount" json:"pg_clearing_amount"`
	Reference        string `xml:"pg_reference" json:"pg_reference"`
	CardName         string `xml:"pg_card_name" json:"pg_card_name"`
	CardPan          string `xml:"pg_card_pan" json:"pg_card_pan"`
	CardToken        string `xml:"pg_card_token" json:"pg_card_token"`
	RefundAmount     string `xml:"pg_refund_amount" json:"pg_refund_amount"`
	Captured         string `xml:"pg_captured" json:"pg_captured"`
	CreateDate       string `xml:"pg_create_date" json:"pg_create_date"`
	Salt             string `xml:"pg_salt" json:"pg_salt"`
	Signature        string `xml:"pg_sig" json:"pg_sig"`
	OrderID          string `xml:"pg_order_id" json:"pg_order_id"`
	ErrorCode        string `xml:"pg_error_code" json:"pg_error_code"`
	ErrorDescription string `xml:"pg_error_description" json:"pg_error_description"`
}

type PaymentStatusResponse struct {
	Exists        bool   `json:"exists"`
	PaymentID     string `json:"payment_id"`
	Amount        string `json:"amount,omitempty"`
	Currency      string `json:"currency,omitempty"`
	Status        string `json:"status,omitempty"`
	PaymentMethod string `json:"payment_method,omitempty"`
	CreateDate    string `json:"create_date,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

type FreedomPayRefundResponse struct {
	Status           string `xml:"pg_status"`
	ErrorCode        string `xml:"pg_error_code"`
	ErrorDescription string `xml:"pg_error_description"`
	Salt             string `xml:"pg_salt"`
	Signature        string `xml:"pg_sig"`
}

type RefundRequest struct {
	PaymentID    string `json:"payment_id" binding:"required"`
	RefundAmount *int   `json:"refund_amount,omitempty"`
}

type RefundResponse struct {
	Success   bool   `json:"success"`
	PaymentID string `json:"payment_id"`
	Message   string `json:"message"`
}

type ReceiptAmounts struct {
	TotalPrice int    `json:"total_price"`
	ServiceFee int    `json:"service_fee"`
	FinalPrice int    `json:"final_price"`
	Currency   string `json:"currency"`
}

type ReceiptBookingDetails struct {
	ApartmentAddress string `json:"apartment_address"`
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	DurationHours    int    `json:"duration_hours"`
	RentalType       string `json:"rental_type"`
}

type PaymentReceipt struct {
	ReceiptID      string                `json:"receipt_id"`
	BookingNumber  string                `json:"booking_number"`
	PaymentID      string                `json:"payment_id"`
	OrderID        string                `json:"order_id"`
	Status         string                `json:"status"`
	PaymentDate    string                `json:"payment_date"`
	PaymentMethod  string                `json:"payment_method"`
	CardPan        string                `json:"card_pan"`
	Amounts        ReceiptAmounts        `json:"amounts"`
	BookingDetails ReceiptBookingDetails `json:"booking_details"`
	CreatedAt      string                `json:"created_at"`
}

type FreedomPayService interface {
	GetPaymentStatus(paymentID string) (*FreedomPayStatusResponse, error)
	GetPaymentStatusByOrderID(orderID string) (*FreedomPayStatusResponse, error)
	RefundPayment(paymentID string, refundAmount *int) (*FreedomPayRefundResponse, error)
}

type PaymentUseCase interface {
	CheckPaymentStatus(paymentID string) (*PaymentStatusResponse, error)
	CheckPaymentStatusByOrderID(orderID string, bookingID int64) (*PaymentStatusResponse, error)
	CheckPaymentStatusWithBooking(paymentID string, bookingID int64) (*PaymentStatusResponse, error)
	RefundPayment(paymentID string, refundAmount *int) (*RefundResponse, error)
}

type PaymentRepository interface {
	Create(payment *Payment) error
	GetByID(id int64) (*Payment, error)
	GetByPaymentID(paymentID string) (*Payment, error)
	GetByBookingID(bookingID int64) ([]*Payment, error)
	Update(payment *Payment) error
	GetAll(filters map[string]interface{}, page, pageSize int) ([]*Payment, int, error)
}

type PaymentLogRepository interface {
	Create(log *PaymentLog) error
	GetByPaymentID(paymentID int64) ([]*PaymentLog, error)
	GetByBookingID(bookingID int64) ([]*PaymentLog, error)
	GetByFPPaymentID(fpPaymentID string) ([]*PaymentLog, error)
	GetAll(filters map[string]interface{}, page, pageSize int) ([]*PaymentLog, int, error)
}
