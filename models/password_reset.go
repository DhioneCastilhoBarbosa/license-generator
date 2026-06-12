package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordResetToken struct {
	gorm.Model
	UsuarioID uint      `gorm:"index;not null"`
	TokenHash string    `gorm:"size:64;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
}
