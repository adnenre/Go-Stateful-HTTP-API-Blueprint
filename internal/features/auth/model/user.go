package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username      string         `gorm:"not null"`
	FirstName     string         `gorm:"default:''"`
	LastName      string         `gorm:"default:''"`
	Email         string         `gorm:"uniqueIndex;not null"`
	Password      string         `gorm:"not null"` // hashed
	EmailVerified bool           `gorm:"default:false"`
	Status        string         `gorm:"default:'pending'"`
	Role          string         `gorm:"default:'user'"`
	Avatar        *string        `gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}
