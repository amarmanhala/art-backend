CREATE TABLE IF NOT EXISTS product_variants (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    size VARCHAR(50) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT product_variants_product_size_key UNIQUE (product_id, size)
);

CREATE INDEX IF NOT EXISTS product_variants_product_id_default_idx
ON product_variants (product_id, is_default DESC, id);

INSERT INTO product_variants (product_id, size, price, stock_quantity, is_default)
SELECT
    p.id,
    p.size,
    p.price,
    p.stock_quantity,
    TRUE
FROM products p
WHERE NOT EXISTS (
    SELECT 1
    FROM product_variants pv
    WHERE pv.product_id = p.id
);

ALTER TABLE cart_items
ADD COLUMN IF NOT EXISTS product_variant_id BIGINT;

UPDATE cart_items ci
SET product_variant_id = pv.id
FROM products p
JOIN product_variants pv ON pv.product_id = p.id
WHERE ci.product_id = p.id
  AND ci.product_variant_id IS NULL
  AND pv.size = p.size;

UPDATE cart_items ci
SET product_variant_id = pv.id
FROM product_variants pv
WHERE ci.product_variant_id IS NULL
  AND ci.product_id = pv.product_id
  AND pv.is_default = TRUE;

ALTER TABLE cart_items
ALTER COLUMN product_variant_id SET NOT NULL;

ALTER TABLE cart_items
DROP CONSTRAINT IF EXISTS cart_items_cart_product_unique;

ALTER TABLE cart_items
DROP CONSTRAINT IF EXISTS cart_items_cart_product_variant_unique;

ALTER TABLE cart_items
ADD CONSTRAINT cart_items_cart_product_variant_unique UNIQUE (cart_id, product_variant_id);

ALTER TABLE cart_items
DROP CONSTRAINT IF EXISTS cart_items_product_variant_id_fkey;

ALTER TABLE cart_items
ADD CONSTRAINT cart_items_product_variant_id_fkey
FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE CASCADE;
