PACKAGE_NAME = dockpulse
VERSION = $(shell cat VERSION)
BINARY = dockpulse

.PHONY: all clean clean-all build install reinstall remove purge check test lint info files run help version version-patch version-minor version-major update-changelog release release-check release-help release-tag release-patch release-minor release-major

all: build

# Show the current version.
version:
	@echo "Current version: $(VERSION)"

# Increment the version: patch for fixes, minor for features, major for breaking changes.
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

# Run tests.
test:
	@echo "Running Go tests..."
	go test -v ./...

# Run style and quality checks.
lint:
	@echo "Running code style and quality checks..."
	go vet ./...
	@if command -v revive >/dev/null 2>&1; then \
		revive -config revive.toml ./...; \
	elif [ -f "$$(go env GOPATH)/bin/revive" ]; then \
		$$(go env GOPATH)/bin/revive -config revive.toml ./...; \
	else \
		echo "revive not found, run: go install github.com/mgechev/revive@latest"; \
	fi

# Build the standalone Go binary.
build:
	@echo "Building $(BINARY) v$(VERSION)..."
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v$(VERSION)" -o $(BINARY) ./cmd/dockpulse
	@echo "Build complete: ./$(BINARY)"

# Run the application locally.
run:
	@echo "Running $(BINARY)..."
	go run ./cmd/dockpulse

# Update the changelog.
update-changelog:
	@echo "Updating debian/changelog with version $(VERSION)"
	@chmod +x update-changelog.sh
	@./update-changelog.sh

# Remove build artifacts.
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY)
	rm -rf dist bin deb_pkg_*
	@echo "Clean complete"

clean-all: clean
	@echo "Clean-all complete"

# Check release readiness.
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

# Create the release tag.
release-tag: release-check
	@CUR_VER=$$(cat VERSION); \
	echo "Creating git tag v$${CUR_VER}..."; \
	git tag -a "v$${CUR_VER}" -m "Release v$${CUR_VER}"; \
	echo "Tag v$${CUR_VER} successfully created!"; \
	echo ""; \
	echo "To publish the release and automatically build Linux binaries & .deb packages:"; \
	echo "  git push origin main --tags"

# Release the current version.
release: release-tag

# Release with a patch version increment.
release-patch: version-patch update-changelog
	@git add VERSION debian/changelog
	@git commit -m "Bump version to $$(cat VERSION)"
	@$(MAKE) release-tag

# Release with a minor version increment.
release-minor: version-minor update-changelog
	@git add VERSION debian/changelog
	@git commit -m "Bump version to $$(cat VERSION)"
	@$(MAKE) release-tag

# Release with a major version increment.
release-major: version-major update-changelog
	@git add VERSION debian/changelog
	@git commit -m "Bump version to $$(cat VERSION)"
	@$(MAKE) release-tag

# Show help.
help:
	@echo "Available commands:"
	@echo "  make version         - Show current version"
	@echo "  make test            - Run Go unit tests"
	@echo "  make lint            - Run Go style and quality checks (go vet + revive)"
	@echo "  make build           - Build standalone Go binary"
	@echo "  make run             - Run the application locally via go run"
	@echo "  make release         - Create release tag for current version v$(VERSION)"
	@echo "  make release-patch   - Bump patch version and create release tag"
	@echo "  make release-minor   - Bump minor version and create release tag"
	@echo "  make release-major   - Bump major version and create release tag"
	@echo "  make clean           - Clean build artifacts"
