package validator

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	AlphaNumberSpaceRegex = regexp.MustCompile("^[a-zA-Z0-9 ]+$")
)

// Validator is a validator that validates the given struct.
type Validator interface {
	// Validate validates the given struct
	Validate(s any) error
}

type DefaultValidator struct {
	v *validator.Validate
}

// NewDefaultValidator creates a new default validator.
// It returns a new DefaultValidator and an error if the validator registration fails.
func NewDefaultValidator() (*DefaultValidator, error) {
	v := validator.New()

	// Register custom validators
	if err := v.RegisterValidation("alphanumspace", validateAlphanumspace); err != nil {
		return nil, fmt.Errorf("register alphanumspace validator: %w", err)
	}

	if err := v.RegisterValidation("enum", validateEnum); err != nil {
		return nil, fmt.Errorf("register enum validator: %w", err)
	}

	return &DefaultValidator{v: v}, nil
}

func (v DefaultValidator) Validate(s any) error {
	return v.v.Struct(s)
}

// IsValidationError checks if the given error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(validator.ValidationErrors)
	return ok
}

func ValidationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "uuid":
		return "must be a valid UUID"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters long", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", fe.Param())
	case "alphanumspace":
		return "must contain only alphanumeric characters and spaces"
	case "ip":
		return "must be a valid IP address"
	case "enum":
		return fmt.Sprintf("invalid enum value: %s", fe.Value())
	case "sort":
		return fmt.Sprintf("must contain only allowed sort fields: [%s]", fe.Param())
	default:
		return "is invalid"
	}
}

func validateAlphanumspace(fl validator.FieldLevel) bool {
	return AlphaNumberSpaceRegex.MatchString(fl.Field().String())
}

func validateEnum(fl validator.FieldLevel) bool {
	type Enum interface {
		Validate() error
	}

	value, ok := fl.Field().Interface().(Enum)
	if !ok {
		return false
	}

	return value.Validate() == nil
}
