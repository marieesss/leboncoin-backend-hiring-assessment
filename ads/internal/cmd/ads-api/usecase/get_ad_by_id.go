package usecase

import (
	"context"
	"fmt"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	// GetAdByIDFunc returns a single ad with its owner's username. It fails with ErrorCodeNotFound when no ad matches the identifier.
	GetAdByIDFunc func(ctx context.Context, id domain.AdID) (AdDetails, error)

	// AdDetails is an ad enriched with its owner's username. Username is empty when the user service cannot provide it.
	AdDetails struct {
		Ad       domain.Ad
		Username string
	}

	getAdByID struct {
		adDAO adsdao.DAO
	}
)

// NewGetAdByIDFunc returns a GetAdByIDFunc backed by the given DAO.
func NewGetAdByIDFunc(adDAO adsdao.DAO) GetAdByIDFunc {
	return getAdByID{adDAO: adDAO}.getAdByID
}

func (uc getAdByID) getAdByID(ctx context.Context, id domain.AdID) (AdDetails, error) {
	found, err := uc.adDAO.GetAdByID(ctx, id)
	if err != nil {
		return AdDetails{}, fmt.Errorf("unable to get ad %s: %w", id, err)
	}

	return AdDetails{Ad: found}, nil
}
