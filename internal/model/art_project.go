package model

import (
	"time"

	"github.com/google/uuid"
)

type ArtProject struct {
	ID                  uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Title               string     `gorm:"type:varchar(255);not null" json:"title"`
	CreatedAt           time.Time  `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"type:timestamp;default:now()" json:"updated_at"`
	ContentType         string     `gorm:"type:varchar(255)" json:"content_type"`
	Filename            string     `gorm:"type:varchar(255)" json:"filename"`
	Public              bool       `gorm:"type:boolean;default:false" json:"public"`
	LatestRevisionID    uuid.UUID  `gorm:"type:uuid" json:"latest_revision_id"`
	PublishedRevisionID *uuid.UUID `gorm:"type:uuid" json:"published_revision_id"`
	Tags                []*Tag     `gorm:"many2many:art_project_tags;" json:"tags"`
	StashID             uuid.UUID  `gorm:"type:uuid;not null" json:"stash_id"`
	Stash               Stash      `gorm:"foreignKey:StashID;constraint:OnDelete:CASCADE" json:"-"`
	UserID              uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User                User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Name        string    `gorm:"type:varchar(255);unique;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
}

type Tag struct {
	ID          uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Name        string        `gorm:"type:varchar(255);unique;not null" json:"name"`
	ArtProjects []*ArtProject `gorm:"many2many:art_project_tags;" json:"art_projects"`
}
