PACKAGE_NAME = system-monitor
VERSION = $(shell cat VERSION)
DEB_FILE = ../$(PACKAGE_NAME)_$(VERSION)_all.deb

.PHONY: all clean clean-all build install reinstall remove purge check test info files run dev dev-patch dev-minor help version version-patch version-minor version-major update-changelog release release-check release-help release-tag release-patch release-minor release-major

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
	@echo "Running tests..."
	python3 -m unittest discover tests

# Сборка пакета (на Linux через dpkg-buildpackage)
build:
	@echo "Building $(PACKAGE_NAME) version $(VERSION)"
	@if command -v dpkg-buildpackage >/dev/null 2>&1; then \
		if ! grep -q "$(PACKAGE_NAME) ($(VERSION))" debian/changelog; then \
			echo "WARNING: Version $(VERSION) not found in debian/changelog"; \
			echo "Run: make update-changelog"; \
		fi; \
		dpkg-buildpackage -us -uc -b; \
	else \
		echo "NOTE: dpkg-buildpackage is not available on this OS."; \
		echo "The .deb package will be built automatically by GitHub Actions when tag v$(VERSION) is pushed."; \
	fi

# Обновление changelog
update-changelog:
	@echo "Updating debian/changelog with version $(VERSION)"
	@chmod +x update-changelog.sh
	@./update-changelog.sh

# Очистка
clean:
	@echo "Cleaning build files..."
	rm -rf ../$(PACKAGE_NAME)_$(VERSION)_*.deb
	rm -rf ../$(PACKAGE_NAME)_$(VERSION)_*.buildinfo
	rm -rf ../$(PACKAGE_NAME)_$(VERSION)_*.changes
	rm -rf debian/files
	rm -rf debian/.debhelper
	rm -rf debian/$(PACKAGE_NAME)
	rm -rf debian/*.log
	@echo "Clean complete"

# Полная очистка (включая старые версии)
clean-all:
	@echo "Cleaning all build files..."
	rm -rf ../$(PACKAGE_NAME)_*.deb
	rm -rf ../$(PACKAGE_NAME)_*.buildinfo
	rm -rf ../$(PACKAGE_NAME)_*.changes
	rm -rf debian/files
	rm -rf debian/.debhelper
	rm -rf debian/$(PACKAGE_NAME)
	rm -rf debian/*.log
	@echo "Clean complete"

# Установка (только для Debian/Ubuntu)
install: build
	@if [ -f "$(DEB_FILE)" ]; then \
		echo "Installing $(PACKAGE_NAME) version $(VERSION)"; \
		sudo dpkg -i $(DEB_FILE); \
		sudo apt-get install -f -y; \
		echo "Installation complete"; \
	else \
		echo "Package $(DEB_FILE) not found. Run 'make build' on Debian/Ubuntu first."; \
	fi

# Проверка пакета
check:
	@echo "Checking package..."
	lintian $(DEB_FILE) || true

# Показать информацию о пакете
info:
	@echo "Package: $(PACKAGE_NAME)"
	@echo "Version: $(VERSION)"
	@echo "File: $(DEB_FILE)"
	@if [ -f "$(DEB_FILE)" ]; then \
		echo "Package exists: Yes"; \
		dpkg --info $(DEB_FILE); \
	else \
		echo "Package exists: No (run 'make build' on Linux first)"; \
	fi

# Проверка готовности к релизу
release-check:
	@echo "Checking release readiness..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Uncommitted changes exist. Commit or stash them before releasing."; \
		git status --short; \
		exit 1; \
	fi
	@if git tag | grep -q "^v$(VERSION)$$"; then \
		echo "Warning: Tag v$(VERSION) already exists."; \
	fi
	@echo "Working tree is clean. Ready for release v$(VERSION)."

# Создание git-тега релиза
release-tag: release-check
	@echo "Creating git tag v$(VERSION)..."
	git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@echo "Tag v$(VERSION) successfully created!"
	@echo ""
	@echo "To publish the release and automatically build .deb package:"
	@echo "  git push origin v$(VERSION)"

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

# Показать установленные файлы
files:
	@dpkg -L $(PACKAGE_NAME) 2>/dev/null || echo "Package not installed"

# Запуск приложения
run:
	@echo "Running $(PACKAGE_NAME)..."
	python3 main.py

# Помощь
help:
	@echo "Available commands:"
	@echo "  make version         - Show current version"
	@echo "  make test            - Run unit tests"
	@echo "  make run             - Run the application locally"
	@echo "  make release         - Create release tag for current version v$(VERSION)"
	@echo "  make release-patch   - Bump patch version (e.g. 0.1.8 -> 0.1.9) and create release tag"
	@echo "  make release-minor   - Bump minor version (e.g. 0.1.8 -> 0.2.0) and create release tag"
	@echo "  make release-major   - Bump major version (e.g. 0.1.8 -> 1.0.0) and create release tag"
	@echo "  make build           - Build DEB package (on Linux)"
	@echo "  make clean           - Clean build files"
