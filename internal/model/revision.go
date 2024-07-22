package model

import (
	"time"

	"github.com/google/uuid"
)

type Revision struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	ArtID        string     `gorm:"type:varchar(255);not null" json:"art_id"`
	Version      int        `gorm:"type:int" json:"version"`
	FilePath     string     `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time  `gorm:"type:timestamp;default:now()" json:"created_at"`
	Comment      string     `gorm:"type:text" json:"comment"`
	Size         int64      `gorm:"type:bigint;not null;default:0" json:"size"`
	ArtProjectID uuid.UUID  `gorm:"type:uuid;not null" json:"art_project_id"`
	ArtProject   ArtProject `gorm:"foreignKey:ArtProjectID;constraint:OnDelete:CASCADE" json:"-"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User         User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

type ArtLink struct {
	Token      string    `gorm:"primaryKey"`
	RevisionID uuid.UUID `gorm:"not null"`
	ExpiresAt  time.Time `gorm:"not null"`
	OneTime    bool      `gorm:"not null;default:false"`
	Used       bool      `gorm:"not null;default:false"`
	Revision   Revision  `gorm:"foreignKey:RevisionID;constraint:OnDelete:CASCADE" json:"-"`
}
