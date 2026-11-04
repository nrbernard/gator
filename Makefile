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

hooks:
	git config core.hooksPath .githooks

lint:
	golangci-lint run ./...

# The checks the pre-commit hook and CI both run.
check:
	@unformatted=$$(gofmt -l $$(git ls-files '*.go')); \
	if [ -n "$$unformatted" ]; then echo "files need gofmt -w:"; echo "$$unformatted"; exit 1; fi
	go build ./...
	golangci-lint run ./...
	go test ./...
