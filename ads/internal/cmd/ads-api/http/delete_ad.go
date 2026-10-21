package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/httpresponse"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/middleware"
)

// handlerDeleteAd serves DELETE /ads/{adID}.
func handlerDeleteAd(deleteAd usecase.DeleteAdFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		adID, err := adIDFromURL(r)
		if err != nil {
			httpresponse.WriteError(rw, http.StatusBadRequest, err.Error())
			return
		}

		userID := middleware.UserIDFromRequest(r)

		if err := deleteAd(r.Context(), adID, userID); err != nil {
			switch {
			case errors.Is(err, adsdao.ErrNotFound):
				httpresponse.WriteError(rw, http.StatusNotFound, fmt.Sprintf("ad %d not found", adID))
			case errors.Is(err, usecase.ErrForbidden):
				httpresponse.WriteError(rw, http.StatusForbidden, fmt.Sprintf("ad %d does not belong to the user", adID))
			default:
				slog.Error("unable to delete ad", "ad_id", adID, "user_id", userID, "error", err)
				httpresponse.WriteError(rw, http.StatusInternalServerError, "internal error")
			}

			return
		}

		rw.WriteHeader(http.StatusNoContent)
	}
}
