package model

import (
	"github.com/google/uuid"
)

type StorageUsage struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	UsedSpace int64     `gorm:"type:bigint;default:0" json:"used_space"`
	Quota     int64     `gorm:"type:bigint;default:104857600" json:"quota"` // Default 100 MB quota
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
}
