-- Banners can have a picture behind their text.
--
-- image_url is a path on this site, such as /banners/keys.webp. NULL means
-- no picture: the banner is then drawn as its two-color gradient, as before.
ALTER TABLE banners ADD COLUMN image_url TEXT;

-- The three banners that exist get their pictures here, so a database that
-- was seeded before this migration does not need to be seeded again. They are
-- found by position, which is how the seed numbers them. A fresh database has
-- no banners yet at this point, and the seed then inserts them with pictures.
UPDATE banners SET image_url = '/banners/keys.webp?v=1' WHERE position = 1;
UPDATE banners SET image_url = '/banners/new.webp?v=1'  WHERE position = 2;
UPDATE banners SET image_url = '/banners/play.webp?v=1' WHERE position = 3;
