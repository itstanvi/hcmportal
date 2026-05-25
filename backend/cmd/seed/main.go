package main

import (
	"encoding/json"
	"log"
	"time"

	"hcm-backend/internal/auth"
	"hcm-backend/internal/config"
	"hcm-backend/internal/database"
	"hcm-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	if err := seed(db); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
	log.Println("seed completed successfully")
}

func seed(db *gorm.DB) error {
	tenants := []models.Tenant{
		{ID: uuid.New(), Name: "Acme Corp", Slug: "acme"},
		{ID: uuid.New(), Name: "Globex Inc", Slug: "globex"},
	}
	for i := range tenants {
		var existing models.Tenant
		if err := db.Where("slug = ?", tenants[i].Slug).First(&existing).Error; err == nil {
			tenants[i] = existing
			continue
		}
		if err := db.Create(&tenants[i]).Error; err != nil {
			return err
		}
	}

	// Users: 2 per tenant — one ADMIN, one HR. Default password: Password@123
	pwHash, _ := auth.HashPassword("Password@123")
	type seedUser struct {
		Email string
		Name  string
		Role  string
	}
	for _, t := range tenants {
		users := []seedUser{
			{Email: "admin@" + t.Slug + ".com", Name: t.Name + " Admin", Role: models.RoleAdmin},
			{Email: "hr@" + t.Slug + ".com", Name: t.Name + " HR", Role: models.RoleHR},
		}
		for _, u := range users {
			var existing models.User
			if err := db.Where("LOWER(email) = LOWER(?)", u.Email).First(&existing).Error; err == nil {
				continue
			}
			user := models.User{
				ID:           uuid.New(),
				TenantID:     t.ID,
				Email:        u.Email,
				PasswordHash: pwHash,
				Name:         u.Name,
				Role:         u.Role,
			}
			if err := db.Create(&user).Error; err != nil {
				return err
			}
		}
	}

	// Custom field samples for Acme only.
	acme := tenants[0]
	options, _ := json.Marshal([]string{"Office", "Hybrid", "Remote"})
	sampleFields := []models.EmployeeCustomField{
		{
			ID: uuid.New(), TenantID: acme.ID, FieldName: "Work Mode", FieldKey: "work_mode",
			FieldType: models.FieldTypeDropdown, Required: true, Active: true,
			Options: datatypes.JSON(options), DisplayOrder: 1,
		},
		{
			ID: uuid.New(), TenantID: acme.ID, FieldName: "PAN Number", FieldKey: "pan_number",
			FieldType: models.FieldTypeText, Required: false, Active: true, DisplayOrder: 2,
		},
		{
			ID: uuid.New(), TenantID: acme.ID, FieldName: "Has Laptop", FieldKey: "has_laptop",
			FieldType: models.FieldTypeBoolean, Required: false, Active: true, DisplayOrder: 3,
		},
	}
	for _, f := range sampleFields {
		var existing models.EmployeeCustomField
		if err := db.Where("tenant_id = ? AND field_key = ?", f.TenantID, f.FieldKey).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&f).Error; err != nil {
			return err
		}
	}

	// Sample employee in Acme
	var existingEmp models.Employee
	if err := db.Where("tenant_id = ? AND email = ?", acme.ID, "alice@acme.com").First(&existingEmp).Error; err != nil {
		dob := time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)
		emp := models.Employee{
			ID:             uuid.New(),
			TenantID:       acme.ID,
			Name:           "Alice Walker",
			Email:          "alice@acme.com",
			Phone:          "+1 555 123 4567",
			EmployeeCode:   "ACME-001",
			Department:     "Engineering",
			Designation:    "Senior Engineer",
			DateOfJoining:  &dob,
			EmploymentType: "Full-Time",
			Status:         models.EmployeeStatusActive,
		}
		if err := db.Create(&emp).Error; err != nil {
			return err
		}
	}

	return nil
}
