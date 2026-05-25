package handlers

import (
	"encoding/json"
	"net/http"

	"hcm-backend/internal/dto"
	"hcm-backend/internal/middleware"
	"hcm-backend/internal/models"
	"hcm-backend/internal/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FieldHandler struct {
	DB *gorm.DB
}

func NewFieldHandler(db *gorm.DB) *FieldHandler {
	return &FieldHandler{DB: db}
}

// List GET /api/employee-fields?active_only=true
// Both ADMIN and HR can read; HR only needs them to render forms.
func (h *FieldHandler) List(c *gin.Context) {
	tenantID := middleware.TenantID(c)
	activeOnly := c.Query("active_only") == "true"

	var fields []models.EmployeeCustomField
	q := h.DB.Where("tenant_id = ?", tenantID).Order("display_order ASC, created_at ASC")
	if activeOnly {
		q = q.Where("active = ?", true)
	}
	if err := q.Find(&fields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fields)
}

// Create POST /api/employee-fields  (ADMIN only)
func (h *FieldHandler) Create(c *gin.Context) {
	tenantID := middleware.TenantID(c)

	var req dto.CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateFieldKey(req.FieldKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateFieldType(req.FieldType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.FieldType == models.FieldTypeDropdown && len(req.Options) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dropdown fields must have at least one option"})
		return
	}

	// Reject duplicates within the tenant.
	var existing int64
	h.DB.Model(&models.EmployeeCustomField{}).
		Where("tenant_id = ? AND field_key = ?", tenantID, req.FieldKey).
		Count(&existing)
	if existing > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "field_key already exists"})
		return
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	var optionsJSON datatypes.JSON
	if len(req.Options) > 0 {
		b, _ := json.Marshal(req.Options)
		optionsJSON = b
	}

	field := models.EmployeeCustomField{
		ID:           uuid.New(),
		TenantID:     tenantID,
		FieldName:    req.FieldName,
		FieldKey:     req.FieldKey,
		FieldType:    req.FieldType,
		Required:     req.Required,
		Active:       active,
		Options:      optionsJSON,
		DisplayOrder: req.DisplayOrder,
	}

	if err := h.DB.Create(&field).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, field)
}

// Update PUT /api/employee-fields/:id  (ADMIN only)
// Note: field_key is intentionally NOT mutable to keep stored values consistent.
func (h *FieldHandler) Update(c *gin.Context) {
	tenantID := middleware.TenantID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var field models.EmployeeCustomField
	if err := h.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&field).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "field not found"})
		return
	}

	var req dto.UpdateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.FieldName != nil {
		field.FieldName = *req.FieldName
	}
	if req.FieldType != nil {
		if err := validator.ValidateFieldType(*req.FieldType); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		field.FieldType = *req.FieldType
	}
	if req.Required != nil {
		field.Required = *req.Required
	}
	if req.Active != nil {
		field.Active = *req.Active
	}
	if req.DisplayOrder != nil {
		field.DisplayOrder = *req.DisplayOrder
	}
	if req.Options != nil {
		b, _ := json.Marshal(req.Options)
		field.Options = b
	}
	if field.FieldType == models.FieldTypeDropdown {
		var opts []string
		if len(field.Options) > 0 {
			_ = json.Unmarshal(field.Options, &opts)
		}
		if len(opts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dropdown fields must have at least one option"})
			return
		}
	}

	if err := h.DB.Save(&field).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, field)
}
