package model

import (
	"github.com/google/uuid"
)

type CollectionArtProject struct {
	CollectionID uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"collection_id"`
	RevisionID   uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"revision_id"`
	Collection   Collection `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE" json:"-"`
	Revision     Revision   `gorm:"foreignKey:RevisionID;constraint:OnDelete:CASCADE" json:"-"`
}
