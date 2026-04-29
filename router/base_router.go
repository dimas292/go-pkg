package router

import "github.com/gin-gonic/gin"

// Module is the interface every feature module must implement
// to register its routes with the central router.
type Module interface {
	// RegisterRoutes registers all routes for this module
	// under the provided router group (e.g., /api/v1).
	RegisterRoutes(rg *gin.RouterGroup)
}

// RegisterModules registers multiple modules under a common route group.
func RegisterModules(r *gin.Engine, prefix string, modules ...Module) {
	group := r.Group(prefix)
	for _, m := range modules {
		m.RegisterRoutes(group)
	}
}
