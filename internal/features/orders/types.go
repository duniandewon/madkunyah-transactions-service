package orders

import (
	"context"
	"errors"
	"time"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrUnauthorizedAccess = errors.New("unauthorized access to order")
	ErrInvalidOrderStatus = errors.New("invalid order status for this operation")
	ErrEmptyOrderItems    = errors.New("order must have at least one item")
)

type Order struct {
	ID                int       `json:"id"`
	UserID            *int      `json:"user_id,omitempty"`
	CustomerName      string    `json:"customer_name"`
	Phone             string    `json:"phone"`
	Address           string    `json:"address"`
	Total             int       `json:"total"`
	PaymentStatus     string    `json:"payment_status"`
	FulfillmentStatus string    `json:"fulfillment_status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID        int                 `json:"id"`
	MenuName  string              `json:"menu_name"`
	UnitPrice int                 `json:"unit_price"`
	Quantity  int                 `json:"quantity"`
	ItemTotal int                 `json:"item_total"`
	Modifiers []OrderItemModifier `json:"modifiers"`
}

type OrderItemModifier struct {
	ID                int    `json:"id"`
	ModifierName      string `json:"modifier_name"`
	ModifierGroupName string `json:"modifier_group"`
	ModifierPrice     int    `json:"modifier_price"`
	Quantity          int    `json:"quantity"`
}

type OrderDetail struct {
	Items []OrderItem `json:"items"`
}

type CreateOrderInput struct {
	UserID       *int                   `json:"user_id,omitempty"`
	CustomerName string                 `json:"customer_name"`
	Phone        string                 `json:"phone"`
	Address      string                 `json:"address"`
	Items        []CreateOrderItemInput `json:"items" validate:"required,min=1"`
}

type CreateOrderItemInput struct {
	MenuID    int                            `json:"menu_id"`
	MenuName  string                         `json:"menu_name"`
	Quantity  int                            `json:"quantity"`
	Price     int                            `json:"price"`
	Modifiers []CreateOrderItemModifierInput `json:"modifiers"`
}

type CreateOrderItemModifierInput struct {
	ModifierID        int    `json:"modifier_id"`
	ModifierItemName  string `json:"modifier_item_name"`
	ModifierGroupName string `json:"modifier_group_name"`
	Price             int    `json:"price"`
}

type OrderRequest struct {
	Customer CustomerRequest   `json:"customer"`
	Items    []MenuItemRequest `json:"items"`
}

type CustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type MenuItemRequest struct {
	MenuID           int64 `json:"menu_id"`
	Quantity         int64 `json:"quantity"`
	ModifiersItemsID []int `json:"modifiers-items-id"`
}

type OrderRepository interface {
	// Order operations
	Create(ctx context.Context, params CreateOrderInput) (*Order, error)
	GetAll(ctx context.Context, offset, limit int) ([]*Order, error)
	GetAllByUserID(ctx context.Context, userID int) ([]*Order, error)
	GetUserOrderDetails(ctx context.Context, userID, orderID int) (*OrderDetail, error)

	// Kitchen workflow queries
	// GetOrdersReadyToPrepare(ctx context.Context) ([]*Order, error)
	// GetOrdersInPreparation(ctx context.Context) ([]*Order, error)
	// GetOrdersForDelivery(ctx context.Context) ([]*Order, error)

	// Status transition
	// MarkOrderPreparing(ctx context.Context, orderId int) error
	// MarkOrderDelivering(ctx context.Context, orderId int) error
	// MarkOrderCompleted(ctx context.Context, orderId int) error
}
