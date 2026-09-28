package gormz_test

import (
	"fmt"
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ═══════════════════════════════════════════════
// Benchmark Setup
// ═══════════════════════════════════════════════

// setupBenchDB creates a new DB for benchmarks.
func setupBenchDB(b *testing.B) {
	b.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(b, err)
	require.NoError(b, db.AutoMigrate(&User{}, &Order{}))

	gormz.SetDB(db)
	gormz.ClearRegistry()

	b.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
		gormz.ResetDB()
		gormz.ClearRegistry()
	})
}

// ═══════════════════════════════════════════════
// Single Operation Benchmarks
// ═══════════════════════════════════════════════

func BenchmarkCreate(b *testing.B) {
	setupBenchDB(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		user := &User{
			Name:  "User",
			Email: fmt.Sprintf("user%d@test.com", i),
		}
		if err := gormz.New[User]().Create(user); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateMany(b *testing.B) {
	setupBenchDB(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		users := make([]User, 100)
		for j := range users {
			users[j] = User{
				Name:  "User",
				Email: fmt.Sprintf("user%d_%d@test.com", i, j),
			}
		}
		if err := gormz.New[User]().CreateMany(users); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFirst(b *testing.B) {
	setupBenchDB(b)

	if err := gormz.New[User]().Create(&User{
		Name:  "Ali",
		Email: "ali@test.com",
	}); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gormz.New[User]().First()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFilter(b *testing.B) {
	setupBenchDB(b)

	// Seed 1000 users
	for i := 0; i < 1000; i++ {
		if err := gormz.New[User]().Create(&User{
			Name:   "User",
			Email:  fmt.Sprintf("user%d@test.com", i),
			Age:    20 + (i % 50),
			Active: i%2 == 0,
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gormz.New[User]().
			Filter("active", true).
			Filter("age__gte", 30).
			Limit(10).
			All()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaginate(b *testing.B) {
	setupBenchDB(b)

	for i := 0; i < 1000; i++ {
		if err := gormz.New[User]().Create(&User{
			Name:  "User",
			Email: fmt.Sprintf("user%d@test.com", i),
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gormz.New[User]().OrderBy("id").Paginate(1, 20)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCount(b *testing.B) {
	setupBenchDB(b)

	for i := 0; i < 1000; i++ {
		if err := gormz.New[User]().Create(&User{
			Name:  "User",
			Email: fmt.Sprintf("u%d@test.com", i),
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gormz.New[User]().Count()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExists(b *testing.B) {
	setupBenchDB(b)

	if err := gormz.New[User]().Create(&User{
		Name:  "Ali",
		Email: "ali@test.com",
	}); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gormz.New[User]().Filter("name", "Ali").Exists()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ═══════════════════════════════════════════════
// Chained Query Benchmarks
// ═══════════════════════════════════════════════

func BenchmarkQuerySet_Chain(b *testing.B) {
	setupBenchDB(b)

	for i := 0; i < 500; i++ {
		if err := gormz.New[User]().Create(&User{
			Name:   "User",
			Email:  fmt.Sprintf("u%d@test.com", i),
			Age:    20 + (i % 50),
			Active: true,
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := gormz.New[User]().
			Filter("active", true).
			Filter("age__gte", 30).
			OrderBy("-age").
			Limit(20)

		_, err := q.All()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ═══════════════════════════════════════════════
// Registry Benchmarks
// ═══════════════════════════════════════════════

func BenchmarkRegistry_Lookup(b *testing.B) {
	setupBenchDB(b)

	gormz.Register[User]("user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := gormz.Lookup[User]("user")
		if !ok {
			b.Fatal("not found")
		}
	}
}

// ═══════════════════════════════════════════════
// Comparison with GORM
// ═══════════════════════════════════════════════

// BenchmarkVsGORM compares gormz vs raw GORM.
func BenchmarkVsGORM(b *testing.B) {
	setupBenchDB(b)

	for i := 0; i < 1000; i++ {
		if err := gormz.New[User]().Create(&User{
			Name:   "User",
			Email:  fmt.Sprintf("u%d@test.com", i),
			Age:    20 + (i % 50),
			Active: true,
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.Run("gormz", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := gormz.New[User]().
				Filter("active", true).
				Filter("age__gte", 30).
				Limit(10).
				All()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("gorm", func(b *testing.B) {
		db := gormz.DB()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var users []User
			err := db.Model(&User{}).
				Where("active = ?", true).
				Where("age >= ?", 30).
				Limit(10).
				Find(&users).Error
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}