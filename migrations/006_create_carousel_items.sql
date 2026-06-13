CREATE TABLE IF NOT EXISTS carousel_items (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(150) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL,
    blob_name TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS carousel_items_active_sort_order_idx
ON carousel_items (is_active, sort_order, id);
