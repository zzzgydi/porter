# Porter - Private Docker Registry Platform

A lightweight private Docker image platform built on top of CNCF Distribution (`registry:3.1.0`), Cloudflare R2 (or filesystem for dev), Go API Server, Postgres, Redis, and a Vite React console.

## Architecture

```
Docker CLI / CI / CD
        |
        | docker login / push / pull
        v
    Host Nginx (HTTPS, optional)
        |
        +-- registry.example.com  -> registry:3.1.0 (Token Auth, R2/FS storage)
        |
        +-- console.example.com   -> Vite React + Go API Server
                                        |
                                        +-- Postgres (business data)
                                        +-- Redis (cache, rate limit)
                                        +-- Registry HTTP API (webhook, tag delete)
```

## Quick Start (Local Dev)

### Prerequisites

- Docker Desktop (macOS/Windows) or Docker Engine (Linux)
- For macOS: configure Docker Engine to accept the insecure registry:

```json
{
  "insecure-registries": ["localhost:5000"]
}
```

(Settings -> Docker Engine -> edit JSON -> Apply & Restart)

### 1. Generate auth certificates

```bash
bash scripts/generate-auth-cert.sh
```

This creates `registry/certs/auth.key`, `registry/certs/auth.crt`, and `registry/certs/jwks.json`.
If `jwks.json` is missing, the API server generates it from `auth.crt` on startup.

### 2. Prepare environment

```bash
cp .env.example .env
# Review and update values if needed. For local dev, defaults are fine.
```

### 3. Start the dev stack

```bash
docker compose -f docker-compose.dev.yml up -d --build
```

Services will be available at:
- Registry: `http://localhost:5000`
- Console: `http://localhost:4173`
- API: `http://localhost:3000`

### 4. Verify registry challenge

```bash
curl -I http://localhost:5000/v2/
```

Expected: `401 Unauthorized` with `WWW-Authenticate: Bearer ...`

### 5. Log in to the console

Open `http://localhost:4173` in your browser.

Default admin credentials (from `.env`):
- Email: `admin@example.com`
- Password: `change_me_admin_password`

### 6. Create a project and robot token

1. Go to **Projects** -> **New Project** -> name: `demo`
2. Go to **Robot Tokens** -> **New Token**
   - Project: `demo`
   - Token Name: `ci`
   - Permissions: check `pull` and `push`
3. Copy the token (shown only once).

### 7. Docker login & push/pull

```bash
# Login
docker login localhost:5000
# Username: robot$demo-ci
# Password: <paste token>

# Push test
docker pull alpine:latest
docker tag alpine:latest localhost:5000/demo/alpine:latest
docker push localhost:5000/demo/alpine:latest

# Pull test
docker rmi localhost:5000/demo/alpine:latest
docker pull localhost:5000/demo/alpine:latest
```

### 8. Check the console

Go to **Projects** -> **demo** -> click repository **alpine**.
You should see the `latest` tag with digest, size, and pushed time.

## Production Deployment

### 1. Prepare R2 bucket

Create a Cloudflare R2 bucket and API token with Object Read & Write permissions.

Record:
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET_NAME`

### 2. Prepare TLS certificates

Use certbot or any CA on your host machine. Configure your host nginx (or another reverse proxy) to terminate TLS and proxy to the Docker services:

```
registry.example.com  -> 127.0.0.1:5000
console.example.com   -> 127.0.0.1:4173
  /api/*              -> 127.0.0.1:3000
```

See `nginx.example.conf` for a reference nginx configuration.

### 3. Generate registry auth cert

```bash
bash scripts/generate-auth-cert.sh
```

### 4. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and replace all `change_me_*` placeholders with strong secrets.
Set `REGISTRY_PUBLIC_URL` and `CONSOLE_API_URL` to your real public URLs.

### 5. Review registry config

`registry/config.yml` is a template rendered at container startup via `envsubst`
— all values (R2 credentials, domains, secrets) come from `.env`, so there is
nothing to edit in the YAML itself. If you use different variable names, keep
`auth.token.realm` pointing to your public API token endpoint.

### 6. Deploy

```bash
docker compose up -d --build
```

## Directory Structure

```
.
├── docker-compose.yml              # Production stack
├── docker-compose.dev.yml          # Dev override (filesystem storage, HTTP)
├── .env.example                    # Environment template
├── nginx.example.conf              # Reference nginx config for host
├── registry/
│   ├── config.yml                  # Production registry config (R2)
│   ├── config.dev.yml              # Dev registry config (filesystem)
│   └── certs/                      # auth.key, auth.crt, jwks.json
├── scripts/
│   ├── generate-auth-cert.sh
│   ├── bootstrap-dev.sh
│   ├── backup-postgres.sh
│   └── gc.sh
├── api/                            # Go API Server
│   ├── Dockerfile
│   ├── go.mod
│   └── cmd/server/main.go
│   └── internal/...
└── console/                        # Vite React Console
    ├── Dockerfile
    ├── package.json
    └── src/...
```

## Key Design Decisions

- **Registry:3.1.0**: Uses `/etc/distribution/config.yml` (v3 path).
- **Storage**: Production uses R2 S3-compatible API; dev uses filesystem.
- **Auth**: Go API Server issues RS256 JWTs for registry token auth. Console uses HMAC session cookies.
- **Robot Tokens**: Username format `robot$<project>-<name>`. Password is a 32-byte random hex string, bcrypt-hashed in DB (legacy SHA-256 hashes still verify). Only shown once on creation. Scoped strictly to their own project.
- **Webhook**: Registry pushes events to Go API Server, which upserts projects/repositories/tags into Postgres.
- **Delete tag**: Console triggers registry manifest delete + Postgres soft delete.
- **GC**: Manual only in MVP. Use `scripts/gc.sh` as a guide.

## Security Checklist

- [ ] Replace all `change_me_*` secrets in `.env`
- [ ] Use real TLS certificates in production (host nginx)
- [ ] Restrict R2 API token to the registry bucket only
- [ ] Do not expose Postgres/Redis ports publicly (services bind to 127.0.0.1)
- [ ] Rotate `auth.key` periodically (requires reconfiguring registry)
- [ ] Enable 2FA for console users (future enhancement)

## API Endpoints

### Registry Token
```
GET /api/registry/token?service=...&scope=...
Authorization: Basic base64(username:password)
```

### Console Auth
```
POST /api/auth/login
POST /api/auth/logout
GET  /api/me
```

### Management
```
GET    /api/projects
POST   /api/projects
GET    /api/projects/:project
PATCH  /api/projects/:project
DELETE /api/projects/:project
GET    /api/projects/:project/members
POST   /api/projects/:project/members
DELETE /api/projects/:project/members/:userId
GET    /api/projects/:project/repositories
GET    /api/projects/:project/repositories/:repo
GET    /api/projects/:project/repositories/:repo/tags
GET    /api/projects/:project/repositories/:repo/tags/:tag
DELETE /api/projects/:project/repositories/:repo/tags/:tag
GET    /api/robot-tokens
POST   /api/robot-tokens
DELETE /api/robot-tokens/:id
GET    /api/users
POST   /api/users
DELETE /api/users/:id
GET    /api/audit-logs
```

### Internal Webhook
```
POST /internal/registry/events
Authorization: Bearer <WEBHOOK_SECRET>
```

## License

MIT
