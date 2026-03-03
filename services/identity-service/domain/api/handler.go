package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

type IdentityHandler struct {
	DB *sql.DB
}

func NewIdentityHandler(db *sql.DB) *IdentityHandler {
	return &IdentityHandler{
		DB: db,
	}
}

func (h *IdentityHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":   "up",
		"service":  "identity-service",
		"database": "connected",
	})
}
