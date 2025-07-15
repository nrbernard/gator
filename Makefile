build-server:
	go build -o bin/gator cmd/main.go

build-css:
	./tailwindcss -i static/css/input.css -o static/css/output.css

watch-css:
	./tailwindcss -i static/css/input.css -o static/css/output.css --watch

build: build-server build-css

run:
	$(MAKE) watch-css & air

migrate-up:
	goose -dir sql/schema sqlite3 data/gator.db up

migrate-down:
	goose -dir sql/schema sqlite3 data/gator.db down

reset:
	$(MAKE) migrate-down
	$(MAKE) migrate-up

generate:
	sqlc generate

test:
	go test -v ./...
