-- Indexes for the filters of GET /ads 

-- Api is mostly read, so indexes are useful and doesnt cost much on writes

-- Filter max_price_cents/min_price_cents
CREATE INDEX CONCURRENTLY IF NOT EXISTS ads_price_cents_idx ON ads (price_cents);

-- user_id
CREATE INDEX CONCURRENTLY IF NOT EXISTS ads_user_id_idx ON ads (user_id);

-- title filter (ILIKE '%...%') pg_trgm used to speed up LIKE searches
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX CONCURRENTLY IF NOT EXISTS ads_title_trgm_idx ON ads USING gin (title gin_trgm_ops);
