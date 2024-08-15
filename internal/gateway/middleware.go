package gateway

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/shenjing023/rbac-api-gateway/internal/rbac"
	"github.com/shenjing023/rbac-api-gateway/pkg/jwt"
)

// CORSMiddleware 返回一个 CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有来源，您可以根据需要限制特定域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 排除不需要认证的路由
		if isExcludedPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证token"})
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func isExcludedPath(path string) bool {
	log.Printf("path: %v\n", path)
	excludedPaths := []string{
		"/auth/register",
		"/auth/login",
		"/posts",
		"/posts/:id",
		// 可以添加其他不需要认证的路径
	}

	for _, excludedPath := range excludedPaths {
		if path == excludedPath {
			return true
		}
	}
	return false
}

func RBACMiddleware(permissionChecker *rbac.PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 排除不需要认证的路由
		if isExcludedPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")

		input := &rbac.PermissionInput{
			Action: c.Request.Method + ":" + c.FullPath(),
			Resource: struct {
				Type    string `json:"type"`
				ID      string `json:"id"`
				IsOwner bool   `json:"is_owner,omitempty"`
			}{
				Type: getResourceTypeFromPath(c.Request.URL.Path),
				ID:   getParamOrDefault(c, "id", "0"),
			},
			User: struct {
				ID   uint   `json:"id"`
				Role string `json:"role"`
			}{
				ID:   userID.(uint),
				Role: role.(string),
			},
		}

		log.Printf("input: %+v\n", input)

		allowed, err := permissionChecker.CheckPermission(c, input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "权限检查失败"})
			log.Printf("err: %v\n", err)
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "没有权限执行此操作"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getParamOrDefault(c *gin.Context, key, defaultValue string) string {
	value := c.Param(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getResourceTypeFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		return parts[1] // 假设路径格式为 "/posts/:id"
	}
	return ""
}
