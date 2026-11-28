DATABASE_URL ?= postgres://admin:adminpassword@localhost:5434/murmur?sslmode=disable&connect_timeout=5
REDIS_URL ?= redis://localhost:6381/0
export DATABASE_URL
export REDIS_URL

.PHONY: up down run test seed load explain redis

up:
	docker compose up -d --wait

down:
	docker compose down

run:
	go build -o bin/api ./cmd/api
	./bin/api

test:
	go test -race ./...

#fill the database with a graph of 100,000 users (it empties it first)
seed:
	go run ./cmd/seed --users 100000 --seed 1

#run a k6 script from load/ , in Docker: make load S=feed
load:
	docker compose run --rm -e VUS -e READERS k6 run /load/$(S).js
#what postgres does to build one user's feed : make explain U=42
explain:
	@docker compose exec -T postgres psql -U admin -d murmur -v user=$(U) < db/explain/feed.sql \
		| grep -E "Scan|Execution"

# one Redis command, run in the container: make redis C="dbsize"
redis:
	@docker compose exec redis redis-cli $(C)
