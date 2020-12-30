package models

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user
type User struct {
	ID        uuid.UUID
	Username  string
	Password  []byte
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Create a user in the database
func (u *User) Create() error {
	sql := ` -- name: UserCreate :one
		INSERT INTO users
		(id, username, password, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
		RETURNING id, username, password, role, created_at, updated_at
		;`

	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	err = DB.QueryRow(context.Background(), sql, uuid, u.Username, u.Password, u.Role).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// Find a user in the database
func (u *User) Find(id uuid.UUID) error {
	sql := `SELECT id, username, role, created_at, updated_at FROM users WHERE id=$1;`
	err := DB.QueryRow(context.Background(), sql, id).Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// FindByUsername finds a user by username
func (u *User) FindByUsername(username string) error {
	sql := `SELECT id, username, role, created_at, updated_at FROM users WHERE username=$1;`
	err := DB.QueryRow(context.Background(), sql, username).Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// Authenticate compares a password to the hashed password in the database.
func (u *User) Authenticate(password string) error {
	sql := `SELECT password FROM users WHERE id=$1`
	var passwordHash []byte
	err := DB.QueryRow(context.Background(), sql, u.ID).Scan(&passwordHash)
	if err != nil {
		return err
	}

	if password == "" {
		return errors.New("Empty password")
	}

	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(password)); err != nil {
		return err
	}

	return nil
}
