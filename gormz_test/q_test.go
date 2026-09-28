package gormz_test

import (
	"testing"

	"github.com/light-tech-dev/gormz"
	"github.com/stretchr/testify/assert"
)

func TestQ_SimpleOr(t *testing.T) {
	q := gormz.QOr(
		gormz.Eq("status", "active"),
		gormz.Eq("status", "pending"),
	)

	sql, args := q.ToSQL()
	assert.Contains(t, sql, "status = ?")
	assert.Contains(t, sql, " OR ")
	assert.Len(t, args, 2)
}

func TestQ_SimpleAnd(t *testing.T) {
	q := gormz.QAnd(
		gormz.Eq("status", "active"),
		gormz.Gt("age", 18),
	)

	sql, _ := q.ToSQL()
	assert.Contains(t, sql, " AND ")
}

func TestQ_Nested(t *testing.T) {
	q := gormz.Qb().And(
		gormz.QOr(
			gormz.Eq("status", "active"),
			gormz.Eq("status", "pending"),
		),
		gormz.Gt("age", 18),
	)

	sql, _ := q.ToSQL()
	assert.Contains(t, sql, "(")
	assert.Contains(t, sql, ")")
	assert.Contains(t, sql, " AND ")
	assert.Contains(t, sql, " OR ")
}

func TestQ_Not(t *testing.T) {
	q := gormz.QOr(
		gormz.Not(gormz.Eq("status", "deleted")),
		gormz.Eq("status", "active"),
	)

	sql, _ := q.ToSQL()
	assert.Contains(t, sql, "NOT (")
}

func TestQ_NotWithNestedQ(t *testing.T) {
	inner := gormz.QOr(
		gormz.Eq("a", 1),
		gormz.Eq("b", 2),
	)

	q := gormz.Qb().And(
		gormz.Not(inner),
		gormz.Eq("c", 3),
	)

	sql, _ := q.ToSQL()
	assert.Contains(t, sql, "NOT (")
}

func TestQ_Immutability(t *testing.T) {
	base := gormz.Qb().And(gormz.Eq("a", 1))
	extended := base.And(gormz.Eq("b", 2))

	sqlBase, _ := base.ToSQL()
	sqlExt, _ := extended.ToSQL()

	assert.NotEqual(t, sqlBase, sqlExt)
	assert.NotContains(t, sqlBase, "b = ?")
}

func TestQ_OrAfterAnd(t *testing.T) {
	q := gormz.Qb().
		And(gormz.Eq("a", 1)).
		Or(gormz.Eq("b", 2))

	sql, _ := q.ToSQL()
	assert.Contains(t, sql, " OR ")
}

func TestQ_Empty(t *testing.T) {
	q := gormz.Qb()
	assert.True(t, q.IsEmpty())
	assert.Equal(t, 0, q.Len())

	sql, _ := q.ToSQL()
	assert.Equal(t, "1=1", sql)
}

func TestQ_AndGroup(t *testing.T) {
	q := gormz.Qb().
		And(gormz.Eq("x", 1)).
		AndGroup(
			gormz.Eq("a", 2),
			gormz.Eq("b", 3),
		)

	sql, _ := q.ToSQL()
	// تحقق من وجود AND مع الأقواس (مقبولة)
	assert.Contains(t, sql, " AND ")
	assert.Contains(t, sql, "a = ?")
	assert.Contains(t, sql, "b = ?")
	// تحقق من البنية: مجموعة AND بين a و b
	assert.Regexp(t, `\(a = \?\) AND \(b = \?\)`, sql)
}

func TestQ_OrGroup(t *testing.T) {
	q := gormz.Qb().
		And(gormz.Eq("x", 1)).
		OrGroup(
			gormz.Eq("a", 2),
			gormz.Eq("b", 3),
		)

	sql, _ := q.ToSQL()
	assert.Contains(t, sql, "a = ?")
	assert.Contains(t, sql, "b = ?")
	assert.Regexp(t, `\(a = \?\) OR \(b = \?\)`, sql)
}
