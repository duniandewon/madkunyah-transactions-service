package paymentgateway

import "context"

type PaymentGateway interface {
	CreatePaymentRequest(ctx context.Context, amount int, externalID string) (url string, gatewayID string, err error)
}
