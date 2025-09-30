package routes

import (
	"orderease/handlers"
	"orderease/routes/backend"
	"orderease/routes/frontend"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, h *handlers.Handler) {
	backend.SetupAdminRoutes(r, h)
	backend.SetupNoAuthRoutes(r, h)
	backend.SetupShopRoutes(r, h)

	frontend.SetupFrontRoutes(r, h)
}
