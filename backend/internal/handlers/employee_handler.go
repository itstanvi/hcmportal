package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"hcm-backend/internal/dto"
	"hcm-backend/internal/middleware"
	"hcm-backend/internal/models"
	"hcm-backend/internal/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeHandler struct {
	DB *gorm.DB
}

func NewEmployeeHandler(db *gorm.DB) *EmployeeHandler {
	return &EmployeeHandler{DB: db}
}

// ---------- helpers ----------

func (h *EmployeeHandler) loadActiveFieldsForTenant(tenantID uuid.UUID) ([]models.EmployeeCustomField, error) {
	var fields []models.EmployeeCustomField
	err := h.DB.Where("tenant_id = ? AND active = ?", tenantID, true).
		Order("display_order ASC, created_at ASC").
		Find(&fields).Error
	return fields, err
}

func toEmployeeResponse(emp models.Employee, fieldsByID map[uuid.UUID]models.EmployeeCustomField) dto.EmployeeResponse {
	custom := map[string]interface{}{}
	for _, v := range emp.CustomValues {
		f, ok := fieldsByID[v.FieldID]
		if !ok {
			continue
		}
		custom[f.FieldKey] = parseStoredValue(f.FieldType, v.Value)
	}
	return dto.EmployeeResponse{
		ID:             emp.ID,
		Name:           emp.Name,
		Email:          emp.Email,
		Phone:          emp.Phone,
		EmployeeCode:   emp.EmployeeCode,
		Department:     emp.Department,
		Designation:    emp.Designation,
		DateOfJoining:  emp.DateOfJoining,
		EmploymentType: emp.EmploymentType,
		Status:         emp.Status,
		CustomFields:   custom,
		CreatedAt:      emp.CreatedAt,
		UpdatedAt:      emp.UpdatedAt,
	}
}

// parseStoredValue converts the stored string back to a typed JSON-friendly value.
func parseStoredValue(fieldType, raw string) interface{} {
	if raw == "" {
		return nil
	}
	switch fieldType {
	case models.FieldTypeNumber:
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			return f
		}
	case models.FieldTypeBoolean:
		return raw == "true"
	}
	return raw
}

// validateAndCollect validates the supplied custom_fields against the active field config.
// It returns the rows to upsert and rejects unknown keys.
func validateAndCollect(
	tenantID uuid.UUID,
	employeeID uuid.UUID,
	supplied map[string]interface{},
	activeFields []models.EmployeeCustomField,
) ([]models.EmployeeCustomFieldValue, error) {
	if supplied == nil {
		supplied = map[string]interface{}{}
	}

	byKey := map[string]models.EmployeeCustomField{}
	for _, f := range activeFields {
		byKey[f.FieldKey] = f
	}

	// Reject unknown keys.
	for k := range supplied {
		if _, ok := byKey[k]; !ok {
			return nil, fmt.Errorf("unknown field: %s", k)
		}
	}

	var rows []models.EmployeeCustomFieldValue
	for _, f := range activeFields {
		raw, present := supplied[f.FieldKey]
		if !present {
			if f.Required {
				return nil, fmt.Errorf("%s is required", f.FieldName)
			}
			continue
		}
		var opts []string
		if len(f.Options) > 0 {
			_ = json.Unmarshal(f.Options, &opts)
		}
		val, err := validator.NormalizeAndValidateCustomValue(f, raw, opts)
		if err != nil {
			return nil, err
		}
		if val == "" {
			continue
		}
		rows = append(rows, models.EmployeeCustomFieldValue{
			ID:         uuid.New(),
			TenantID:   tenantID,
			EmployeeID: employeeID,
			FieldID:    f.ID,
			Value:      val,
		})
	}
	return rows, nil
}

// ---------- handlers ----------

