package gormz

import "time"

// ═══════════════════════════════════════════════
// Config — إعدادات الاتصال
// ═══════════════════════════════════════════════

// Config يضبط إعدادات connection pool والسلوك العام.
type Config struct {
	// Connection pool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Logging
	LogLevel  LogLevel
	SlowQuery time.Duration

	// Behavior
	PrepareStmt bool // تحضير الاستعلامات (أسرع)
	DryRun      bool // عدم التنفيذ فعليًا (للتشخيص)
}

// LogLevel مستوى تسجيل الاستعلامات.
type LogLevel int

const (
	LogSilent LogLevel = iota
	LogError
	LogWarn
	LogInfo
)

// String يعرض اسم المستوى.
func (l LogLevel) String() string {
	switch l {
	case LogSilent:
		return "silent"
	case LogError:
		return "error"
	case LogWarn:
		return "warn"
	case LogInfo:
		return "info"
	}
	return "unknown"
}

// DefaultConfig إعدادات افتراضية معقولة للإنتاج.
func DefaultConfig() Config {
	return Config{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        LogSilent,
		SlowQuery:       200 * time.Millisecond,
		PrepareStmt:     false,
		DryRun:          false,
	}
}

// DevelopmentConfig إعدادات مناسبة للتطوير.
func DevelopmentConfig() Config {
	return Config{
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		LogLevel:        LogInfo,
		SlowQuery:       100 * time.Millisecond,
		PrepareStmt:     false,
		DryRun:          false,
	}
}

// TestingConfig إعدادات للاختبارات.
func TestingConfig() Config {
	return Config{
		MaxOpenConns:    5,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Minute,
		LogLevel:        LogSilent,
		SlowQuery:       0,
		PrepareStmt:     false,
		DryRun:          false,
	}
}