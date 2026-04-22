package model

import "time"

// HealthEntity is a placeholder – health has no persistent data.
// This file exists for consistency; real features would define GORM models.
type HealthEntity struct {
	ID        uint      `gorm:"primaryKey"`
	Status    string    `gorm:"not null"`
	CheckedAt time.Time `gorm:"autoCreateTime"`
}
