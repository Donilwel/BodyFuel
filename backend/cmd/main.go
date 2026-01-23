package main

import (
	"backend/internal/app"
	"flag"
)

var (
	configPath = flag.String("config", "./config/config.yaml", "path to config file. default: ./config/config.yaml")
)

// these are for the right swagger tags order

// @tag.name					Auth
// @tag.name					User Weight
// @tag.name					User Params
// @tag.name					User Info
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func main() {
	flag.Parse()

	application := app.NewApp(*configPath)
	application.Run()
}
