include .env
export

.PHONY: run build db-up db-down

## Запустить сервер (схема БД создаётся автоматически)
run:
	go run ./cmd/api/...

## Сборка бинарника
build:
	go build -o bin/karvon ./cmd/api/...

## Запустить PostgreSQL
db-up:
	docker compose up -d postgres

## Остановить PostgreSQL
db-down:
	docker compose down
