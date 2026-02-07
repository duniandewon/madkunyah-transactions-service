package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/duniandewon/madkunyah-transactions-service/internal/features/orders"
	mw "github.com/duniandewon/madkunyah-transactions-service/internal/middleware"
)

type OrderHandler struct {
	repo       orders.OrderRepository
	menuClient *orders.MenuClient
}

func NewOrderHandler(repo orders.OrderRepository, menuClient *orders.MenuClient) *OrderHandler {
	return &OrderHandler{
		repo:       repo,
		menuClient: menuClient,
	}
}

func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req orders.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	orderItems, err := h.buildOrderItems(r.Context(), req.Items)
	if err != nil {
		http.Error(w, "failed to build order items: "+err.Error(), http.StatusBadRequest)
		return
	}

	orderInput := orders.CreateOrderInput{
		CustomerName: req.Customer.Name,
		Phone:        req.Customer.Phone,
		Address:      req.Customer.Address,
		Items:        orderItems,
	}

	order, err := h.repo.Create(r.Context(), orderInput)
	if err != nil {
		http.Error(w, "failed to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) GetAllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetClaims(r.Context())
	if !ok || claims.Role != "admin" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "failed to get orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetOrdersByUserIdHandler(w http.ResponseWriter, r *http.Request) {
	userID := 1
	orders, err := h.repo.GetAllByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetUserOrderDetailsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetClaims(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid category ID", http.StatusBadRequest)
		return
	}

	orderItemsRows, err := h.repo.GetUserOrderDetails(r.Context(), claims.UserID, orderID)
	if err != nil {
		http.Error(w, "failed to get order details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orderItemsRows)
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
