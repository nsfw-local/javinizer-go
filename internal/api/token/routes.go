package token

import (
	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/api/core"
)

func RegisterRoutes(protected *gin.RouterGroup, deps *core.ServerDependencies) {
	protected.GET("/tokens", listTokens(deps))
	protected.POST("/tokens", createToken(deps))
	protected.DELETE("/tokens/:id", revokeToken(deps))
	protected.POST("/tokens/:id/regenerate", regenerateToken(deps))
}
