// Package main is the entry point of the application.
package main

import (
	"github.com/ilaziness/orange-tv/cmd"
	_ "github.com/ilaziness/orange-tv/docs/swagger"
)

// @title App Template API
// @version 1.0
// @description 这是一个灵活的 Go 应用模板框架，支持多种服务类型和可选组件集成
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
func main() {
	cmd.Execute()
}
