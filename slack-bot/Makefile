BINARY := slack-bot
BUILD_DIR := build

# OS/ARCH targets to cross-compile for
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	linux/arm \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

.PHONY: build build-all clean

# Build for the current OS/architecture
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) .

# Build for every platform in PLATFORMS
build-all:
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		out=$(BUILD_DIR)/$(BINARY)-$$os-$$arch$$ext; \
		echo "Building $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -o $$out . || exit 1; \
	done

clean:
	rm -rf $(BUILD_DIR)
