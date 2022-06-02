package main

import (
	"fmt"
	"github.com/homework3/comments/internal/app"
	"github.com/homework3/comments/internal/config"
	"github.com/homework3/comments/internal/database"
	"github.com/homework3/comments/internal/repository/pgx_repository"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

func main() {
	if err := config.ReadConfigYML("config.yml"); err != nil {
		log.Fatal(err)
	}

	a := app.App{Config: config.GetConfigInstance()}

	add := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		a.Config.Database.Host,
		a.Config.Database.Port,
		a.Config.Database.User,
		a.Config.Database.Password,
		a.Config.Database.Name,
	)

	adp, err := database.NewPgxPool(context.Background(), add)
	if err != nil {
		log.Fatalf("Db connect failed: %s", err)
	}
	defer adp.Close()

	a.Repo = pgx_repository.New(adp)

	s := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", a.Config.Rest.Port),
		Handler: a.Routes(),
	}

	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
