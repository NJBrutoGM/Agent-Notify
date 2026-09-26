BINARY_NAME=agent-notify
DAEMON_NAME=agent-notifyd
VERSION?=dev
BUILD_DIR=build
DIST_DIR=dist

GO=go
GOFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all build build-daemon clean install test lint fmt vet tidy release

all: build build-daemon

build:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/agent-notify

build-daemon:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(DAEMON_NAME) ./cmd/agent-notifyd

build-all: build build-daemon
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/agent-notify
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/agent-notify
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/agent-notify
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/agent-notify
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/agent-notify
	GOOS=windows GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd/agent-notify

	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(DAEMON_NAME)-linux-amd64 ./cmd/agent-notifyd
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(DAEMON_NAME)-linux-arm64 ./cmd/agent-notifyd
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(DAEMON_NAME)-darwin-amd64 ./cmd/agent-notifyd
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(DAEMON_NAME)-darwin-arm64 ./cmd/agent-notifyd
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(DAEMON_NAME)-windows-amd64.exe ./cmd/agent-notifyd
	GOOS=windows GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(DAEMON_NAME)-windows-arm64.exe ./cmd/agent-notifyd

install: build
	install -Dm755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	install -Dm755 $(BUILD_DIR)/$(DAEMON_NAME) /usr/local/bin/$(DAEMON_NAME)
	mkdir -p /usr/local/share/agent-notify/assets/sounds
	cp -r assets/sounds/* /usr/local/share/agent-notify/assets/sounds/ 2>/dev/null || true
	mkdir -p /etc/agent-notify
	cp configs/config.yaml.example /etc/agent-notify/config.yaml.example

install-user: build
	mkdir -p ~/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	cp $(BUILD_DIR)/$(DAEMON_NAME) ~/.local/bin/
	mkdir -p ~/.config/agent-notify
	cp configs/config.yaml.example ~/.config/agent-notify/config.yaml.example
	mkdir -p ~/.local/share/agent-notify/assets/sounds
	cp -r assets/sounds/* ~/.local/share/agent-notify/assets/sounds/ 2>/dev/null || true

clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)

test:
	$(GO) test ./...

lint:
	golangci-lint run

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

deps:
	$(GO) mod download

# Generate default sounds (placeholder - replace with actual sound files)
generate-sounds:
	mkdir -p assets/sounds
	@echo "Place your .wav files in assets/sounds/"
	@echo "Required: success.wav, alert.wav, error.wav, warning.wav, info.wav"

release: tidy build-all
	mkdir -p $(DIST_DIR)
	cp $(BUILD_DIR)/* $(DIST_DIR)/
	cd $(DIST_DIR) && sha256sum * > checksums.txt

dev-run: build
	$(BUILD_DIR)/$(BINARY_NAME) work_done

dev-daemon: build-daemon
	$(BUILD_DIR)/$(DAEMON_NAME) -foreground