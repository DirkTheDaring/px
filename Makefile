# Define variables
include config.mk
GO := go
MODULE := github.com/$(GITHUB_USERNAME)/$(REPO)
SRC_DIR := src
BUILD_DIR := build
RELEASE_DIR := release

BINARY_NAME := px

# Read the version from a shell script
VERSION=$(file < VERSION.txt)
GITHUB_TOKEN=$(file < .token)

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	#$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	cd $(SRC_DIR) && $(GO) build -o ../$(BUILD_DIR)/$(BINARY_NAME) .

# Run tests
test:
	$(GO) test -v $(MODULE)/...

# Create a release build
release:
	@echo "Building release $(BINARY_NAME)..."
	@mkdir -p $(RELEASE_DIR)
	cd $(SRC_DIR) && GOARCH=amd64 GOOS=linux   $(GO) build -o ../$(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64 .
	cd $(SRC_DIR) && GOARCH=arm64 GOOS=linux   $(GO) build -o ../$(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64 .
	cd $(SRC_DIR) && GOARCH=amd64 GOOS=darwin  $(GO) build -o ../$(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64 .
	cd $(SRC_DIR) && GOARCH=amd64 GOOS=windows $(GO) build -o ../$(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_windows_amd64.exe .

# Publish the release on GitHub using the official gh CLI
publish: release
	@echo "Authenticating with GitHub..."
	gh auth login --with-token < .token
	gh release create v$(VERSION) $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64 \
		$(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64 \
		$(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64 \
		$(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_windows_amd64.exe \
		--title "Release $(VERSION)" \
		--notes "Release version $(VERSION)"

# Clean up build and release directories
distclean:
	rm -rf $(BUILD_DIR)
	rm -rf $(RELEASE_DIR)
	rm -f src/px

.PHONY: build test release publish
