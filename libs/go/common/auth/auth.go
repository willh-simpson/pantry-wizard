package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func (v *TokenValidator) AuthWorker(jwksURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: missing or malformed token",
			})

			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		keySet, err := v.JWKSCache.Get(c.Request.Context(), v.JWKS_URL)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "could not verify identity: " + err.Error(),
			})

			return
		}

		token, err := jwt.ParseString(tokenString, jwt.WithKeySet(keySet))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: invalid token",
			})

			return
		}

		sub, found := token.Get("sub")
		if !found {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: identity missing from token",
			})

			return
		}

		c.Set("user_external_id", sub.(string))

		c.Next()
	}
}
