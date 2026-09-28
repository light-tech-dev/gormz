package advanced_test

import (
	"context"
	"errors"
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/light-tech-dev/gormz/advanced"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ═══════════════════════════════════════════════
// Basic Transaction Tests
// ═══════════════════════════════════════════════

func TestWithTransaction_Commit(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
		func(tx *advanced.Tx) error {
			return advanced.Query[User](tx).Create(&User{
				Name:  "Ali",
				Email: "ali@test.com",
			})
		})
	require.NoError(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(1), count)
}

func TestWithTransaction_Rollback(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
		func(tx *advanced.Tx) error {
			if err := advanced.Query[User](tx).Create(&User{
				Name:  "Ali",
				Email: "ali@test.com",
			}); err != nil {
				return err
			}
			return errors.New("intentional")
		})
	assert.Error(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(0), count)
}

func TestWithTransaction_Panic(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
		func(tx *advanced.Tx) error {
			if err := advanced.Query[User](tx).Create(&User{
				Name:  "Ali",
				Email: "ali@test.com",
			}); err != nil {
				return err
			}
			panic("oops")
		})
	assert.Error(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(0), count)
}

// ═══════════════════════════════════════════════
// Nested Transaction Tests
// ═══════════════════════════════════════════════

func TestNested_Savepoint(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
		func(tx *advanced.Tx) error {
			// Outer insert — should succeed
			if err := advanced.Query[User](tx).Create(&User{
				Name:  "A",
				Email: "a@test.com",
			}); err != nil {
				return err
			}

			// Nested insert — should rollback
			_ = tx.Nested(func(inner *advanced.Tx) error {
				if err := advanced.Query[User](inner).Create(&User{
					Name:  "B",
					Email: "b@test.com",
				}); err != nil {
					return err
				}
				return errors.New("nested fails")
			})

			return nil
		})
	require.NoError(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(1), count)

	a, _ := gormz.New[User]().Find("email", "a@test.com")
	assert.NotNil(t, a)

	b, _ := gormz.New[User]().Find("email", "b@test.com")
	assert.Nil(t, b)
}

// ═══════════════════════════════════════════════
// Convenience Wrappers
// ═══════════════════════════════════════════════

func TestSimpleTransaction(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.SimpleTransaction(ctx, func(tx *advanced.Tx) error {
		return advanced.Query[User](tx).Create(&User{
			Name:  "Ali",
			Email: "ali@test.com",
		})
	})
	require.NoError(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(1), count)
}

// ═══════════════════════════════════════════════
// Manual Transaction Control
// ═══════════════════════════════════════════════

func TestBegin_Commit(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	tx, err := advanced.Begin(ctx, advanced.DefaultTxConfig())
	require.NoError(t, err)
	defer tx.RollbackIfActive()

	if err := advanced.Query[User](tx).Create(&User{
		Name:  "Ali",
		Email: "ali@test.com",
	}); err != nil {
		t.Fatal(err)
	}

	require.NoError(t, tx.Commit())

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(1), count)
}

func TestBegin_Rollback(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	tx, err := advanced.Begin(ctx, advanced.DefaultTxConfig())
	require.NoError(t, err)

	if err := advanced.Query[User](tx).Create(&User{
		Name:  "Ali",
		Email: "ali@test.com",
	}); err != nil {
		t.Fatal(err)
	}

	require.NoError(t, tx.Rollback())

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(0), count)
}

func TestBegin_RollbackIfActive(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	tx, err := advanced.Begin(ctx, advanced.DefaultTxConfig())
	require.NoError(t, err)

	if err := advanced.Query[User](tx).Create(&User{
		Name:  "Ali",
		Email: "ali@test.com",
	}); err != nil {
		t.Fatal(err)
	}

	// Auto rollback — don't call Commit
	tx.RollbackIfActive()

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(0), count)
}

// ═══════════════════════════════════════════════
// Error Handling
// ═══════════════════════════════════════════════

func TestWithTransaction_NilCallback(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(), nil)
	assert.Error(t, err)
}

func TestTx_CommitAfterCommit(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	tx, err := advanced.Begin(ctx, advanced.DefaultTxConfig())
	require.NoError(t, err)

	require.NoError(t, tx.Commit())

	// Second commit should error
	err = tx.Commit()
	assert.Error(t, err)
}

func TestTx_RollbackAfterCommit(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	tx, err := advanced.Begin(ctx, advanced.DefaultTxConfig())
	require.NoError(t, err)

	require.NoError(t, tx.Commit())

	// Rollback after commit should error
	err = tx.Rollback()
	assert.Error(t, err)
}

// ═══════════════════════════════════════════════
// Query / QueryWithContext Functions
// ═══════════════════════════════════════════════

func TestQueryFunction(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
		func(tx *advanced.Tx) error {
			q := advanced.Query[User](tx)
			return q.Create(&User{Name: "Ali", Email: "ali@test.com"})
		})
	require.NoError(t, err)
}

func TestQueryWithContextFunction(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
		func(tx *advanced.Tx) error {
			q := advanced.QueryWithContext[User](tx, ctx)
			return q.Create(&User{Name: "Ali", Email: "ali@test.com"})
		})
	require.NoError(t, err)
}

// ═══════════════════════════════════════════════
// Tx.Instance() Tests
// ═══════════════════════════════════════════════

func TestTx_Instance(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := advanced.WithTransaction(ctx, advanced.DefaultTxConfig(),
		func(tx *advanced.Tx) error {
			instance := tx.Instance()
			q := gormz.NewWith[User](instance)
			return q.Create(&User{Name: "Ali", Email: "ali@test.com"})
		})
	require.NoError(t, err)

	count, _ := gormz.New[User]().Count()
	assert.Equal(t, int64(1), count)
}
