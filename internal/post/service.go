package post

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shenjing023/rbac-api-gateway/pkg/cache"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreatePost(title, content string, authorID uint) error {
	post := Post{
		Title:    title,
		Content:  content,
		AuthorID: authorID,
	}
	return s.db.Create(&post).Error
}

func (s *Service) GetPost(id uint) (*Post, error) {
	var post Post
	if err := s.db.First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (s *Service) UpdatePost(id uint, title, content string) error {
	err := s.db.Model(&Post{}).Where("id = ?", id).Updates(map[string]interface{}{
		"title":   title,
		"content": content,
	}).Error
	if err == nil {
		go s.updateCache(id)
	}
	return err
}

func (s *Service) DeletePost(id uint) error {
	err := s.db.Delete(&Post{}, id).Error
	if err == nil {
		go s.deleteCache(id)
	}
	return err
}

func (s *Service) updateCache(id uint) {
	// 实现更新缓存的逻辑
}

func (s *Service) deleteCache(id uint) {
	cache := cache.GetInstance()
	cacheKey := fmt.Sprintf("post:%d:author", id)
	cache.Delete(cacheKey)
}

func (s *Service) ListPosts(page, pageSize int) ([]Post, error) {
	var posts []Post
	offset := (page - 1) * pageSize
	if err := s.db.Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

type PostChecker struct {
	service *Service
	cache   *cache.Cache
}

func NewPostChecker(service *Service, cache *cache.Cache) *PostChecker {
	return &PostChecker{service: service, cache: cache}
}

func (pc *PostChecker) CheckResourceOwnership(ctx context.Context, resourceID string, userID uint) (bool, error) {
	postID, err := strconv.ParseUint(resourceID, 10, 32)
	if err != nil {
		return false, err
	}

	cacheKey := fmt.Sprintf("post:%d:author", postID)
	cachedAuthorID, found := pc.cache.Get(cacheKey)
	if found {
		return cachedAuthorID.(uint) == userID, nil
	}

	post, err := pc.service.GetPost(uint(postID))
	if err != nil {
		return false, err
	}
	if post == nil {
		return false, nil
	}

	pc.cache.Set(cacheKey, post.AuthorID, 5*time.Minute)
	return post.AuthorID == userID, nil
}
