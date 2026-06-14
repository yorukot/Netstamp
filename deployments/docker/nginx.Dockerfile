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
COPY web web
COPY packages/brand packages/brand
COPY packages/ui packages/ui

ARG VITE_NETSTAMP_REGISTRATION_ENABLED=true
ARG VITE_NETSTAMP_PROJECT_CREATION_ENABLED=true
ARG VITE_NETSTAMP_USER_CREDENTIAL_CHANGES_ENABLED=true
ENV VITE_NETSTAMP_REGISTRATION_ENABLED=$VITE_NETSTAMP_REGISTRATION_ENABLED
ENV VITE_NETSTAMP_PROJECT_CREATION_ENABLED=$VITE_NETSTAMP_PROJECT_CREATION_ENABLED
ENV VITE_NETSTAMP_USER_CREDENTIAL_CHANGES_ENABLED=$VITE_NETSTAMP_USER_CREDENTIAL_CHANGES_ENABLED

RUN pnpm --filter @netstamp/web build
RUN pnpm --filter @netstamp/docs build

FROM nginx:1.27-alpine

COPY deployments/docker/nginx.conf /etc/nginx/nginx.conf
COPY --from=app-build /app/web/dist /usr/share/nginx/web
COPY --from=app-build /app/docs/dist /usr/share/nginx/docs

EXPOSE 80
