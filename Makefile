PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin
VERSION = 2.0.0

# Build flags for smaller binary
LDFLAGS = -ldflags="-s -w -X main.VERSION=$(VERSION)"

.PHONY: all build clean install uninstall deps

all: build

deps:
	@echo "Initializing Go module..."
	@go mod init outputbuddy 2>/dev/null || true
	@echo "Getting dependencies..."
	@go get github.com/creack/pty
	@go get golang.org/x/term
	@go mod tidy

build: deps
	@echo "Building outputbuddy..."
	go build $(LDFLAGS) -o outputbuddy outputbuddy.go
	@echo "Build complete!"

install: build
	@echo "Installing to $(BINDIR)..."
	@install -d $(BINDIR)
	@install -m 755 outputbuddy $(BINDIR)/outputbuddy
	@ln -sf $(BINDIR)/outputbuddy $(BINDIR)/ob
	@echo "Installation complete!"
	@echo "You can now use 'outputbuddy' or 'ob'"

uninstall:
	@echo "Removing outputbuddy..."
	@rm -f $(BINDIR)/outputbuddy $(BINDIR)/ob
	@echo "Uninstall complete!"

clean:
	@rm -f outputbuddy
	@rm -f go.mod go.sum
	@echo "Cleaned build artifacts"

# Development helpers
test: build
	@echo "Running basic tests..."
	./outputbuddy 2+1=test.log 2+1 -- echo "Test message"
	@cat test.log
	@rm -f test.log
	@echo "Tests passed!"

# Cross-compilation targets
build-all: build-linux build-macos build-windows

build-linux: deps
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o outputbuddy-linux-amd64 outputbuddy.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o outputbuddy-linux-arm64 outputbuddy.go

build-macos: deps
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o outputbuddy-darwin-amd64 outputbuddy.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o outputbuddy-darwin-arm64 outputbuddy.go

build-windows: deps
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o outputbuddy-windows-amd64.exe outputbuddy.go