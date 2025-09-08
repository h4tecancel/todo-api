package main

import (
	"log"
	"todo-api/internal/config"
	"todo-api/internal/lib/logger"
	"todo-api/internal/server"
	"todo-api/internal/server/handlers"
	"todo-api/internal/storage/sqlite"
)

func main() {
	logger := logger.Init()
	logger.Info("Server started!")
	cfg := config.MustLoad()
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Fatal(err)
	}

	//handlersContext, handlerCancel := context.WithTimeout(context.Background(), cfg.HTTPServer.Timeout)
	server := server.New(handlers.New(logger, storage))
	//defer handlerCancel()

	if err := server.Start(cfg.HTTPServer.Address, cfg.HTTPServer.IdleTimeout, cfg.HTTPServer.Timeout); err != nil {
		log.Fatal(err)
	}

}
