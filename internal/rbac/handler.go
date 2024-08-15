package rbac

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateRole(req.Name, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建角色失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "角色创建成功"})
}

func (h *Handler) AssignRoleToUser(c *gin.Context) {
	var req struct {
		UserID uint   `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AssignRoleToUser(req.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分配角色失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色分配成功"})
}

func (h *Handler) CreatePermission(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreatePermission(req.Name, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建权限失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "权限创建成功"})
}

func (h *Handler) AssignPermissionToRole(c *gin.Context) {
	var req struct {
		Role       string `json:"role" binding:"required"`
		Permission string `json:"permission" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AssignPermissionToRole(req.Role, req.Permission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分配权限失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限分配成功"})
}

func (h *Handler) CheckUserPermission(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	permission := c.Query("permission")

	if userID == 0 || permission == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID或权限"})
		return
	}

	hasPermission, err := h.service.CheckUserPermission(uint(userID), permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_permission": hasPermission})
}

func RegisterRoutes(r *gin.Engine, service *Service) {
	handler := NewHandler(service)

	rbac := r.Group("/rbac")
	{
		rbac.POST("/roles", handler.CreateRole)
		rbac.POST("/assign-role", handler.AssignRoleToUser)
		rbac.POST("/permissions", handler.CreatePermission)
		rbac.POST("/assign-permission", handler.AssignPermissionToRole)
		rbac.GET("/check-permission", handler.CheckUserPermission)
	}
}
