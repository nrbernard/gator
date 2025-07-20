build-server:
	go build -o bin/gator cmd/main.go

build-css:
	./tailwindcss -i static/css/input.css -o static/css/output.css

watch-css:
	./tailwindcss -i static/css/input.css -o static/css/output.css --watch

build: build-server build-css

run:
	$(MAKE) build-css
	$(MAKE) watch-css &
	air

migrate-up:
	goose -dir sql/schema sqlite3 $(DATABASE_PATH) up

migrate-down:
	goose -dir sql/schema sqlite3 $(DATABASE_PATH) down

reset:
	$(MAKE) migrate-down
	$(MAKE) migrate-up

generate:
	sqlc generate

test:
	go test -v ./...
