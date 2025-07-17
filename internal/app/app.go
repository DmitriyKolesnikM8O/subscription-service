package app

import (
	"context"
	"fmt"

	"github.com/DmitriyKolesnikM8O/subscription-service/config"
	"github.com/DmitriyKolesnikM8O/subscription-service/internal/repo"
	"github.com/DmitriyKolesnikM8O/subscription-service/internal/service"
	log "github.com/sirupsen/logrus"

	"github.com/DmitriyKolesnikM8O/subscription-service/pkg/client/postgres"
	"github.com/DmitriyKolesnikM8O/subscription-service/pkg/logger"
)

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
		log.Fatal(fmt.Errorf("Error when connecting DB: %w", err))
	}
	defer pool.Close()

	log.Info("Initializing repositories")
	repositories := repo.NewRepositories(pool)

	log.Info("Initializing services")
	_ = service.NewSubscriptionService(repositories)

}
