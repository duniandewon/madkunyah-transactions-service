-- name: GetAllOrderItems :many
SELECT
    oi.id AS order_item_id,
    oi.menu_name_snapshot,
    oi.unit_price,
    oi.quantity AS item_quantity,
    oi.item_total,
    oim.modifier_group_name_snapshot,
    oim.modifier_item_name_snapshot,
    oim.modifier_price,
    oim.quantity AS modifier_quantity
FROM order_items oi
LEFT JOIN order_item_modifiers oim
    ON oi.id = oim.order_item_id
WHERE oi.order_id = sqlc.arg('order_id')
ORDER BY oi.id;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id,
    menu_id,
    menu_name_snapshot,
    unit_price,
    quantity,
    modifiers_total,
    item_total
) VALUES (
    sqlc.arg('order_id'),
    sqlc.arg('menu_id'),
    sqlc.arg('menu_name_snapshot'),
    sqlc.arg('unit_price'),
    sqlc.arg('quantity'),
    sqlc.arg('modifiers_total'),
    sqlc.arg('item_total')
) RETURNING *;