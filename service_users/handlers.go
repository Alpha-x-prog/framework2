package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// POST /v1/users/register
func handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	existing, err := getUserByEmail(req.Email)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to query user")
		return
	}
	if existing != nil {
		fail(c, http.StatusConflict, "EMAIL_TAKEN", "User with this email already exists")
		return
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		fail(c, http.StatusInternalServerError, "HASH_ERROR", "Failed to hash password")
		return
	}

	user := &User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: passwordHash,
		Roles:        []string{"user"},
	}

	if err := insertUser(user); err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to save user")
		return
	}

	resp := gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"name":      user.Name,
		"roles":     user.Roles,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
	}
	success(c, resp)
}

// POST /v1/users/login
func handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	user, err := getUserByEmail(req.Email)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to query user")
		return
	}
	if user == nil || !checkPassword(user.PasswordHash, req.Password) {
		fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email or password is incorrect")
		return
	}

	token, err := generateToken(user)
	if err != nil {
		fail(c, http.StatusInternalServerError, "TOKEN_ERROR", "Failed to generate token")
		return
	}

	success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"roles": user.Roles,
		},
	})
}

// GET /v1/users/me
func handleMe(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		fail(c, http.StatusInternalServerError, "CONTEXT_ERROR", "User ID missing in context")
		return
	}
	userID, _ := userIDVal.(string)

	user, err := getUserByID(userID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to query user")
		return
	}
	if user == nil {
		fail(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	success(c, gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"name":      user.Name,
		"roles":     user.Roles,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
	})
}
