ALTER TABLE art_styles
ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}';

UPDATE art_styles
SET tags = '{}'
WHERE tags IS NULL;
