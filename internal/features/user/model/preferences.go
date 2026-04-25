package model

import (
	"time"
)

type UserPreferences struct {
	UserID        string    `gorm:"primaryKey;type:uuid"`
	Notifications *bool     `gorm:"default:true"`
	Language      *string   `gorm:"default:'en'"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}
