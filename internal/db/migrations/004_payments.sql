-- +goose up
CREATE TABLE IF NOT EXISTS payments (
  id SERIAL PRIMARY KEY,
  order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  external_id VARCHAR(255) NOT NULL UNIQUE,
  gateway_transaction_id VARCHAR(255),
  gateway_name VARCHAR(50) NOT NULL,
  amount INTEGER NOT NULL,
  payment_channel VARCHAR(50),
  status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (
    status IN (
      'pending',
      'paid',
      'failed',
      'expired',
      'settled'
    )
  ),
  paid_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_payments_external_id ON payments(external_id);
-- +goose down
DROP TABLE payments;