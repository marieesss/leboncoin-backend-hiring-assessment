package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	mockdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
)

func TestDeleteAd(t *testing.T) {
	ctx := context.Background()
	owned := domain.Ad{ID: 42, Title: "bike", Price: 5000, UserID: "user-1"}

	t.Run("nominal", func(t *testing.T) {
		m := mockdao.New(t)
		m.On("GetAdByID", ctx, domain.AdID(42)).Return(owned, nil)
		m.On("DeleteAd", ctx, domain.AdID(42)).Return(nil)

		err := usecase.NewDeleteAdFunc(m)(ctx, 42, "user-1")

		require.NoError(t, err)
	})

	t.Run("error no ad found", func(t *testing.T) {
		m := mockdao.New(t)
		m.On("GetAdByID", ctx, domain.AdID(99)).Return(domain.Ad{}, adsdao.ErrNotFound)

		err := usecase.NewDeleteAdFunc(m)(ctx, 99, "user-1")

		require.Error(t, err)
		assert.ErrorIs(t, err, adsdao.ErrNotFound)
	})

	t.Run("error ad belongs to another user", func(t *testing.T) {
		m := mockdao.New(t)
		m.On("GetAdByID", ctx, domain.AdID(42)).Return(owned, nil)

		err := usecase.NewDeleteAdFunc(m)(ctx, 42, "user-2")

		require.Error(t, err)
		assert.ErrorIs(t, err, usecase.ErrForbidden)
	})

	t.Run("error from dao on delete", func(t *testing.T) {
		m := mockdao.New(t)
		dbErr := errors.New("connection reset")
		m.On("GetAdByID", ctx, domain.AdID(42)).Return(owned, nil)
		m.On("DeleteAd", ctx, domain.AdID(42)).Return(dbErr)

		err := usecase.NewDeleteAdFunc(m)(ctx, 42, "user-1")

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
	})
}
