-- name: CreatePayment :one
INSERT INTO payments (
    order_id,
    external_id,
    gateway_name,
    payment_channel,
    amount,
    status
  )
VALUES (
    sqlc.arg(order_id),
    sqlc.arg(external_id),
    sqlc.arg(gateway_name),
    sqlc.arg(payment_channel),
    sqlc.arg(amount),
    'pending'
  )
RETURNING *;
-- name: GetAllPayments :many
SELECT *
FROM payments
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
-- name: GetPaymentByExternalID :many
SELECT *
FROM payments
WHERE external_id = sqlc.arg('external_id')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
-- name: GetPaymentsByOrderID :many
SELECT *
FROM payments
WHERE order_id = sqlc.arg('order_id')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
-- name: MarkPaymentPaid :exec
UPDATE payments
SET status = 'paid',
  payment_channel = sqlc.arg('payment_channel'),
  gateway_transaction_id = sqlc.arg('gateway_transaction_id'),
  paid_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE external_id = sqlc.arg('external_id')
  AND status = 'pending';
-- name: MarkPaymentFailed :exec
UPDATE payments
SET status = 'failed',
  updated_at = CURRENT_TIMESTAMP
WHERE external_id = sqlc.arg('external_id')
  AND status = 'pending';
-- name: MarkPaymentExpired :exec
UPDATE payments
SET status = 'expired',
  updated_at = CURRENT_TIMESTAMP
WHERE external_id = sqlc.arg('external_id')
  AND status = 'pending';
-- name: MarkPaymentSettled :exec
UPDATE payments
SET status = 'settled',
  updated_at = CURRENT_TIMESTAMP
WHERE external_id = sqlc.arg('external_id')
  AND status = 'paid';