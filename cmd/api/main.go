package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	logOption "github.com/digitalrealmforgestudios/d-logger/option"

	"idas-video/internal/app"
	"idas-video/internal/infrastructure/config"
	"idas-video/internal/infrastructure/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conf := config.Load()
	log := logger.Init()
	log.Info("application bootstrapping", logOption.Attribute("address", conf.Address))

	application, err := app.New(ctx, conf)
	if err != nil {
		log.Fatal("application bootstrap failed", logOption.Error(err))
		return
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Error("application close failed", logOption.Error(err))
		}
	}()

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("server listening", logOption.Attribute("address", conf.Address))
		serverErrors <- application.Run()
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := application.Server.Shutdown(shutdownCtx); err != nil {
			log.Fatal("server shutdown failed", logOption.Error(err))
			return
		}
		log.Info("server shutdown completed")
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server stopped with error", logOption.Error(err))
			return
		}
		log.Info("server stopped")
	}
}
