package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	SessionUserIDKey  = "user_id"
	SessionCookieName = "session-name"
)

type LoginRequest struct {
	Username     string `json:"username"      validate:"required"`
	Password     string `json:"password"      validate:"required"`
	KeepSignedIn bool   `json:"keepSignedIn"`
}

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Username  string    `gorm:"type:varchar(255);unique;not null" json:"username"`
	Password  string    `gorm:"type:varchar(255);not null" json:"password"`
	Role      string    `gorm:"type:varchar(50);not null" json:"role"`
	CreatedAt time.Time `gorm:"type:timestamp;default:now()" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:now()" json:"-"`
}
