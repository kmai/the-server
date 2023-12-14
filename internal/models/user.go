package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID string `gorm:"primaryKey" json:"id"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"     json:"deletedAt"`
}
