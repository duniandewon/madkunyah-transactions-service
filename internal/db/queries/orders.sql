-- name: GetAllOrders :many
SELECT *
FROM orders
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
-- name: GetOrdersByUserId :many
SELECT *
FROM orders
WHERE user_id = sqlc.arg('user_id')
ORDER BY created_at DESC;
-- name: GetOrderById :one
SELECT *
FROM orders
WHERE id = sqlc.arg('id');
-- name: CreateOrder :one
INSERT INTO orders (
    user_id,
    customer_name,
    customer_phone,
    delivery_address,
    order_total,
    payment_status,
    fulfillment_status
  )
VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('customer_name'),
    sqlc.arg('customer_phone'),
    sqlc.arg('delivery_address'),
    sqlc.arg('order_total'),
    sqlc.arg('payment_status'),
    sqlc.arg('fulfillment_status')
  )
RETURNING *;
-- name: UpdateOrderTotal :exec
UPDATE orders
SET order_total = sqlc.arg('order_total'),
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');
-- name: CancelOrder :exec
UPDATE orders
SET fulfillment_status = 'canceled',
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
  AND payment_status = 'pending'
  AND fulfillment_status = 'new';
-- name: MarkOrderPaid :exec
UPDATE orders
SET payment_status = 'paid',
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
  AND payment_status = 'pending';
-- name: MarkOrderPaymentFailed :exec
UPDATE orders
SET payment_status = 'failed',
  fulfillment_status = 'canceled',
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
  AND payment_status = 'pending'
  AND fulfillment_status = 'new';
-- name: MarkOrderPaymentExpired :exec
UPDATE orders
SET payment_status = 'expired',
  fulfillment_status = 'canceled',
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
  AND payment_status = 'pending'
  AND fulfillment_status = 'new';
-- name: StartPreparingOrder :exec
UPDATE orders
SET fulfillment_status = 'preparing',
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
  AND payment_status = 'paid'
  AND fulfillment_status = 'new';
-- name: MarkOrderDelivering :exec
UPDATE orders
SET fulfillment_status = 'delivering',
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
  AND payment_status = 'paid'
  AND fulfillment_status = 'preparing';
-- name: CompleteOrder :exec
UPDATE orders
SET fulfillment_status = 'completed',
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
  AND payment_status = 'paid'
  AND fulfillment_status = 'delivering';