package orders

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/duniandewon/madkunyah-transactions-service/internal/db/sqlc"
)

type svc struct {
	*db.Queries
	connPool *sql.DB
}

func NewService(connPool *sql.DB) *svc {
	return &svc{
		Queries:  db.New(connPool),
		connPool: connPool,
	}
}

func calculateOrderTotal(items []CreateOrderItemInput) int {
	total := 0
	for _, item := range items {
		unitPrice := item.Price

		for _, mod := range item.Modifiers {
			unitPrice += mod.Price
		}

		total += unitPrice * item.Quantity
	}
	return total
}

func (s *svc) Create(ctx context.Context, params CreateOrderInput) (*Order, error) {
	total := calculateOrderTotal(params.Items)

	tx, err := s.connPool.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	var userIDParam sql.NullInt32
	if params.UserID != nil && *params.UserID > 0 {
		userIDParam = sql.NullInt32{Int32: int32(*params.UserID), Valid: true}
	}

	dbOrder, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
		UserID:            userIDParam,
		CustomerName:      params.CustomerName,
		CustomerPhone:     params.Phone,
		DeliveryAddress:   params.Address,
		OrderTotal:        int32(total),
		PaymentStatus:     "pending",
		FulfillmentStatus: "new",
	})
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	for _, item := range params.Items {
		dbOrderItem, err := qtx.CreateOrderItem(ctx, db.CreateOrderItemParams{
			OrderID:          dbOrder.ID,
			MenuID:           int32(item.MenuID),
			MenuNameSnapshot: item.MenuName,
			UnitPrice:        int32(item.Price),
			Quantity:         int32(item.Quantity),
			ItemTotal:        int32(item.Price * item.Quantity),
		})
		if err != nil {
			return nil, fmt.Errorf("create order item: %w", err)
		}

		for _, mod := range item.Modifiers {
			_, err := qtx.CreateOrderItemModifier(ctx, db.CreateOrderItemModifierParams{
				OrderItemID:               dbOrderItem.ID,
				ModifierGroupNameSnapshot: mod.ModifierGroupName,
				ModifierItemNameSnapshot:  mod.ModifierItemName,
				ModifierPrice:             int32(mod.Price),
				Quantity:                  1,
			})
			if err != nil {
				return nil, fmt.Errorf("create order item modifier: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &Order{
		ID:                int(dbOrder.ID),
		UserID:            nil,
		CustomerName:      dbOrder.CustomerName,
		Phone:             dbOrder.CustomerPhone,
		Address:           dbOrder.DeliveryAddress,
		Total:             int(dbOrder.OrderTotal),
		PaymentStatus:     dbOrder.PaymentStatus,
		FulfillmentStatus: dbOrder.FulfillmentStatus,
		CreatedAt:         dbOrder.CreatedAt,
		UpdatedAt:         dbOrder.UpdatedAt,
	}, nil
}

func (s *svc) GetAll(ctx context.Context, offset, limit int) ([]*Order, error) {
	dbOrders, err := s.Queries.GetAllOrders(ctx, db.GetAllOrdersParams{
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}

	orders := make([]*Order, 0, len(dbOrders))
	for _, dbOrder := range dbOrders {
		var userIDPtr *int
		if dbOrder.UserID.Valid {
			uid := int(dbOrder.UserID.Int32)
			userIDPtr = &uid
		}

		order := &Order{
			ID:                int(dbOrder.ID),
			UserID:            userIDPtr,
			CustomerName:      dbOrder.CustomerName,
			Phone:             dbOrder.CustomerPhone,
			Address:           dbOrder.DeliveryAddress,
			Total:             int(dbOrder.OrderTotal),
			PaymentStatus:     dbOrder.PaymentStatus,
			FulfillmentStatus: dbOrder.FulfillmentStatus,
			CreatedAt:         dbOrder.CreatedAt,
			UpdatedAt:         dbOrder.UpdatedAt,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *svc) GetAllByUserID(ctx context.Context, userID int) ([]*Order, error) {
	dbOrders, err := s.Queries.GetOrdersByUserId(ctx, sql.NullInt32{Int32: int32(userID), Valid: true})
	if err != nil {
		return nil, fmt.Errorf("get orders by user id: %w", err)
	}

	orders := make([]*Order, 0, len(dbOrders))
	for _, dbOrder := range dbOrders {
		var userIDPtr *int
		if dbOrder.UserID.Valid {
			uid := int(dbOrder.UserID.Int32)
			userIDPtr = &uid
		}

		order := &Order{
			ID:                int(dbOrder.ID),
			UserID:            userIDPtr,
			CustomerName:      dbOrder.CustomerName,
			Phone:             dbOrder.CustomerPhone,
			Address:           dbOrder.DeliveryAddress,
			Total:             int(dbOrder.OrderTotal),
			PaymentStatus:     dbOrder.PaymentStatus,
			FulfillmentStatus: dbOrder.FulfillmentStatus,
			CreatedAt:         dbOrder.CreatedAt,
			UpdatedAt:         dbOrder.UpdatedAt,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *svc) GetUserOrderDetails(ctx context.Context, userID, orderID int) (*OrderDetail, error) {
	dbOrderItems, err := s.Queries.GetAllOrderItems(ctx, int32(orderID))
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}

	orderItems := TransformOrderRows(dbOrderItems)

	orderDetail := OrderDetail{
		Items: orderItems,
	}

	return &orderDetail, nil
}

func TransformOrderRows(rows []db.GetAllOrderItemsRow) []OrderItem {
	itemMap := make(map[int32]*OrderItem)
	var order []int32

	for _, row := range rows {
		if _, exists := itemMap[row.OrderItemID]; !exists {
			itemMap[row.OrderItemID] = &OrderItem{
				ID:        int(row.OrderItemID),
				MenuName:  row.MenuNameSnapshot,
				UnitPrice: int(row.UnitPrice),
				Quantity:  int(row.ItemQuantity),
				ItemTotal: int(row.ItemTotal),
				Modifiers: []OrderItemModifier{},
			}
			order = append(order, row.OrderItemID)
		}

		if row.ModifierItemNameSnapshot.Valid {
			modifier := OrderItemModifier{
				ID:                int(row.ModifierID.Int32),
				ModifierName:      row.ModifierItemNameSnapshot.String,
				ModifierGroupName: row.ModifierGroupNameSnapshot.String,
				ModifierPrice:     int(row.ModifierPrice.Int32),
				Quantity:          int(row.ModifierQuantity.Int32),
			}
			itemMap[row.OrderItemID].Modifiers = append(itemMap[row.OrderItemID].Modifiers, modifier)
		}
	}

	result := make([]OrderItem, 0, len(order))
	for _, id := range order {
		result = append(result, *itemMap[id])
	}

	return result
}
