CREATE TABLE IF NOT EXISTS liked_arts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT liked_arts_status_check CHECK (status IN ('liked', 'disliked')),
    CONSTRAINT liked_arts_user_product_unique UNIQUE (user_id, product_id)
);

CREATE INDEX IF NOT EXISTS liked_arts_user_status_updated_at_idx
ON liked_arts (user_id, status, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS liked_arts_product_id_idx
ON liked_arts (product_id);
