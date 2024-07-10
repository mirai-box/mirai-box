package models

import (
	"time"

	"github.com/google/uuid"
)

type Sale struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	ArtProjectID uuid.UUID  `gorm:"type:uuid;not null" json:"art_project_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Price        float64    `gorm:"type:numeric(10,2);not null" json:"price"`
	SoldAt       time.Time  `gorm:"type:timestamp;default:now()" json:"sold_at"`
	ArtProject   ArtProject `gorm:"foreignKey:ArtProjectID;constraint:OnDelete:CASCADE" json:"-"`
	User         User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
