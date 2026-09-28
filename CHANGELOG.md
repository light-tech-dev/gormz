# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Hooks System (BeforeCreate, AfterCreate, etc.)
- Relations API (HasMany, BelongsTo, ManyToMany)
- Multi-tenancy support
- Audit integration

### Changed
- `Filter` now validates field names more strictly

### Deprecated
- `MustRegister` → use `Register`

### Removed
- Nothing

### Fixed
- Nil pointer when passing nil to `Filter`

### Security
- Updated dependency X to fix CVE-XXXX-XXXX

## [0.1.0] - 2025-01-XX

### Added
- Initial release
- QuerySet[T] with generics
- modern-style lookups
- Q Builder
- Registry
- Pagination
- Advanced features (CTE, Window, Union, Lock, Retry)
- Multi-DB support
- 90% test coverage
- Complete documentation

[Unreleased]: https://github.com/light-tech-dev/gormz/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/light-tech-dev/gormz/releases/tag/v0.1.0