package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"hcm-backend/internal/models"
)

var (
	emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	phoneRe = regexp.MustCompile(`^\+?[0-9\s\-]{7,20}$`)
	keyRe   = regexp.MustCompile(`^[a-z][a-z0-9_]{0,99}$`)
)

// ValidateFieldKey makes sure the key is snake_case-friendly.
func ValidateFieldKey(key string) error {
	if !keyRe.MatchString(key) {
		return fmt.Errorf("field_key must be lowercase, start with a letter, contain only letters, digits or underscores")
	}
	return nil
}

// ValidateFieldType ensures the field type is supported.
func ValidateFieldType(t string) error {
	switch t {
	case models.FieldTypeText, models.FieldTypeNumber, models.FieldTypeDate,
		models.FieldTypeDropdown, models.FieldTypeBoolean, models.FieldTypeEmail,
		models.FieldTypePhone:
		return nil
	}
	return fmt.Errorf("unsupported field_type: %s", t)
}

// NormalizeAndValidateCustomValue verifies the supplied value matches the field's type
// and (for dropdowns) is in the allowed options. Returns the canonical string form
// that should be stored in the database.
func NormalizeAndValidateCustomValue(field models.EmployeeCustomField, raw interface{}, options []string) (string, error) {
	// nil / empty handling
	if raw == nil {
		if field.Required {
			return "", fmt.Errorf("%s is required", field.FieldName)
		}
		return "", nil
	}
	str, isStr := raw.(string)
	if isStr && strings.TrimSpace(str) == "" {
		if field.Required {
			return "", fmt.Errorf("%s is required", field.FieldName)
		}
		return "", nil
	}

	switch field.FieldType {
	case models.FieldTypeText:
		s := toString(raw)
		return s, nil

	case models.FieldTypeNumber:
		switch v := raw.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		case int:
			return strconv.Itoa(v), nil
		case string:
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				return "", fmt.Errorf("%s must be a number", field.FieldName)
			}
			return v, nil
		}
		return "", fmt.Errorf("%s must be a number", field.FieldName)

	case models.FieldTypeDate:
		s := toString(raw)
		// Accept YYYY-MM-DD or RFC3339
		if _, err := time.Parse("2006-01-02", s); err == nil {
			return s, nil
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t.Format("2006-01-02"), nil
		}
		return "", fmt.Errorf("%s must be a date in YYYY-MM-DD format", field.FieldName)

	case models.FieldTypeBoolean:
		switch v := raw.(type) {
		case bool:
			return strconv.FormatBool(v), nil
		case string:
			lower := strings.ToLower(v)
			if lower == "true" || lower == "false" {
				return lower, nil
			}
		}
		return "", fmt.Errorf("%s must be a boolean", field.FieldName)

	case models.FieldTypeEmail:
		s := toString(raw)
		if !emailRe.MatchString(s) {
			return "", fmt.Errorf("%s must be a valid email", field.FieldName)
		}
		return s, nil

	case models.FieldTypePhone:
		s := toString(raw)
		if !phoneRe.MatchString(s) {
			return "", fmt.Errorf("%s must be a valid phone number", field.FieldName)
		}
		return s, nil

	case models.FieldTypeDropdown:
		s := toString(raw)
		for _, opt := range options {
			if opt == s {
				return s, nil
			}
		}
		return "", fmt.Errorf("%s must be one of: %s", field.FieldName, strings.Join(options, ", "))
	}
	return "", fmt.Errorf("unknown field type for %s", field.FieldName)
}

func toString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case bool:
		return strconv.FormatBool(x)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ValidateEmail returns nil if email looks valid.
func ValidateEmail(s string) error {
	if !emailRe.MatchString(s) {
		return fmt.Errorf("invalid email")
	}
	return nil
}

// ValidatePhone returns nil if phone is empty or looks valid.
func ValidatePhone(s string) error {
	if s == "" {
		return nil
	}
	if !phoneRe.MatchString(s) {
		return fmt.Errorf("invalid phone number")
	}
	return nil
}

// ValidateEmployeeStatus checks if status is one of allowed values.
func ValidateEmployeeStatus(s string) error {
	if s == "" {
		return nil
	}
	switch s {
	case models.EmployeeStatusActive, models.EmployeeStatusInactive,
		models.EmployeeStatusOnLeave, models.EmployeeStatusTerminated:
		return nil
	}
	return fmt.Errorf("invalid status: %s", s)
}
