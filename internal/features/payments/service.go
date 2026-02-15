package payments

import (
	"database/sql"

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
func (s *svc) CreatePayment(input CreatePaymentInput) (*Payment, error) {
	// code...
	return nil, nil
}
