# 部署与验证指南

## 一键自动部署（推荐）

仓库提供 `deploy/install.sh` 脚本，可在全新 Ubuntu 20.04/22.04 VPS 上自动完成依赖安装、数据库与对象存储初始化、构建和服务部署。

1. 克隆仓库并进入 `deploy/`：
   ```bash
   git clone https://github.com/your/repo.git nodeimage-clone
   cd nodeimage-clone/deploy
   ```
2. 复制并编辑配置文件：
   ```bash
   cp config.sample.env config.env
   nano config.env   # 配置数据库口令、MinIO 密钥等；DOMAIN 可留空自行配置反代
   ```
3. 执行安装脚本（需 root 或 sudo）：
   ```bash
   sudo bash install.sh
   ```
   脚本将安装 Go、Node.js、PostgreSQL、Redis、MinIO，生成配置并构建项目；若填写了 `DOMAIN`，会额外创建 Nginx 站点及（可选）Let's Encrypt 证书。
4. 如果配置了域名，安装完成后可访问 `https://<DOMAIN>` 验证站点；如未设置域名，前端默认监听 `http://<服务器IP>:4173`，可结合自有反向代理。查看日志可用 `journalctl -u nodeimage-api -f`。

如需手动部署或自定义环境，可参考以下详细步骤。

## 1. 基础环境

- Go 1.23+
- Node.js 20+（包含 npm 10+）
- Redis 7+
- PostgreSQL 15+
- MinIO 或任何兼容 S3 的对象存储

> 当前仓库未执行 `go mod tidy` 与 `npm install`，请在本地安装 Go 后补充依赖。

## 2. 配置文件

- `config/config.example.yaml`：API 服务配置，复制为 `config/config.yaml` 后按需修改。
- `config/worker.example.yaml`：Worker 配置。
- `.env`（可选）：通过环境变量覆盖敏感信息，例如 `NODEIMAGE_SECURITY_JWTACCESSSECRET`。

## 3. 数据库迁移

使用 `goose` 或其他工具执行 `apps/api/internal/database/migrations` 内的 SQL：

```bash
export DATABASE_URL=postgres://user:pass@localhost:5432/nodeimage?sslmode=disable
goose -dir apps/api/internal/database/migrations postgres up
```

## 4. 启动顺序

1. **对象存储**：启动 MinIO，并创建 `nodeimage-originals` 与 `nodeimage-variants` bucket。
2. **Redis & PostgreSQL**：确保服务可访问。
3. **API 服务**：
   ```bash
   cd nodeimage-clone/apps/api
   go mod tidy
   go run ./cmd/api
   ```
4. **Worker**：
   ```bash
   cd nodeimage-clone/apps/worker
   go mod tidy
   go run ./cmd/worker
   ```
5. **前端**：
   ```bash
   cd nodeimage-clone/apps/web
   npm install
   npm run dev
   ```

## 5. 域名与反代

建议使用 Traefik / Nginx 在同一域名（如 `example.com`）下反代：

- `/` → 前端（SvelteKit 或静态站点）
- `/api/*` → API 服务（Gin）
- `/media/*` → 图片签名网关（待实现）

## 6. 验证清单

1. **健康检查**：`GET /api/healthz` 返回 200，包含数据库与 Redis 状态。
2. **注册与登录**：使用 `POST /api/v1/auth/register`、`/login`、`/refresh`、`/auth/me`，验证多设备会话限制。
3. **签名校验**：调用任何受保护接口时需携带 `Authorization`、`X-Codex-*` 头，确保 401/403 逻辑正确。
4. **上传流程**：`POST /api/v1/media/upload` 应写入 MinIO，并在 Redis `media:ingest` 流新增任务。
5. **后台管理**：`GET /api/v1/admin/images` 需管理员权限，验证分页返回。
6. **定时任务**：观察 Redis `media:ingest` 流在每日 00:00 插入清理任务、在整点插入 NSFW 复检任务。

## 7. TODO / 已知限制

- 未接入真实的图像处理（缩略图、NSFW）与 Lightbox 查看器，Worker 目前为占位。
- 前端上传暂为模拟流程，需要接入 API 签名逻辑。
- 缺少 OpenAPI / gRPC 定义与自动化测试套件。
- 未拆分独立的图片签名网关服务（`services/image-proxy` 目录待实现）。

完成上述步骤后，即可进入 UI 微调、动效复刻与后台功能拓展阶段。
