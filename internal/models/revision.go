package models

import (
	"time"

	"github.com/google/uuid"
)

type Revision struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	ArtID        uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4()" json:"art_id"`
	ArtProjectID uuid.UUID  `gorm:"type:uuid;not null" json:"art_project_id"`
	Version      int        `gorm:"type:int" json:"version"`
	FilePath     string     `gorm:"type:varchar(255);not null" json:"file_path"`
	CreatedAt    time.Time  `gorm:"type:timestamp;default:now()" json:"created_at"`
	Comment      string     `gorm:"type:text" json:"comment"`
	Size         int64      `gorm:"type:bigint;not null;default:0" json:"size"`
	ArtProject   ArtProject `gorm:"foreignKey:ArtProjectID;constraint:OnDelete:CASCADE" json:"-"`
}
