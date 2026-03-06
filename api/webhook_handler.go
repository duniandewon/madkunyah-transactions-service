package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/duniandewon/madkunyah-transactions-service/internal/features/orders"
	"github.com/duniandewon/madkunyah-transactions-service/internal/features/payments"
	"github.com/duniandewon/madkunyah-transactions-service/internal/response"
)

type WebhookHandler struct {
	oerderService  orders.OrderRepository
	paymentService payments.PaymentService
	webhookSecret  string
}

func NewWebhookHandler(
	orderService orders.OrderRepository,
	paymentService payments.PaymentService,
	webhookSecret string,
) *WebhookHandler {
	return &WebhookHandler{
		oerderService:  orderService,
		paymentService: paymentService,
		webhookSecret:  webhookSecret,
	}
}

type XenditWebhookPayload struct {
	Event      string                 `json:"event"`
	BusinessID string                 `json:"business_id"`
	Created    string                 `json:"created"`
	Data       map[string]interface{} `json:"data"`
}

func (h *WebhookHandler) XenditPaymentWebhook(w http.ResponseWriter, r *http.Request) {
	callbackToken := r.Header.Get("X-CALLBACK-TOKEN")
	if callbackToken != h.webhookSecret {
		response.Unauthorized(w, "unauthorized")
		return
	}

	var payload XenditWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.BadRequest(w, "invalid payload")
		return
	}

	data := payload.Data

	gatewayTransactionID, ok := data["id"].(string)
	if !ok {
		fmt.Println("Error: id field missing")
		response.BadRequest(w, "missing id")
		return
	}
	orderId, ok := data["reference_id"].(string)
	if !ok {
		fmt.Println("Error: order_id field missing")
		response.BadRequest(w, "missing order_id")
		return
	}
	paymentRequestId, ok := data["payment_request_id"].(string)
	if !ok {
		fmt.Println("Error: payment_request_id field missing")
		response.BadRequest(w, "missing payment_request_id")
		return
	}
	status, ok := data["status"].(string)
	if !ok {
		fmt.Println("Warning: status field missing or invalid")
		status = "unknown"
	}
	channelCode, ok := data["channel_code"].(string)
	if !ok {
		fmt.Println("Warning: channel_code field missing or invalid")
		channelCode = "unknown"
	}

	orderIdInt, err := strconv.Atoi(orderId)
	if err != nil {
		fmt.Println("Error: order_id is not a valid integer")
		response.BadRequest(w, "invalid order_id")
		return
	}

	var internalStatus string
	switch status {
	case "SUCCEEDED", "COMPLETED":
		internalStatus = "paid"
	case "EXPIRED":
		internalStatus = "expired"
	case "FAILED":
		internalStatus = "failed"
	default:
		internalStatus = "pending"
	}

	fmt.Printf("Payment Status: %s,payment method: %s, payment request ID: %s, payment internal status: %v\n", status, channelCode, paymentRequestId, internalStatus)

	if err := h.paymentService.UpdatePaymentStatus(r.Context(), payments.UpdatePaymentStatusInput{
		OrderID:              orderIdInt,
		PaymentRequestID:     paymentRequestId,
		PaymentChannel:       channelCode,
		GatewayTransactionID: gatewayTransactionID,
		Status:               internalStatus,
	}); err != nil {
		fmt.Printf("Error updating payment status: %v\n", err)
		response.InternalServerError(w, "failed to update payment status: "+err.Error())
		return
	}

	response.Ok(w, "webhook processed successfully", nil)
}
