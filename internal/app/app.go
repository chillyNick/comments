package app

import (
	"errors"
	"fmt"
	"github.com/homework3/comments/internal/config"
	"github.com/homework3/comments/internal/http_server"
	"github.com/homework3/comments/internal/kafka"
	"github.com/homework3/comments/internal/metrics"
	"github.com/homework3/comments/internal/repository"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	repo     repository.Repository
	producer kafka.Producer
}

func New(repo repository.Repository) *App {
	return &App{
		repo: repo,
	}
}

func (a *App) Start(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := kafka.StartProcessMessages(ctx, a.repo, &cfg.Kafka); err != nil {
			log.Error().Err(err).Msg("Failed to start kafka consumer")
			cancel()
		}
	}()

	producer, err := kafka.CreateProducer(&cfg.Kafka)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start kafka producer")

		return err
	}
	a.producer = producer

	restAddr := fmt.Sprintf("%s:%v", cfg.Rest.Host, cfg.Rest.Port)
	metricsAddr := fmt.Sprintf("%s:%v", cfg.Metrics.Host, cfg.Metrics.Port)

	restServer := http_server.CreateRestServer(restAddr, a.getRouter())

	go func() {
		log.Info().Msgf("Rest http_server is running on %s", restAddr)
		if err := restServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Failed running gateway http_server")
			cancel()
		}
	}()

	metricsServer := metrics.CreateMetricsServer(metricsAddr, cfg)

	go func() {
		log.Info().Msgf("Metrics http_server is running on %s", metricsAddr)
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Failed running metrics http_server")
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
		log.Info().Msgf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		log.Info().Msgf("ctx.Done: %v", done)
	}

	if err := restServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("restServer.Shutdown")
	} else {
		log.Info().Msg("restServer shut down correctly")
	}

	if err := metricsServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("metricsServer.Shutdown")
	} else {
		log.Info().Msg("metricsServer shut down correctly")
	}

	return nil
}
