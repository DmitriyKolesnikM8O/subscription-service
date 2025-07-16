package app

import (
	"github.com/DmitriyKolesnikM8O/subscription-service/config"
	log "github.com/sirupsen/logrus"
)

func Run(configPath string) {
	_, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

}
