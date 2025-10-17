# 架构设计细节

## 1. 服务划分

| 服务 | 语言 | 主要职责 |
| ---- | ---- | -------- |
| `apps/api` | Go 1.23 | 对外 REST API；对内 gRPC façade；鉴权、会话管理、上传入口、统计。 |
| `apps/web` | SvelteKit | 对标 nodeimage 主页、上传页、图库、文档等所有前端内容。 |
| `apps/admin` | SvelteKit | 审核后台、数据总览、用户管理。 |
| `services/image-proxy` | Go | 签名校验、图片读取、缓存控制、Range 支持。 |
| `apps/worker` | Go | 异步队列消费者：缩略图、NSFW、过期清理。 |
| `third-party` | Redis/MinIO/Postgres | 统一通过 docker-compose / K8s 编排。 |

## 2. 数据模型概要

```text
users(id, email, password_hash, display_name, role, status, created_at, updated_at)
user_sessions(id, user_id, device_id, device_name, refresh_token_hash, ip, ua, last_seen_at, created_at, expires_at)
images(id, user_id, bucket, object_key, format, width, height, frames, size_bytes, nsfw_score, status, expire_at, created_at, updated_at)
image_variants(id, image_id, variant, bucket, object_key, width, height, size_bytes, created_at)
image_audit_logs(id, image_id, reviewer_id, action, reason, created_at)
api_keys(id, user_id, name, hash, scopes, created_at, last_used_at)
webhooks(id, user_id, url, secret, status, created_at, updated_at)
reports(id, image_id, reporter_id, reason, notes, status, created_at, handled_at)
```

所有 ID 均使用 `ksuid` 派生并附加签名后再暴露给前端，避免枚举。

## 3. 鉴权流程

1. 登录后生成 `access_token`（JWT，15 分钟）和 `refresh_token`（64 字节随机数）。
2. Refresh token 存入 `user_sessions`，并携带 `device_id`。同一用户上限 10 条记录，超出即使旧记录失效。
3. 所有 API 请求需携带：
   - `Authorization: Bearer <access_token>`
   - `X-Codex-Date`：RFC3339 时间戳
   - `X-Codex-Nonce`：一次性随机数，Redis 缓存 5 分钟防重放
   - `X-Codex-Signature`：`HMAC-SHA256(access_token_id + path + body + date + nonce)` 实现 V4 签名
4. 图片直链使用短期签名 URL（5 分钟），由前端在需要时向 API 申请。

## 4. 上传与处理流程

```
Client -> API /upload (multipart)
         └─> Pre-flight: 读取前 512 bytes 校验魔数 (jpeg/png/webp/gif/apng/avif/svg)
         └─> Storage: 将原始文件写入 MinIO (bucket: originals/)
         └─> Queue: Redis Stream 推送处理任务 {imageID, objectKey}
Worker -> 监听处理任务
         ├─> NSFW 检测 (onnxruntime)
         ├─> SVG 安全过滤（WASM，输出新 svg）
         ├─> 动图识别（libvips -> `n-pages`）
         ├─> 生成缩略图（sm/md/lg/hero WebP，保留动图；另生成 avif）
         └─> 元数据写入 Postgres；通过 webhook 通知客户端
```

## 5. 缓存与 CDN

- 所有图片通过 `services/image-proxy` 返回，负责校验签名、设置缓存头、兼容 Range 请求。
- 对接 CDN（Cloudflare/Akamai/自建）时，只需缓存统一域名（例如 `example.com`）即可。
- 缩略图命名 `https://example.com/media/{imageId}/{variant}.{format}?sig=...`

## 6. 任务调度

- **过期图片清理**：`00:00` 读取 `images.expire_at <= now()`，执行软删除并 enqueue 7 天后硬删除任务。
- **NSFW 复检**：间隔 7 天对灰区图片复检，降低误判。

## 7. 部署建议

- **开发环境**：`docker-compose.dev.yml` 启动 Postgres、Redis、MinIO、反向代理；前后端 `npm run dev`/`air`.
- **生产环境**：Kubernetes + Helm Charts；Traefik 或 Nginx Ingress 提供统一域名入口；MinIO 可替换为公有云对象存储。
- **CI/CD**：GitHub Actions → 构建 Docker 镜像 → 推送到镜像仓库 → ArgoCD 部署。

## 8. 安全要点

- 上传前后都执行 MIME 检测，拒绝任何包含 `application/octet-stream` 或 `text/html` 的伪装文件。
- SVG 白名单标签：`svg`, `g`, `path`, `rect`, `circle`, `linearGradient`, `stop`, `defs`, `use`；属性需过滤 `on*` 事件。
- 所有外链使用 Content-Security-Policy 严格限制。
- 管理后台需二次验证（TOTP）。
- 关键配置密钥使用 `.env` 加载 + SOPS 加密存储。

## 9. 后续迭代

- 引入审计轨迹（Audit Trail）系统，使用 `immutable_logs` 表配合哈希链。
- 计划支持图片集（相册）与 API 限流（基于 Redis Token Bucket）。
- 与对象存储的直传（Pre-signed）将在 MVP 之后迭代。
