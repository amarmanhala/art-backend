CREATE TABLE IF NOT EXISTS product_sizes (
    id BIGSERIAL PRIMARY KEY,
    label VARCHAR(50) NOT NULL UNIQUE,
    width_in NUMERIC(10,2) NOT NULL,
    height_in NUMERIC(10,2) NOT NULL,
    width_cm NUMERIC(10,2) NOT NULL,
    height_cm NUMERIC(10,2) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO product_sizes (label, width_in, height_in, width_cm, height_cm, sort_order)
VALUES
    ('12 x 16 in', 12, 16, 31, 41, 1),
    ('16 x 20 in', 16, 20, 41, 51, 2),
    ('20 x 28 in', 20, 28, 50, 70, 3),
    ('24 x 36 in', 24, 36, 61, 91, 4)
ON CONFLICT (label) DO NOTHING;
