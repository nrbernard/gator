build-cli:
	go build -o bin/gator-cli cmd/cli/main.go

build-server:
	go build -o bin/gator-server cmd/server/main.go

build: build-cli build-server

run-cli:
	go run cmd/cli/main.go

run-server:
	air

migrate-up:
	goose -dir sql/schema postgres "postgres://nick.bernard:@localhost:5432/gator" up

migrate-down:
	goose -dir sql/schema postgres "postgres://nick.bernard:@localhost:5432/gator" down

reset:
	$(MAKE) migrate-down
	$(MAKE) migrate-up

generate:
	sqlc generate
