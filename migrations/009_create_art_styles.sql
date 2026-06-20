CREATE TABLE IF NOT EXISTS art_styles (
    id BIGSERIAL PRIMARY KEY,
    style VARCHAR(180) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL,
    blob_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS art_styles_style_idx
ON art_styles (style);
