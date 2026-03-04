package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func (v *TokenValidator) AuthWorker(jwksURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("processing request for %s", c.Request.URL.Path)

		if v == nil || v.JWKSCache == nil {
			log.Printf("CRITICAL: TokenValidator or cache is nil")

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "auth validator not initizalized",
			})

			return
		}

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("failed to parse header")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: missing or malformed token",
			})

			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		keySet, err := v.JWKSCache.Get(c.Request.Context(), v.JWKS_URL)
		if err != nil {
			log.Printf("jwks cache error: %v", err)

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "could not verify identity: " + err.Error(),
			})

			return
		}

		token, err := jwt.ParseString(tokenString, jwt.WithKeySet(keySet))
		if err != nil {
			log.Printf("token validation failed: %v", err)

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: invalid token",
			})

			return
		}

		sub, found := token.Get("sub")
		if !found {
			log.Printf("could not find identity from token")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: identity missing from token",
			})

			return
		}

		c.Set("user_external_id", sub.(string))

		c.Next()
	}
}
