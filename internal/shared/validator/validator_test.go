package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharederrors "github.com/kamil5b/go-ptse-monolith/internal/shared/errors"
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

func TestValidateStructUUID(t *testing.T) {
	type TestStruct struct {
		ID string `json:"id" validate:"required,uuid"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid UUID", TestStruct{ID: "550e8400-e29b-41d4-a716-446655440000"}, true},
		{"Invalid UUID", TestStruct{ID: "not-a-uuid"}, false},
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

func TestValidateStructURL(t *testing.T) {
	type TestStruct struct {
		Website string `json:"website" validate:"required,url"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid URL", TestStruct{Website: "https://example.com"}, true},
		{"Invalid URL", TestStruct{Website: "not a url"}, false},
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

func TestValidateStructAlpha(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required,alpha"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Name: "John"}, true},
		{"WithNumber", TestStruct{Name: "John123"}, false},
		{"WithSymbol", TestStruct{Name: "John-Doe"}, false},
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

func TestValidateStructGte(t *testing.T) {
	type TestStruct struct {
		Age int `json:"age" validate:"required,gte=18"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Age: 25}, true},
		{"Boundary", TestStruct{Age: 18}, true},
		{"Below", TestStruct{Age: 17}, false},
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

func TestValidateStructGt(t *testing.T) {
	type TestStruct struct {
		Age int `json:"age" validate:"required,gt=18"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Age: 25}, true},
		{"Equal", TestStruct{Age: 18}, false},
		{"Below", TestStruct{Age: 17}, false},
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

func TestValidateStructLte(t *testing.T) {
	type TestStruct struct {
		Age int `json:"age" validate:"required,lte=65"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Age: 25}, true},
		{"Boundary", TestStruct{Age: 65}, true},
		{"Above", TestStruct{Age: 66}, false},
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

func TestValidateStructLt(t *testing.T) {
	type TestStruct struct {
		Age int `json:"age" validate:"required,lt=65"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Age: 25}, true},
		{"Equal", TestStruct{Age: 65}, false},
		{"Above", TestStruct{Age: 66}, false},
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

func TestValidateStructEqfield(t *testing.T) {
	type TestStruct struct {
		Password        string `json:"password" validate:"required"`
		PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Match", TestStruct{Password: "secret", PasswordConfirm: "secret"}, true},
		{"NoMatch", TestStruct{Password: "secret", PasswordConfirm: "different"}, false},
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

func TestValidateStructNefield(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1" validate:"required"`
		Field2 string `json:"field2" validate:"required,nefield=Field1"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Different", TestStruct{Field1: "value1", Field2: "value2"}, true},
		{"Same", TestStruct{Field1: "value", Field2: "value"}, false},
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

func TestValidateStructMinNumeric(t *testing.T) {
	type TestStruct struct {
		Count int `json:"count" validate:"required,min=10"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Count: 15}, true},
		{"TooSmall", TestStruct{Count: 5}, false},
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

func TestValidateStructMaxNumeric(t *testing.T) {
	type TestStruct struct {
		Count int `json:"count" validate:"required,max=100"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Count: 50}, true},
		{"TooLarge", TestStruct{Count: 150}, false},
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

func TestGetValidatorSingleton(t *testing.T) {
	v1 := GetValidator()
	v2 := GetValidator()
	v3 := GetValidator()

	// All calls should return the same instance
	assert.True(t, v1 == v2)
	assert.True(t, v2 == v3)
}

func TestValidateStructMinStringType(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required,min=5"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Name: "John Doe"}, true},
		{"Boundary", TestStruct{Name: "John"}, false},
		{"Minimum", TestStruct{Name: "Hello"}, true},
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

func TestValidateStructMaxStringType(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" validate:"required,max=10"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Name: "John"}, true},
		{"Boundary", TestStruct{Name: "JohnDoeXXX"}, true},
		{"TooLong", TestStruct{Name: "JohnDoeXXXX"}, false},
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

func TestGetErrorMessageWithDefault(t *testing.T) {
	type TestStruct struct {
		Value string `json:"value" validate:"required"`
	}

	// Test default case handling in getErrorMessage
	err := ValidateStruct(TestStruct{Value: ""})
	ve := err.(*sharederrors.ValidationError)
	assert.True(t, ve.HasErrors())
}

func TestValidateStructNumericString(t *testing.T) {
	type TestStruct struct {
		Value string `json:"value" validate:"required,numeric"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		isValid bool
	}{
		{"Valid", TestStruct{Value: "12345"}, true},
		{"Invalid", TestStruct{Value: "abc123"}, false},
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

func TestValidateStructWithJSONTagPriority(t *testing.T) {
	type TestStruct struct {
		UserEmail string `json:"email" validate:"required,email"`
	}

	err := ValidateStruct(TestStruct{UserEmail: ""})
	ve := err.(*sharederrors.ValidationError)

	// Error should use JSON tag name not field name
	errMap := ve.ToMap()
	// Check that the 'fields' key contains the error info
	assert.NotNil(t, errMap)
	assert.Equal(t, "VALIDATION_ERROR", errMap["code"])
	fieldsData := errMap["fields"]
	assert.NotNil(t, fieldsData)
}

func TestValidateStructJSONTagDash(t *testing.T) {
	type TestStruct struct {
		Skip  string `json:"-" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	err := ValidateStruct(TestStruct{Skip: "", Email: ""})
	ve := err.(*sharederrors.ValidationError)

	// Dash means skip this field, so field name should be used as fallback
	assert.True(t, ve.HasErrors())
}

func TestGetErrorMessageCoverage(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		tag      string
		expected string
	}{
		{
			name: "uuid error",
			value: struct {
				ID string `json:"id" validate:"uuid"`
			}{ID: "invalid"},
			tag:      "uuid",
			expected: "must be a valid UUID",
		},
		{
			name: "url error",
			value: struct {
				Site string `json:"site" validate:"url"`
			}{Site: "not-url"},
			tag:      "url",
			expected: "must be a valid URL",
		},
		{
			name: "alpha error",
			value: struct {
				Name string `json:"name" validate:"alpha"`
			}{Name: "Name123"},
			tag:      "alpha",
			expected: "must contain only alphabetic characters",
		},
		{
			name: "gte error",
			value: struct {
				Age int `json:"age" validate:"gte=18"`
			}{Age: 10},
			tag:      "gte",
			expected: "must be greater than or equal to",
		},
		{
			name: "gt error",
			value: struct {
				Age int `json:"age" validate:"gt=18"`
			}{Age: 18},
			tag:      "gt",
			expected: "must be greater than",
		},
		{
			name: "lte error",
			value: struct {
				Age int `json:"age" validate:"lte=65"`
			}{Age: 70},
			tag:      "lte",
			expected: "must be less than or equal to",
		},
		{
			name: "lt error",
			value: struct {
				Age int `json:"age" validate:"lt=65"`
			}{Age: 65},
			tag:      "lt",
			expected: "must be less than",
		},
		{
			name: "eqfield error",
			value: struct {
				Password        string `json:"password" validate:"required"`
				PasswordConfirm string `json:"password_confirm" validate:"eqfield=Password"`
			}{Password: "sec", PasswordConfirm: "diff"},
			tag:      "eqfield",
			expected: "must be equal to",
		},
		{
			name: "nefield error",
			value: struct {
				Field1 string `json:"field1" validate:"required"`
				Field2 string `json:"field2" validate:"nefield=Field1"`
			}{Field1: "same", Field2: "same"},
			tag:      "nefield",
			expected: "must not be equal to",
		},
		{
			name: "default error case",
			value: struct {
				Field string `json:"field" validate:"required"`
			}{Field: ""},
			tag:      "required",
			expected: "is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.value)
			require.Error(t, err)
			ve := err.(*sharederrors.ValidationError)
			assert.True(t, ve.HasErrors())
			errStr := ve.Error()
			assert.Contains(t, errStr, tt.expected)
		})
	}
}
