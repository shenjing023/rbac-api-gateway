package auth

import (
	"errors"

	"github.com/shenjing023/rbac-api-gateway/internal/user"
	"github.com/shenjing023/rbac-api-gateway/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Register(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	newUser := user.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     string(user.RoleUser),
	}

	result := s.db.Create(&newUser)
	return result.Error
}

func (s *Service) Login(username, password string) (string, error) {
	var u user.User
	if err := s.db.Where("username = ?", username).First(&u).Error; err != nil {
		return "", errors.New("用户不存在")
	}
	// get user role
	var userRole user.Role
	if err := s.db.Table("user_roles").
		Select("roles.name").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", u.ID).
		Scan(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userRole = user.RoleUser // 如果没有找到角色，默认为普通用户
		} else {
			return "", errors.New("获取用户角色失败")
		}
	}
	u.Role = string(userRole)

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", errors.New("密码错误")
	}

	token, err := jwt.GenerateToken(u.ID, u.Username, u.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) Logout(token string) error {
	// 在实际应用中，你可能需要将token加入黑名单或实现其他登出逻辑
	// 这里我们简单地返回nil，因为JWT本身是无状态的
	return nil
}
