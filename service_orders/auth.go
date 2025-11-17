package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(jwtSecretString)

func parseToken(tokenString string) (*UserClaims, error) {
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

	return claims, nil
}

func AuthRequired() gin.HandlerFunc {
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

		claims, err := parseToken(parts[1])
		if err != nil {
			fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired")
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

func getUserID(c *gin.Context) (string, bool) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		return "", false
	}
	userID, ok := userIDVal.(string)
	return userID, ok
}

func getRoles(c *gin.Context) []string {
	rolesVal, ok := c.Get("roles")
	if !ok {
		return nil
	}
	if roles, ok := rolesVal.([]string); ok {
		return roles
	}
	return nil
}

func hasAdminRole(c *gin.Context) bool {
	for _, r := range getRoles(c) {
		if r == "admin" {
			return true
		}
	}
	return false
}

func hasRole(c *gin.Context, target string) bool {
	for _, r := range getRoles(c) {
		if r == target {
			return true
		}
	}
	return false
}

func isEngineer(c *gin.Context) bool { return hasRole(c, "engineer") }
func isManager(c *gin.Context) bool  { return hasRole(c, "manager") }
func isDirector(c *gin.Context) bool { return hasRole(c, "director") }
func isCustomer(c *gin.Context) bool { return hasRole(c, "customer") }
