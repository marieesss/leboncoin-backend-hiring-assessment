# Backend hiring assessment summary

## Instructions

### 1. Bug in the GetAdByID usecase
The usecase always called the DAO with id 0 instead of the requested id, so no ad was ever found.
Fix: the id is passed to the DAO, and the commented-out legacy code was removed.

### 2. Delete ad by id
DELETE /ads/{id} with its handler, usecase and DAO method, all tested.
Only the owner can delete an ad: if not 403, 404 when the ad does not exist.

### 3. username field in GetAdByID
The username comes from the user service through the provided client
The username is secondary: when the user service fails, the ad is still returned without username, and the error is logged as a warning.

### 4. Rename price_cents to price as a float
No breaking change: price (in euros) is added next to price_cents, which is marked deprecated until web and mobile clients migrate.
The float only exists at the API: the database and the domain stay in integer cents, to avoid a large table migration.
Sending both price and price_cents, or none, returns 400. The min_price / max_price filters are also available in euros.

### 5. Bulk import failing and creating duplicates
Cause: each ad was saved on its own, so a failure in the middle left a partial import, and a retry created the ads again.
Fix: a single transaction, and every price is validated before anything is written.
Safe retries with an Idempotency-Key header: a retry with the same key returns the ids already created, and the same key with other ads returns 422.

### 6. Quota of 5 ads per rolling 24 hours on POST /ads
Not done.

## What I added

- GitHub Actions CI : build, tests and golangci-lint on every PR, with the local lint command documented in the README.
- Swagger UI in development only: APP_ENV defaults to production, and the OpenAPI spec is embedded in the binary.
- Error handling and logs: shared JSON error responses, slog.Error only on 500 with an identifier, and the original error is never returned to the client. Rules written in CONTRIBUTING.md.
- Indexes on the GET /ads filters: the API is read-heavy, so the price, owner and title filters are indexed.
- user_id validation in GET /ads: a malformed UUID returns 400 instead of an empty list.
- Configuration: .env.example and environment variables in docker-compose.yml.
- Running without the user service: when IDENTITY_URL is empty, GET /ads/{id} returns the ad without username instead of crashing.
- OpenAPI and README up to date for every endpoint, including POST /ads/import, which was not documented

## Nice to have

- Validate the length of the Idempotency-Key sent to POST /ads/import, so an oversized key returns 400 instead of 500
- Replace the plain user UUID in the Authorization header with a signed token (for example a JWT) carrying the user UUID in its payload
- Require at least one reviewer on every PR, as in production, for example (for this exercise) by adding Copilot code review to the pipeline
- Limit the size of a bulk import (number of ads and body size), to keep its transaction short
- Paginate GET /ads: the query has no LIMIT and returns the whole table, while reads are the heavy load
- Validate title and photo_url: an empty title or an invalid URL is accepted today, both in POST /ads and in the bulk import
- Purge the ad_imports table: idempotency keys are never deleted, a job could deleter the older ones every 24h
- Add more logs to resolve problem (500) : an ID on each request to trace issue, and an Info log when a bulk import is done or retried