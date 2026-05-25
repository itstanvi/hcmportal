package handlers

import (
	"net/http"

	"hcm-backend/internal/dto"
	"hcm-backend/internal/middleware"
	"hcm-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MeHandler struct {
	DB *gorm.DB
}

func NewMeHandler(db *gorm.DB) *MeHandler {
	return &MeHandler{DB: db}
}

// Me GET /api/me — returns the authenticated user info
func (h *MeHandler) Me(c *gin.Context) {
	userID := middleware.UserID(c)
	tenantID := middleware.TenantID(c)

	var user models.User
	if err := h.DB.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, dto.UserInfo{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     user.Role,
		TenantID: user.TenantID,
	})
}
