package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWK represents a JSON Web Key in a JWKS set.
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Crv string `json:"crv"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

var (
	jwksCache     *JWKS
	jwksCacheTime time.Time
	jwksMutex     sync.RWMutex
)

// fetchJWKS downloads the JSON Web Key Set from Supabase, caching it in memory.
func fetchJWKS(supabaseURL string) (*JWKS, error) {
	jwksMutex.RLock()
	// Cache for 1 hour to avoid calling Supabase API on every single request
	if jwksCache != nil && time.Since(jwksCacheTime) < 1*time.Hour {
		defer jwksMutex.RUnlock()
		return jwksCache, nil
	}
	jwksMutex.RUnlock()

	jwksMutex.Lock()
	defer jwksMutex.Unlock()

	// Double check cache
	if jwksCache != nil && time.Since(jwksCacheTime) < 1*time.Hour {
		return jwksCache, nil
	}

	url := fmt.Sprintf("%s/auth/v1/jwks", strings.TrimSuffix(supabaseURL, "/"))
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned HTTP %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	jwksCache = &jwks
	jwksCacheTime = time.Now()
	return jwksCache, nil
}

// parseECDSAPublicKey converts JWK coordinates into an ecdsa.PublicKey.
func parseECDSAPublicKey(jwk JWK) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Y: %w", err)
	}

	var curve elliptic.Curve
	switch jwk.Crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", jwk.Crv)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

// parseRSAPublicKey converts JWK parameters into an rsa.PublicKey.
func parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	var eVal int
	for _, b := range eBytes {
		eVal = (eVal << 8) | int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eVal,
	}, nil
}

// AuthMiddleware validates the Supabase JWT token using JWKS public keys or a legacy secret.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		supabaseURL := os.Getenv("SUPABASE_URL")
		jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")

		// In local development, if neither URL nor Secret are set, bypass auth with a warning
		if supabaseURL == "" && jwtSecret == "" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format. Use 'Bearer <token>'"})
			return
		}

		tokenStr := parts[1]

		// Parse and validate token using JWKS (ES256/RS256) or fallback HS256
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			kid, _ := token.Header["kid"].(string)

			// 1. Try JWKS first if we have a Supabase URL and a key ID (kid) in token header
			if supabaseURL != "" && kid != "" {
				jwks, err := fetchJWKS(supabaseURL)
				if err == nil {
					for _, key := range jwks.Keys {
						if key.Kid == kid {
							if key.Kty == "EC" {
								return parseECDSAPublicKey(key)
							} else if key.Kty == "RSA" {
								return parseRSAPublicKey(key)
							}
						}
					}
				}
			}

			// 2. Fallback: HMAC HS256 verification using legacy secret
			if jwtSecret != "" {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
					return []byte(jwtSecret), nil
				}
			}

			return nil, fmt.Errorf("no matching key found for verification (alg: %v)", token.Header["alg"])
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid or expired token: %v", err)})
			return
		}

		// Save claims to Gin context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID, _ := claims["sub"].(string)
			email, _ := claims["email"].(string)

			c.Set("userID", userID)
			c.Set("email", email)
		}

		c.Next()
	}
}
