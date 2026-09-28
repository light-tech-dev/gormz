package gormz

import (
	"errors"
	"fmt"

	"github.com/light-tech-dev/gormz/internal"
	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════
// Errors — الأخطاء الأساسية
// ═══════════════════════════════════════════════

var (
	// ErrNotFound يُرجع عند البحث عن سجل غير موجود.
	ErrNotFound = gorm.ErrRecordNotFound

	// ErrNotInitialized يُرجع لو SetDB لم يُنادى.
	ErrNotInitialized = errors.New("gormz: DB not initialized")

	// ErrNilDB يُرجع لو تم تمرير nil.
	ErrNilDB = errors.New("gormz: nil DB")

	// ErrInvalidField يُرجع عند حقل غير صحيح.
	ErrInvalidField = errors.New("gormz: invalid field")

	// ErrInvalidQuery يُرجع عند استعلام غير صحيح.
	ErrInvalidQuery = errors.New("gormz: invalid query")

	// ErrDangerousOperation يُرجع عند عملية خطرة بدون conditions.
	ErrDangerousOperation = errors.New("gormz: dangerous operation without conditions")

	// ErrAlreadyRegistered يُرجع عند تسجيل موديل بنفس الاسم مرتين.
	ErrAlreadyRegistered = errors.New("gormz: model already registered")

	// ErrNotFoundInRegistry يُرجع عند البحث عن موديل غير مسجّل.
	ErrNotFoundInRegistry = errors.New("gormz: model not found in registry")
)

// ═══════════════════════════════════════════════
// Typed Errors — أخطاء مع سياق
// ═══════════════════════════════════════════════

// NotFoundError يحمل معلومات عن السجل غير الموجود.
type NotFoundError struct {
	Model string
	ID    any
}

func (e *NotFoundError) Error() string {
	if e.ID == nil {
		return fmt.Sprintf("gormz: %s not found", e.Model)
	}
	return fmt.Sprintf("gormz: %s with id %v not found", e.Model, e.ID)
}

// Unwrap يسمح باستخدام errors.Is(err, ErrNotFound).
func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

// NewNotFoundError ينشئ NotFoundError.
func NewNotFoundError(model string, id any) *NotFoundError {
	return &NotFoundError{Model: model, ID: id}
}

// ValidationError يحمل معلومات عن حقل غير صحيح.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("gormz: invalid field %q: %s", e.Field, e.Reason)
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidField
}

// NewValidationError ينشئ ValidationError.
func NewValidationError(field, reason string) *ValidationError {
	return &ValidationError{Field: field, Reason: reason}
}

// DangerousOperationError يحمل معلومات عن عملية خطرة.
type DangerousOperationError struct {
	Operation string
	Reason    string
}

func (e *DangerousOperationError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("gormz: dangerous operation %q", e.Operation)
	}
	return fmt.Sprintf("gormz: dangerous operation %q: %s", e.Operation, e.Reason)
}

func (e *DangerousOperationError) Unwrap() error {
	return ErrDangerousOperation
}

// NewDangerousError ينشئ DangerousOperationError.
func NewDangerousError(operation, reason string) *DangerousOperationError {
	return &DangerousOperationError{Operation: operation, Reason: reason}
}

// ═══════════════════════════════════════════════
// Error Checks — أدوات فحص
// ═══════════════════════════════════════════════

// IsNotFound يفحص إذا كان الخطأ ErrNotFound.
//
// يدعم NotFoundError أيضًا.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsValidation يفحص إذا كان الخطأ ValidationError.
func IsValidation(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve) || errors.Is(err, ErrInvalidField)
}

// IsDangerous يفحص إذا كان الخطأ DangerousOperationError.
func IsDangerous(err error) bool {
	var de *DangerousOperationError
	return errors.As(err, &de) || errors.Is(err, ErrDangerousOperation)
}

// IsAlreadyRegistered يفحص إذا كان الخطأ ErrAlreadyRegistered.
func IsAlreadyRegistered(err error) bool {
	return errors.Is(err, ErrAlreadyRegistered)
}

// IsNotFoundInRegistry يفحص إذا كان الخطأ ErrNotFoundInRegistry.
func IsNotFoundInRegistry(err error) bool {
	return errors.Is(err, ErrNotFoundInRegistry)
}

// AsNotFound يحوّل الخطأ إلى NotFoundError إذا أمكن.
func AsNotFound(err error) (*NotFoundError, bool) {
	var ne *NotFoundError
	if errors.As(err, &ne) {
		return ne, true
	}
	return nil, false
}

// AsValidation يحوّل الخطأ إلى ValidationError إذا أمكن.
func AsValidation(err error) (*ValidationError, bool) {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve, true
	}
	return nil, false
}

// AsDangerous يحوّل الخطأ إلى DangerousOperationError إذا أمكن.
func AsDangerous(err error) (*DangerousOperationError, bool) {
	var de *DangerousOperationError
	if errors.As(err, &de) {
		return de, true
	}
	return nil, false
}

// ═══════════════════════════════════════════════
// Field Validation — Wrappers للاستخدام الخارجي
// ═══════════════════════════════════════════════

// ValidateField يتحقق من صحة اسم حقل.
//
// يقبل:
//   - snake_case: user_id
//   - dotted: users.id
//   - prefixed: t1.name
//
// يرفض:
//   - أسماء فارغة
//   - أحرف غير صالحة (;، '، ")
//   - كلمات SQL محجوزة (select, drop, ...)
//
// مثال:
//
//	if err := gormz.ValidateField(field); err != nil {
//	    return err
//	}
func ValidateField(field string) error {
	return internal.ValidateField(field)
}

// ValidateFields يتحقق من عدة حقول.
func ValidateFields(fields ...string) error {
	return internal.ValidateFields(fields...)
}

// ValidateLookup يتحقق من صحة اسم lookup.
//
// مثال:
//
//	gormz.ValidateLookup("gt")     // nil
//	gormz.ValidateLookup("bad")    // error
func ValidateLookup(lookup string) error {
	return internal.ValidateLookup(lookup)
}

// ValidateTableName يتحقق من صحة اسم جدول.
func ValidateTableName(table string) error {
	return internal.ValidateTableName(table)
}

// ValidateOperator يتحقق من صحة operator.
func ValidateOperator(op string) error {
	return internal.ValidateOperator(op)
}

// ═══════════════════════════════════════════════
// Must Wrappers — للاستخدام الداخلي/السريع
// ═══════════════════════════════════════════════

// MustValidateField يتحقق وpanic عند الخطأ.
//
// ⚠️ استخدم بحذر — للاستخدام الداخلي.
func MustValidateField(field string) {
	if err := ValidateField(field); err != nil {
		panic(err)
	}
}

// MustValidateFields يتحقق وpanic عند الخطأ.
func MustValidateFields(fields ...string) {
	if err := ValidateFields(fields...); err != nil {
		panic(err)
	}
}
