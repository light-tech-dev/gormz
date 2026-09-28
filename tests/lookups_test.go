package tests

import (
	"testing"

	"github.com/light-tech-dev/gormz/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildLookup_Equality(t *testing.T) {
	l, err := internal.BuildLookup("name", "Ali", false)
	require.NoError(t, err)
	assert.Equal(t, "name = ?", l.SQL)
	assert.Equal(t, []any{"Ali"}, l.Args)
	assert.False(t, l.Negate)
}

func TestBuildLookup_GreaterThan(t *testing.T) {
	l, err := internal.BuildLookup("age__gt", 18, false)
	require.NoError(t, err)
	assert.Equal(t, "age > ?", l.SQL)
	assert.Equal(t, []any{18}, l.Args)
}

func TestBuildLookup_Contains(t *testing.T) {
	l, err := internal.BuildLookup("name__contains", "ali", false)
	require.NoError(t, err)
	assert.Equal(t, "name LIKE ?", l.SQL)
	assert.Equal(t, []any{"%ali%"}, l.Args)
}

func TestBuildLookup_IContains(t *testing.T) {
	l, err := internal.BuildLookup("name__icontains", "ALI", false)
	require.NoError(t, err)
	assert.Contains(t, l.SQL, "LOWER(name)")
	assert.Contains(t, l.SQL, "LIKE")
	assert.Equal(t, []any{"%ali%"}, l.Args)
}

func TestBuildLookup_In(t *testing.T) {
	l, err := internal.BuildLookup("id__in", []int{1, 2, 3}, false)
	require.NoError(t, err)
	assert.Equal(t, "id IN (?,?,?)", l.SQL)
	assert.Len(t, l.Args, 3)
}

func TestBuildLookup_InEmpty(t *testing.T) {
	l, err := internal.BuildLookup("id__in", []int{}, false)
	require.NoError(t, err)
	assert.Equal(t, "1=0", l.SQL)
}

func TestBuildLookup_IsNull(t *testing.T) {
	l, err := internal.BuildLookup("deleted_at__isnull", true, false)
	require.NoError(t, err)
	assert.Equal(t, "deleted_at IS NULL", l.SQL)

	l, err = internal.BuildLookup("deleted_at__isnull", false, false)
	require.NoError(t, err)
	assert.Equal(t, "deleted_at IS NOT NULL", l.SQL)
}

func TestBuildLookup_Between(t *testing.T) {
	l, err := internal.BuildLookup("age__between", []int{18, 65}, false)
	require.NoError(t, err)
	assert.Equal(t, "age BETWEEN ? AND ?", l.SQL)
	assert.Equal(t, []any{18, 65}, l.Args)
}

func TestBuildLookup_InvalidField(t *testing.T) {
	_, err := internal.BuildLookup("bad;field", "x", false)
	assert.Error(t, err)
}

func TestBuildLookup_InvalidLookup(t *testing.T) {
	_, err := internal.BuildLookup("name__bad", "x", false)
	assert.Error(t, err)
}

func TestBuildLookup_DateYear(t *testing.T) {
	l, err := internal.BuildLookupWithDialect("created_at__year", 2025, false, internal.DialectSQLite)
	require.NoError(t, err)
	assert.Contains(t, l.SQL, "strftime")
	assert.Equal(t, []any{2025}, l.Args)
}

func TestBuildLookupWithDialect_Postgres(t *testing.T) {
	l, err := internal.BuildLookupWithDialect("created_at__year", 2025, false, internal.DialectPostgres)
	require.NoError(t, err)
	assert.Contains(t, l.SQL, "EXTRACT")
	assert.Contains(t, l.SQL, "YEAR")
}

func TestBuildLookupWithDialect_MySQL(t *testing.T) {
	l, err := internal.BuildLookupWithDialect("created_at__year", 2025, false, internal.DialectMySQL)
	require.NoError(t, err)
	assert.Contains(t, l.SQL, "YEAR(")
}

// ═══════════════════════════════════════════════
// Regex Tests (جديد)
// ═══════════════════════════════════════════════

func TestBuildLookup_Regex_Postgres(t *testing.T) {
	l, err := internal.BuildLookupWithDialect("name__regex", "^ali", false, internal.DialectPostgres)
	require.NoError(t, err)
	assert.Equal(t, "name ~ ?", l.SQL)
	assert.Equal(t, []any{"^ali"}, l.Args)
}

func TestBuildLookup_IRegEx_Postgres(t *testing.T) {
	l, err := internal.BuildLookupWithDialect("name__iregex", "^ALI", false, internal.DialectPostgres)
	require.NoError(t, err)
	assert.Equal(t, "LOWER(name) ~ ?", l.SQL)
	assert.Equal(t, []any{"^ali"}, l.Args)
}

func TestBuildLookup_Regex_MySQL(t *testing.T) {
	l, err := internal.BuildLookupWithDialect("name__regex", "^ali", false, internal.DialectMySQL)
	require.NoError(t, err)
	assert.Equal(t, "name REGEXP ?", l.SQL)
}

func TestBuildLookup_Regex_SQLite(t *testing.T) {
	l, err := internal.BuildLookupWithDialect("name__regex", "^ali", false, internal.DialectSQLite)
	require.NoError(t, err)
	assert.Equal(t, "name REGEXP ?", l.SQL)
}
