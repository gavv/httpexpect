GO111MODULE := on
export GO111MODULE

all: test check

test:
	go test . ./_examples

check:
	golangci-lint run ./...

fmt:
	gofmt -s -w . ./_examples
