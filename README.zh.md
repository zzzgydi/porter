# Porter - 私有 Docker 镜像仓库平台

一个轻量级的私有 Docker 镜像平台，基于 CNCF Distribution (`registry:3.1.0`)、Cloudflare R2（或开发环境本地文件系统）、Go API 服务、Postgres、Redis 和 Vite React 控制台构建。

## 架构

```
Docker CLI / CI / CD
        |
        | docker login / push / pull
        v
    宿主机 Nginx (HTTPS, 可选)
        |
        +-- registry.example.com  -> registry:3.1.0 (Token 认证, R2/FS 存储)
        |
        +-- console.example.com   -> Vite React + Go API 服务
                                        |
                                        +-- Postgres (业务数据)
                                        +-- Redis (缓存, 限流)
                                        +-- Registry HTTP API (Webhook, 删除 Tag)
```

## 快速开始（本地开发）

### 前置条件

- Docker Desktop (macOS/Windows) 或 Docker Engine (Linux)
- macOS 用户：配置 Docker Engine 允许非安全仓库：

```json
{
  "insecure-registries": ["localhost:5000"]
}
```

（设置 -> Docker Engine -> 编辑 JSON -> 应用并重启）

### 1. 生成认证证书

```bash
bash scripts/generate-auth-cert.sh
```

这会创建 `registry/certs/auth.key`、`registry/certs/auth.crt` 和 `registry/certs/jwks.json`。
如果 `jwks.json` 缺失，API 服务启动时会根据 `auth.crt` 自动生成。

### 2. 准备环境变量

```bash
cp .env.example .env
# 检查并更新数值。本地开发使用默认值即可。
```

### 3. 启动开发环境

```bash
docker compose -f docker-compose.dev.yml up -d --build
```

服务将可通过以下地址访问：
- Registry: `http://localhost:5000`
- Console: `http://localhost:4173`
- API: `http://localhost:3000`

### 4. 验证 Registry 挑战

```bash
curl -I http://localhost:5000/v2/
```

预期结果：`401 Unauthorized`，响应头包含 `WWW-Authenticate: Bearer ...`

### 5. 登录控制台

浏览器打开 `http://localhost:4173`。

默认管理员账号（来自 `.env`）：
- 邮箱：`admin@example.com`
- 密码：`change_me_admin_password`

### 6. 创建项目和机器人令牌

1. 进入 **Projects** -> **New Project** -> 名称：`demo`
2. 进入 **Robot Tokens** -> **New Token**
   - Project: `demo`
   - Token Name: `ci`
   - Permissions: 勾选 `pull` 和 `push`
3. 复制令牌（仅显示一次）

### 7. Docker 登录与推送/拉取

```bash
# 登录
docker login localhost:5000
# Username: robot$demo-ci
# Password: <粘贴令牌>

# 推送测试
docker pull alpine:latest
docker tag alpine:latest localhost:5000/demo/alpine:latest
docker push localhost:5000/demo/alpine:latest

# 拉取测试
docker rmi localhost:5000/demo/alpine:latest
docker pull localhost:5000/demo/alpine:latest
```

### 8. 查看控制台

进入 **Projects** -> **demo** -> 点击仓库 **alpine**。
你应该能看到 `latest` 标签的摘要、大小和推送时间。

## 生产环境部署

### 1. 准备 R2 存储桶

在 Cloudflare R2 创建一个存储桶，并创建具有「对象读写」权限的 API 令牌。

记录以下信息：
- `R2_ACCOUNT_ID`
- `R2_ACCESS_KEY_ID`
- `R2_SECRET_ACCESS_KEY`
- `R2_BUCKET_NAME`

### 2. 准备 TLS 证书

在宿主机上使用 certbot 或任意 CA 签发证书。配置宿主机 Nginx（或其他反向代理）终止 TLS，并将流量代理到 Docker 服务：

```
registry.example.com  -> 127.0.0.1:5000
console.example.com   -> 127.0.0.1:4173
  /api/*              -> 127.0.0.1:3000
```

参考 `nginx.example.conf` 获取 Nginx 配置示例。

### 3. 生成 Registry 认证证书

```bash
bash scripts/generate-auth-cert.sh
```

### 4. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env`，将所有 `change_me_*` 替换为强密码。
将 `REGISTRY_PUBLIC_URL` 和 `CONSOLE_API_URL` 设置为你的真实公网地址。

### 5. 检查 Registry 配置

`registry/config.yml` 是模板文件，容器启动时通过 `envsubst` 渲染——所有取值
（R2 凭证、域名、密钥）都来自 `.env`，无需手动编辑 YAML。如果你修改了变量名，
请确保 `auth.token.realm` 指向你的公网 API 令牌端点。

### 6. 部署

```bash
docker compose -f docker-compose.yml up -d --build
```

## 目录结构

```
.
├── docker-compose.yml              # 生产环境
├── docker-compose.dev.yml          # 开发环境（文件系统存储，HTTP）
├── .env.example                    # 环境变量模板
├── nginx.example.conf              # 宿主机 Nginx 配置参考
├── registry/
│   ├── config.yml                  # 生产 Registry 配置（R2）
│   ├── config.dev.yml              # 开发 Registry 配置（文件系统）
│   └── certs/                      # auth.key, auth.crt, jwks.json
├── scripts/
│   ├── generate-auth-cert.sh
│   ├── bootstrap-dev.sh
│   ├── backup-postgres.sh
│   └── gc.sh
├── api/                            # Go API 服务
│   ├── Dockerfile
│   ├── go.mod
│   └── cmd/server/main.go
│   └── internal/...
└── console/                        # Vite React 控制台
    ├── Dockerfile
    ├── package.json
    └── src/...
```

## 关键设计决策

- **Registry:3.1.0**：使用 `/etc/distribution/config.yml`（v3 路径）。
- **存储**：生产环境使用 R2 S3 兼容 API；开发环境使用文件系统。
- **认证**：Go API 服务为 Registry 颁发 RS256 JWT；控制台使用 HMAC Session Cookie。
- **机器人令牌**：用户名格式 `robot$<project>-<name>`。密码为 32 字节随机十六进制字符串，数据库中存储 bcrypt 哈希（旧的 SHA-256 哈希仍可验证）。创建时仅显示一次。权限严格限定在所属项目内。
- **Webhook**：Registry 推送事件到 Go API 服务，API 将项目/仓库/标签信息写入 Postgres。
- **删除标签**：控制台触发 Registry 清单删除 + Postgres 软删除。
- **GC**：MVP 阶段仅支持手动清理。参考 `scripts/gc.sh`。

## 安全清单

- [ ] 替换 `.env` 中所有 `change_me_*` 密码
- [ ] 生产环境使用真实 TLS 证书（宿主机 Nginx）
- [ ] R2 API 令牌仅授权给 Registry 存储桶
- [ ] 不将 Postgres/Redis 端口暴露到公网（服务绑定在 127.0.0.1）
- [ ] 定期轮换 `auth.key`（需要重新配置 Registry）
- [ ] 为控制台用户启用双因素认证（未来功能）

## API 端点

### Registry Token
```
GET /api/registry/token?service=...&scope=...
Authorization: Basic base64(username:password)
```

### 控制台认证
```
POST /api/auth/login
POST /api/auth/logout
GET  /api/me
```

### 管理接口
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

### 内部 Webhook
```
POST /internal/registry/events
Authorization: Bearer <WEBHOOK_SECRET>
```

## 许可证

MIT
