ARG GO_VERSION=1

FROM alpine:3.19 as tailwind
WORKDIR /app
RUN apk add --no-cache nodejs npm
RUN npm install tailwindcss @tailwindcss/cli
COPY static/css/input.css ./static/css/input.css
COPY internal/views ./internal/views
RUN npx @tailwindcss/cli -i ./static/css/input.css -o ./static/css/output.css --minify

FROM golang:${GO_VERSION}-bookworm as builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o bin/gator cmd/main.go

FROM golang:${GO_VERSION}-bookworm as goose-builder
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y make ca-certificates && rm -rf /var/lib/apt/lists/* && \
    mkdir -p /data && chmod 755 /data

COPY --from=builder /usr/src/app/bin/gator ./main
COPY --from=goose-builder /go/bin/goose /usr/local/bin/goose
COPY --from=tailwind /app/static/css/output.css ./static/css/output.css
COPY static ./static
COPY internal/views ./internal/views
COPY Makefile ./
COPY sql/schema ./sql/schema

ENV DATABASE_PATH=/data/gator.db
EXPOSE 8080
CMD ["./main"]
