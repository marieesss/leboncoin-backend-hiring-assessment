package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
)

func TestImportAds(t *testing.T) {
	ctx := context.Background()

	req := usecase.ImportAdsRequest{
		OwnerID: "user-1",
		Ads: []usecase.ImportedAd{
			{Title: "bike", PriceCents: 5000, PhotoURL: "https://example.com/bike.jpg"},
			{Title: "scooter", PriceCents: 8000, PhotoURL: "https://example.com/scooter.jpg"},
		},
	}
	bike := adsdao.CreateAdRequest{Title: "bike", PriceCents: 5000, PhotoURL: "https://example.com/bike.jpg", OwnerID: "user-1"}
	scooter := adsdao.CreateAdRequest{Title: "scooter", PriceCents: 8000, PhotoURL: "https://example.com/scooter.jpg", OwnerID: "user-1"}

	t.Run("nominal", func(t *testing.T) {
		m := mock.New(t)
		m.On("Begin", ctx).Return(nil).Once()
		m.On("CreateAd", ctx, bike).Return(domain.AdID(10), nil).Once()
		m.On("CreateAd", ctx, scooter).Return(domain.AdID(11), nil).Once()
		m.On("Commit").Return(nil).Once()
		m.On("Rollback").Return(nil).Once()

		got, err := usecase.NewImportAdsFunc(m)(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, []domain.AdID{10, 11}, got)
	})

	t.Run("error when the transaction cannot be opened", func(t *testing.T) {
		m := mock.New(t)
		dbErr := errors.New("db down")
		m.On("Begin", ctx).Return(dbErr).Once()

		_, err := usecase.NewImportAdsFunc(m)(ctx, req)

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("rolls back and stops on first create error", func(t *testing.T) {
		m := mock.New(t)
		dbErr := errors.New("connection reset")
		m.On("Begin", ctx).Return(nil).Once()
		m.On("CreateAd", ctx, bike).Return(domain.AdID(0), dbErr).Once()
		m.On("Rollback").Return(nil).Once()

		_, err := usecase.NewImportAdsFunc(m)(ctx, req)

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("error when the commit fails", func(t *testing.T) {
		m := mock.New(t)
		dbErr := errors.New("commit failed")
		m.On("Begin", ctx).Return(nil).Once()
		m.On("CreateAd", ctx, bike).Return(domain.AdID(10), nil).Once()
		m.On("CreateAd", ctx, scooter).Return(domain.AdID(11), nil).Once()
		m.On("Commit").Return(dbErr).Once()
		m.On("Rollback").Return(nil).Once()

		_, err := usecase.NewImportAdsFunc(m)(ctx, req)

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
	})
}
