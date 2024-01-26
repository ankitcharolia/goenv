BIN_NAME := goenv
PKG := github.com/ankitcharolia/goenv
GOFILES := $(shell go list -f '{{range .GoFiles}}{{.}} {{end}}' ./...)

build: bin/$(BIN_NAME)

bin/$(BIN_NAME): $(GOFILES)
	go build -o $@ $(PKG)

build-cross: clean
	GOOS=linux  GOARCH=amd64 go build -o dist/$(BIN_NAME)_amd64_linux $(PKG)
	GOOS=linux  GOARCH=arm64 go build -o dist/$(BIN_NAME)_arm64_linux $(PKG)
	GOOS=linux  GOARCH=arm go build -o dist/$(BIN_NAME)_arm_linux $(PKG)
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BIN_NAME)_amd64_darwin $(PKG)
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BIN_NAME)_arm64_darwin $(PKG)

clean:
	rm -rf bin dist

deps:
	go get -t ./...
	go mod tidy

# test:
# 	go test -race -v ./...
