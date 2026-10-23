package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/httpresponse"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/middleware"
)

// idempotencyKeyHeader carries the key a client sends again when it retries an import, so the ads are not created twice.
const idempotencyKeyHeader = "Idempotency-Key"

type importAdsBody struct {
	Ads []createAdBody `json:"ads"`
}

type importAdsResponse struct {
	IDs []int64 `json:"ids"`
}

// Import ads handler serves POST /ads/import.
func handlerImportAds(importAds usecase.ImportAdsFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		idempotencyKey := strings.TrimSpace(r.Header.Get(idempotencyKeyHeader))
		var body importAdsBody
		if err := decodeBody(r, &body); err != nil {
			httpresponse.WriteError(rw, http.StatusBadRequest, err.Error())
			return
		}

		ownerID := domain.UserID(middleware.UserIDFromRequest(r))
		ads := make([]usecase.ImportedAd, 0, len(body.Ads))

		for i, item := range body.Ads {
			priceCents, err := item.priceInCents()
			if err != nil {
				httpresponse.WriteError(rw, http.StatusBadRequest, fmt.Sprintf("ad %d: %s", i, err))
				return
			}

			ads = append(ads, usecase.ImportedAd{
				Title:      item.Title,
				PriceCents: priceCents,
				PhotoURL:   item.PhotoURL,
			})
		}

		adIDs, err := importAds(r.Context(), usecase.ImportAdsRequest{
			OwnerID:        ownerID,
			IdempotencyKey: idempotencyKey,
			Ads:            ads,
		})
		if errors.Is(err, usecase.ErrIdempotencyKeyReused) {
			httpresponse.WriteError(rw, http.StatusUnprocessableEntity, fmt.Sprintf("%s already used with other ads", idempotencyKeyHeader))
			return
		}

		if err != nil {
			slog.Error("unable to import ads", "user_id", ownerID, "idempotency_key", idempotencyKey, "ads_count", len(ads), "error", err)
			httpresponse.WriteError(rw, http.StatusInternalServerError, "internal error")
			return
		}

		ids := make([]int64, 0, len(adIDs))
		for _, id := range adIDs {
			ids = append(ids, int64(id))
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(rw).Encode(importAdsResponse{IDs: ids})
	}
}
