package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/homework3/comments/internal/cache"
	"github.com/homework3/comments/internal/config"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

type r struct {
	client *redis.Client
}

func New(cfg *config.Redis) cache.Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
	})

	return &r{client: client}
}

func (r *r) SetComments(ctx context.Context, itemId int32, comments []byte) error {
	status := r.client.Set(ctx, string(itemId), comments, time.Hour)
	if status.Err() != nil {
		log.Error().Err(status.Err()).Msg("Failed to save comments in redis")
	}

	return status.Err()
}

func (r *r) GetComments(ctx context.Context, itemId int32) (comments []byte, err error) {
	res := r.client.Get(ctx, string(itemId))
	if err = res.Err(); err != nil {
		log.Error().Err(err).Msg("Failed to get comments from redis")

		return
	}

	if comments, err = res.Bytes(); err != nil {
		log.Error().Err(err).Msg("Failed to get []bytes from redis result")
	}

	return
}

func (r *r) Close() error {
	return r.client.Close()
}
