package usecase

import (
	"context"
	"fmt"
	"log/slog"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/model"
	usersclient "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg"
)

type (
	// GetAdByIDFunc returns a single ad with its owner's username. It fails with ErrorCodeNotFound when no ad matches the identifier.
	GetAdByIDFunc func(ctx context.Context, id domain.AdID) (AdDetails, error)

	// AdDetails is an ad enriched with its owner's username. Username is empty when the user service cannot provide it.
	AdDetails struct {
		Ad       domain.Ad
		Username string
	}

	// UserClient is the part of the user service GetAdByID needs.
	UserClient interface {
		GetUserByID(ctx context.Context, userID model.UserID) (usersclient.User, error)
	}

	getAdByID struct {
		adDAO      adsdao.DAO
		userClient UserClient
	}
)

// NewGetAdByIDFunc returns a GetAdByIDFunc backed by the given DAO and user
func NewGetAdByIDFunc(adDAO adsdao.DAO, userClient UserClient) GetAdByIDFunc {
	return getAdByID{adDAO: adDAO, userClient: userClient}.getAdByID
}

func (uc getAdByID) getAdByID(ctx context.Context, id domain.AdID) (AdDetails, error) {
	found, err := uc.adDAO.GetAdByID(ctx, id)
	if err != nil {
		return AdDetails{}, fmt.Errorf("unable to get ad %s: %w", id, err)
	}

	// The username is secondary: the ad is still returned when the user service fails
	user, err := uc.userClient.GetUserByID(ctx, found.UserID)
	if err != nil {
		slog.Warn("unable to get ad owner username", "ad_id", id, "user_id", found.UserID, "error", err)

		return AdDetails{Ad: found}, nil
	}

	return AdDetails{Ad: found, Username: user.Nickname}, nil
}
