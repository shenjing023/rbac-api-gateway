package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/shenjing023/rbac-api-gateway/internal/auth"
	"github.com/shenjing023/rbac-api-gateway/internal/gateway"
	"github.com/shenjing023/rbac-api-gateway/internal/post"
	"github.com/shenjing023/rbac-api-gateway/internal/rbac"
	"github.com/shenjing023/rbac-api-gateway/internal/user"
	"github.com/shenjing023/rbac-api-gateway/pkg/cache"
	"github.com/shenjing023/rbac-api-gateway/pkg/database"
)

func main() {
	// 初始化数据库连接
	db, err := database.InitDB("127.0.0.1", "postgres", "123456", "postgres", 5432)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// 数据库迁移
	err = db.AutoMigrate(&user.User{}, &rbac.Role{}, &rbac.Permission{}, &rbac.UserRole{}, &post.Post{})
	if err != nil {
		log.Fatalf("Failed to perform database migration: %v", err)
	}

	// 初始化 OPA
	if err := rbac.InitOPA(); err != nil {
		log.Fatalf("Failed to initialize OPA: %v", err)
	}

	// 初始化路由
	r := gin.Default()

	// 初始化服务
	authService := auth.NewService(db)
	userService := user.NewService(db)
	rbacService := rbac.NewService(db)
	// postService := post.NewService(db)

	permissionChecker := rbac.NewPermissionChecker()

	postService := post.NewService(db)
	postChecker := post.NewPostChecker(postService, cache.GetInstance())
	permissionChecker.RegisterResourceChecker("posts", postChecker)

	// 添加网关中间件
	r.Use(gateway.CORSMiddleware())
	r.Use(gateway.AuthMiddleware())
	r.Use(gateway.RBACMiddleware(permissionChecker))

	// 设置路由
	auth.RegisterRoutes(r, authService)
	user.RegisterRoutes(r, userService)
	rbac.RegisterRoutes(r, rbacService)
	post.RegisterRoutes(r, postService)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
