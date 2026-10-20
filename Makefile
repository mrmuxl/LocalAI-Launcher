BIN_DIR       := bin
LAUNCHER_NAME := LocalAI-launcher
LAUNCHER_CMD  := ./cmd/launcher

.PHONY: build-launcher build-release test lint fmt vet tidy

build-launcher:
	go build -o $(BIN_DIR)/$(LAUNCHER_NAME).exe $(LAUNCHER_CMD)

# Windows GUI 发布构建：-H windowsgui 隐藏控制台（cmd）窗口，双击 exe 直接出界面。
build-release:
	go build -ldflags "-H windowsgui" -o $(BIN_DIR)/$(LAUNCHER_NAME).exe $(LAUNCHER_CMD)

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
