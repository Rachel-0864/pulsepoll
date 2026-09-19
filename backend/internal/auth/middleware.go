package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth protects routes that require a valid JWT.
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c)

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization token is required",
			})
			return
		}

		userID, err := getUserIDFromToken(tokenString, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// OptionalAuth attempts to authenticate the request when a JWT is present.
//
// Unlike RequireAuth, this middleware does NOT reject anonymous requests.
// This is useful for public endpoints such as voting:
//
//   - logged-in users get a userID
//   - anonymous users continue without a userID
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c)

		if tokenString != "" {
			if userID, err := getUserIDFromToken(tokenString, jwtSecret); err == nil {
				c.Set("userID", userID)
			}
		}

		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	header := strings.TrimSpace(c.GetHeader("Authorization"))

	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)

	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func getUserIDFromToken(tokenString string, jwtSecret string) (string, error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			// Make sure the token was signed using an HMAC algorithm.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(jwtSecret), nil
		},
	)

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}

	userID, ok := claims["sub"].(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", jwt.ErrTokenInvalidClaims
	}

	return userID, nil
}