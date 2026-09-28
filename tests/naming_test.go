package tests

import (
	"testing"

	"github.com/light-tech-dev/gormz/internal"
	"github.com/stretchr/testify/assert"
)

func TestToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"FirstName", "first_name"},
		{"UserID", "user_id"},
		{"HTTPServer", "http_server"},
		{"APIKey", "api_key"},
		{"ID", "id"},
		{"", ""},
		{"already_snake", "already_snake"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, internal.ToSnake(tt.input))
		})
	}
}

// ... (نفس الشيء لـ ToCamel, ToPascal, Pluralize, Singularize)
