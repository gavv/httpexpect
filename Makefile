all: tidy gen build lint test

tidy:
	go mod tidy -v
	cd _examples && go mod tidy -v

gen:
	go generate

build:
	go build
	cd _examples && go build

lint:
	golangci-lint run .

test:
	gotest
	cd _examples && gotest

fmt:
	gofmt -s -w . ./_examples
