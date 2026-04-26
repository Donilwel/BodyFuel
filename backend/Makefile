.PHONY: run stop restart build swag test lint

## Запустить всё (пересобрать образ если нужно)
run:
	docker compose up --build

## Запустить в фоне
up:
	docker compose up --build -d

## Остановить и удалить контейнеры
stop:
	docker compose down

## Перезапустить
restart:
	docker compose down && docker compose up --build

## Собрать только Docker-образ
build:
	docker compose build

## Сгенерировать Swagger-документацию
swag:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/main.go -o api-schemas
	@echo "Done"

## Запустить тесты
test:
	go test -race ./...

## Запустить линтер
lint:
	golangci-lint run --timeout=5m ./...
