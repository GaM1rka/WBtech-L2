BINARY=calendar

.PHONY: all test fmt vet build run docker docker-run

all: fmt vet test build

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./... -count=1

build:
	go build -o $(BINARY) ./cmd/server

run: build
	./$(BINARY)

docker:
	sudo docker build -t calendar:latest .

docker-run: docker
	sudo docker run -p 8080:8080 --rm calendar:latest
