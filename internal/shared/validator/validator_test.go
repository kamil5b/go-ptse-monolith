package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharederrors "go-modular-monolith/internal/shared/errors"
)

func TestGetValidator(t *testing.T) {
	v1 := GetValidator()
	v2 := GetValidator()

	require.NotNil(t, v1)
	assert.Equal(t, v1, v2) // Should be singleton
}

func TestValidateStructRequired(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Name: "John"}, true},
		{"Empty", TestStruct{Name: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				ve, ok := err.(*sharederrors.ValidationError)
				assert.True(t, ok)
				assert.True(t, ve.HasErrors())
			}
		})
	}
}

func TestValidateStructEmail(t *testing.T) {
	type TestStruct struct {
		Email string `json:"email" validate:"required,email"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Email: "test@example.com"}, true},
		{"Invalid", TestStruct{Email: "not-an-email"}, false},
		{"Empty", TestStruct{Email: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateStructMin(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required,min=3"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Name: "John"}, true},
		{"TooShort", TestStruct{Name: "Jo"}, false},
		{"Minimum", TestStruct{Name: "Joh"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				ve := err.(*sharederrors.ValidationError)
				assert.Contains(t, ve.Error(), "at least")
			}
		})
	}
}

func TestValidateStructMax(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required,max=5"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Name: "John"}, true},
		{"TooLong", TestStruct{Name: "Jonathan"}, false},
		{"Maximum", TestStruct{Name: "Johns"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				ve := err.(*sharederrors.ValidationError)
				assert.Contains(t, ve.Error(), "at most")
			}
		})
	}
}

func TestValidateStructLen(t *testing.T) {
	type TestStruct struct {
		Code string `json:"code" validate:"required,len=3"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Code: "ABC"}, true},
		{"TooShort", TestStruct{Code: "AB"}, false},
		{"TooLong", TestStruct{Code: "ABCD"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				ve := err.(*sharederrors.ValidationError)
				assert.Contains(t, ve.Error(), "exactly")
			}
		})
	}
}

func TestValidateStructNumeric(t *testing.T) {
	type TestStruct struct {
		Age int `json:"age" validate:"required,numeric,gte=0,lte=150"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Age: 25}, true},
		{"TooSmall", TestStruct{Age: -1}, false},
		{"TooLarge", TestStruct{Age: 200}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateStructAlphanum(t *testing.T) {
	type TestStruct struct {
		Username string `json:"username" validate:"required,alphanum"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Username: "user123"}, true},
		{"InvalidSymbol", TestStruct{Username: "user-123"}, false},
		{"InvalidSpace", TestStruct{Username: "user 123"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateStructMultipleErrors(t *testing.T) {
	type TestStruct struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
		Age      int    `json:"age" validate:"required,gte=18"`
	}

	input := TestStruct{
		Email:    "invalid",
		Password: "short",
		Age:      15,
	}

	err := ValidateStruct(input)
	require.Error(t, err)

	ve := err.(*sharederrors.ValidationError)
	assert.True(t, ve.HasErrors())
	assert.True(t, len(ve.Fields) >= 2) // At least email and password are invalid
}

func TestValidateConvenienceFunction(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required"`
	}

	err := Validate(TestStruct{Name: ""})
	assert.Error(t, err)

	err = Validate(TestStruct{Name: "Valid"})
	assert.NoError(t, err)
}

func TestValidateStructNoValidationTag(t *testing.T) {
	type TestStruct struct {
		Name string
	}

	err := ValidateStruct(TestStruct{Name: ""})
	assert.NoError(t, err) // No validation tags, so no errors
}

func TestValidateStructWithJSONTag(t *testing.T) {
	type TestStruct struct {
		UserName string `json:"user_name" validate:"required"`
	}

	input := TestStruct{UserName: ""}
	err := ValidateStruct(input)

	require.Error(t, err)
	ve := err.(*sharederrors.ValidationError)
	assert.Contains(t, ve.Error(), "user_name")
}

func TestValidateStructNilInput(t *testing.T) {
	err := ValidateStruct(nil)
	assert.Error(t, err)
}

func TestValidateStructOneof(t *testing.T) {
	type TestStruct struct {
		Status string `json:"status" validate:"oneof=active inactive pending"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Status: "active"}, true},
		{"Invalid", TestStruct{Status: "unknown"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateStructComplex(t *testing.T) {
	type Address struct {
		Street string `json:"street" validate:"required"`
		City   string `json:"city" validate:"required"`
	}

	type Person struct {
		Name    string  `json:"name" validate:"required,min=2"`
		Email   string  `json:"email" validate:"required,email"`
		Age     int     `json:"age" validate:"required,gte=0,lte=150"`
		Address Address `json:"address" validate:"required"`
	}

	valid := Person{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
		Address: Address{
			Street: "123 Main St",
			City:   "New York",
		},
	}

	err := ValidateStruct(valid)
	assert.NoError(t, err)

	invalid := Person{
		Name:  "J",
		Email: "invalid",
		Age:   200,
		Address: Address{
			Street: "",
			City:   "",
		},
	}

	err = ValidateStruct(invalid)
	require.Error(t, err)
	ve := err.(*sharederrors.ValidationError)
	assert.True(t, ve.HasErrors())
}

func TestErrorMessageFormat(t *testing.T) {
	type TestStruct struct {
		Email string `json:"email" validate:"required,email"`
	}

	err := ValidateStruct(TestStruct{Email: ""})
	ve := err.(*sharederrors.ValidationError)

	// Check that error messages are human-readable
	assert.True(t, ve.HasErrors())
	errStr := ve.Error()
	assert.Contains(t, errStr, "email")
	assert.Contains(t, errStr, "VALIDATION_ERROR")
}
