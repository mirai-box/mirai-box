package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	ArtProjectID uuid.UUID  `gorm:"type:uuid;not null" json:"art_project_id"`
	Content      string     `gorm:"type:text;not null" json:"content"`
	CreatedAt    time.Time  `gorm:"type:timestamp;default:now()" json:"created_at"`
	User         User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	ArtProject   ArtProject `gorm:"foreignKey:ArtProjectID;constraint:OnDelete:CASCADE" json:"-"`
}
