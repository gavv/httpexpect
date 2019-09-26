GO111MODULE := on
export GO111MODULE

all: test check

test:
	go test
	cd _examples && go test

check:
	golangci-lint run .

fmt:
	gofmt -s -w . ./_examples

tidy:
	go mod tidy -v
	mv _examples examples && ( \
		cd examples ; \
		go mod tidy -v ; \
		go get -v -u=patch github.com/gavv/httpexpect ; \
			) && mv examples _examples
