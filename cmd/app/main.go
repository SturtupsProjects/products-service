package main

import (
	"crm-admin/config"
	"crm-admin/internal/app"
)

func main() {
	cfg := config.NewConfig()

	app.Run(cfg)
}
