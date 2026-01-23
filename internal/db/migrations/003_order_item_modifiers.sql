-- +goose up
CREATE TABLE order_item_modifiers (
    id SERIAL PRIMARY KEY,
    order_item_id INTEGER NOT NULL,
    modifier_group_name_snapshot VARCHAR(255) NOT NULL,
    modifier_item_name_snapshot VARCHAR(255) NOT NULL,
    modifier_price INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    CONSTRAINT fk_order_item
        FOREIGN KEY (order_item_id)
        REFERENCES order_items(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_order_item_modifiers_order_item_id ON order_item_modifiers(order_item_id);

-- +goose down
DROP TABLE order_item_modifiers;