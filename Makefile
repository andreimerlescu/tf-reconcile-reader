# Variables
MAIN_PATH=.
APP_NAME := $(shell basename "$(shell realpath $(MAIN_PATH))")
BIN_DIR=bin

# Go build flags
# -s: Strip symbols (reduces binary size)
# -w: Omit DWARF debugging information
LDFLAGS=-ldflags "-s -w"

.PHONY: all clean summary install darwin-amd64 darwin-amd64 linux-amd64 linux-arm64 windows-amd64

# Create build directory if it doesn't exist
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

# Build for all platforms
all: summary darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64 install

summary:
	@if ! command -v summarize > /dev/null; then \
		go install github.com/andreimerlescu/summarize@latest; \
	fi
	@summarize -i "go,Makefile,md,mod,LICENSE,gitignore" \
			   -s "summary.,.min.js,.min.css,.git/,.svn/,.vscode/,.vs/,.idea/,logs/,secrets/,summaries/,bin/"

install: $(BIN_DIR)
	@if [[ "$(shell go env GOOS)" == "windows" ]]; then \
		cp "$(BIN_DIR)/$(APP_NAME)-$(shell go env GOOS)-$(shell go env GOARCH).exe" "$(shell go env GOBIN)/$(APP_NAME).exe"; \
	else \
		cp "$(BIN_DIR)/$(APP_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)" "$(shell go env GOBIN)/$(APP_NAME)"; \
	fi
	@echo "NEW: $(shell which $(APP_NAME))"

# Build for macOS Intel (amd64)
darwin-amd64: $(BIN_DIR) summary
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "NEW: $(BIN_DIR)/$(APP_NAME)-darwin-amd64"

# Build for macOS Silicon (arm64)
darwin-arm64: $(BIN_DIR) summary
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "NEW: $(BIN_DIR)/$(APP_NAME)-darwin-amd64"

# Build for Linux ARM64
linux-arm64: $(BIN_DIR) summary
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "NEW: $(BIN_DIR)/$(APP_NAME)-darwin-arm64"

# Build for Linux AMD64
linux-amd64: $(BIN_DIR) summary
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "NEW: $(BIN_DIR)/$(APP_NAME)-linux-amd64"

# Build for Windows AMD64
windows-amd64: $(BIN_DIR) summary
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME).exe $(MAIN_PATH)
	@echo "NEW: $(BIN_DIR)/$(APP_NAME).exe"

# Clean build artifacts
clean:
	@rm -rf $(BIN_DIR)
	@echo "REMOVED: $(BIN_DIR)"

