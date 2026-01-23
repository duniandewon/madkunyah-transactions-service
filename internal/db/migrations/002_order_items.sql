-- +goose up
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL,
    menu_id INTEGER NOT NULL,
    menu_name_snapshot VARCHAR(255) NOT NULL,
    unit_price INTEGER NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    modifiers_total INTEGER NOT NULL DEFAULT 0,
    item_total INTEGER NOT NULL CHECK (item_total = (unit_price * quantity) + modifiers_total),
    CONSTRAINT fk_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- +goose down
DROP TABLE order_items;