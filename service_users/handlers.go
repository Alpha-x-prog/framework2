package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"required"` // engineer / manager / director / customer
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// POST /v1/users/register
// POST /v1/users/register
func handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// проверим, что роль валидная
	baseRole, ok := normalizeRole(req.Role)
	if !ok {
		fail(c, http.StatusBadRequest, "INVALID_ROLE",
			"Role must be one of: engineer, manager, director, customer")
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

	// считаем админов
	adminCount, err := getAdminsCount()
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to count admins")
		return
	}

	roles := []string{baseRole}
	if adminCount == 0 {
		// первый админ в системе — добавляем техническую роль admin
		roles = append(roles, "admin")
	}

	user := &User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: passwordHash,
		Roles:        roles,
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

// GET /v1/users (только admin)
func handleGetUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	total, err := getUsersCount()
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to count users")
		return
	}

	users, err := listUsers(limit, offset)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to list users")
		return
	}

	items := make([]gin.H, 0, len(users))
	for _, u := range users {
		items = append(items, gin.H{
			"id":        u.ID,
			"email":     u.Email,
			"name":      u.Name,
			"roles":     u.Roles,
			"createdAt": u.CreatedAt,
			"updatedAt": u.UpdatedAt,
		})
	}

	success(c, gin.H{
		"items": items,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func normalizeRole(role string) (string, bool) {
	switch role {
	case "engineer", "manager", "director", "customer", "admin":
		return role, true
	default:
		return "", false
	}
}
