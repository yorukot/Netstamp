FROM golang:1.25.7-alpine AS openapi-build

WORKDIR /src/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server ./
RUN mkdir -p /out \
    && go run ./cmd/openapi -output /out/openapi.json

FROM node:24-alpine AS app-build

WORKDIR /app

RUN npm install -g pnpm@11.0.8

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY docs/package.json docs/package.json
COPY web/package.json web/package.json
COPY packages/brand/package.json packages/brand/package.json
COPY packages/ui/package.json packages/ui/package.json

RUN pnpm install --frozen-lockfile --filter @netstamp/docs... --filter @netstamp/web... --filter @netstamp/ui

COPY docs docs
COPY --from=openapi-build /out/openapi.json docs/public/openapi.json
COPY web web
COPY packages/brand packages/brand
COPY packages/ui packages/ui

RUN pnpm --filter @netstamp/web build
RUN pnpm --filter @netstamp/docs build

FROM nginx:1.27-alpine

COPY deployments/docker/nginx.conf /etc/nginx/nginx.conf
COPY --from=app-build /app/web/dist /usr/share/nginx/web
COPY --from=app-build /app/docs/dist /usr/share/nginx/docs

EXPOSE 80
