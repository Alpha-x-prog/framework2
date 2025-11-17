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

// число админов
func getAdminsCount() (int, error) {
	row := db.QueryRow(`SELECT COUNT(*) FROM users WHERE roles LIKE '%admin%'`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// обновление профиля (сейчас только name)
func updateUserProfile(id string, name string) (*User, error) {
	now := time.Now()

	_, err := db.Exec(
		`UPDATE users SET name = ?, updated_at = ? WHERE id = ?`,
		name, now, id,
	)
	if err != nil {
		return nil, err
	}

	return getUserByID(id)
}

// список пользователей с фильтрами и пагинацией
func getUsersCountFiltered(email, role string) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE 1=1`
	var args []any

	if email != "" {
		query += ` AND email LIKE ?`
		args = append(args, "%"+email+"%")
	}
	if role != "" {
		query += ` AND roles LIKE ?`
		args = append(args, "%"+role+"%")
	}

	row := db.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func listUsersFiltered(email, role string, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, email, name, password_hash, roles, created_at, updated_at
		FROM users
		WHERE 1=1
	`
	var args []any

	if email != "" {
		query += ` AND email LIKE ?`
		args = append(args, "%"+email+"%")
	}
	if role != "" {
		query += ` AND roles LIKE ?`
		args = append(args, "%"+role+"%")
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		var u User
		var rolesStr string

		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &rolesStr, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Roles = rolesFromString(rolesStr)
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
