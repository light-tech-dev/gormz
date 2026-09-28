package tests

import (
	"testing"

	"github.com/light-tech-dev/gormz/internal"
	"github.com/stretchr/testify/assert"
)

func TestValidateField_Valid(t *testing.T) {
	valid := []string{
		"name", "user_id", "first_name",
		"users.id", "t1.name", "field_1",
	}

	for _, f := range valid {
		t.Run(f, func(t *testing.T) {
			assert.NoError(t, internal.ValidateField(f))
		})
	}
}

func TestValidateField_Invalid(t *testing.T) {
	invalid := []string{
		"", "name; DROP TABLE", "name'", "name\"",
		"name--", "1field", "field name", "select", "drop",
	}

	for _, f := range invalid {
		t.Run(f, func(t *testing.T) {
			assert.Error(t, internal.ValidateField(f))
		})
	}
}
