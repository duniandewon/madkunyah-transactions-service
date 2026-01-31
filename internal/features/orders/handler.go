package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	mw "github.com/duniandewon/madkunyah-transactions-service/internal/middleware"
)

type handler struct {
	repo       OrderRepository
	menuClient *MenuClient
}

func NewHandler(repo OrderRepository, menuClient *MenuClient) *handler {
	return &handler{
		repo:       repo,
		menuClient: menuClient,
	}
}

func (h *handler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	orderItems, err := h.buildOrderItems(r.Context(), req.Items)
	if err != nil {
		http.Error(w, "failed to build order items: "+err.Error(), http.StatusBadRequest)
		return
	}

	orderInput := CreateOrderInput{
		CustomerName: req.Customer.Name,
		Phone:        req.Customer.Phone,
		Address:      req.Customer.Address,
		Currency:     "IDR",
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

func (h *handler) GetAllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetClaims(r.Context())
	if !ok || claims.Role != "admin" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "failed to get orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *handler) GetOrdersByUserIdHandler(w http.ResponseWriter, r *http.Request) {
	userID := 1
	orders, err := h.repo.GetAllByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *handler) GetUserOrderDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) buildOrderItems(ctx context.Context, items []MenuItemRequest) ([]CreateOrderItemInput, error) {
	resChan := make(chan CreateOrderItemInput, len(items))
	errChan := make(chan error, len(items))
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)

		go func(item MenuItemRequest) {
			defer wg.Done()

			menu, err := h.menuClient.FetchMenu(ctx, int(item.MenuID))
			if err != nil {
				errChan <- err
				return
			}

			var selectedModifiersItems []CreateOrderItemModifierInput
			for _, modReq := range item.ModifiersItemsID {
				found := false
				for _, group := range menu.ModifierGroups {
					for _, modItem := range group.Items {
						if modItem.ID == modReq {
							selectedModifiersItems = append(selectedModifiersItems, CreateOrderItemModifierInput{
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

			resChan <- CreateOrderItemInput{
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

	var orderItems []CreateOrderItemInput
	for item := range resChan {
		orderItems = append(orderItems, item)
	}

	return orderItems, nil
}
