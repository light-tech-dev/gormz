# gormz

> Type-safe, immutable ORM for Go — built on top of GORM.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/light-tech-dev/gormz.svg)](https://pkg.go.dev/github.com/light-tech-dev/gormz)
[![Test Coverage](https://img.shields.io/badge/coverage-90%25-brightgreen)]()

---

## 🎯 What is gormz?

**gormz** is a type-safe, immutable ORM layer on top of [GORM](https://gorm.io). It keeps GORM's power while offering a cleaner, more intuitive API.

```go
// ❌ GORM — verbose
var users []User
db.Model(&User{}).
    Where("age > ?", 18).
    Where("active = ?", true).
    Order("created_at DESC").
    Limit(10).
    Find(&users)

// ✅ gormz — clean, type-safe
users, _ := gormz.New[User]().
    Filter("age__gt", 18).
    Filter("active", true).
    OrderBy("-created_at").
    Limit(10).
    All()