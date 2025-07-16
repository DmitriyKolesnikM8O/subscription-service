package main

import "github.com/DmitriyKolesnikM8O/subscription-service/internal/app"

const configPath = "config/config.yaml"

func main() {
	app.Run(configPath)
}
