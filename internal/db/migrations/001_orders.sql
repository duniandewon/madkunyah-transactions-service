-- +goose up
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    customer_name VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(50) NOT NULL,
    delivery_address TEXT NOT NULL,
    order_total INTEGER NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'IDR',
    payment_status VARCHAR(20) NOT NULL CHECK (payment_status IN ('pending', 'paid', 'failed', 'expired')),
    fulfillment_status VARCHAR(20) NOT NULL CHECK (fulfillment_status IN ('new', 'preparing', 'delivering', 'completed', 'canceled')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_payment_fulfillment 
        CHECK (
        (payment_status = 'pending' AND fulfillment_status = 'new') OR
        (payment_status = 'paid' AND fulfillment_status IN ('new','preparing','delivering','completed'))
        )
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_customer_name ON orders(customer_name);
CREATE INDEX idx_orders_customer_phone ON orders(customer_phone);

-- +goose down
DROP TABLE orders;
