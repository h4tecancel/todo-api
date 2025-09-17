package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-api/internal/config"
	"todo-api/internal/lib/logger"
	"todo-api/internal/server"
	"todo-api/internal/server/handlers"
	"todo-api/internal/storage/sqlite"
)

func main() {
	logger := logger.Init()
	cfg := config.MustLoad()
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Fatal(err)
	}
	srv := server.New(handlers.New(logger, storage))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(cfg.HTTPServer.Address, cfg.HTTPServer.IdleTimeout, cfg.HTTPServer.Timeout); err != nil {
			log.Fatal(err)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Fatal(err)
	}
}

