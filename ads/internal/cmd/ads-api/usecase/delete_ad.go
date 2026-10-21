package usecase

import (
	"context"
	"errors"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	// DeleteAdFunc deletes an ad owned by the calling user. It fails with dao.ErrNotFound when no ad matches the identifier, and with ErrForbidden when the ad belongs to someone else.
	DeleteAdFunc func(ctx context.Context, id domain.AdID, userID domain.UserID) error

	deleteAd struct {
		adDAO adsdao.DAO
	}
)

// ErrForbidden is returned when the calling user is not allowed to act on the ad.
var ErrForbidden = errors.New("forbidden")

// NewDeleteAdFunc returns a DeleteAdFunc backed by the given DAO.
func NewDeleteAdFunc(adDAO adsdao.DAO) DeleteAdFunc {
	return deleteAd{adDAO: adDAO}.deleteAd
}

func (uc deleteAd) deleteAd(_ context.Context, _ domain.AdID, _ domain.UserID) error {
	return errors.New("delete ad is not implemented yet")
}
