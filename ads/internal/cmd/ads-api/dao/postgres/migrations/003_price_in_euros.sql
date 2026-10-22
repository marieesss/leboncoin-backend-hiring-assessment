-- Store the price in euros as an exact decimal instead of cents.
ALTER TABLE ads RENAME COLUMN price_cents TO price;
ALTER TABLE ads ALTER COLUMN price TYPE NUMERIC(12, 2) USING price / 100.0;
