package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Tenant represents a company in the system.
type Tenant struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:200;not null" json:"name"`
	Slug      string    `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Role enum values used across users.
const (
	RoleAdmin = "ADMIN"
	RoleHR    = "HR"
)

// User belongs to a single tenant. Password hash is never serialized.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index;index:idx_users_tenant_email,unique,priority:1" json:"tenant_id"`
	Email        string    `gorm:"size:255;not null;index:idx_users_tenant_email,unique,priority:2" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Name         string    `gorm:"size:200;not null" json:"name"`
	Role         string    `gorm:"size:20;not null" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Field type enums for EmployeeCustomField.
const (
	FieldTypeText     = "text"
	FieldTypeNumber   = "number"
	FieldTypeDate     = "date"
	FieldTypeDropdown = "dropdown"
	FieldTypeBoolean  = "boolean"
	FieldTypeEmail    = "email"
	FieldTypePhone    = "phone"
)

// EmployeeCustomField is the per-tenant configuration for additional employee fields.
type EmployeeCustomField struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID     uuid.UUID      `gorm:"type:uuid;not null;index;index:idx_fields_tenant_key,unique,priority:1" json:"tenant_id"`
	FieldName    string         `gorm:"size:200;not null" json:"field_name"`
	FieldKey     string         `gorm:"size:100;not null;index:idx_fields_tenant_key,unique,priority:2" json:"field_key"`
	FieldType    string         `gorm:"size:20;not null" json:"field_type"`
	Required     bool           `gorm:"not null;default:false" json:"required"`
	Active       bool           `gorm:"not null;default:true" json:"active"`
	Options      datatypes.JSON `gorm:"type:jsonb" json:"options,omitempty"`
	DisplayOrder int            `gorm:"not null;default:0" json:"display_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Employee status enum.
const (
	EmployeeStatusActive    = "ACTIVE"
	EmployeeStatusInactive  = "INACTIVE"
	EmployeeStatusOnLeave   = "ON_LEAVE"
	EmployeeStatusTerminated = "TERMINATED"
)

// Employee is the core employee record. Common fields are columns; custom values
// are stored in EmployeeCustomFieldValue.
type Employee struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index;index:idx_emp_tenant_email,unique,priority:1;index:idx_emp_tenant_code,unique,priority:1" json:"tenant_id"`
	Name           string     `gorm:"size:200;not null" json:"name"`
	Email          string     `gorm:"size:255;not null;index:idx_emp_tenant_email,unique,priority:2" json:"email"`
	Phone          string     `gorm:"size:30" json:"phone"`
	EmployeeCode   string     `gorm:"size:50;not null;index:idx_emp_tenant_code,unique,priority:2" json:"employee_code"`
	Department     string     `gorm:"size:100" json:"department"`
	Designation    string     `gorm:"size:100" json:"designation"`
	DateOfJoining  *time.Time `json:"date_of_joining"`
	EmploymentType string     `gorm:"size:50" json:"employment_type"`
	Status         string     `gorm:"size:30;not null;default:'ACTIVE'" json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	CustomValues []EmployeeCustomFieldValue `gorm:"foreignKey:EmployeeID" json:"custom_values,omitempty"`
}

// EmployeeCustomFieldValue stores a value for a single custom field on a single employee.
type EmployeeCustomFieldValue struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;index:idx_ecfv_emp_field,unique,priority:1" json:"employee_id"`
	FieldID    uuid.UUID `gorm:"type:uuid;not null;index:idx_ecfv_emp_field,unique,priority:2" json:"field_id"`
	Value      string    `gorm:"type:text" json:"value"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
