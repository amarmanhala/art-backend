ALTER TABLE products
ADD COLUMN IF NOT EXISTS original_url TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS product_images (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    alt_text TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS product_images_product_id_sort_order_idx
ON product_images (product_id, sort_order);

INSERT INTO products (
    title,
    slug,
    description,
    price,
    currency,
    category,
    style,
    theme,
    orientation,
    size,
    image_url,
    thumbnail_url,
    original_url,
    stock_quantity,
    is_active
) VALUES
(
    'Azure Abstract Study 1',
    'azure-abstract-study-1',
    'Original art print with a gallery-ready presentation.',
    129.99,
    'USD',
    'Print',
    'Abstract',
    'Modern',
    'Portrait',
    '18x24 in',
    'https://artbackendmedia.blob.core.windows.net/product-thumbnails/ChatGPT%20Image%20Jun%2010,%202026,%2011_33_17%20PM.png',
    'https://artbackendmedia.blob.core.windows.net/product-thumbnails/ChatGPT%20Image%20Jun%2010,%202026,%2011_33_17%20PM.png',
    '',
    10,
    TRUE
),
(
    'Azure Abstract Study 2',
    'azure-abstract-study-2',
    'Contemporary wall art for home, studio, or office spaces.',
    139.99,
    'USD',
    'Print',
    'Abstract',
    'Contemporary',
    'Portrait',
    '18x24 in',
    'https://artbackendmedia.blob.core.windows.net/product-thumbnails/ChatGPT%20Image%20Jun%2010,%202026,%2011_54_50%20PM.png',
    'https://artbackendmedia.blob.core.windows.net/product-thumbnails/ChatGPT%20Image%20Jun%2010,%202026,%2011_54_50%20PM.png',
    '',
    10,
    TRUE
),
(
    'Azure Abstract Study 3',
    'azure-abstract-study-3',
    'A bold art piece designed for a clean modern catalog.',
    149.99,
    'USD',
    'Print',
    'Abstract',
    'Modern',
    'Landscape',
    '24x36 in',
    'https://artbackendmedia.blob.core.windows.net/product-thumbnails/ChatGPT%20Image%20Jun%2010,%202026,%2011_54_50%20PM.png',
    'https://artbackendmedia.blob.core.windows.net/product-thumbnails/ChatGPT%20Image%20Jun%2010,%202026,%2011_54_50%20PM.png',
    '',
    10,
    TRUE
),
(
    'Azure Abstract Study 4',
    'azure-abstract-study-4',
    'Expressive product artwork prepared for storefront display.',
    159.99,
    'USD',
    'Print',
    'Abstract',
    'Gallery',
    'Portrait',
    '18x24 in',
    'https://artbackendmedia.blob.core.windows.net/product-thumbnails/ChatGPT%20Image%20Jun%2011,%202026,%2012_05_26%20AM.png',
    'https://artbackendmedia.blob.core.windows.net/product-thumbnails/ChatGPT%20Image%20Jun%2011,%202026,%2012_05_26%20AM.png',
    '',
    10,
    TRUE
)
ON CONFLICT (slug) DO NOTHING;
