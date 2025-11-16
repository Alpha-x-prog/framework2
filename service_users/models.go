package main

import (
	"database/sql"
	"strings"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Roles        []string  `json:"roles"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func rolesToString(roles []string) string {
	return strings.Join(roles, ",")
}

func rolesFromString(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

func insertUser(u *User) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	_, err := db.Exec(
		`INSERT INTO users (id, email, name, password_hash, roles, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.Name, u.PasswordHash, rolesToString(u.Roles), u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func getUserByEmail(email string) (*User, error) {
	row := db.QueryRow(
		`SELECT id, email, name, password_hash, roles, created_at, updated_at
		 FROM users WHERE email = ?`,
		email,
	)

	var u User
	var rolesStr string

	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &rolesStr, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	u.Roles = rolesFromString(rolesStr)
	return &u, nil
}

func getUserByID(id string) (*User, error) {
	row := db.QueryRow(
		`SELECT id, email, name, password_hash, roles, created_at, updated_at
		 FROM users WHERE id = ?`,
		id,
	)

	var u User
	var rolesStr string

	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &rolesStr, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	u.Roles = rolesFromString(rolesStr)
	return &u, nil
}
