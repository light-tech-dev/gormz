package advanced_test

import (
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/advanced"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnion(t *testing.T) {
	setupTestDB(t)

	users := []User{
		{Name: "Ali", Email: "ali@test.com", Age: 30, Active: true},
		{Name: "Sara", Email: "sara@test.com", Age: 25, Active: true},
		{Name: "Omar", Email: "omar@test.com", Age: 17, Active: false},
	}
	require.NoError(t, gormz.New[User]().CreateMany(users))

	q1 := gormz.New[User]().Filter("age__lt", 18)
	q2 := gormz.New[User]().Filter("active", true)

	result, err := advanced.Union(q1, q2).All()
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestUnionAll(t *testing.T) {
	setupTestDB(t)

	users := []User{
		{Name: "Ali", Email: "ali@test.com", Age: 30, Active: true},
		{Name: "Sara", Email: "sara@test.com", Age: 25, Active: true},
	}
	require.NoError(t, gormz.New[User]().CreateMany(users))

	q1 := gormz.New[User]().Filter("active", true)
	q2 := gormz.New[User]().Filter("age__gt", 20)

	result, err := advanced.UnionAll(q1, q2).All()
	require.NoError(t, err)
	// 2 + 2 = 4 مع التكرار
	assert.Len(t, result, 4)
}