build:
	go build -o bin/gator main.go

run:
	go run main.go

migrate-up:
	goose -dir sql/schema postgres "postgres://nick.bernard:@localhost:5432/gator" up

migrate-down:
	goose -dir sql/schema postgres "postgres://nick.bernard:@localhost:5432/gator" down

reset:
	$(MAKE) migrate-down
	$(MAKE) migrate-up

generate:
	sqlc generate
