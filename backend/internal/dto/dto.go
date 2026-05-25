package dto

import (
	"time"

	"github.com/google/uuid"
)

// ---------- Auth ----------

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=4"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	TenantID uuid.UUID `json:"tenant_id"`
}

// ---------- Custom Fields ----------

type CreateFieldRequest struct {
	FieldName    string   `json:"field_name" binding:"required,min=1,max=200"`
	FieldKey     string   `json:"field_key" binding:"required,min=1,max=100"`
	FieldType    string   `json:"field_type" binding:"required"`
	Required     bool     `json:"required"`
	Active       *bool    `json:"active"`
	Options      []string `json:"options"`
	DisplayOrder int      `json:"display_order"`
}

type UpdateFieldRequest struct {
	FieldName    *string  `json:"field_name"`
	FieldType    *string  `json:"field_type"`
	Required     *bool    `json:"required"`
	Active       *bool    `json:"active"`
	Options      []string `json:"options"`
	DisplayOrder *int     `json:"display_order"`
}

// ---------- Employees ----------

type CreateEmployeeRequest struct {
	Name           string                 `json:"name" binding:"required"`
	Email          string                 `json:"email" binding:"required,email"`
	Phone          string                 `json:"phone"`
	EmployeeCode   string                 `json:"employee_code" binding:"required"`
	Department     string                 `json:"department"`
	Designation    string                 `json:"designation"`
	DateOfJoining  *time.Time             `json:"date_of_joining"`
	EmploymentType string                 `json:"employment_type"`
	Status         string                 `json:"status"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
}

type UpdateEmployeeRequest struct {
	Name           *string                `json:"name"`
	Email          *string                `json:"email"`
	Phone          *string                `json:"phone"`
	EmployeeCode   *string                `json:"employee_code"`
	Department     *string                `json:"department"`
	Designation    *string                `json:"designation"`
	DateOfJoining  *time.Time             `json:"date_of_joining"`
	EmploymentType *string                `json:"employment_type"`
	Status         *string                `json:"status"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
}

type EmployeeResponse struct {
	ID             uuid.UUID              `json:"id"`
	Name           string                 `json:"name"`
	Email          string                 `json:"email"`
	Phone          string                 `json:"phone"`
	EmployeeCode   string                 `json:"employee_code"`
	Department     string                 `json:"department"`
	Designation    string                 `json:"designation"`
	DateOfJoining  *time.Time             `json:"date_of_joining"`
	EmploymentType string                 `json:"employment_type"`
	Status         string                 `json:"status"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type ListEmployeesResponse struct {
	Items      []EmployeeResponse `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}
