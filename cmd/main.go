package main

import (
	"fmt"
	"github.com/homework3/comments/internal/app"
	"github.com/homework3/comments/internal/config"
	"github.com/homework3/comments/internal/database"
	"github.com/homework3/comments/internal/repository/pgx_repository"
	"github.com/homework3/comments/internal/tracer"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

func main() {
	if err := config.ReadConfigYML("config.yml"); err != nil {
		log.Fatal().Err(err).Msg("Failed to read configuration")
	}

	cfg := config.GetConfigInstance()

	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	adr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	if cfg.Database.MigrationsOn {
		err := runMigration("pgx", adr, cfg.Database.Migrations)
		if err != nil {
			return
		}
	}

	adp, err := database.NewPgxPool(context.Background(), adr)
	if err != nil {
		log.Fatal().Err(err).Msg("Db connect failed: %s")
	}
	defer adp.Close()

	repo := pgx_repository.New(adp)

	tracing, err := tracer.NewTracer(&cfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed init tracing")

		return
	}
	defer tracing.Close()

	if err = app.New(repo).Start(&cfg); err != nil {
		log.Error().Err(err).Msg("Failed to start app")
	}
}

func runMigration(driver, adr, migration string) error {
	conn, err := goose.OpenDBWithDriver(driver, adr)
	if err != nil {
		log.Error().Err(err).Msg("db connection failed")

		return err
	}
	defer conn.Close()

	if err = goose.Up(conn, migration); err != nil {
		log.Error().Err(err).Msg("Migration failed")

		return err
	}

	return nil
}
