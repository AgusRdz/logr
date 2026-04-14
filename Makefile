.PHONY: build test coverage clean cross install changelog release release-patch release-minor release-major _require-git-cliff

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	docker compose run --rm dev go build -ldflags="$(LDFLAGS)" -o bin/logr .

test:
	docker compose run --rm dev go test ./... -v

coverage:
	docker compose run --rm dev go test -coverprofile=coverage.out ./...
	docker compose run --rm dev go tool cover -func=coverage.out

clean:
	rm -rf bin/

# Platform detection
UNAME_S := $(shell uname -s)
ifeq ($(findstring MINGW,$(UNAME_S)),MINGW)
  GOOS ?= windows
else ifeq ($(findstring MSYS,$(UNAME_S)),MSYS)
  GOOS ?= windows
else ifeq ($(findstring Darwin,$(UNAME_S)),Darwin)
  GOOS ?= darwin
else
  GOOS ?= linux
endif
GOARCH ?= $(if $(filter arm64 aarch64,$(shell uname -m)),arm64,amd64)
EXT := $(if $(filter windows,$(GOOS)),.exe,)
BINARY := bin/logr$(EXT)

ifeq ($(GOOS),windows)
  INSTALL_DIR ?= $(LOCALAPPDATA)/Programs/logr
else
  INSTALL_DIR ?= $(HOME)/.local/bin
endif

install:
	docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='$(LDFLAGS)' -o $(BINARY) ."
	@mkdir -p "$(INSTALL_DIR)"
	cp $(BINARY) "$(INSTALL_DIR)/logr$(EXT)"
	@echo "installed logr $(VERSION) ($(GOOS)/$(GOARCH)) to $(INSTALL_DIR)/logr$(EXT)"

cross:
	docker compose run --rm dev sh -c "\
		CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o bin/logr-linux-amd64 . && \
		CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o bin/logr-linux-arm64 . && \
		CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o bin/logr-darwin-amd64 . && \
		CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o bin/logr-darwin-arm64 . && \
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o bin/logr-windows-amd64.exe ."

# Changelog - requires git-cliff on host (only for local releases; CI uses the action)
.PHONY: _require-git-cliff
_require-git-cliff:
	@command -v git-cliff >/dev/null 2>&1 || { echo "git-cliff required. See https://git-cliff.org/docs/installation"; exit 1; }

changelog: _require-git-cliff
	git-cliff --output CHANGELOG.md

CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
MAJOR := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | cut -d. -f1)
MINOR := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | cut -d. -f2)
PATCH := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | cut -d. -f3)

# Release targets: tag locally and push. CI handles CHANGELOG.md via git-cliff-action.
# git-cliff is NOT required on the host.

release:
	@BUMP=patch; \
	if git log $$(git describe --tags --abbrev=0 2>/dev/null || echo HEAD)..HEAD --format="%s" | grep -qE '^feat(\(.*\))?!:'; then BUMP=major; \
	elif git log $$(git describe --tags --abbrev=0 2>/dev/null || echo HEAD)..HEAD --format="%B" | grep -q 'BREAKING CHANGE'; then BUMP=major; \
	elif git log $$(git describe --tags --abbrev=0 2>/dev/null || echo HEAD)..HEAD --format="%s" | grep -qE '^feat'; then BUMP=minor; fi; \
	echo "detected: $$BUMP"; \
	$(MAKE) release-$$BUMP

release-patch:
	@NEXT=v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1))); \
	echo "$(CURRENT_TAG) -> $$NEXT"; \
	git tag $$NEXT && \
	{ git push origin HEAD $$NEXT && echo "released $$NEXT"; } || { git tag -d $$NEXT; echo "push failed - tag rolled back"; exit 1; }

release-minor:
	@NEXT=v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0; \
	echo "$(CURRENT_TAG) -> $$NEXT"; \
	git tag $$NEXT && \
	{ git push origin HEAD $$NEXT && echo "released $$NEXT"; } || { git tag -d $$NEXT; echo "push failed - tag rolled back"; exit 1; }

release-major:
	@NEXT=v$(shell echo $$(($(MAJOR)+1))).0.0; \
	echo "$(CURRENT_TAG) -> $$NEXT"; \
	git tag $$NEXT && \
	{ git push origin HEAD $$NEXT && echo "released $$NEXT"; } || { git tag -d $$NEXT; echo "push failed - tag rolled back"; exit 1; }
