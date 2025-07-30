package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/DmitriyKolesnikM8O/subscription-service/config"
	v1 "github.com/DmitriyKolesnikM8O/subscription-service/internal/controller/http/v1"
	"github.com/DmitriyKolesnikM8O/subscription-service/internal/repo"
	"github.com/DmitriyKolesnikM8O/subscription-service/internal/service"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"github.com/DmitriyKolesnikM8O/subscription-service/pkg/client/postgres"
	"github.com/DmitriyKolesnikM8O/subscription-service/pkg/httpserver"
	"github.com/DmitriyKolesnikM8O/subscription-service/pkg/logger"
	"github.com/DmitriyKolesnikM8O/subscription-service/pkg/validator"

	_ "github.com/DmitriyKolesnikM8O/subscription-service/docs"
)

// @title           Subscription Service API
// @version         1.0
// @description     REST-сервис для агрегирования данных о подписках пользователей

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @schemes   http
func Run(configPath string) {

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	logger.SetLogrus(cfg.Log.Level)
	log.Info("Configuration successfully read")

	log.Info("Connect to BD")
	pool, err := postgres.NewClient(context.Background(), cfg.Storage, 3)
	if err != nil {
		log.Fatalf("Error when connecting DB: %v", err)
	}
	defer func() {
		pool.Close()

	}()

	log.Info("Initializing repositories")
	repositories := repo.NewRepositories(pool)

	log.Info("Initializing services")
	services := service.NewSubscriptionService(repositories)

	log.Info("Initializing controllers")
	handler := echo.New()
	handler.Validator = validator.NewValidator()
	v1.NewRouter(handler, services, log.StandardLogger())

	log.Info("Initializing HTTP-server")
	log.Debugf("Server port: %d", cfg.HTTP.Port)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	log.Info("Initializing graceful shutdown")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	log.Info("Shutting down...")
	err = httpServer.Shutdown()
	if err != nil {
		log.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
