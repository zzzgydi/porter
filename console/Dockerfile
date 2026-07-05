FROM node:22-alpine AS builder

RUN corepack enable && corepack prepare pnpm@9.5.0 --activate

WORKDIR /app

COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile

COPY . .
ARG VITE_API_BASE_URL
ARG VITE_REGISTRY_PUBLIC_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_REGISTRY_PUBLIC_URL=${VITE_REGISTRY_PUBLIC_URL}
RUN pnpm run build

FROM node:22-alpine

WORKDIR /app

RUN npm install -g serve

COPY --from=builder /app/dist ./dist

EXPOSE 4173

CMD ["serve", "-s", "dist", "-l", "4173"]
