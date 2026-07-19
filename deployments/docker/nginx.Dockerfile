FROM node:24-alpine AS app-build

WORKDIR /app

RUN npm install -g pnpm@11.0.8

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

RUN pnpm --filter @netstamp/i18n build \
    && pnpm --filter @netstamp/web build

FROM nginx:1.27-alpine

COPY deployments/docker/nginx.conf /etc/nginx/nginx.conf
COPY --from=app-build /app/web/dist /usr/share/nginx/web

EXPOSE 80
