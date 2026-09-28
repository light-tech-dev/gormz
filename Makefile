# ═══════════════════════════════════════════════════════════════════
#  Makefile for gormz
#  Type-safe, immutable ORM for Go
# ═══════════════════════════════════════════════════════════════════

# ─── متغيرات ────────────────────────────────────────────────────────
BINARY_NAME    := gormz
MODULE         := github.com/light-tech-dev/gormz
GO             := go
GOFLAGS        :=
GOTEST         := $(GO) test
GOBENCH        := $(GO) test -bench
GOVET          := $(GO) vet
GOFMT          := gofmt
GOFUMPT        := gofumpt
GOIMPORTS      := goimports
GOLANGCI       := golangci-lint
GOSEC          := gosec
GOVULNCHECK    := govulncheck
STATICCHECK    := staticcheck
REVIVE         := revive

# إصدار Go
GO_VERSION     := 1.22

# ألوان للطباعة
RED            := \033[0;31m
GREEN          := \033[0;32m
YELLOW         := \033[1;33m
BLUE           := \033[0;34m
MAGENTA        := \033[0;35m
CYAN           := \033[0;36m
WHITE          := \033[0;37m
BOLD           := \033[1m
NC             := \033[0m # No Color

# مسارات
COVERAGE_FILE  := coverage.out
COVERAGE_HTML  := coverage.html
BENCH_FILE     := benchmark.txt
CPU_PROFILE    := cpu.prof
MEM_PROFILE    := mem.prof
DIST_DIR       := dist
BUILD_DIR      := build

# ═══════════════════════════════════════════════════════════════════
#  Help — عرض المساعدة
# ═══════════════════════════════════════════════════════════════════

.PHONY: help
help: ## عرض المساعدة
	@echo ""
	@echo "$(BOLD)$(CYAN)╔══════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BOLD)$(CYAN)║                      gormz Makefile                          ║$(NC)"
	@echo "$(BOLD)$(CYAN)╚══════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(BOLD)الأوامر المتاحة:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-22s$(NC) %s\n", $$1, $$2}'
	@echo ""

.DEFAULT_GOAL := help

# ═══════════════════════════════════════════════════════════════════
#  Setup — التهيئة
# ═══════════════════════════════════════════════════════════════════

.PHONY: setup
setup: ## تثبيت الأدوات المطلوبة
	@echo "$(BOLD)$(BLUE)🔧 Setting up development environment...$(NC)"
	@echo "$(YELLOW)▸ Installing Go tools...$(NC)"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install mvdan.cc/gofumpt@latest
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	@$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	@$(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	@$(GO) install github.com/mgechev/revive@latest
	@$(GO) install github.com/rakyll/hey@latest
	@$(GO) install github.com/cosmtrek/air@latest
	@echo "$(GREEN)✅ Setup complete!$(NC)"

.PHONY: install
install: ## تثبيت الاعتمادات
	@echo "$(BOLD)$(BLUE)📦 Installing dependencies...$(NC)"
	@$(GO) mod download
	@$(GO) mod verify
	@echo "$(GREEN)✅ Dependencies installed!$(NC)"

.PHONY: tools-check
tools-check: ## التحقق من الأدوات المطلوبة
	@echo "$(BOLD)$(BLUE)🔍 Checking tools...$(NC)"
	@command -v $(GOLANGCI) >/dev/null 2>&1 || { echo "$(RED)❌ golangci-lint not found$(NC)"; exit 1; }
	@command -v $(GOFUMPT) >/dev/null 2>&1 || { echo "$(RED)❌ gofumpt not found$(NC)"; exit 1; }
	@command -v $(GOSEC) >/dev/null 2>&1 || { echo "$(RED)❌ gosec not found$(NC)"; exit 1; }
	@command -v $(GOVULNCHECK) >/dev/null 2>&1 || { echo "$(RED)❌ govulncheck not found$(NC)"; exit 1; }
	@echo "$(GREEN)✅ All tools available!$(NC)"

# ═══════════════════════════════════════════════════════════════════
#  Build — البناء
# ═══════════════════════════════════════════════════════════════════

.PHONY: build
build: ## بناء المشروع
	@echo "$(BOLD)$(BLUE)🔨 Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./...
	@echo "$(GREEN)✅ Build complete!$(NC)"

.PHONY: build-all
build-all: ## بناء لكل الأنظمة
	@echo "$(BOLD)$(BLUE)🌍 Building for multiple platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@GOOS=linux   GOARCH=amd64 $(GO) build -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64     ./...
	@GOOS=linux   GOARCH=arm64 $(GO) build -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64     ./...
	@GOOS=darwin  GOARCH=amd64 $(GO) build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64    ./...
	@GOOS=darwin  GOARCH=arm64 $(GO) build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64    ./...
	@GOOS=windows GOARCH=amd64 $(GO) build -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./...
	@echo "$(GREEN)✅ Cross-build complete!$(NC)"

.PHONY: build-verbose
build-verbose: ## بناء مع تفاصيل
	@echo "$(BOLD)$(BLUE)🔨 Building (verbose)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -v $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./...
	@echo "$(GREEN)✅ Build complete!$(NC)"

.PHONY: rebuild
rebuild: clean build ## إعادة البناء من الصفر

# ═══════════════════════════════════════════════════════════════════
#  Test — الاختبارات
# ═══════════════════════════════════════════════════════════════════

.PHONY: test
test: ## تشغيل الاختبارات
	@echo "$(BOLD)$(BLUE)🧪 Running tests...$(NC)"
	@$(GOTEST) -v -race -timeout=5m ./...
	@echo "$(GREEN)✅ Tests passed!$(NC)"

.PHONY: test-short
test-short: ## اختبارات سريعة (بدون integration)
	@echo "$(BOLD)$(BLUE)⚡ Running short tests...$(NC)"
	@$(GOTEST) -short -v ./...
	@echo "$(GREEN)✅ Short tests passed!$(NC)"

.PHONY: test-cover
test-cover: ## اختبارات مع تقرير التغطية
	@echo "$(BOLD)$(BLUE)📊 Running tests with coverage...$(NC)"
	@$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -1
	@echo "$(GREEN)✅ Coverage report generated!$(NC)"

.PHONY: test-cover-html
test-cover-html: test-cover ## تقرير التغطية HTML
	@echo "$(BOLD)$(BLUE)🌐 Generating HTML coverage report...$(NC)"
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)✅ HTML report: $(COVERAGE_HTML)$(NC)"
	@command -v xdg-open >/dev/null 2>&1 && xdg-open $(COVERAGE_HTML) || true

.PHONY: test-cover-check
test-cover-check: test-cover ## التحقق من حد التغطية
	@echo "$(BOLD)$(BLUE)🎯 Checking coverage threshold...$(NC)"
	@COVERAGE=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	THRESHOLD=80; \
	echo "Coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < $$THRESHOLD" | bc -l) -eq 1 ]; then \
		echo "$(RED)❌ Coverage $$COVERAGE% is below threshold $$THRESHOLD%$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)✅ Coverage is above threshold!$(NC)"

.PHONY: test-bench
test-bench: ## تشغيل Benchmarks
	@echo "$(BOLD)$(BLUE)⚡ Running benchmarks...$(NC)"
	@$(GOBENCH)=. -benchmem -benchtime=1s ./... | tee $(BENCH_FILE)
	@echo "$(GREEN)✅ Benchmarks complete!$(NC)"

.PHONY: test-bench-compare
test-bench-compare: ## مقارنة الأداء (يحتاج benchstat)
	@command -v benchstat >/dev/null 2>&1 || $(GO) install golang.org/x/perf/cmd/benchstat@latest
	@$(GOBENCH)=. -benchmem -count=10 ./... > new.txt
	@benchstat old.txt new.txt 2>/dev/null || echo "$(YELLOW)⚠️ No old.txt found, save current as old.txt$(NC)"

.PHONY: test-race
test-race: ## اختبارات مع Race Detector
	@echo "$(BOLD)$(BLUE)🏁 Running tests with race detector...$(NC)"
	@$(GOTEST) -race -v -timeout=10m ./...
	@echo "$(GREEN)✅ Race tests passed!$(NC)"

.PHONY: test-verbose
test-verbose: ## اختبارات مفصلة
	@echo "$(BOLD)$(BLUE)🧪 Running verbose tests...$(NC)"
	@$(GOTEST) -v -count=1 -timeout=5m ./...

.PHONY: test-clean
test-clean: ## اختبارات بدون cache
	@echo "$(BOLD)$(BLUE)🧹 Running tests (no cache)...$(NC)"
	@$(GOTEST) -count=1 ./...

.PHONY: test-all
test-all: test-race test-cover-check test-bench ## كل الاختبارات

# ═══════════════════════════════════════════════════════════════════
#  Coverage — التغطية
# ═══════════════════════════════════════════════════════════════════

.PHONY: coverage
coverage: test-cover-html ## تقرير التغطية (alias)

.PHONY: coverage-func
coverage-func: ## عرض التغطية حسب الدوال
	@echo "$(BOLD)$(BLUE)📊 Coverage by function:$(NC)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | sort -t'%' -k2 -n | tail -20

.PHONY: coverage-pkg
coverage-pkg: ## التغطية حسب الحزمة
	@echo "$(BOLD)$(BLUE)📊 Coverage by package:$(NC)"
	@$(GOTEST) -coverprofile=$(COVERAGE_FILE).tmp ./... >/dev/null 2>&1
	@$(GO) tool cover -func=$(COVERAGE_FILE).tmp | grep -v "total:" | awk '{print $$1}' | sed 's|/[^/]*$$||' | sort | uniq -c | sort -rn
	@rm -f $(COVERAGE_FILE).tmp

# ═══════════════════════════════════════════════════════════════════
#  Lint — الفحص
# ═══════════════════════════════════════════════════════════════════

.PHONY: lint
lint: ## تشغيل كل الفحوصات
	@echo "$(BOLD)$(BLUE)🔍 Running linters...$(NC)"
	@$(GOLANGCI) run --config=.golangci.yml
	@echo "$(GREEN)✅ Linting complete!$(NC)"

.PHONY: lint-fix
lint-fix: ## إصلاح المشاكل تلقائياً
	@echo "$(BOLD)$(BLUE)🔧 Auto-fixing linting issues...$(NC)"
	@$(GOLANGCI) run --config=.golangci.yml --fix
	@echo "$(GREEN)✅ Auto-fix complete!$(NC)"

.PHONY: lint-verbose
lint-verbose: ## فحص مفصل
	@echo "$(BOLD)$(BLUE)🔍 Running verbose linters...$(NC)"
	@$(GOLANGCI) run --config=.golangci.yml -v

.PHONY: lint-fast
lint-fast: ## فحص سريع (بدون كل الـ linters)
	@echo "$(BOLD)$(BLUE)⚡ Running fast lint...$(NC)"
	@$(GOLANGCI) run --config=.golangci.yml --fast

.PHONY: lint-new
lint-new: ## فحص الأسطر الجديدة فقط
	@echo "$(BOLD)$(BLUE)🆕 Linting new code only...$(NC)"
	@$(GOLANGCI) run --config=.golangci.yml --new-from-rev=HEAD~1

.PHONY: staticcheck
staticcheck: ## تشغيل staticcheck
	@echo "$(BOLD)$(BLUE)🔍 Running staticcheck...$(NC)"
	@$(STATICCHECK) ./...

.PHONY: revive
revive: ## تشغيل revive
	@echo "$(BOLD)$(BLUE)🔍 Running revive...$(NC)"
	@$(REVIVE) -config revive.toml -formatter friendly ./...

.PHONY: vet
vet: ## تشغيل go vet
	@echo "$(BOLD)$(BLUE)🔍 Running go vet...$(NC)"
	@$(GOVET) ./...

.PHONY: sec
sec: ## فحص أمني
	@echo "$(BOLD)$(BLUE)🔒 Running security checks...$(NC)"
	@$(GOSEC) -fmt=text -out=$(BUILD_DIR)/gosec.txt ./... || true
	@echo "$(YELLOW)⚠️ Full report: $(BUILD_DIR)/gosec.txt$(NC)"

.PHONY: vulncheck
vulncheck: ## فحص الثغرات
	@echo "$(BOLD)$(BLUE)🛡️  Running govulncheck...$(NC)"
	@$(GOVULNCHECK) ./...

.PHONY: lint-all
lint-all: lint vet staticcheck revive sec vulncheck ## كل الفحوصات

# ═══════════════════════════════════════════════════════════════════
#  Format — التنسيق
# ═══════════════════════════════════════════════════════════════════

.PHONY: fmt
fmt: ## تنسيق الكود
	@echo "$(BOLD)$(BLUE)🎨 Formatting code...$(NC)"
	@$(GOFMT) -w -s .
	@$(GOFUMPT) -w .
	@$(GOIMPORTS) -w .
	@echo "$(GREEN)✅ Formatting complete!$(NC)"

.PHONY: fmt-check
fmt-check: ## التحقق من التنسيق (CI)
	@echo "$(BOLD)$(BLUE)🔍 Checking formatting...$(NC)"
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "$(RED)❌ Files need formatting:$(NC)"; \
		$(GOFMT) -l .; \
		exit 1; \
	fi
	@if [ -n "$$($(GOFUMPT) -l .)" ]; then \
		echo "$(RED)❌ Files need gofumpt:$(NC)"; \
		$(GOFUMPT) -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ All files formatted!$(NC)"

.PHONY: fmt-diff
fmt-diff: ## عرض الفروقات في التنسيق
	@echo "$(BOLD)$(BLUE)📝 Formatting diff:$(NC)"
	@gofmt -d . | head -50

# ═══════════════════════════════════════════════════════════════════
#  Mod — إدارة الوحدات
# ═══════════════════════════════════════════════════════════════════

.PHONY: mod-download
mod-download: ## تحميل الاعتمادات
	@$(GO) mod download

.PHONY: mod-tidy
mod-tidy: ## تنظيف go.mod
	@echo "$(BOLD)$(BLUE)🧹 Tidying go.mod...$(NC)"
	@$(GO) mod tidy
	@echo "$(GREEN)✅ go.mod tidied!$(NC)"

.PHONY: mod-verify
mod-verify: ## التحقق من الاعتمادات
	@$(GO) mod verify

.PHONY: mod-update
mod-update: ## تحديث الاعتمادات
	@echo "$(BOLD)$(BLUE)⬆️  Updating dependencies...$(NC)"
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "$(GREEN)✅ Dependencies updated!$(NC)"

.PHONY: mod-graph
mod-graph: ## عرض شجرة الاعتمادات
	@$(GO) mod graph | head -30

# ═══════════════════════════════════════════════════════════════════
#  Clean — التنظيف
# ═══════════════════════════════════════════════════════════════════

.PHONY: clean
clean: ## تنظيف الملفات المؤقتة
	@echo "$(BOLD)$(BLUE)🧹 Cleaning up...$(NC)"
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -f $(BENCH_FILE)
	@rm -f $(CPU_PROFILE) $(MEM_PROFILE)
	@rm -f old.txt new.txt
	@$(GO) clean -cache -testcache -modcache
	@echo "$(GREEN)✅ Cleanup complete!$(NC)"

.PHONY: clean-test
clean-test: ## تنظيف cache الاختبارات
	@$(GO) clean -testcache

.PHONY: clean-all
clean-all: clean ## تنظيف كل شيء

# ═══════════════════════════════════════════════════════════════════
#  Dev — التطوير
# ═══════════════════════════════════════════════════════════════════

.PHONY: watch
watch: ## مراقبة التغييرات وإعادة التشغيل
	@command -v air >/dev/null 2>&1 || $(GO) install github.com/cosmtrek/air@latest
	@air -c .air.toml

.PHONY: dev
dev: lint test ## فحص + اختبار
	@echo "$(GREEN)✅ Dev checks passed!$(NC)"

.PHONY: ci
ci: fmt-check lint test-cover-check build ## كل ما يشغله CI
	@echo "$(GREEN)✅ CI checks passed!$(NC)"

.PHONY: pre-commit
pre-commit: fmt lint-fast test-short ## قبل commit
	@echo "$(GREEN)✅ Pre-commit checks passed!$(NC)"

.PHONY: pre-push
pre-push: fmt-check lint test-race ## قبل push
	@echo "$(GREEN)✅ Pre-push checks passed!$(NC)"

# ═══════════════════════════════════════════════════════════════════
#  Profile — التحليل
# ═══════════════════════════════════════════════════════════════════

.PHONY: profile-cpu
profile-cpu: ## CPU profiling
	@echo "$(BOLD)$(BLUE)🔥 CPU profiling...$(NC)"
	@$(GOTEST) -bench=. -cpuprofile=$(CPU_PROFILE) ./...
	@$(GO) tool pprof -top $(CPU_PROFILE) | head -30
	@echo "$(YELLOW)💡 Interactive: go tool pprof $(CPU_PROFILE)$(NC)"

.PHONY: profile-mem
profile-mem: ## Memory profiling
	@echo "$(BOLD)$(BLUE)💾 Memory profiling...$(NC)"
	@$(GOTEST) -bench=. -memprofile=$(MEM_PROFILE) ./...
	@$(GO) tool pprof -top $(MEM_PROFILE) | head -30
	@echo "$(YELLOW)💡 Interactive: go tool pprof $(MEM_PROFILE)$(NC)"

.PHONY: profile-web
profile-web: ## فتح web interface للـ profile
	@command -v graphviz >/dev/null 2>&1 || echo "$(YELLOW)⚠️ Install graphviz for web view$(NC)"
	@$(GO) tool pprof -http=:8080 $(CPU_PROFILE)

.PHONY: profile-trace
profile-trace: ## Execution tracing
	@echo "$(BOLD)$(BLUE)📊 Creating trace...$(NC)"
	@$(GOTEST) -bench=. -trace=trace.out ./...
	@$(GO) tool trace trace.out
	@echo "$(YELLOW)💡 Opening trace viewer...$(NC)"

# ═══════════════════════════════════════════════════════════════════
#  Docs — التوثيق
# ═══════════════════════════════════════════════════════════════════

.PHONY: docs
docs: ## عرض التوثيق محلياً
	@echo "$(BOLD)$(BLUE)📚 Starting godoc server...$(NC)"
	@command -v godoc >/dev/null 2>&1 || $(GO) install golang.org/x/tools/cmd/godoc@latest
	@echo "$(GREEN)✅ Open: http://localhost:6060/pkg/$(MODULE)/$(NC)"
	@godoc -http=:6060

.PHONY: docs-serve
docs-serve: docs ## alias

.PHONY: docs-generate
docs-generate: ## توليد التوثيق
	@echo "$(BOLD)$(BLUE)📝 Generating documentation...$(NC)"
	@command -v gomarkdoc >/dev/null 2>&1 || $(GO) install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
	@mkdir -p docs/api
	@gomarkdoc -o docs/api/README.md ./...
	@echo "$(GREEN)✅ Docs generated!$(NC)"

.PHONY: docs-check
docs-check: ## التحقق من التوثيق
	@echo "$(BOLD)$(BLUE)🔍 Checking documentation...$(NC)"
	@if grep -r "TODO" --include="*.go" . | grep -v "_test.go" | grep -v "testdata"; then \
		echo "$(YELLOW)⚠️ Found TODOs in code$(NC)"; \
	fi
	@echo "$(GREEN)✅ Docs check complete!$(NC)"

# ═══════════════════════════════════════════════════════════════════
#  Release — الإصدارات
# ═══════════════════════════════════════════════════════════════════

.PHONY: version
version: ## عرض الإصدار الحالي
	@echo "$(BOLD)$(CYAN)Version:$(NC) $$(git describe --tags --always 2>/dev/null || echo 'dev')"
	@echo "$(BOLD)$(CYAN)Commit:$(NC)  $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "$(BOLD)$(CYAN)Date:$(NC)    $$(date +'%Y-%m-%d %H:%M:%S')"

.PHONY: changelog
changelog: ## عرض Changelog
	@echo "$(BOLD)$(CYAN)Recent changes:$(NC)"
	@git log --oneline -20 2>/dev/null || echo "No git history"

.PHONY: tag
tag: ## إنشاء tag جديد (usage: make tag VERSION=v0.2.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)❌ Usage: make tag VERSION=v0.2.0$(NC)"; \
		exit 1; \
	fi
	@echo "$(BOLD)$(BLUE)🏷️  Creating tag $(VERSION)...$(NC)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "$(GREEN)✅ Tag $(VERSION) created and pushed!$(NC)"

.PHONY: release-check
release-check: fmt-check lint test-race ## فحوصات قبل الإصدار
	@echo "$(GREEN)✅ Ready for release!$(NC)"

# ═══════════════════════════════════════════════════════════════════
#  Docker — الحاويات
# ═══════════════════════════════════════════════════════════════════

.PHONY: docker-build
docker-build: ## بناء Docker image
	@echo "$(BOLD)$(BLUE)🐳 Building Docker image...$(NC)"
	@docker build -t $(BINARY_NAME):latest .
	@echo "$(GREEN)✅ Docker image built!$(NC)"

.PHONY: docker-run
docker-run: ## تشغيل Docker container
	@docker run --rm -it $(BINARY_NAME):latest

# ═══════════════════════════════════════════════════════════════════
#  All — كل الأوامر
# ═══════════════════════════════════════════════════════════════════

.PHONY: all
all: clean install fmt lint test-cover-check build ## كل شيء
	@echo "$(GREEN)✅ All tasks complete!$(NC)"

.PHONY: quick
quick: fmt lint-fast test-short ## فحص سريع
	@echo "$(GREEN)✅ Quick check complete!$(NC)"

# ═══════════════════════════════════════════════════════════════════
#  CI Helpers — مساعدات CI
# ═══════════════════════════════════════════════════════════════════

.PHONY: ci-setup
ci-setup: ## تهيئة بيئة CI
	@$(GO) version
	@$(GO) mod download
	@$(GO) mod verify

.PHONY: ci-test
ci-test: ## اختبارات CI
	@$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...

.PHONY: ci-build
ci-build: ## بناء CI
	@$(GO) build -v ./...

# ═══════════════════════════════════════════════════════════════════
#  Utilities — أدوات مساعدة
# ═══════════════════════════════════════════════════════════════════

.PHONY: loc
loc: ## عدد الأسطر
	@echo "$(BOLD)$(CYAN)Lines of code:$(NC)"
	@find . -name "*.go" -not -path "./vendor/*" -not -path "./testdata/*" | xargs wc -l | tail -1

.PHONY: loc-detail
loc-detail: ## تفاصيل الأسطر
	@echo "$(BOLD)$(CYAN)LOC by file:$(NC)"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | sort -rn | head -20

.PHONY: deps
deps: ## عرض الاعتمادات
	@echo "$(BOLD)$(CYAN)Dependencies:$(NC)"
	@$(GO) list -m all | tail -n +2

.PHONY: outdated
outdated: ## الاعتمادات القديمة
	@echo "$(BOLD)$(CYAN)Outdated dependencies:$(NC)"
	@$(GO) list -u -m all | grep '\[' || echo "$(GREEN)✅ All up to date!$(NC)"

.PHONY: todo
todo: ## عرض TODOs
	@echo "$(BOLD)$(CYAN)TODOs:$(NC)"
	@grep -rn "TODO\|FIXME\|XXX\|HACK" --include="*.go" . | grep -v "_test.go" || echo "$(GREEN)✅ No TODOs!$(NC)"

.PHONY: stats
stats: ## إحصائيات المشروع
	@echo ""
	@echo "$(BOLD)$(CYAN)╔══════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BOLD)$(CYAN)║                    Project Statistics                        ║$(NC)"
	@echo "$(BOLD)$(CYAN)╚══════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(BOLD)Files:$(NC)       $$(find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "$(BOLD)Lines:$(NC)       $$(find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1 | awk '{print $$1}')"
	@echo "$(BOLD)Packages:$(NC)    $$($(GO) list ./... | wc -l)"
	@echo "$(BOLD)Tests:$(NC)       $$(find . -name '*_test.go' | wc -l)"
	@echo "$(BOLD)Deps:$(NC)        $$($(GO) list -m all | tail -n +2 | wc -l)"
	@echo ""

# ═══════════════════════════════════════════════════════════════════
#  Git Hooks — Hooks
# ═══════════════════════════════════════════════════════════════════

.PHONY: install-hooks
install-hooks: ## تثبيت git hooks
	@echo "$(BOLD)$(BLUE)🪝 Installing git hooks...$(NC)"
	@mkdir -p .git/hooks
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make pre-commit' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo '#!/bin/sh' > .git/hooks/pre-push
	@echo 'make pre-push' >> .git/hooks/pre-push
	@chmod +x .git/hooks/pre-push
	@echo "$(GREEN)✅ Git hooks installed!$(NC)"

.PHONY: uninstall-hooks
uninstall-hooks: ## إزالة git hooks
	@rm -f .git/hooks/pre-commit .git/hooks/pre-push
	@echo "$(GREEN)✅ Git hooks removed!$(NC)"