package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/duniandewon/madkunyah-transactions-service/internal/features/orders"
	"github.com/duniandewon/madkunyah-transactions-service/internal/features/payments"
	mw "github.com/duniandewon/madkunyah-transactions-service/internal/middleware"
	"github.com/duniandewon/madkunyah-transactions-service/internal/platform/paymentgateway"
	"github.com/duniandewon/madkunyah-transactions-service/internal/response"
)

type OrderHandler struct {
	repo           orders.OrderRepository
	paymentService payments.PaymentService
	menuClient     *orders.MenuClient
	paymentgateway paymentgateway.PaymentGateway
}

func NewOrderHandler(
	repo orders.OrderRepository,
	paymentService payments.PaymentService,
	menuClient *orders.MenuClient,
	paymentGateway paymentgateway.PaymentGateway,
) *OrderHandler {
	return &OrderHandler{
		repo:           repo,
		menuClient:     menuClient,
		paymentService: paymentService,
		paymentgateway: paymentGateway,
	}
}

type CreateOrderResponse struct {
	OrderID   int    `json:"order_id"`
	URL       string `json:"url"`
	Total     int    `json:"total"`
	GatewayID string `json:"gateway_id"`
}

func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req orders.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	orderItems, err := h.buildOrderItems(r.Context(), req.Items)
	if err != nil {
		response.InternalServerError(w, "failed to build order items: "+err.Error())
		return
	}

	order, err := h.repo.Create(r.Context(), orders.CreateOrderInput{
		CustomerName: req.Customer.Name,
		Phone:        req.Customer.Phone,
		Address:      req.Customer.Address,
		Items:        orderItems,
	})
	if err != nil {
		response.InternalServerError(w, "failed to create order: "+err.Error())
		return
	}

	url, gatewayID, err := h.paymentgateway.CreatePaymentRequest(r.Context(), order.Total, fmt.Sprint(order.ID))
	if err != nil {
		response.InternalServerError(w, "Failed to create payment request: "+err.Error())
		return
	}

	_, err = h.paymentService.CreatePayment(r.Context(), payments.CreatePaymentInput{
		OrderID:     order.ID,
		ExternalID:  gatewayID,
		GatewayName: "xendit",
		Amount:      order.Total,
	})
	if err != nil {
		response.InternalServerError(w, "failed to create payment record: "+err.Error())
		return
	}

	res := CreateOrderResponse{
		OrderID:   order.ID,
		URL:       url,
		Total:     order.Total,
		GatewayID: gatewayID,
	}

	response.Ok(w, "order created successfully", res)
}

func (h *OrderHandler) GetAllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetClaims(r.Context())
	if !ok || claims.Role != "admin" {
		response.Unauthorized(w, "unauthorized")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	orders, err := h.repo.GetAll(r.Context(), offset, limit)
	if err != nil {
		response.InternalServerError(w, "failed to get orders: "+err.Error())
		return
	}

	response.Ok(w, "orders retrieved successfully", orders)
}

func (h *OrderHandler) GetOrdersByUserIdHandler(w http.ResponseWriter, r *http.Request) {
	userID := 1
	orders, err := h.repo.GetAllByUserID(r.Context(), userID)
	if err != nil {
		response.InternalServerError(w, "failed to get orders: "+err.Error())
		return
	}

	response.Ok(w, "orders retrieved successfully", orders)
}

func (h *OrderHandler) GetUserOrderDetailsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetClaims(r.Context())
	if !ok {
		response.Unauthorized(w, "unauthorized")
		return
	}

	orderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "invalid category ID")
		return
	}

	orderItemsRows, err := h.repo.GetUserOrderDetails(r.Context(), claims.UserID, orderID)
	if err != nil {
		response.InternalServerError(w, "failed to get order details: "+err.Error())
		return
	}

	response.Ok(w, "order details retrieved successfully", orderItemsRows)
}

func (h *OrderHandler) buildOrderItems(ctx context.Context, items []orders.MenuItemRequest) ([]orders.CreateOrderItemInput, error) {
	resChan := make(chan orders.CreateOrderItemInput, len(items))
	errChan := make(chan error, len(items))
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)

		go func(item orders.MenuItemRequest) {
			defer wg.Done()

			menu, err := h.menuClient.FetchMenu(ctx, int(item.MenuID))
			if err != nil {
				errChan <- err
				return
			}

			var selectedModifiersItems []orders.CreateOrderItemModifierInput
			for _, modReq := range item.ModifiersItemsID {
				found := false
				for _, group := range menu.ModifierGroups {
					for _, modItem := range group.Items {
						if modItem.ID == modReq {
							selectedModifiersItems = append(selectedModifiersItems, orders.CreateOrderItemModifierInput{
								ModifierID:        modReq,
								ModifierItemName:  modItem.Name,
								ModifierGroupName: group.Name,
								Price:             int(modItem.Price),
							})
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					errChan <- fmt.Errorf("modifier item ID %d not found for menu %s", modReq, menu.Name)
					return
				}

			}

			resChan <- orders.CreateOrderItemInput{
				MenuID:    int(item.MenuID),
				MenuName:  menu.Name,
				Quantity:  int(item.Quantity),
				Price:     int(menu.Price),
				Modifiers: selectedModifiersItems,
			}
		}(item)
	}
	wg.Wait()
	close(resChan)
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan
	}

	var orderItems []orders.CreateOrderItemInput
	for item := range resChan {
		orderItems = append(orderItems, item)
	}

	return orderItems, nil
}
