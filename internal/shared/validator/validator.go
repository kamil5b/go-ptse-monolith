package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"

	sharederrors "go-modular-monolith/internal/shared/errors"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// GetValidator returns a singleton validator instance
func GetValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()

		// Use json tag names in error messages
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return fld.Name
			}
			return name
		})

		// Register custom validations here
		// validate.RegisterValidation("customtag", customValidationFunc)
	})

	return validate
}

// ValidateStruct validates a struct and returns a ValidationError if invalid
func ValidateStruct(s any) error {
	v := GetValidator()
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return sharederrors.ErrValidation.WithError(err)
	}

	ve := sharederrors.NewValidationError()
	for _, fieldError := range validationErrors {
		field := fieldError.Field()
		message := getErrorMessage(fieldError)
		ve.AddFieldError(field, message)
	}

	return ve
}

// Validate is a convenience function for ValidateStruct
func Validate(s any) error {
	return ValidateStruct(s)
}

// getErrorMessage returns a human-readable error message for a validation error
func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		if fe.Kind() == reflect.String {
			return "must be at least " + fe.Param() + " characters"
		}
		return "must be at least " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String {
			return "must be at most " + fe.Param() + " characters"
		}
		return "must be at most " + fe.Param()
	case "len":
		return "must be exactly " + fe.Param() + " characters"
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "oneof":
		return "must be one of: " + fe.Param()
	case "alphanum":
		return "must contain only alphanumeric characters"
	case "alpha":
		return "must contain only alphabetic characters"
	case "numeric":
		return "must be a valid number"
	case "gte":
		return "must be greater than or equal to " + fe.Param()
	case "gt":
		return "must be greater than " + fe.Param()
	case "lte":
		return "must be less than or equal to " + fe.Param()
	case "lt":
		return "must be less than " + fe.Param()
	case "eqfield":
		return "must be equal to " + fe.Param()
	case "nefield":
		return "must not be equal to " + fe.Param()
	default:
		return "is invalid"
	}
}
