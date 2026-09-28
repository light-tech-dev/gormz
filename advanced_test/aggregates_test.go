package advanced_test

import (
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/advanced"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupBy_Count(t *testing.T) {
	setupTestDB(t)

	orders := []Order{
		{UserID: 1, Total: 100, Status: "paid"},
		{UserID: 1, Total: 200, Status: "paid"},
		{UserID: 2, Total: 150, Status: "paid"},
		{UserID: 2, Total: 50, Status: "pending"},
		{UserID: 3, Total: 300, Status: "paid"},
	}
	require.NoError(t, gormz.New[Order]().CreateMany(orders))

	results, err := advanced.GroupBy[Order]("user_id").
		Count("id", "order_count").
		Sum("total", "total_amount").
		All()

	require.NoError(t, err)
	assert.Len(t, results, 3) // 3 user IDs
}

func TestGroupBy_WithHaving(t *testing.T) {
	setupTestDB(t)

	orders := []Order{
		{UserID: 1, Total: 100, Status: "paid"},
		{UserID: 1, Total: 200, Status: "paid"},
		{UserID: 1, Total: 150, Status: "paid"},
		{UserID: 2, Total: 100, Status: "paid"},
	}
	require.NoError(t, gormz.New[Order]().CreateMany(orders))

	type UserStats struct {
		UserID     uint
		OrderCount int64
	}

	var stats []UserStats
	err := advanced.GroupBy[Order]("user_id").
		Count("id", "order_count").
		Having("COUNT(id) > ?", 2).
		ScanInto(&stats)

	require.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, uint(1), stats[0].UserID)
	assert.Equal(t, int64(3), stats[0].OrderCount)
}

func TestGroupBy_WithFilter(t *testing.T) {
	setupTestDB(t)

	orders := []Order{
		{UserID: 1, Total: 100, Status: "paid"},
		{UserID: 1, Total: 200, Status: "pending"},
		{UserID: 2, Total: 150, Status: "paid"},
	}
	require.NoError(t, gormz.New[Order]().CreateMany(orders))

	results, err := advanced.GroupBy[Order]("user_id").
		Count("id", "paid_count").
		Filter("status", "paid").
		All()

	require.NoError(t, err)
	assert.Len(t, results, 2) // 2 users with paid orders
}

func TestDistinct(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	emails, err := advanced.Distinct[User]("email")
	require.NoError(t, err)
	assert.Len(t, emails, 5)
}

func TestCountDistinctValues(t *testing.T) {
	setupTestDB(t)
	seedUsers(t)

	count, err := advanced.CountDistinctValues[User]("name")
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}