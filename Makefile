BINARY  := pirates-gold
VERSION := 1.2
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"
BUILD   := CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -trimpath

.PHONY: build clean tidy

build:
	$(BUILD) -o $(BINARY) .

clean:
	rm -f $(BINARY)

tidy:
	go mod tidy
