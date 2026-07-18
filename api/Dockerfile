FROM node:22-alpine AS console-builder

RUN corepack enable && corepack prepare pnpm@9.5.0 --activate

WORKDIR /console

COPY console/package.json console/pnpm-lock.yaml console/.npmrc ./
RUN pnpm install --frozen-lockfile

COPY console/ ./

# The console is served by the API at the same origin, so its default API base
# URL is empty. Keep the argument for installations that deliberately use a
# separate API origin.
ARG VITE_API_BASE_URL
ARG VITE_REGISTRY_PUBLIC_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_REGISTRY_PUBLIC_URL=${VITE_REGISTRY_PUBLIC_URL}
RUN pnpm run build

FROM golang:1.23-alpine AS api-builder

WORKDIR /api

RUN apk add --no-cache git ca-certificates

COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -ldflags='-s -w' -o /registry-api ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=api-builder /registry-api /usr/local/bin/registry-api
COPY --from=console-builder /console/dist /app/console

EXPOSE 3000

CMD ["registry-api"]
