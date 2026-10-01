package model

import "time"

// Project groups tasks under a single owner.
type Project struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	OwnerID     int64     `db:"owner_id" json:"ownerId"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"-"`
}
