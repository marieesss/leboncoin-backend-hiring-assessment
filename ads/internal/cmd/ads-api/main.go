package main

import (
	"context"
	_ "embed"
	"log/slog"
	"os"

	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/postgres"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/http"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/userclient"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/httpserver"
)

//go:embed specs/openapi.yaml
var openAPISpec []byte

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	cfg := Load()

	adDAO, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if err := adDAO.Close(); err != nil {
			slog.Error("unable to close the database", "error", err)
		}
	}()

	// Without the user service, ads are returned without username.
	var userClient usecase.UserClient
	if cfg.IdentityURL == "" {
		slog.Warn("users service url not found, ads are returned without username")
	} else {
		userClient = userclient.New(cfg.IdentityURL, cfg.IdentityTimeout)
	}

	// Swagger UI is only exposed in development.
	var swaggerSpec []byte
	if cfg.IsDevelopment() {
		swaggerSpec = openAPISpec
	}

	router := http.NewRouter(http.Usecases{
		CreateAd:  usecase.NewCreateAdFunc(adDAO),
		DeleteAd:  usecase.NewDeleteAdFunc(adDAO),
		GetAd:     usecase.NewGetAdByIDFunc(adDAO, userClient),
		ListAds:   usecase.NewListAdsFunc(adDAO),
		ImportAds: usecase.NewImportAdsFunc(adDAO),
	}, swaggerSpec)

	return httpserver.Run(ctx, cfg.HTTPAddr, router)
}
