package main

import (
	"bookmark/api/router"
	"bookmark/db/connection"
	"bookmark/util"
	"bookmark/util/logger"
	"bookmark/util/validator"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config", err)
	}

	l := logger.New(config.Debug)
	v := validator.New()
	// var logLevel zerolog.Level
	// if config.Debug {
	// 	logLevel = zerolog.InfoLevel
	// } else {
	// 	logLevel = zerolog.ErrorLevel
	// }
	// logger := zerolog.New(os.Stdout).Level(logLevel)

	db := connection.ConnectDB()
	defer db.Close()
	r := router.Router(l, v, db, &config)

	server := &http.Server{
		Addr:         config.PORT,
		Handler:      r,
		ReadTimeout:  config.TimeoutRead,
		WriteTimeout: config.TimeoutWrite,
		IdleTimeout:  config.TimeoutIdle,
	}

	closed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		l.Info().Msgf("Shutting down server at %s", server.Addr)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			l.Error().Err(err).Msg("Server shutdown failure")
		}

		close(closed)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.Fatal().Err(err).Msg("Server startup failure")
	}

	<-closed
	l.Info().Msg("Server shutdown successfully")
}
