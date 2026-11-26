-- Search: matching product names despite typos.
--
-- pg_trgm is a standard PostgreSQL extension. It splits text into trigrams
-- (three-letter slices: "kart" gives " ka", "kar", "art", "rt ") and measures
-- how many trigrams two texts share. A misspelled word still shares most of
-- its trigrams with the right one, which is how "circut" finds "Circuit".
--
-- The search query calls its word_similarity() function. See
-- backend/catalog/search.go.
CREATE EXTENSION IF NOT EXISTS pg_trgm;
