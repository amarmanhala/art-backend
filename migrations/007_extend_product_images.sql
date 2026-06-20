ALTER TABLE product_images
ADD COLUMN IF NOT EXISTS thumbnail_url TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS original_url TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS blob_name TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS thumbnail_blob_name TEXT NOT NULL DEFAULT '';

UPDATE product_images
SET original_url = image_url
WHERE original_url = '';

UPDATE product_images
SET thumbnail_url = image_url
WHERE thumbnail_url = '';
