package paymentgateway

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/xendit/xendit-go/v7"
	"github.com/xendit/xendit-go/v7/payment_request"
)

type XenditGateway struct {
	client *xendit.APIClient
}

func NewXenditGateway(secretKey string) *XenditGateway {
	return &XenditGateway{
		client: xendit.NewClient(secretKey),
	}
}

func (x *XenditGateway) CreatePaymentRequest(ctx context.Context, amount int, externalID string) (string, string, error) {
	req := *payment_request.NewPaymentRequestParameters(payment_request.PAYMENTREQUESTCURRENCY_IDR)
	req.SetAmount(float64(amount))
	req.SetReferenceId(externalID)

	paymentMethod := payment_request.NewPaymentMethodParameters(
		payment_request.PAYMENTMETHODTYPE_QR_CODE,
		payment_request.PAYMENTMETHODREUSABILITY_ONE_TIME_USE,
	)

	qrParams := payment_request.NewQRCodeParameters()
	qrParams.SetChannelCode(payment_request.QRCODECHANNELCODE_QRIS)

	paymentMethod.SetQrCode(*qrParams)
	req.SetPaymentMethod(*paymentMethod)

	resp, r, err := x.client.PaymentRequestApi.CreatePaymentRequest(ctx).
		IdempotencyKey(externalID).
		PaymentRequestParameters(req).
		Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PaymentRequestApi.CreatePaymentRequest``: %v\n", err.Error())

		b, _ := json.Marshal(err.FullError())
		fmt.Fprintf(os.Stderr, "Full Error Struct: %v\n", string(b))

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	qrString := *resp.PaymentMethod.QrCode.Get().ChannelProperties.QrString

	return qrString, resp.GetId(), nil
}
