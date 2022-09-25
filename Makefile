lint:
	go mod tidy
	golangci-lint run ./...
run:
	go run ./...
up:
	docker-compose up -d
down:
	docker-compose down
test: up
	go test -failfast -v ./...