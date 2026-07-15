// Package router provides centralized route registration and management.
package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterRoutes registers all routes to the gin engine.
func RegisterRoutes(engine *gin.Engine, h *Handlers) error {
	if err := h.validateForRoutes(); err != nil {
		return err
	}

	engine.GET(PathSwagger, ginSwagger.WrapHandler(swaggerFiles.Handler))
	registerSystemRoutes(engine, h.Health)
	registerClientRoutes(engine, h)
	registerAdminRoutes(engine, h)
	registerOpenRoutes(engine, h)
	registerInternalRoutes(engine, h)
	return nil
}
