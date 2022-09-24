lint:
	go mod tidy
	golangci-lint run ./...
run:
	go run ./...
up:
	docker-compose up -d

down:
	docker-compose down
test:
	go test -failfast -v ./...