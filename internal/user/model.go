package user

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null;default:'user'"`
}

type Role string

const (
	RoleUser  Role = "user"
	RoleMod   Role = "moderator"
	RoleAdmin Role = "admin"
)
