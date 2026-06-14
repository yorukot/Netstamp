# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS server-build

WORKDIR /src/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN mkdir -p /out/agents \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/controller ./cmd/controller \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/agents/netstamp-agent-linux-amd64 ./cmd/agent \
    && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o /out/agents/netstamp-agent-linux-arm64 ./cmd/agent \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM node:24-alpine AS web-build

WORKDIR /src

RUN npm install -g pnpm@11.4.0

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json web/package.json
COPY packages/brand/package.json packages/brand/package.json
COPY packages/ui/package.json packages/ui/package.json

RUN pnpm install --frozen-lockfile --filter @netstamp/web... --filter @netstamp/ui

COPY web web
COPY packages/brand packages/brand
COPY packages/ui packages/ui

ARG VITE_NETSTAMP_REGISTRATION_ENABLED=true
ARG VITE_NETSTAMP_PROJECT_CREATION_ENABLED=true
ARG VITE_NETSTAMP_USER_CREDENTIAL_CHANGES_ENABLED=true
ARG VITE_NETSTAMP_DEMO_MODE=false
ARG VITE_NETSTAMP_DEMO_EMAIL=
ARG VITE_NETSTAMP_DEMO_PASSWORD=
ENV VITE_NETSTAMP_REGISTRATION_ENABLED=$VITE_NETSTAMP_REGISTRATION_ENABLED
ENV VITE_NETSTAMP_PROJECT_CREATION_ENABLED=$VITE_NETSTAMP_PROJECT_CREATION_ENABLED
ENV VITE_NETSTAMP_USER_CREDENTIAL_CHANGES_ENABLED=$VITE_NETSTAMP_USER_CREDENTIAL_CHANGES_ENABLED
ENV VITE_NETSTAMP_DEMO_MODE=$VITE_NETSTAMP_DEMO_MODE
ENV VITE_NETSTAMP_DEMO_EMAIL=$VITE_NETSTAMP_DEMO_EMAIL
ENV VITE_NETSTAMP_DEMO_PASSWORD=$VITE_NETSTAMP_DEMO_PASSWORD

RUN pnpm --filter @netstamp/web build

FROM alpine:3.22

RUN addgroup -S netstamp && adduser -S -G netstamp netstamp

WORKDIR /app

COPY --from=server-build /out/controller /app/controller
COPY --from=server-build /out/agents /app/agents
COPY --from=server-build /out/migrate /app/migrate
COPY --from=server-build /src/server/db/migrations /app/db/migrations
COPY --from=web-build /src/web/dist /app/web

ENV WEB_DIR=/app/web

USER netstamp:netstamp

EXPOSE 8080

ENTRYPOINT ["/app/controller"]
