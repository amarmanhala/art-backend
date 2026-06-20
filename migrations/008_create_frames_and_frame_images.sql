CREATE TABLE IF NOT EXISTS frames (
    id BIGSERIAL PRIMARY KEY,
    vendor_name VARCHAR(120) NOT NULL,
    frame_name VARCHAR(180) NOT NULL,
    color VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    article_number VARCHAR(120) NOT NULL,
    product_detail TEXT NOT NULL DEFAULT '',
    material_description TEXT NOT NULL DEFAULT '',
    care TEXT NOT NULL DEFAULT '',
    price NUMERIC(10,2) NOT NULL,
    picture_width_cm NUMERIC(10,2) NOT NULL,
    picture_width_in NUMERIC(10,2) NOT NULL,
    picture_height_cm NUMERIC(10,2) NOT NULL,
    picture_height_in NUMERIC(10,2) NOT NULL,
    frame_width_cm NUMERIC(10,2) NOT NULL,
    frame_width_in NUMERIC(10,2) NOT NULL,
    frame_height_cm NUMERIC(10,2) NOT NULL,
    frame_height_in NUMERIC(10,2) NOT NULL,
    frame_depth_cm NUMERIC(10,2) NOT NULL,
    frame_depth_in NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT frames_vendor_name_article_number_key UNIQUE (vendor_name, article_number)
);

CREATE TABLE IF NOT EXISTS frame_images (
    id BIGSERIAL PRIMARY KEY,
    frame_id BIGINT NOT NULL REFERENCES frames(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL,
    original_url TEXT NOT NULL,
    blob_name TEXT NOT NULL,
    thumbnail_blob_name TEXT NOT NULL,
    alt_text TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS frame_images_frame_id_sort_order_idx
ON frame_images (frame_id, sort_order);
