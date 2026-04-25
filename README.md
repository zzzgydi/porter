# Porter - Private Docker Registry Platform

A lightweight private Docker image platform built on top of CNCF Distribution (`registry:3.1.0`), Cloudflare R2 (or filesystem for dev), Go API Server, Postgres, Redis, Nginx, and a Vite React console.

## Architecture

```
Docker CLI / CI / CD
        |
        | docker login / push / pull
        v
    Nginx (HTTPS)
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
  "insecure-registries": ["localhost:5080"]
}
```

(Settings -> Docker Engine -> edit JSON -> Apply & Restart)

### 1. Generate auth certificates

```bash
bash scripts/generate-auth-cert.sh
```

This creates `registry/certs/auth.key` and `registry/certs/auth.crt`.
The API server will auto-generate `jwks.json` on startup.

### 2. Prepare environment

```bash
cp .env.example .env
# Review and update values if needed. For local dev, defaults are fine.
```

### 3. Start the dev stack

```bash
docker compose --profile dev up -d --build
```

Services will be available at:
- Registry: `http://localhost:5080`
- Console: `http://localhost:5081`

### 4. Verify registry challenge

```bash
curl -I http://localhost:5080/v2/
```

Expected: `401 Unauthorized` with `WWW-Authenticate: Bearer ...`

### 5. Log in to the console

Open `http://localhost:5081` in your browser.

Default admin credentials (from `.env`):
- Email: `admin@example.com`
- Password: `change_me_admin_password`

### 6. Create a project and robot token

1. Go to **Projects** -> **New Project** -> name: `demo`
2. Go to **Robot Tokens** -> **New Token**
   - Project Name: `demo`
   - Token Name: `ci`
   - Permissions: `pull,push`
3. Copy the token (shown only once).

### 7. Docker login & push/pull

```bash
# Login
docker login localhost:5080
# Username: robot$demo-ci
# Password: <paste token>

# Push test
docker pull alpine:latest
docker tag alpine:latest localhost:5080/demo/alpine:latest
docker push localhost:5080/demo/alpine:latest

# Pull test
docker rmi localhost:5080/demo/alpine:latest
docker pull localhost:5080/demo/alpine:latest
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

Use certbot or any CA:

```bash
sudo certbot certonly --standalone \
  -d registry.example.com \
  -d console.registry.example.com
```

Copy to:
```
nginx/certs/fullchain.pem
nginx/certs/privkey.pem
```

### 3. Generate registry auth cert

```bash
bash scripts/generate-auth-cert.sh
```

### 4. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and replace all `change_me_*` placeholders with strong secrets.
Set `REGISTRY_DOMAIN` and `CONSOLE_DOMAIN` to your real domains.

### 5. Update registry config

Edit `registry/config.yml` (production config):
- Replace `YOUR_R2_*` placeholders with real values.
- Ensure `auth.token.realm` points to your public console domain.

### 6. Update Nginx config

Edit `nginx/conf.d/registry.conf`:
- Replace `registry.example.com` and `console.registry.example.com` with your domains.

### 7. Deploy

```bash
docker compose up -d --build
```

## Directory Structure

```
.
├── docker-compose.yml              # Production stack
├── docker-compose.dev.yml          # Dev override (filesystem storage, HTTP)
├── .env.example                    # Environment template
├── nginx/
│   ├── conf.d/registry.conf        # Production Nginx (HTTPS)
│   └── conf.d/registry.dev.conf    # Dev Nginx (HTTP)
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
- **Robot Tokens**: Username format `robot$<project>-<name>`. Password is a 32-byte random hex string, SHA-256 hashed in DB. Only shown once on creation.
- **Webhook**: Registry pushes events to Go API Server, which upserts projects/repositories/tags into Postgres.
- **Delete tag**: Console triggers registry manifest delete + Postgres soft delete.
- **GC**: Manual only in MVP. Use `scripts/gc.sh` as a guide.

## Security Checklist

- [ ] Replace all `change_me_*` secrets in `.env`
- [ ] Use real TLS certificates in production
- [ ] Restrict R2 API token to the registry bucket only
- [ ] Do not expose Postgres/Redis ports publicly
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
