CREATE TABLE IF NOT EXISTS contact_requests (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(150) NOT NULL,
    order_number VARCHAR(120),
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE IF EXISTS contact_messages RENAME TO contact_requests;

ALTER TABLE IF EXISTS contact_requests
ADD COLUMN IF NOT EXISTS order_number VARCHAR(120);

UPDATE contact_requests
SET order_number = NULLIF(BTRIM(order_number), '');
