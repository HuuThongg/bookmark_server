package api

import (
	"bookmark/util"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type API struct {
	logger    *zerolog.Logger
	validator *validator.Validate
	db        *pgxpool.Pool
	config    *util.Config
	redis     *redis.Client
}

func NewAPI(logger *zerolog.Logger, validator *validator.Validate, db *pgxpool.Pool, config *util.Config, redis *redis.Client) *API {
	return &API{
		logger:    logger,
		validator: validator,
		db:        db,
		config:    config,
		redis:     redis,
	}
}
