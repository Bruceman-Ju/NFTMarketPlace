.PHONY: build run test docker-build docker-run clean

APP_NAME = nft-marketplace-backend

build:
	go build -o $(APP_NAME) ./main.go

run: build
	./$(APP_NAME)

test:
	go test ./... -v

docker-build:
	docker build -t $(APP_NAME) .

docker-run: docker-build
	docker run -p 8080:8080 --name $(APP_NAME) -d $(APP_NAME)

clean:
	rm -f $(APP_NAME)
	docker rm -f $(APP_NAME) 2>/dev/null || true

migrate:
	go run cmd/migrate.go  # 可选：单独迁移脚本

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build        - Build binary"
	@echo "  make run          - Build and run"
	@echo "  make test         - Run tests"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run in Docker"
	@echo "  make clean        - Clean binary and container"