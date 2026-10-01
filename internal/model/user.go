package model

import "time"

// User is an account that can authenticate and own/be assigned to work.
type User struct {
	ID           int64     `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"-"`
}
