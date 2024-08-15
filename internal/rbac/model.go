package rbac

import (
	"gorm.io/gorm"
)

type Permission struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
}

type Role struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

type UserRole struct {
	gorm.Model
	UserID uint `gorm:"uniqueIndex:idx_user_role"`
	RoleID uint `gorm:"uniqueIndex:idx_user_role"`
}
