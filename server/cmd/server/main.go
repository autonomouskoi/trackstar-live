package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/autonomouskoi/trackstar-live/server"
	"github.com/autonomouskoi/trackstar-live/server/store"
	"github.com/autonomouskoi/trackstar-live/server/store/sqlite3"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(-1)
}

func fatalIfError(err error, msg string) {
	if err != nil {
		fatal("error: ", msg, ": ", err)
	}
}

func main() {
	if len(os.Args) != 2 {
		fatal("usage: ", os.Args[0], "<config path>")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := server.LoadConfig(os.Args[1])
	fatalIfError(err, "loading config")

	logLevel := slog.LevelInfo
	if cfg.LogDebug {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	db, err := sqlite3.New(cfg.DBPath)
	fatalIfError(err, "opening database")
	store := store.New(db)

	handler, err := server.New(cfg, logger, store)
	fatalIfError(err, "creating handlers")

	srv := &http.Server{
		Addr:    cfg.Listen,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutting down server")
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error("shutting down server", "error", err.Error())
		}
	}()

	logger.Info("listening", "address", cfg.Listen)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("listening", "address", cfg.Listen, "error", err.Error())
	}

	if err := db.Close(); err != nil {
		logger.Error("closing database", "error", err.Error())
	}
}
