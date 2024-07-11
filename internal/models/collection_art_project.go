package models

import (
	"github.com/google/uuid"
)

type CollectionArtProject struct {
	CollectionID uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"collection_id"`
	ArtProjectID uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"art_project_id"`
	Collection   Collection `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE" json:"collection"`
	ArtProject   ArtProject `gorm:"foreignKey:ArtProjectID;constraint:OnDelete:CASCADE" json:"art_project"`
}
