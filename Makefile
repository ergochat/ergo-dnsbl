.PHONY: all build test gofmt

all: build

build:
	go build -v .

test:
	go test . && go vet .
	./.check-gofmt.sh

gofmt:
	./.check-gofmt.sh --fix