// List GET /api/employees?page=1&page_size=20&search=&department=&status=
func (h *EmployeeHandler) List(c *gin.Context) {
	tenantID := middleware.TenantID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	q := h.DB.Model(&models.Employee{}).Where("tenant_id = ?", tenantID)

	if s := strings.TrimSpace(c.Query("search")); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(employee_code) LIKE ?", like, like, like)
	}
	if dept := strings.TrimSpace(c.Query("department")); dept != "" {
		q = q.Where("department = ?", dept)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		q = q.Where("status = ?", status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var employees []models.Employee
	if err := q.Preload("CustomValues").
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load ALL fields (active + inactive) so we can still surface values
	// belonging to deactivated fields without losing data.
	var allFields []models.EmployeeCustomField
	h.DB.Where("tenant_id = ?", tenantID).Find(&allFields)
	fieldsByID := map[uuid.UUID]models.EmployeeCustomField{}
	for _, f := range allFields {
		fieldsByID[f.ID] = f
	}

	items := make([]dto.EmployeeResponse, 0, len(employees))
	for _, e := range employees {
		items = append(items, toEmployeeResponse(e, fieldsByID))
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	c.JSON(http.StatusOK, dto.ListEmployeesResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// Get GET /api/employees/:id
func (h *EmployeeHandler) Get(c *gin.Context) {
	tenantID := middleware.TenantID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var emp models.Employee
	if err := h.DB.Preload("CustomValues").
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}
	var allFields []models.EmployeeCustomField
	h.DB.Where("tenant_id = ?", tenantID).Find(&allFields)
	fieldsByID := map[uuid.UUID]models.EmployeeCustomField{}
	for _, f := range allFields {
		fieldsByID[f.ID] = f
	}
	c.JSON(http.StatusOK, toEmployeeResponse(emp, fieldsByID))
}

// Create POST /api/employees
func (h *EmployeeHandler) Create(c *gin.Context) {
	tenantID := middleware.TenantID(c)

	var req dto.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidatePhone(req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateEmployeeStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = models.EmployeeStatusActive
	}

	activeFields, err := h.loadActiveFieldsForTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	emp := models.Employee{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Name:           req.Name,
		Email:          strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:          req.Phone,
		EmployeeCode:   strings.TrimSpace(req.EmployeeCode),
		Department:     req.Department,
		Designation:    req.Designation,
		DateOfJoining:  req.DateOfJoining,
		EmploymentType: req.EmploymentType,
		Status:         req.Status,
	}

	customRows, err := validateAndCollect(tenantID, emp.ID, req.CustomFields, activeFields)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&emp).Error; err != nil {
			return err
		}
		if len(customRows) > 0 {
			if err := tx.Create(&customRows).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		// Friendly message for unique conflicts.
		if isUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "email or employee_code already exists for this tenant"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload to return fresh data with custom values
	h.DB.Preload("CustomValues").First(&emp, "id = ?", emp.ID)
	var allFields []models.EmployeeCustomField
	h.DB.Where("tenant_id = ?", tenantID).Find(&allFields)
	fieldsByID := map[uuid.UUID]models.EmployeeCustomField{}
	for _, f := range allFields {
		fieldsByID[f.ID] = f
	}
	c.JSON(http.StatusCreated, toEmployeeResponse(emp, fieldsByID))
}

// Update PUT /api/employees/:id
func (h *EmployeeHandler) Update(c *gin.Context) {
	tenantID := middleware.TenantID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var emp models.Employee
	if err := h.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}

	var req dto.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		emp.Name = *req.Name
	}
	if req.Email != nil {
		if err := validator.ValidateEmail(*req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		emp.Email = strings.ToLower(strings.TrimSpace(*req.Email))
	}
	if req.Phone != nil {
		if err := validator.ValidatePhone(*req.Phone); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		emp.Phone = *req.Phone
	}
	if req.EmployeeCode != nil {
		emp.EmployeeCode = strings.TrimSpace(*req.EmployeeCode)
	}
	if req.Department != nil {
		emp.Department = *req.Department
	}
	if req.Designation != nil {
		emp.Designation = *req.Designation
	}
	if req.DateOfJoining != nil {
		emp.DateOfJoining = req.DateOfJoining
	}
	if req.EmploymentType != nil {
		emp.EmploymentType = *req.EmploymentType
	}
	if req.Status != nil {
		if err := validator.ValidateEmployeeStatus(*req.Status); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		emp.Status = *req.Status
	}

	activeFields, err := h.loadActiveFieldsForTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// custom_fields: only replace values for active fields supplied.
	// Unknown keys (not in active config) are rejected.
	var customRows []models.EmployeeCustomFieldValue
	if req.CustomFields != nil {
		customRows, err = validateAndCollect(tenantID, emp.ID, req.CustomFields, activeFields)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&emp).Error; err != nil {
			return err
		}
		if req.CustomFields != nil {
			// Only delete values for fields that the request actually carried, plus active fields that are required.
			// Simpler & safer: delete values for all currently-active fields, then insert the supplied set.
			// Inactive-field values stay untouched (data preserved).
			activeIDs := make([]uuid.UUID, 0, len(activeFields))
			for _, f := range activeFields {
				activeIDs = append(activeIDs, f.ID)
			}
			if len(activeIDs) > 0 {
				if err := tx.Where("tenant_id = ? AND employee_id = ? AND field_id IN ?", tenantID, emp.ID, activeIDs).
					Delete(&models.EmployeeCustomFieldValue{}).Error; err != nil {
					return err
				}
			}
			if len(customRows) > 0 {
				if err := tx.Create(&customRows).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "email or employee_code already exists for this tenant"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.DB.Preload("CustomValues").First(&emp, "id = ?", emp.ID)
	var allFields []models.EmployeeCustomField
	h.DB.Where("tenant_id = ?", tenantID).Find(&allFields)
	fieldsByID := map[uuid.UUID]models.EmployeeCustomField{}
	for _, f := range allFields {
		fieldsByID[f.ID] = f
	}
	c.JSON(http.StatusOK, toEmployeeResponse(emp, fieldsByID))
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
