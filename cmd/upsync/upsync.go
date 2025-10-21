package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"upsync/internal/adapter/logger"
	"upsync/internal/adapter/storage"
	"upsync/internal/core/config"
	"upsync/internal/core/upsync"

	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, os.Kill)
	defer cancel()

	cfg, err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	lgr, err := logger.New(logger.SetLevel(cfg.LogLevel), logger.SetLogPath(cfg.LogPath))
	if err != nil {
		log.Fatal(err)
	}

	store, err := storage.New()
	if err != nil {
		log.Fatal(err)
	}

	ups, err := upsync.New(ctx, cfg.Upsync, store, lgr)
	if err != nil {
		log.Fatal(err)
	}

	if err := ups.Sync(ctx); err != nil {
		lgr.Error("failed sync", zap.Error(err))
	}

	lgr.Info("Stopping app ...")
	ups.Close()
	lgr.Info("Stopped app")
}
