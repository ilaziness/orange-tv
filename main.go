// Package main is the entry point of the application.
package main

import (
	"github.com/ilaziness/orange-tv/cmd"
	_ "github.com/ilaziness/orange-tv/docs/swagger"
)

// @title 小橘TV
// @version 1.0
// @description 小橘TV - 影视站系统
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @host localhost:8080
// @BasePath /
func main() {
	cmd.Execute()
}
