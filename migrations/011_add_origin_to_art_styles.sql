ALTER TABLE art_styles
ADD COLUMN IF NOT EXISTS origin VARCHAR(120);

UPDATE art_styles
SET origin = COALESCE(NULLIF(BTRIM(origin), ''), 'japanese');

ALTER TABLE art_styles
ALTER COLUMN origin SET NOT NULL;

ALTER TABLE art_styles
ALTER COLUMN origin SET DEFAULT 'japanese';

ALTER TABLE art_styles
DROP CONSTRAINT IF EXISTS art_styles_style_key;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'art_styles_origin_style_key'
    ) THEN
        ALTER TABLE art_styles
        ADD CONSTRAINT art_styles_origin_style_key UNIQUE (origin, style);
    END IF;
END $$;
