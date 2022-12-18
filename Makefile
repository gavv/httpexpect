all: tidy gen build lint test spell

tidy:
	go mod tidy -v
	cd _examples && go get -v -u github.com/gavv/httpexpect/v2
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
ifneq ($(shell which gotest),)
	gotest
	cd _examples && gotest
else
	go test
	cd _examples && go test
endif

short:
ifneq ($(shell which gotest),)
	gotest -short
else
	go test -short
endif

spell:
ifneq ($(shell which mdspell),)
	mdspell -a README.md
	sort .spelling -o .spelling
endif
