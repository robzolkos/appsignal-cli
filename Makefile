.PHONY: build test lint e2e install clean

build:
	go build -o appsignal ./cmd/appsignal

test:
	go test ./...

lint:
	golangci-lint run

e2e:
	go test -tags=e2e ./e2e/...

install:
	go install ./cmd/appsignal

clean:
	rm -f appsignal
