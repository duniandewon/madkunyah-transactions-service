-- name: CreateOrderItemModifier :one
INSERT INTO order_item_modifiers (
    order_item_id,
    modifier_group_name_snapshot,
    modifier_item_name_snapshot,
    modifier_price,
    quantity
) VALUES (
    sqlc.arg('order_item_id'),
    sqlc.arg('modifier_group_name_snapshot'),
    sqlc.arg('modifier_item_name_snapshot'),
    sqlc.arg('modifier_price'),
    sqlc.arg('quantity')
) RETURNING *;