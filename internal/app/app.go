package app

import (
	"github.com/homework3/comments/internal/config"
	"github.com/homework3/comments/internal/repository"
)

type App struct {
	Config config.Config
	Repo   repository.Repository
}
