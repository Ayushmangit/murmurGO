DATABASE_URL ?= postgres://admin:adminpassword@localhost:5434/murmur?sslmode=disable&connect_timeout=5
export DATABASE_URL

.PHONY: up down run test

up:
	docker compose up -d --wait

down:
	docker compose down

run:
	go build -o bin/api ./cmd/api
	./bin/api

test:
	go test -race ./...
