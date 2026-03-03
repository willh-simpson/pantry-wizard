package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/services/identity-service/auth/client"
	"github.com/willh-simpson/pantry-wizard/services/identity-service/domain/database"
	"github.com/willh-simpson/pantry-wizard/services/identity-service/domain/model"
)

type IdentityHandler struct {
	DB            *sql.DB
	CognitoClient *client.CognitoClient
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

func (h *IdentityHandler) GetUserProfile(c *gin.Context) {
	externalID, exists := c.Get("user_external_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "identity context missing",
		})

		return
	}

	user, err := database.GetUserByExternalID(h.DB, c.Request.Context(), externalID.(string))
	if err != nil {
		log.Printf("user %s not found in local database: %v", externalID, err)

		c.JSON(http.StatusNotFound, gin.H{
			"error": "user profile not found",
		})

		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *IdentityHandler) Register(c *gin.Context) {
	var req model.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	externalID, err := h.CognitoClient.SignUp(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "auth provider failure: " + err.Error(),
		})

		return
	}

	user, err := database.CreateOrUpdateUser(h.DB, c.Request.Context(), req.Email, externalID, req.DisplayName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "database sync failure",
			"error":   err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *IdentityHandler) ConfirmRegistration(c *gin.Context) {
	var req model.ConfirmRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	err := h.CognitoClient.ConfirmSignUp(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		log.Printf("confirmation failed for %s: %v", req.Email, err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid or expired confirmation code",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "account confirmed successfully",
	})
}

func (h *IdentityHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	authResult, err := h.CognitoClient.SignIn(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("login failed for %s: %v", req.Email, err)

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})

		return
	}

	user, err := database.GetUserByEmail(h.DB, c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "profile sync error",
			"error":   err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  authResult.AccessToken,
		"id_token":      authResult.IdToken,
		"refresh_token": authResult.RefreshToken,
		"expires_in":    authResult.ExpiresIn,
		"user":          user,
	})
}
