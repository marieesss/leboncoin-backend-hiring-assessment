# Contributing

## Tests

Every handler and every usecase must have tests.

- **Handlers** (`ads/internal/cmd/ads-api/http`): test each response status the handler can return replacing the usecase with a stub function.
- **Usecases** (`ads/internal/cmd/ads-api/usecase`): test the nominal case and each error case, using the DAO mock (`dao/mock`) with the exact arguments expected, do not use `mock.Anything`.

Run all the tests with:

```shell
go test ./...
```

## Logs

- A serious error is any error returned as a `500 Internal Server Error`.
- Log it with `slog.Error`, with a short message, a identifier, and the original error:

  ```go
  slog.Error("unable to get ad", "ad_id", adID, "error", err)
  ```

- Do not log client errors (400, 401, 404)
- Never return the original error to the client,  it stays in the logs.