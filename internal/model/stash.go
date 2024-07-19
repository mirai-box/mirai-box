package model

import (
	"time"

	"github.com/google/uuid"
)

type Stash struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ArtProjects uint64    `gorm:"type:bigint;default:0" json:"art_projects"`
	Files       uint64    `gorm:"type:bigint;default:0" json:"files"`
	UsedSpace   int64     `gorm:"type:bigint;default:0" json:"used_space"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`
	User        User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
