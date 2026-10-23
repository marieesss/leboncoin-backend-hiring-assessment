package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	// ImportAdsFunc creates every ad of the request for the calling user, or none of them.
	// A retry with the same idempotency key returns the ads created the first time instead of creating them again.
	ImportAdsFunc func(ctx context.Context, req ImportAdsRequest) ([]domain.AdID, error)

	// ImportAdsRequest describes a bulk import. OwnerID is the authenticated user and owns every imported ad.
	// IdempotencyKey is optional: when empty, every call creates the ads.
	ImportAdsRequest struct {
		OwnerID        domain.UserID
		IdempotencyKey string
		Ads            []ImportedAd
	}

	// ImportedAd is a single ad of a bulk import.
	ImportedAd struct {
		Title      string
		PriceCents int64
		PhotoURL   string
	}

	importAds struct {
		adDAO adsdao.DAO
	}
)

// ErrIdempotencyKeyReused is returned when an idempotency key is sent again with other ads.
var ErrIdempotencyKeyReused = errors.New("idempotency key already used with another ads")

// NewImportAdsFunc returns an ImportAdsFunc backed by the given DAO.
func NewImportAdsFunc(adDAO adsdao.DAO) ImportAdsFunc {
	return importAds{adDAO: adDAO}.importAds
}

func (uc importAds) importAds(ctx context.Context, req ImportAdsRequest) ([]domain.AdID, error) {
	tx, err := uc.adDAO.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to import ads: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	key := adsdao.ImportKey{OwnerID: req.OwnerID, IdempotencyKey: req.IdempotencyKey}

	if req.IdempotencyKey != "" {
		hash, err := hashImportedAds(req.Ads)
		if err != nil {
			return nil, fmt.Errorf("unable to import ads: %w", err)
		}

		reported, err := tx.ReportImport(ctx, key, hash)
		if err != nil {
			return nil, fmt.Errorf("unable to import ads: %w", err)
		}

		// The key was already used: replay the first import instead of creating the ads again.
		if !reported {
			return replayImport(ctx, tx, key, hash)
		}
	}

	ids := make([]domain.AdID, 0, len(req.Ads))

	for i, ad := range req.Ads {
		id, err := tx.CreateAd(ctx, adsdao.CreateAdRequest{
			Title:      ad.Title,
			PriceCents: ad.PriceCents,
			PhotoURL:   ad.PhotoURL,
			OwnerID:    req.OwnerID,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to import ad %d: %w", i, err)
		}

		ids = append(ids, id)
	}

	if req.IdempotencyKey != "" {
		if err := tx.CompleteImport(ctx, key, ids); err != nil {
			return nil, fmt.Errorf("unable to import ads: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("unable to import ads: %w", err)
	}

	return ids, nil
}

// replayImport returns the ads created by the first import made with the key, as long as it carried the same ads.
func replayImport(ctx context.Context, tx adsdao.Tx, key adsdao.ImportKey, hash string) ([]domain.AdID, error) {
	imp, err := tx.GetImport(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to replay import: %w", err)
	}

	if imp.RequestHash != hash {
		return nil, fmt.Errorf("unable to replay import: %w", ErrIdempotencyKeyReused)
	}

	return imp.AdIDs, nil
}

// hashImportedAds identifies the content of an import. Prices are already in cents.
func hashImportedAds(ads []ImportedAd) (string, error) {
	payload, err := json.Marshal(ads)
	if err != nil {
		return "", fmt.Errorf("unable to hash imported ads: %w", err)
	}

	sum := sha256.Sum256(payload)

	return hex.EncodeToString(sum[:]), nil
}
