package validator

import (
	"encoding/json"
	"testing"

	"hcm-backend/internal/models"

	"gorm.io/datatypes"
)

func mkField(t string, required bool, opts []string) models.EmployeeCustomField {
	f := models.EmployeeCustomField{
		FieldName: "Test", FieldKey: "test", FieldType: t, Required: required, Active: true,
	}
	if len(opts) > 0 {
		b, _ := json.Marshal(opts)
		f.Options = datatypes.JSON(b)
	}
	return f
}

func TestNumberAndBooleanCoercion(t *testing.T) {
	v, err := NormalizeAndValidateCustomValue(mkField("number", true, nil), 42, nil)
	if err != nil || v != "42" {
		t.Fatalf("number: got %v, %v", v, err)
	}

	v, err = NormalizeAndValidateCustomValue(mkField("boolean", true, nil), true, nil)
	if err != nil || v != "true" {
		t.Fatalf("boolean: got %v, %v", v, err)
	}
}

func TestDropdownOptionEnforcement(t *testing.T) {
	opts := []string{"Office", "Hybrid", "Remote"}
	if _, err := NormalizeAndValidateCustomValue(mkField("dropdown", true, opts), "Hybrid", opts); err != nil {
		t.Fatalf("expected hybrid to be allowed, got %v", err)
	}
	if _, err := NormalizeAndValidateCustomValue(mkField("dropdown", true, opts), "Beach", opts); err == nil {
		t.Fatalf("expected error for invalid dropdown option")
	}
}

func TestRequiredEnforcement(t *testing.T) {
	if _, err := NormalizeAndValidateCustomValue(mkField("text", true, nil), "", nil); err == nil {
		t.Fatalf("expected required error for empty text")
	}
	if v, err := NormalizeAndValidateCustomValue(mkField("text", false, nil), "", nil); err != nil || v != "" {
		t.Fatalf("optional empty should pass, got %v %v", v, err)
	}
}

func TestEmailAndPhoneTypes(t *testing.T) {
	if _, err := NormalizeAndValidateCustomValue(mkField("email", true, nil), "not-an-email", nil); err == nil {
		t.Fatalf("expected email validation error")
	}
	if _, err := NormalizeAndValidateCustomValue(mkField("phone", true, nil), "abc", nil); err == nil {
		t.Fatalf("expected phone validation error")
	}
}

func TestFieldKeyFormat(t *testing.T) {
	if err := ValidateFieldKey("Bad-Key"); err == nil {
		t.Fatalf("uppercase should fail")
	}
	if err := ValidateFieldKey("good_key1"); err != nil {
		t.Fatalf("snake_case should pass, got %v", err)
	}
}
