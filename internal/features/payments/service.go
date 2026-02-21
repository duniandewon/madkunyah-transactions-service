package payments

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

func (s *svc) CreatePayment(ctx context.Context, input CreatePaymentInput) (*Payment, error) {
	payment, err := s.Queries.CreatePayment(ctx, db.CreatePaymentParams{
		OrderID:     int32(input.OrderID),
		ExternalID:  input.ExternalID,
		GatewayName: input.GatewayName,
		Amount:      int32(input.Amount),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	return &Payment{
		ID:                   int(payment.ID),
		OrderID:              int(payment.OrderID),
		ExternalID:           payment.ExternalID,
		GatewayTransactionID: payment.GatewayTransactionID.String,
		GatewayName:          payment.GatewayName,
		Amount:               int(payment.Amount),
		PaymentChannel:       payment.PaymentChannel.String,
		Status:               payment.Status,
		PaidAt:               payment.PaidAt.Time,
		CreatedAt:            payment.CreatedAt,
		UpdatedAt:            payment.UpdatedAt,
	}, nil
}

func (s *svc) UpdatePaymentStatus(ctx context.Context, input UpdatePaymentStatusInput) error {
	tx, err := s.connPool.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	switch input.Status {
	case "paid":
		if err := qtx.MarkOrderPaid(ctx, int32(input.OrderID)); err != nil {
			return fmt.Errorf("mark order paid failed: %w", err)
		}

		if err := qtx.MarkPaymentPaid(ctx, db.MarkPaymentPaidParams{
			PaymentChannel: sql.NullString{
				String: input.PaymentChannel,
				Valid:  true,
			},
			GatewayTransactionID: sql.NullString{
				String: input.GatewayTransactionID,
				Valid:  true,
			},
			ExternalID: input.PaymentRequestID,
		}); err != nil {
			return fmt.Errorf("mark payment paid failed: %w", err)
		}

	case "failed":
		if err := qtx.MarkOrderPaymentFailed(ctx, int32(input.OrderID)); err != nil {
			return fmt.Errorf("mark order payment failed failed: %w", err)
		}

		if err := qtx.MarkOrderPaymentFailed(ctx, int32(input.OrderID)); err != nil {
			return fmt.Errorf("mark order payment failed failed: %w", err)
		}

	case "expired":
		if err := qtx.MarkOrderPaymentExpired(ctx, int32(input.OrderID)); err != nil {
			return fmt.Errorf("mark order payment expired failed: %w", err)
		}

		if err := qtx.MarkOrderPaymentExpired(ctx, int32(input.OrderID)); err != nil {
			return fmt.Errorf("mark order payment expired failed: %w", err)
		}

	case "settled":
		if err := qtx.MarkPaymentSettled(ctx, input.PaymentRequestID); err != nil {
			return fmt.Errorf("mark payment settled failed: %w", err)
		}

	default:
		return fmt.Errorf("unsupported payment status: %s", input.Status)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
