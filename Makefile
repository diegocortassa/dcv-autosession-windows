MAIN_NAME=dcv-autosession

DIST_DIR=dist

# Version information
VERSION=$(shell (git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0-dev") | sed 's/^v//')
# Windows version resource requires numeric X.X.X.X format
VERSION_NUM=$(shell echo $(VERSION) | sed 's/[^0-9.]*$$//')
RELEASE=1
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build variables
BINARY_NAME=$(MAIN_NAME)
GO=$(shell which go)
GOFMT=$(shell which gofmt)
GOFILES=$(shell find . -name "*.go")
LDFLAGS="-X github.com/diegocortassa/dcv-autosession-windows/internal/version.Version=$(VERSION) \
         -X github.com/diegocortassa/dcv-autosession-windows/internal/version.Commit=$(COMMIT) \
         -X github.com/diegocortassa/dcv-autosession-windows/internal/version.BuildTime=$(BUILD_TIME)"

WINDOWS_AMD64_BINARY=$(MAIN_NAME).exe
WINDOWS_AMD64_DIR=$(MAIN_NAME)-v$(VERSION)-windows_amd64

# Windows installer variables
NSIS=makensis
NSIS_SCRIPT=contrib/nsis/installer.nsi
INSTALLER_DIR=installer
INSTALLER_NAME=$(BINARY_NAME)-$(VERSION)-setup.exe

# Build for Windows
.PHONY: build
build:
	mkdir -p $(DIST_DIR)/$(WINDOWS_AMD64_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags $(LDFLAGS) -o $(DIST_DIR)/$(WINDOWS_AMD64_DIR)/$(WINDOWS_AMD64_BINARY) ./cmd/$(MAIN_NAME)
	cp README.md LICENSE.md $(DIST_DIR)/$(WINDOWS_AMD64_DIR)/
	cd $(DIST_DIR) && 7z a -bd -r $(WINDOWS_AMD64_DIR).zip $(WINDOWS_AMD64_DIR)

## audit: run quality control checks
.PHONY: audit
audit:
	$(GO) mod tidy -diff
	$(GO) mod verify
	test -z "$(shell gofmt -l .)"
	$(GO) vet ./...
	$(GO) run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000 ./...
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Show version
.PHONY: version
version:
	@echo $(VERSION)

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(DIST_DIR)

# Run the application
.PHONY: run
run: build
	./$(DIST_DIR)/$(WINDOWS_AMD64_DIR)/$(WINDOWS_AMD64_BINARY)

# Run tests
.PHONY: test
test:
	$(GO) test ./... ;

# Install dependencies
.PHONY: deps
deps:
	$(GO) mod tidy ;
	$(GO) mod verify ;

################# WINDOWS INSTALLER #################

# Build Windows installer
.PHONY: installer
installer: build
	@echo "Building Windows installer..."
	mkdir -p $(DIST_DIR)/$(INSTALLER_DIR)
	cd contrib/nsis && $(NSIS) -DVERSION=$(VERSION) -DVERSION_NUM=$(VERSION_NUM) installer.nsi
	mv contrib/nsis/dcv-autosession-setup.exe $(DIST_DIR)/$(INSTALLER_DIR)/$(INSTALLER_NAME)
	@echo "Installer created: $(DIST_DIR)/$(INSTALLER_DIR)/$(INSTALLER_NAME)"

