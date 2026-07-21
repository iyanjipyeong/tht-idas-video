package http

import (
	"net/http"

	"idas-video/internal/adapter/inbound/http/handler"
	"idas-video/internal/adapter/inbound/http/middleware"
	"idas-video/internal/entity"
)

type Dependencies struct {
	HealthHandler         *handler.HealthHandler
	SwaggerUIHandler      *handler.SwaggerUIHandler
	OpenAPIHandler        *handler.OpenAPIHandler
	AuthHandler           *handler.AuthHandler
	SubscriptionHandler   *handler.SubscriptionHandler
	TierHandler           *handler.TierHandler
	TransactionHandler    *handler.TransactionHandler
	VideoHandler          *handler.VideoHandler
	PaymentHandler        *handler.PaymentHandler
	AuthMiddleware        func(http.Handler) http.Handler
	VideoAccessMiddleware func(http.Handler) http.Handler
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(entity.RouteHealth, deps.HealthHandler)
	mux.Handle(entity.RouteDocs, deps.SwaggerUIHandler)
	mux.Handle(entity.RouteOpenAPI, deps.OpenAPIHandler)
	mux.Handle(entity.RouteAuthLogin, http.HandlerFunc(deps.AuthHandler.Login))
	mux.Handle(entity.RouteSubscriptionsCreate, deps.AuthMiddleware(http.HandlerFunc(deps.SubscriptionHandler.Subscribe)))
	mux.Handle(entity.RouteSubscriptionActive, deps.AuthMiddleware(http.HandlerFunc(deps.SubscriptionHandler.GetActive)))
	mux.Handle(entity.RouteTiersList, http.HandlerFunc(deps.TierHandler.ListTiers))
	mux.Handle(entity.RouteTransactionsList, deps.AuthMiddleware(http.HandlerFunc(deps.TransactionHandler.ListTransactions)))
	mux.Handle(entity.RouteTransactionDetail, deps.AuthMiddleware(http.HandlerFunc(deps.TransactionHandler.GetTransaction)))
	mux.Handle(entity.RouteVideosList, deps.AuthMiddleware(http.HandlerFunc(deps.VideoHandler.ListVideos)))
	mux.Handle(entity.RouteVideoDetail, deps.AuthMiddleware(deps.VideoAccessMiddleware(http.HandlerFunc(deps.VideoHandler.GetVideo))))
	mux.Handle(entity.RoutePaymentCallback, http.HandlerFunc(deps.PaymentHandler.Callback))
	return middleware.RequestLogging(mux)
}
