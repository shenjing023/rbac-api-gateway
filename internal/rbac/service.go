package rbac

import (
	"context"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateRole(name, description string) error {
	role := Role{Name: name, Description: description}
	return s.db.Create(&role).Error
}

func (s *Service) AssignRoleToUser(userID uint, roleName string) error {
	var role Role
	if err := s.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}

	userRole := UserRole{UserID: userID, RoleID: role.ID}
	return s.db.Create(&userRole).Error
}

func (s *Service) CreatePermission(name, description string) error {
	permission := Permission{Name: name, Description: description}
	return s.db.Create(&permission).Error
}

func (s *Service) AssignPermissionToRole(roleName, permissionName string) error {
	var role Role
	if err := s.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}

	var permission Permission
	if err := s.db.Where("name = ?", permissionName).First(&permission).Error; err != nil {
		return err
	}

	return s.db.Model(&role).Association("Permissions").Append(&permission)
}

func (s *Service) CheckUserPermission(userID uint, permissionName string) (bool, error) {
	var userRole UserRole
	if err := s.db.Where("user_id = ?", userID).First(&userRole).Error; err != nil {
		return false, err
	}

	var role Role
	if err := s.db.Preload("Permissions").First(&role, userRole.RoleID).Error; err != nil {
		return false, err
	}

	for _, permission := range role.Permissions {
		if permission.Name == permissionName {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) GetUserRole(userID uint) (string, error) {
	var userRole UserRole
	if err := s.db.Where("user_id = ?", userID).First(&userRole).Error; err != nil {
		return "", err
	}

	var role Role
	if err := s.db.First(&role, userRole.RoleID).Error; err != nil {
		return "", err
	}

	return role.Name, nil
}

type ResourceChecker interface {
	CheckResourceOwnership(ctx context.Context, resourceID string, userID uint) (bool, error)
}

type PermissionChecker struct {
	resourceCheckers sync.Map
}

func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{}
}

func (pc *PermissionChecker) RegisterResourceChecker(resourceType string, checker ResourceChecker) {
	pc.resourceCheckers.Store(resourceType, checker)
}

func (pc *PermissionChecker) CheckPermission(c *gin.Context, input *PermissionInput) (bool, error) {
	if checkerValue, ok := pc.resourceCheckers.Load(input.Resource.Type); ok {
		checker := checkerValue.(ResourceChecker)
		isOwner, err := checker.CheckResourceOwnership(c.Request.Context(), input.Resource.ID, input.User.ID)
		if err != nil {
			return false, fmt.Errorf("error checking resource ownership: %w", err)
		}
		input.Resource.IsOwner = isOwner
	}

	// 这里调用 OPA 进行权限评估
	allowed, err := evaluateOPAPolicy(input)
	if err != nil {
		return false, err
	}

	return allowed, nil
}

type PermissionInput struct {
	Action   string `json:"action"`
	Resource struct {
		Type    string `json:"type"`
		ID      string `json:"id"`
		IsOwner bool   `json:"is_owner,omitempty"`
	} `json:"resource"`
	User struct {
		ID   uint   `json:"id"`
		Role string `json:"role"`
	} `json:"user"`
}
