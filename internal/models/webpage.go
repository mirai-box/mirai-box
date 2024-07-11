package models

import (
	"time"

	"github.com/google/uuid"
)

type WebPage struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"-"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Html      string    `gorm:"type:text;not null" json:"html"`
	PageType  string    `gorm:"type:varchar(255);default:'main'" json:"page_type"`
	Public    bool      `gorm:"type:boolean;default:false" json:"public"`
	CreatedAt time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
