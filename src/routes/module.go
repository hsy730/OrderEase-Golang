package routes

import (
	ordercontextHandlers "orderease/contexts/ordercontext/application/handlers"
	"orderease/routes/backend"
	"orderease/routes/frontend"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, h *ordercontextHandlers.Handler) {
	backend.SetupAdminRoutes(r, h)
	backend.SetupNoAuthRoutes(r, h)
	backend.SetupShopRoutes(r, h)

	frontend.SetupFrontRoutes(r, h)
	frontend.SetupFrontNoAuthRoutes(r, h)
}
