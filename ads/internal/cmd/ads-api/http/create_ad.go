package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/httpresponse"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/middleware"
)

type createAdBody struct {
	Title      string   `json:"title"`
	Price      *float64 `json:"price,omitempty"`
	PriceCents *int64   `json:"price_cents,omitempty"` // Deprecated: kept until web and mobile clients send price.
	PhotoURL   string   `json:"photo_url"`
}

// Price body validator
func (b createAdBody) priceInCents() (int64, error) {
	switch {
	case b.Price != nil && b.PriceCents != nil:
		return 0, errors.New("send either price or price_cents, not both")
	case b.Price != nil:
		return eurosToCents(*b.Price)
	case b.PriceCents != nil:
		return *b.PriceCents, nil
	default:
		return 0, errors.New("price or price_cents is required")
	}
}

type createAdResponse struct {
	ID int64 `json:"id"`
}

// handlerCreateAd serves POST /ads.
func handlerCreateAd(createAd usecase.CreateAdFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var body createAdBody
		if err := decodeBody(r, &body); err != nil {
			httpresponse.WriteError(rw, http.StatusBadRequest, err.Error())
			return
		}

		priceCents, err := body.priceInCents()
		if err != nil {
			httpresponse.WriteError(rw, http.StatusBadRequest, err.Error())
			return
		}

		ownerID := middleware.UserIDFromRequest(r)

		adID, err := createAd(r.Context(), usecase.CreateAdRequest{
			Title:      body.Title,
			PriceCents: priceCents,
			PhotoURL:   body.PhotoURL,
			OwnerID:    ownerID,
		})
		if err != nil {
			slog.Error("unable to create ad", "user_id", ownerID, "error", err)
			httpresponse.WriteError(rw, http.StatusInternalServerError, "internal error")
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(rw).Encode(createAdResponse{ID: int64(adID)})
	}
}
