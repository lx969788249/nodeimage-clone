# NodeImage Clone

高性能的全栈图床解决方案，目标是一比一复刻 [nodeimage.com](https://www.nodeimage.com/) 的视觉与交互，同时在安全性、可靠性与性能层面进行系统化加强。

## 架构概览

```
┌──────────────────────────┐
│           Client          │
│  ──────────┬──────────    │
│  SSR Web   │  Admin Web   │
└──────┬──────┴──────┬──────┘
       │             │
       ▼             ▼
┌───────────────┐  ┌───────────────┐
│  Frontend CDN │  │ Admin CDN     │
│ (apps/web)    │  │ (apps/admin)  │
└────────┬──────┘  └──────┬────────┘
         │                │
         ▼                ▼
               ┌────────────────────┐
               │    API Gateway     │
               │ (apps/api, Go/Gin) │
               └────────┬──────────┘
                        │
        ┌───────────────┴────────────────┐
        │        Service Mesh (gRPC)      │
        │                                  │
┌───────▼────────┐  ┌───────────────┐  ┌─────────────┐
│Auth & Session  │  │Media Service  │  │Task Worker  │
│(Go, Postgres)  │  │(Go, libvips)  │  │(Go + Cron)  │
└──────┬─────────┘  └──────┬────────┘  └────┬────────┘
       │                   │               │
       ▼                   ▼               ▼
┌──────────────┐   ┌──────────────┐  ┌──────────────┐
│ Postgres     │   │ MinIO / S3   │  │ Redis (queue)│
│ (metadata)   │   │ (image blob) │  │              │
└──────────────┘   └──────────────┘  └──────────────┘
```

### 选用技术栈（聚焦运行效率）

| 层级                | 技术方案                                                                                             | 说明 |
| ------------------- | ----------------------------------------------------------------------------------------------------- | ---- |
| 前端 SSR            | **SvelteKit + Vite + TypeScript**                                                                    | 极速首屏、极低 bundle 体积，支持精细动画。 |
| UI                  | TailwindCSS + Motion One                                                                             | 便于还原一比一动画与响应式。 |
| 图像查看器          | 自研 WebGL/Canvas 查看器（基于 `@use-gesture` + `three.js` 精简版本）                                 | 支持滚轮缩放、拖拽、滑块。 |
| API Gateway         | **Go 1.23 + Gin + gRPC**                                                                             | 高性能 HTTP 与内部 RPC 网关。 |
| 鉴权&会话           | Go + Postgres + Redis（多设备、刷新令牌、V4 签名）                                                   | 最高 10 设备会话管理。 |
| 图像处理/缩略图     | Go + libvips（通过 `govips` 绑定）                                                                   | 多尺寸 WebP、AVIF 转码，高效。 |
| SVG 安全过滤        | Rust（WASM）插件 + SVGO + 自定义白名单                                                               | 去除 `<script>` 等危险节点。 |
| NSFW 检测           | ONNXRuntime（Go binding） + `nsfw_model.onnx`                                                         | GPU/CPU 自适应，阈值可配置。 |
| 任务调度            | Go cron worker + Redis Stream                                                                         | 过期清理、NSFW 复检。 |
| 存储                | MinIO（与 S3 兼容，可换成 OSS/COS）                                                                   | 图片主存储与变体管理。 |
| 日志 & 观测         | OpenTelemetry + Loki + Grafana                                                                        | 统一追踪与监控。 |
| 部署与运维          | Docker Compose（开发） / K8s（生产推荐），Nginx/Traefik 提供统一域名反代。                              | 明确分层部署。 |

## 核心功能对照表

| 需求项 | 实现方案 |
| ------ | -------- |
| 严格文件验证 | 后端在上传入口读取前 512 bytes，基于魔数验证 + MIME sniffing，拒绝伪装；SVG 需通过 WASM 过滤。 |
| API 鉴权 | 登录颁发 `access_token`（JWT，15min） + `refresh_token`（随机256bit，绑定设备）；所有 API 必须携带 HMAC-SHA256 V4 签名头。 |
| 多设备登录 | 用户最多 10 个活跃设备，登录时 Device Fingerprint + Refresh Token 存于 `user_sessions`，超出则旧设备强制下线。 |
| 单域部署 | 前后端与媒体服务统一由 `example.com` 提供，通过反向代理区分 `/api`、`/media` 等路径。 |
| 安全 ID | 图片 ID 使用 `base58(timestamp + random + checksum)` + V4 HMAC 签名，防穷举。 |
| URL 隐私 | 外链格式 `https://example.com/media/{imageId}/{variant}?sig=...`，不暴露用户 ID。 |
| 管理后台 | 独立 SvelteKit 应用（apps/admin），调用管理 API，可筛选 NSFW、举报、用户。 |
| SVG/AVIF | libvips 直接处理 AVIF；SVG 经清洗后再写入存储。 |
| 动图处理 | 检测多帧（gif/webp/apng）；对多帧保持原始，生成缩略也保持动图。 |
| NSFW 校准 | 双阈值策略（`review`, `block`），低于 review 直接通过，高于 block 直接拒绝，中间进入人工队列。 |
| 缩略图 | 预生成 `thumb_sm` `thumb_md` `thumb_lg` `hero` 四档 WebP + 一档原比例 AVIF。 |
| 过期清理 | Worker 读取 `images.expire_at`，到期执行删除 + 通知队列；支持软删除保留 7 天。 |
| 图片查看器 | Web 端全屏 Lightbox，支持滚轮缩放、拖拽、键盘导航、缩略尺滑块。 |
| 暗色模式 | Tailwind `class` 模式，监听系统 `prefers-color-scheme`，支持手动切换并存储到 localStorage。 |
| 批量复制 | 上传完成后在队列中生成 Markdown / HTML / BBCode，支持“一键全复制”。 |
| 品牌视觉 | `assets/brand/` 持有 SVG Logo，按原站比例布局。 |
| 设置按钮引导 | 首次访问时播放引导动画 + Tooltip 引导。 |
| 页脚完善 | 静态页：版权、使用条款、隐私政策 Markdown 渲染。 |

## 代码结构草案

```
nodeimage-clone/
├── apps/
│   ├── web/          # 主站（SvelteKit）
│   ├── admin/        # 管理后台
│   ├── api/          # Go API 服务（REST + gRPC）
│   └── worker/       # 后台任务（Go）
├── services/
│   └── image-proxy/  # 图片签名校验与响应头
├── packages/
│   ├── config/       # 统一配置定义（Go + TS）
│   ├── proto/        # gRPC proto
│   └── ui/           # 共享 UI 组件库
├── deployments/
│   ├── docker-compose.dev.yml
│   ├── k8s/          # helm charts
│   └── traefik/      # 反向代理配置
├── docs/
│   ├── architecture.md
│   ├── api.md
│   ├── storage.md
│   └── operations.md
└── Makefile          # 常用脚本（lint/test/build）
```

## 下一步

1. **业务建模**：设计数据库 schema（Postgres），梳理用户、图片、会话、审核日志等表。
2. **API 设计**：完成 OpenAPI + gRPC 定义，覆盖上传、管理、统计等核心流程。
3. **基础服务实现**：从 API 服务入手，搭建身份认证、设备管理、签名校验、上传处理。
4. **媒体管线搭建**：完成 libvips、NSFW 检测、SVG 过滤、缩略图生成与存储层对接。
5. **前端复刻**：SvelteKit 搭建页面，按 nodeimage 原有布局与动效复现，接入 API。

完成上述步骤后将进入部署与性能调优阶段，并补充详细的上线指南。接下来将进入数据库与 API 设计阶段。 

## VPS 一键部署脚本

仓库提供 `deploy/install.sh` 安装脚本，可在全新的 Ubuntu 服务器上自动完成依赖安装、数据库与对象存储初始化、构建与服务部署。

使用方式：

```bash
cd nodeimage-clone/deploy
cp config.sample.env config.env
# 编辑 config.env 配置数据库/MinIO 密钥；若自行配置反代，DOMAIN 可留空
nano config.env

sudo bash install.sh
```

脚本会自动创建 systemd 服务（`nodeimage-api`、`nodeimage-worker`、`nodeimage-web`）并构建必要组件；如果配置了 `DOMAIN`，还会生成 Nginx 站点与可选的 Let's Encrypt 证书。若保留 `DOMAIN` 为空，服务默认监听本地端口，可按需自行接入其他反向代理。
