.PHONY: build test vet clean install lint all

BINARY   := swcr
PACKAGE  := github.com/dengmengmian/swcr-go/cmd/swcr
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -s -w
LDFLAGS  += -X '$(PACKAGE).version=$(VERSION)'
LDFLAGS  += -X '$(PACKAGE).commit=$(COMMIT)'
LDFLAGS  += -X '$(PACKAGE).buildDate=$(DATE)'

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/swcr

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

lint:
	go vet ./...
	test -z "$$(gofmt -l . 2>/dev/null)" || (echo "run: gofmt -w ." && exit 1)

clean:
	rm -f $(BINARY)
	rm -f code.docx

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/swcr
