package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	logOption "github.com/digitalrealmforgestudios/d-logger/option"

	apphttp "idas-video/internal/adapter/inbound/http"
	"idas-video/internal/adapter/inbound/http/handler"
	"idas-video/internal/adapter/inbound/http/middleware"
	repositorypostgres "idas-video/internal/adapter/outbound/postgres"
	"idas-video/internal/infrastructure/buildinfo"
	"idas-video/internal/infrastructure/config"
	"idas-video/internal/infrastructure/logger"
	"idas-video/internal/usecase"
)

type Application struct {
	Server *http.Server
	DB     *sql.DB
}

func New(ctx context.Context, conf config.Config) (*Application, error) {
	log := logger.Child("app")
	log.Info("validating configuration", logger.WithContext(ctx))
	if err := conf.Validate(); err != nil {
		log.Error("configuration validation failed", logger.WithContext(ctx), logOption.Error(err))
		return nil, err
	}

	log.Info("opening database connection", logger.WithContext(ctx), logOption.Attribute("database.host", conf.DatabaseHost), logOption.Attribute("database.name", conf.DatabaseName))
	gormDB, err := repositorypostgres.Open(ctx, conf.DatabaseDSN())
	if err != nil {
		log.Error("database connection failed", logger.WithContext(ctx), logOption.Error(err), logOption.Attribute("database.host", conf.DatabaseHost), logOption.Attribute("database.name", conf.DatabaseName))
		return nil, err
	}
	log.Info("database connection ready", logger.WithContext(ctx))

	db, err := repositorypostgres.SQLDB(gormDB)
	if err != nil {
		log.Error("database sql handle failed", logger.WithContext(ctx), logOption.Error(err))
		return nil, err
	}

	store := repositorypostgres.NewStore(gormDB)
	usecaseLogger := logger.NewUsecaseLogger()
	authUsecase := usecase.NewAuthUsecaseWithLogger(store, conf.JWTSecret, usecaseLogger)
	tierUsecase := usecase.NewTierUsecase(store)
	transactionUsecase := usecase.NewTransactionUsecase(store)
	videoUsecase := usecase.NewVideoUsecaseWithLogger(store, store, usecaseLogger)
	subscriptionUsecase := usecase.NewSubscriptionUsecaseWithLogger(store, store, store, usecaseLogger)
	paymentUsecase := usecase.NewPaymentCallbackUsecaseWithLogger(store, subscriptionUsecase, store, usecaseLogger)

	router := apphttp.NewRouter(apphttp.Dependencies{
		HealthHandler:         handler.NewHealthHandler(buildinfo.BuildID, buildinfo.BuildTime),
		SwaggerUIHandler:      handler.NewSwaggerUIHandler(),
		OpenAPIHandler:        handler.NewOpenAPIHandler(),
		AuthHandler:           handler.NewAuthHandler(authUsecase),
		SubscriptionHandler:   handler.NewSubscriptionHandler(subscriptionUsecase),
		TierHandler:           handler.NewTierHandler(tierUsecase),
		TransactionHandler:    handler.NewTransactionHandler(transactionUsecase),
		VideoHandler:          handler.NewVideoHandler(videoUsecase),
		PaymentHandler:        handler.NewPaymentHandler(paymentUsecase),
		AuthMiddleware:        middleware.AuthMiddleware(authUsecase),
		VideoAccessMiddleware: middleware.VideoAccess(videoUsecase),
	})

	server := &http.Server{
		Addr:              conf.Address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Application{Server: server, DB: db}, nil
}

func (application *Application) Close() error {
	if application == nil || application.DB == nil {
		return nil
	}
	logger.Child("app").Info("closing database connection")
	return application.DB.Close()
}

func (application *Application) Run() error {
	if application == nil || application.Server == nil {
		return fmt.Errorf("application server is not configured")
	}
	logger.Child("app").Info("http server starting", logOption.Attribute("address", application.Server.Addr))
	return application.Server.ListenAndServe()
}
