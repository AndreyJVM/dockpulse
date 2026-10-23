PACKAGE_NAME = system-monitor
VERSION = $(shell cat VERSION)
BINARY = system-monitor

.PHONY: all clean clean-all build install reinstall remove purge check test info files run help version version-patch version-minor version-major update-changelog release release-check release-help release-tag release-patch release-minor release-major

all: build

# Проверка версии
version:
	@echo "Current version: $(VERSION)"

# Увеличение версии (patch - исправления, minor - новые функции, major - критические изменения)
version-patch:
	$(eval NEW_VERSION=$(shell echo $(VERSION) | awk -F. '{print $$1"."$$2"."$$3+1}'))
	@echo $(NEW_VERSION) > VERSION
	@echo "Version bumped to $(NEW_VERSION)"

version-minor:
	$(eval NEW_VERSION=$(shell echo $(VERSION) | awk -F. '{print $$1"."$$2+1".0"}'))
	@echo $(NEW_VERSION) > VERSION
	@echo "Version bumped to $(NEW_VERSION)"

version-major:
	$(eval NEW_VERSION=$(shell echo $(VERSION) | awk -F. '{print $$1+1".0.0"}'))
	@echo $(NEW_VERSION) > VERSION
	@echo "Version bumped to $(NEW_VERSION)"

# Запуск тестов
test:
	@echo "Running Go tests..."
	go test -v ./...

# Сборка автономного бинарника Go
build:
	@echo "Building $(BINARY) v$(VERSION)..."
	go build -ldflags="-s -w -X main.version=v$(VERSION)" -o $(BINARY) ./cmd/system-monitor
	@echo "Build complete: ./$(BINARY)"

# Запуск приложения локально
run:
	@echo "Running $(BINARY)..."
	go run ./cmd/system-monitor

# Обновление changelog
update-changelog:
	@echo "Updating debian/changelog with version $(VERSION)"
	@chmod +x update-changelog.sh
	@./update-changelog.sh

# Очистка
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY)
	rm -rf dist bin deb_pkg_*
	@echo "Clean complete"

clean-all: clean
	@echo "Clean-all complete"

# Проверка готовности к релизу
release-check:
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Uncommitted changes exist. Commit or stash them before releasing."; \
		git status --short; \
		exit 1; \
	fi
	@CUR_VER=$$(cat VERSION); \
	if git tag | grep -q "^v$${CUR_VER}$$"; then \
		echo "Warning: Tag v$${CUR_VER} already exists."; \
	fi; \
	echo "Working tree is clean. Ready for release v$${CUR_VER}."

# Создание git-тега релиза
release-tag: release-check
	@CUR_VER=$$(cat VERSION); \
	echo "Creating git tag v$${CUR_VER}..."; \
	git tag -a "v$${CUR_VER}" -m "Release v$${CUR_VER}"; \
	echo "Tag v$${CUR_VER} successfully created!"; \
	echo ""; \
	echo "To publish the release and automatically build cross-platform binaries & .deb packages:"; \
	echo "  git push origin main --tags"

# Полный релиз текущей версии
release: release-tag

# Релиз с увеличением patch-версии
release-patch: version-patch update-changelog
	@git add VERSION debian/changelog
	@git commit -m "Bump version to $$(cat VERSION)"
	@$(MAKE) release-tag

# Релиз с увеличением minor-версии
release-minor: version-minor update-changelog
	@git add VERSION debian/changelog
	@git commit -m "Bump version to $$(cat VERSION)"
	@$(MAKE) release-tag

# Релиз с увеличением major-версии
release-major: version-major update-changelog
	@git add VERSION debian/changelog
	@git commit -m "Bump version to $$(cat VERSION)"
	@$(MAKE) release-tag

# Помощь
help:
	@echo "Available commands:"
	@echo "  make version         - Show current version"
	@echo "  make test            - Run Go unit tests"
	@echo "  make build           - Build standalone Go binary"
	@echo "  make run             - Run the application locally via go run"
	@echo "  make release         - Create release tag for current version v$(VERSION)"
	@echo "  make release-patch   - Bump patch version and create release tag"
	@echo "  make release-minor   - Bump minor version and create release tag"
	@echo "  make release-major   - Bump major version and create release tag"
	@echo "  make clean           - Clean build artifacts"
