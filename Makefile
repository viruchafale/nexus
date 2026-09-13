.PHONY: build run test fmt clean

build:
	go build -o nexus .

run:
	go run . --help

test:
	go test ./...

fmt:
	gofmt -l -w .
	go vet ./...

clean:
	rm -f nexus
