all: tidy gen build lint test spell

tidy:
	go mod tidy -v
	cd _examples && go mod tidy -v

gen:
	go generate

fmt:
	gofmt -s -w . ./_examples

build:
	go build
	cd _examples && go build

lint:
	golangci-lint run .

test:
	gotest
	cd _examples && gotest

spell:
	mdspell README.md
