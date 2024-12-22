ALTER TABLE links ADD COLUMN IF NOT EXISTS cdn_image_width integer DEFAULT 0;
ALTER TABLE links ADD COLUMN IF NOT EXISTS cdn_image_height integer DEFAULT 0;

---- create above / drop below ----

ALTER TABLE links DROP COLUMN cdn_image_width;
ALTER TABLE links DROP COLUMN cdn_image_height;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
