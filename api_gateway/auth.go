package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(jwtSecretString)

func parseTokenGateway(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrTokenMalformed
			}
			return jwtSecret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	// доп. защита от устаревших токенов
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, jwt.ErrTokenExpired
	}

	return claims, nil
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fail(c, http.StatusUnauthorized, "AUTH_REQUIRED", "Missing Authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			fail(c, http.StatusUnauthorized, "AUTH_REQUIRED", "Invalid Authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := parseTokenGateway(tokenString)
		if err != nil {
			fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired")
			c.Abort()
			return
		}

		// можно прокинуть userId/roles дальше, если понадобится
		c.Set("userId", claims.UserID)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}
