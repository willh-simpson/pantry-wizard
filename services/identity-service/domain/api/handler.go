package api

import (
	"github.com/gin-gonic/gin"
)

type IdentityHandler struct {
}

func NewIdentityHandler() *IdentityHandler {
	return &IdentityHandler{}
}

func (h *IdentityHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":   "up",
		"service":  "identity-service",
		"database": "connected",
	})
}
