package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/duniandewon/madkunyah-transactions-service/internal/features/orders"
	"github.com/duniandewon/madkunyah-transactions-service/internal/features/payments"
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
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload XenditWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	data := payload.Data

	gatewayTransactionID, ok := data["id"].(string)
	if !ok {
		fmt.Println("Error: id field missing")
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	orderId, ok := data["reference_id"].(string)
	if !ok {
		fmt.Println("Error: order_id field missing")
		http.Error(w, "missing order_id", http.StatusBadRequest)
		return
	}
	paymentRequestId, ok := data["payment_request_id"].(string)
	if !ok {
		fmt.Println("Error: payment_request_id field missing")
		http.Error(w, "missing payment_request_id", http.StatusBadRequest)
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
		http.Error(w, "invalid order_id", http.StatusBadRequest)
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

		http.Error(w, "failed to update payment status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
