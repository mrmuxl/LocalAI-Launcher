BIN_DIR       := bin
LAUNCHER_NAME := ninfer-launcher
LAUNCHER_CMD  := ./cmd/launcher

.PHONY: build-launcher test lint fmt vet tidy

build-launcher:
	go build -o $(BIN_DIR)/$(LAUNCHER_NAME).exe $(LAUNCHER_CMD)

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
