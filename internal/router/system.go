package router

import (
	"github.com/gin-gonic/gin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
)

func registerSystemRoutes(engine *gin.Engine, health *httphandler.HealthHandler) {
	engine.GET(PathHealth, health.Health)
	engine.GET(PathReadiness, health.Readiness)
	engine.GET(PathLiveness, health.Liveness)
	engine.GET(PathVersion, health.Version)
}
