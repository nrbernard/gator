ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o bin/gator cmd/main.go


FROM debian:bookworm

COPY --from=builder /usr/src/app/bin/gator /usr/local/bin/
CMD ["gator"]
