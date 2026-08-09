# syntax=docker/dockerfile:1
# check=skip=SecretsUsedInArgOrEnv

ARG NETSTAMP_VERSION=0.1.0

FROM golang:1.26-alpine AS server-build

ARG NETSTAMP_VERSION

WORKDIR /src/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN mkdir -p /out/agents \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/controller ./cmd/controller \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/yorukot/netstamp/internal/agent/runtime.Version=${NETSTAMP_VERSION} -X github.com/yorukot/netstamp/internal/agent/runtime.AgentString=netstamp-probe/${NETSTAMP_VERSION}" -o /out/agents/netstamp-agent-linux-amd64 ./cmd/agent \
    && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w -X github.com/yorukot/netstamp/internal/agent/runtime.Version=${NETSTAMP_VERSION} -X github.com/yorukot/netstamp/internal/agent/runtime.AgentString=netstamp-probe/${NETSTAMP_VERSION}" -o /out/agents/netstamp-agent-linux-arm64 ./cmd/agent \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM node:24-alpine AS web-build

WORKDIR /src

RUN npm install -g pnpm@11.4.0

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json web/package.json
COPY packages/brand/package.json packages/brand/package.json
COPY packages/i18n/package.json packages/i18n/package.json
COPY packages/ui/package.json packages/ui/package.json

RUN pnpm install --frozen-lockfile --filter @netstamp/web... --filter @netstamp/ui

COPY web web
COPY packages/brand packages/brand
COPY packages/i18n packages/i18n
COPY packages/ui packages/ui

ARG VITE_NETSTAMP_DEMO_EMAIL=
ARG VITE_NETSTAMP_DEMO_PASSWORD=
ENV VITE_NETSTAMP_DEMO_EMAIL=$VITE_NETSTAMP_DEMO_EMAIL
ENV VITE_NETSTAMP_DEMO_PASSWORD=$VITE_NETSTAMP_DEMO_PASSWORD

ARG VITE_NETSTAMP_GA_MEASUREMENT_ID=
ARG VITE_NETSTAMP_GOOGLE_TAG_ID=
ARG VITE_NETSTAMP_CLARITY_PROJECT_ID=
ARG VITE_NETSTAMP_META_PIXEL_ID=
ARG VITE_NETSTAMP_FACEBOOK_PIXEL_ID=
ARG VITE_NETSTAMP_POSTHOG_KEY=
ARG VITE_NETSTAMP_POSTHOG_HOST=
ARG VITE_NETSTAMP_PLAUSIBLE_DOMAIN=
ARG VITE_NETSTAMP_PLAUSIBLE_SCRIPT_URL=
ARG VITE_NETSTAMP_UMAMI_WEBSITE_ID=
ARG VITE_NETSTAMP_UMAMI_SCRIPT_URL=
ARG VITE_NETSTAMP_TRACKING_CONSENT_MODE=regional
ARG VITE_NETSTAMP_TRACKING_CONSENT_COUNTRIES=
ENV VITE_NETSTAMP_GA_MEASUREMENT_ID=$VITE_NETSTAMP_GA_MEASUREMENT_ID
ENV VITE_NETSTAMP_GOOGLE_TAG_ID=$VITE_NETSTAMP_GOOGLE_TAG_ID
ENV VITE_NETSTAMP_CLARITY_PROJECT_ID=$VITE_NETSTAMP_CLARITY_PROJECT_ID
ENV VITE_NETSTAMP_META_PIXEL_ID=$VITE_NETSTAMP_META_PIXEL_ID
ENV VITE_NETSTAMP_FACEBOOK_PIXEL_ID=$VITE_NETSTAMP_FACEBOOK_PIXEL_ID
ENV VITE_NETSTAMP_POSTHOG_KEY=$VITE_NETSTAMP_POSTHOG_KEY
ENV VITE_NETSTAMP_POSTHOG_HOST=$VITE_NETSTAMP_POSTHOG_HOST
ENV VITE_NETSTAMP_PLAUSIBLE_DOMAIN=$VITE_NETSTAMP_PLAUSIBLE_DOMAIN
ENV VITE_NETSTAMP_PLAUSIBLE_SCRIPT_URL=$VITE_NETSTAMP_PLAUSIBLE_SCRIPT_URL
ENV VITE_NETSTAMP_UMAMI_WEBSITE_ID=$VITE_NETSTAMP_UMAMI_WEBSITE_ID
ENV VITE_NETSTAMP_UMAMI_SCRIPT_URL=$VITE_NETSTAMP_UMAMI_SCRIPT_URL
ENV VITE_NETSTAMP_TRACKING_CONSENT_MODE=$VITE_NETSTAMP_TRACKING_CONSENT_MODE
ENV VITE_NETSTAMP_TRACKING_CONSENT_COUNTRIES=$VITE_NETSTAMP_TRACKING_CONSENT_COUNTRIES

RUN pnpm --filter @netstamp/web build

FROM alpine:3.22

ARG NETSTAMP_VERSION

RUN addgroup -S netstamp && adduser -S -G netstamp netstamp

WORKDIR /app

COPY --from=server-build /out/controller /app/controller
COPY --from=server-build /out/agents /app/agents
COPY --from=server-build /out/migrate /app/migrate
COPY --from=server-build /src/server/db/migrations /app/db/migrations
COPY --from=web-build /src/web/dist /app/web

ENV APP_VERSION=$NETSTAMP_VERSION \
    WEB_DIR=/app/web

USER netstamp:netstamp

EXPOSE 8080

ENTRYPOINT ["/app/controller"]
