package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/httpresponse"
)

type listAdsItem struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Price      float64   `json:"price"`
	PriceCents int64     `json:"price_cents"` // Deprecated see issue #6
	PhotoURL   string    `json:"photo_url"`
	UserID     string    `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type listAdsResponse struct {
	Ads []listAdsItem `json:"ads"`
}

func adToListAdsItem(a domain.Ad) listAdsItem {
	return listAdsItem{
		ID:         int64(a.ID),
		Title:      a.Title,
		Price:      centsToEuros(a.Price),
		PriceCents: a.Price,
		PhotoURL:   a.PhotoURL,
		UserID:     a.UserID.String(),
		CreatedAt:  a.CreatedAt.UTC(),
	}
}

func handlerListAds(listAds usecase.ListAdsFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		req := usecase.ListAdsRequest{
			Title: q.Get("title"),
		}

		if raw := q.Get("user_id"); raw != "" {
			if err := uuid.Validate(raw); err != nil {
				httpresponse.WriteError(rw, http.StatusBadRequest, fmt.Sprintf("user_id %q is not a valid UUID", raw))
				return
			}

			req.OwnerID = domain.UserID(raw)
		}

		minPriceCents, maxPriceCents, err := priceFiltersInCents(q)
		if err != nil {
			httpresponse.WriteError(rw, http.StatusBadRequest, err.Error())
			return
		}

		req.MinPriceCents = minPriceCents
		req.MaxPriceCents = maxPriceCents

		ads, err := listAds(r.Context(), req)
		if err != nil {
			slog.Error("unable to list ads", "user_id", req.OwnerID, "error", err)
			httpresponse.WriteError(rw, http.StatusInternalServerError, "internal error")
			return
		}

		items := make([]listAdsItem, 0, len(ads))
		for _, a := range ads {
			items = append(items, adToListAdsItem(a))
		}

		rw.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(rw).Encode(listAdsResponse{Ads: items})
	}
}

// Price filters validator: either min_price and max_price in euros, or min_price_cents and max_price_cents, never both
func priceFiltersInCents(q url.Values) (int64, int64, error) {
	inEuros := q.Get("min_price") != "" || q.Get("max_price") != ""
	inCents := q.Get("min_price_cents") != "" || q.Get("max_price_cents") != ""
	if inEuros && inCents {
		return 0, 0, errors.New("send either min_price and max_price or min_price_cents and max_price_cents, not both")
	}

	minCents, err := priceFilter(q, "min_price", "min_price_cents")
	if err != nil {
		return 0, 0, err
	}

	maxCents, err := priceFilter(q, "max_price", "max_price_cents")
	if err != nil {
		return 0, 0, err
	}

	return minCents, maxCents, nil
}

func priceFilter(q url.Values, eurosParam, centsParam string) (int64, error) {
	// Parse string euros into cents
	if raw := q.Get(eurosParam); raw != "" {
		euros, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0, fmt.Errorf("%s %q is not a number", eurosParam, raw)
		}

		cents, err := eurosToCents(euros)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", eurosParam, err)
		}

		return cents, nil
	}

	// Parse string cents
	if raw := q.Get(centsParam); raw != "" {
		cents, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%s %q is not a number", centsParam, raw)
		}

		return cents, nil
	}

	return 0, nil
}
