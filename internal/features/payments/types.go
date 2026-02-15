package payments

import (
	"time"
)

type Payment struct {
	ID                   int       `json:"id"`
	OrderID              int       `json:"order_id"`
	ExternalID           string    `json:"external_id"`
	GatewayTransactionID string    `json:"gateway_transaction_id"`
	GatewayName          string    `json:"gateway_name"`
	Amount               int       `json:"amount"`
	PaymentMethod        string    `json:"payment_method"`
	PaymentChannel       string    `json:"payment_channel"`
	Status               string    `json:"status"`
	PaidAt               time.Time `json:"paid_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type CreatePaymentInput struct {
	OrderID     int    `json:"order_id"`
	ExternalID  string `json:"external_id"`
	GatewayName string `json:"gateway_name"`
	Amount      int    `json:"amount"`
}

type PaymentService interface {
	CreatePayment(input CreatePaymentInput) (*Payment, error)
}
