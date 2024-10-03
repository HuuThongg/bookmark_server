package api

import (
	"bookmark/util"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type API struct {
	logger    *zerolog.Logger
	validator *validator.Validate
	db        *pgxpool.Pool
	config    *util.Config
}

func NewAPI(logger *zerolog.Logger, validator *validator.Validate, db *pgxpool.Pool, config *util.Config) *API {
	return &API{
		logger:    logger,
		validator: validator,
		db:        db,
		config:    config,
	}
}
