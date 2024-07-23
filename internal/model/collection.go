package model

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	CollectionID string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"collection_id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Title        string    `gorm:"type:varchar(255);not null" json:"title"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`
	User         User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
