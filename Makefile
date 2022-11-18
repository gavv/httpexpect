GO111MODULE := on
export GO111MODULE

all: build lint test

build:
	go build
	cd _examples && go build

lint:
	golangci-lint run .

test:
	gotest
	cd _examples && gotest

gen:
	go generate

fmt:
	gofmt -s -w . ./_examples

tidy:
	go mod tidy -v
	mv _examples examples && ( \
		cd examples ; \
		go mod tidy -v ; \
		go get -v -u=patch github.com/gavv/httpexpect ; \
			) && mv examples _examples
