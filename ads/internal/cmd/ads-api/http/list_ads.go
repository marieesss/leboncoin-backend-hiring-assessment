package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/httpresponse"
)

type listAdsItem struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Price      float64   `json:"price"`
	PriceCents int64     `json:"price_cents"` // Deprecated
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
			ownerID := domain.UserID(raw)
			req.OwnerID = ownerID
		}

		if raw := q.Get("min_price_cents"); raw != "" {
			v, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				httpresponse.WriteError(rw, http.StatusBadRequest, err.Error())
				return
			}

			req.MinPriceCents = v
		}

		if raw := q.Get("max_price_cents"); raw != "" {
			v, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				httpresponse.WriteError(rw, http.StatusBadRequest, err.Error())
				return
			}

			req.MaxPriceCents = v
		}

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
