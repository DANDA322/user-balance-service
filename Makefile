lint:
	go mod tidy
	golangci-lint run ./...

build:
	go build ./...

run:
	docker-compose up -d

up:
	docker-compose up -d db

down:
	docker-compose down

test: up
	go test -failfast -v ./...