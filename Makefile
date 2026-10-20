BIN_DIR       := bin
LAUNCHER_NAME := LocalAI-launcher
LAUNCHER_CMD  := ./cmd/launcher

# 版本：构建时从 git 取（有 tag 用 tag 如 v0.2.0，无 tag 用短 commit hash，
# 工作区脏则加 -dirty 后缀）；可用 make build-release VERSION=0.2.0 覆盖。
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.1.0)
LDFLAGS := -X llmlauncher/internal/version.Version=$(VERSION)

.PHONY: build-launcher build-release test lint fmt vet tidy

build-launcher:
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(LAUNCHER_NAME).exe $(LAUNCHER_CMD)

# Windows GUI 发布构建：-H windowsgui 隐藏控制台（cmd）窗口，双击 exe 直接出界面。
build-release:
	go build -ldflags "-H windowsgui $(LDFLAGS)" -o $(BIN_DIR)/$(LAUNCHER_NAME).exe $(LAUNCHER_CMD)

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

vet:
	go vet ./...

tidy:
	go mod tidy
