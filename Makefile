build-cli:
	go build -o bin/gator-cli cmd/cli/main.go

build-server:
	go build -o bin/gator-server cmd/server/main.go

build-css:
	./tailwindcss -i static/css/input.css -o static/css/output.css

watch-css:
	./tailwindcss -i static/css/input.css -o static/css/output.css --watch

build: build-cli build-server build-css

run-cli:
	go run cmd/cli/main.go

run-server:
	$(MAKE) watch-css & air

migrate-up:
	goose -dir sql/schema postgres "postgres://nick.bernard:@localhost:5432/gator" up

migrate-down:
	goose -dir sql/schema postgres "postgres://nick.bernard:@localhost:5432/gator" down

reset:
	$(MAKE) migrate-down
	$(MAKE) migrate-up

generate:
	sqlc generate

test:
	go test -v ./...

db-start:
	brew services start postgresql@15

db-stop:
	brew services stop postgresql@15
