# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build

WORKDIR /src/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN mkdir -p /out \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/controller ./cmd/controller \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.22

RUN addgroup -S netstamp && adduser -S -G netstamp netstamp

WORKDIR /app

COPY --from=build /out/controller /app/controller
COPY --from=build /out/migrate /app/migrate
COPY --from=build /src/server/db/migrations /app/db/migrations

USER netstamp:netstamp

EXPOSE 8080

ENTRYPOINT ["/app/controller"]
