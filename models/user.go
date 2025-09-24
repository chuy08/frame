package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the database
type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
