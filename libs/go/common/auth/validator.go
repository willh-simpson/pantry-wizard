package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

type TokenValidator struct {
	JWKSCache *jwk.Cache
	JWKS_URL  string
}

func NewTokenValidator(region, userPoolID string) *TokenValidator {
	ctx := context.Background()
	url := fmt.Sprintf("http://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)

	cache := jwk.NewCache(ctx)
	cache.Register(url, jwk.WithMinRefreshInterval(15*time.Minute))

	return &TokenValidator{
		JWKSCache: cache,
		JWKS_URL:  url,
	}
}
