package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmd-moradi/goup/configs"
	_ "github.com/mmd-moradi/goup/docs"
	"github.com/mmd-moradi/goup/internal/api"
	"github.com/mmd-moradi/goup/pkg/logger"
)

// @title           Goup API
// @version         1.0
// @description     A photo upload service built with Go.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/mmd-moradi/goup
// @contact.email  mmdmi.workspace@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the access token.

func main() {
	log := logger.New()
	cfg, err := configs.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	server := api.NewServer(cfg, log)
	go func() {
		log.Info().Str("addr", cfg.Server.Addr).Msg("Starting the server...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)

	if err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}
