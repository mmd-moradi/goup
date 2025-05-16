package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, err := range validationErrors {
				field := strings.ToLower(err.Field())
				errorMessages = append(errorMessages, fmt.Sprintf("field %s failed validation: %s", field, getErrorMessage(err)))
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, ", "))
		}
		return err
	}
	return nil
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s charechters long", err.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters long", err.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", err.Param())
	default:
		return fmt.Sprintf("failed validation for tag %s", err.Tag())
	}
}
