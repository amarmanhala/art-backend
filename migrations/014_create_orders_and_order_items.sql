CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    payment_status VARCHAR(30) NOT NULL DEFAULT 'unpaid',
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    subtotal NUMERIC(10,2) NOT NULL DEFAULT 0,
    shipping_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    stripe_session_id TEXT UNIQUE,
    stripe_payment_intent_id TEXT UNIQUE,
    customer_email VARCHAR(150),
    customer_name VARCHAR(150),
    shipping_name VARCHAR(150),
    shipping_phone VARCHAR(50),
    shipping_line1 VARCHAR(200),
    shipping_line2 VARCHAR(200),
    shipping_city VARCHAR(100),
    shipping_state VARCHAR(100),
    shipping_postal_code VARCHAR(50),
    shipping_country VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT orders_status_check CHECK (status IN ('pending', 'paid', 'expired', 'checkout_failed', 'cancelled', 'fulfilled')),
    CONSTRAINT orders_payment_status_check CHECK (payment_status IN ('unpaid', 'paid', 'expired', 'failed', 'refunded', 'no_payment_required'))
);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    product_variant_id BIGINT NOT NULL,
    product_title VARCHAR(180) NOT NULL,
    product_slug VARCHAR(180) NOT NULL,
    variant_size VARCHAR(50) NOT NULL DEFAULT '',
    unit_price NUMERIC(10,2) NOT NULL,
    quantity INTEGER NOT NULL,
    subtotal NUMERIC(10,2) NOT NULL,
    image_url TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT order_items_quantity_check CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS order_items_order_id_idx
ON order_items (order_id);

CREATE INDEX IF NOT EXISTS orders_user_id_created_at_idx
ON orders (user_id, created_at DESC, id DESC);
