BIN := $(shell pwd)/bin
BUF_VERSION := v1.15.1

.PHONY: install-tools
install-tools:
	curl -sSL "https://github.com/bufbuild/buf/releases/download/${BUF_VERSION}/buf-$(shell uname -s)-$(shell uname -m)" -o "${BIN}/buf"
	chmod +x "${BIN}/buf"
	${BIN}/buf mod update
	go mod tidy

.PHONY: generate
	rm -rf ./gen/*
	${BIN}/buf generate

.PHONY: serve
serve: generate
	go run cmd/echo/main.go server

.PHONY: build
build: generate
	go build -o echo-server ./cmd/echo-server/main.go
