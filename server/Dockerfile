# syntax=docker/dockerfile:1

FROM golang:1.25.7-alpine AS build

WORKDIR /src/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/netstamp-api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/netstamp-migrate ./cmd/migrate

FROM alpine:3.22

RUN addgroup -S netstamp && adduser -S -G netstamp netstamp

WORKDIR /app

COPY --from=build /out/netstamp-api /app/netstamp-api
COPY --from=build /out/netstamp-migrate /app/netstamp-migrate
COPY --from=build /src/server/db/migrations /app/db/migrations

USER netstamp:netstamp

EXPOSE 8080

ENTRYPOINT ["/app/netstamp-api"]
