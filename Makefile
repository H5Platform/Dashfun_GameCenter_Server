.PHONY: build-linux build clean

# Output directory
BUILD_DIR := build

# Default target (Linux build)
build-linux:
	CGO_ENABLED=0 GOARCH="amd64" GOOS="linux" go build -o $(BUILD_DIR)/

# Local build
build:
	go build -o $(BUILD_DIR)/

clean:
	rm -rf $(BUILD_DIR)
