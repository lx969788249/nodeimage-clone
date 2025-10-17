#!/usr/bin/env bash
#
# NodeImage 一键部署脚本（针对 Ubuntu 20.04/22.04 设计）
# 使用前请复制 deploy/config.sample.env 为 deploy/config.env 并填写变量。
#
# 本脚本将执行以下操作：
# 1. 安装系统依赖：git、curl、PostgreSQL、Redis、Nginx 等
# 2. 安装 Go 1.23 与 Node.js 20
# 3. 安装并配置 MinIO（对象存储）
# 4. 初始化数据库、生成配置文件
# 5. 构建后端 API / Worker、前端 Web
# 6. 创建 systemd 与 Nginx 服务，自动启用并（可选）申请 HTTPS 证书

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_FILE="${REPO_DIR}/deploy/config.env"

function require_root() {
  if [[ "$(id -u)" -ne 0 ]]; then
    echo "本脚本需要 root 权限，请使用 sudo 执行。" >&2
    exit 1
  fi
}

function load_config() {
  if [[ ! -f "${CONFIG_FILE}" ]]; then
    echo "未找到 ${CONFIG_FILE}，请复制 deploy/config.sample.env 为 deploy/config.env 并填写。" >&2
    exit 1
  fi
  set -a
  source "${CONFIG_FILE}"
  set +a

  DOMAIN="${DOMAIN:-}"
  : "${DB_PASSWORD:?DB_PASSWORD 未设置}"
  : "${MINIO_ACCESS_KEY:?MINIO_ACCESS_KEY 未设置}"
  : "${MINIO_SECRET_KEY:?MINIO_SECRET_KEY 未设置}"
  : "${JWT_ACCESS_SECRET:?JWT_ACCESS_SECRET 未设置}"
  : "${JWT_REFRESH_SECRET:?JWT_REFRESH_SECRET 未设置}"
  : "${SIGNATURE_SECRET:?SIGNATURE_SECRET 未设置}"
  : "${FRONTEND_PORT:=4173}"
  : "${API_PORT:=8080}"
  : "${REDIS_STREAM:=media:ingest}"
  : "${ENABLE_HTTPS:=yes}"
  if [[ -z "${DOMAIN}" ]]; then
    LETSENCRYPT_EMAIL="${LETSENCRYPT_EMAIL:-}"
  else
    : "${LETSENCRYPT_EMAIL:=admin@${DOMAIN}}"
  fi
}

function apt_install() {
  echo ">> 更新 apt 软件源并安装基础依赖..."
  apt-get update
  apt-get install -y build-essential git curl unzip ufw pkg-config \
    postgresql postgresql-contrib redis-server nginx libpq-dev
}

function install_go() {
  if command -v go >/dev/null 2>&1; then
    echo "Go 已安装，跳过。"
    return
  fi

  echo ">> 安装 Go 1.23..."
  curl -OL https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
  rm go1.23.0.linux-amd64.tar.gz
  if ! grep -q '/usr/local/go/bin' /etc/profile; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
  fi
  export PATH=$PATH:/usr/local/go/bin
}

function install_node() {
  if command -v node >/dev/null 2>&1; then
    echo "Node.js 已安装，跳过。"
    return
  fi

  echo ">> 安装 Node.js 20..."
  curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
  apt-get install -y nodejs
}

function configure_postgres() {
  echo ">> 配置 PostgreSQL..."
  sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname='nodeimage';" | grep -q 1 || \
    sudo -u postgres psql -c "CREATE ROLE nodeimage WITH LOGIN PASSWORD '${DB_PASSWORD}';"
  sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='nodeimage';" | grep -q 1 || \
    sudo -u postgres psql -c "CREATE DATABASE nodeimage OWNER nodeimage;"
}

function configure_redis() {
  echo ">> 配置 Redis..."
  sed -i 's/^supervised .*/supervised systemd/' /etc/redis/redis.conf
  systemctl enable --now redis-server
}

function install_minio() {
  if command -v minio >/dev/null 2>&1; then
    echo "MinIO 已安装，跳过。"
  else
    echo ">> 安装 MinIO..."
    curl -o /usr/local/bin/minio https://dl.min.io/server/minio/release/linux-amd64/minio
    chmod +x /usr/local/bin/minio
  fi

  id -u minio-user >/dev/null 2>&1 || useradd -r minio-user -s /sbin/nologin
  mkdir -p /var/minio/data
  chown -R minio-user:minio-user /var/minio

  cat >/etc/systemd/system/minio.service <<EOF
[Unit]
Description=MinIO
After=network.target
Requires=network.target

[Service]
User=minio-user
Group=minio-user
ExecStart=/usr/local/bin/minio server /var/minio/data --console-address ":9001"
Environment="MINIO_ROOT_USER=${MINIO_ACCESS_KEY}"
Environment="MINIO_ROOT_PASSWORD=${MINIO_SECRET_KEY}"
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable --now minio
}

function configure_minio_buckets() {
  echo ">> 创建 MinIO bucket..."
  sleep 5
  MC_BIN=/usr/local/bin/mc
  if [[ ! -f "${MC_BIN}" ]]; then
    curl -o "${MC_BIN}" https://dl.min.io/client/mc/release/linux-amd64/mc
    chmod +x "${MC_BIN}"
  fi

  "${MC_BIN}" alias set nodeimage http://127.0.0.1:9000 "${MINIO_ACCESS_KEY}" "${MINIO_SECRET_KEY}" >/dev/null
  "${MC_BIN}" mb -p nodeimage/nodeimage-originals >/dev/null 2>&1 || true
  "${MC_BIN}" mb -p nodeimage/nodeimage-variants >/dev/null 2>&1 || true
}

function generate_configs() {
  echo ">> 生成配置文件..."
  mkdir -p "${REPO_DIR}/config"

  local cors_entries
  if [[ -n "${DOMAIN}" ]]; then
    cors_entries="  - https://${DOMAIN}\n  - https://www.${DOMAIN}"
  else
    cors_entries="  - http://localhost\n  - http://127.0.0.1"
  fi

  cat >"${REPO_DIR}/config/config.yaml" <<EOF
environment: production

http:
  host: 0.0.0.0
  port: ${API_PORT}
  readTimeout: 10s
  writeTimeout: 15s
  idleTimeout: 60s

postgres:
  dsn: postgres://nodeimage:${DB_PASSWORD}@127.0.0.1:5432/nodeimage?sslmode=disable
  maxOpen: 50
  maxIdle: 10
  connMaxLifetime: 30m

redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0

storage:
  endpoint: http://127.0.0.1:9000
  accessKey: ${MINIO_ACCESS_KEY}
  secretKey: ${MINIO_SECRET_KEY}
  bucketOriginals: nodeimage-originals
  bucketVariants: nodeimage-variants
  useSSL: false
  region: us-east-1

security:
  jwtAccessSecret: ${JWT_ACCESS_SECRET}
  jwtRefreshSecret: ${JWT_REFRESH_SECRET}
  jwtAccessTTL: 15m
  jwtRefreshTTL: 720h
  signatureSecret: ${SIGNATURE_SECRET}
  maxSessions: 10

nsfw:
  modelPath: ./models/nsfw_model.onnx
  thresholdBlock: 0.92
  thresholdReview: 0.75
  recheckInterval: 168h

allowCORSOrigins:
$(printf "%b" "${cors_entries}")
EOF

  cat >"${REPO_DIR}/config/worker.yaml" <<EOF
environment: production

redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0
  stream: ${REDIS_STREAM}
  group: media-workers
  consumer: worker-1

storage:
  endpoint: http://127.0.0.1:9000
  accessKey: ${MINIO_ACCESS_KEY}
  secretKey: ${MINIO_SECRET_KEY}
  bucketOriginals: nodeimage-originals
  bucketVariants: nodeimage-variants
  useSSL: false
  region: us-east-1

queues:
  visibilityTimeout: 2m
  claimInterval: 15s

logging:
  level: info
EOF
}

function run_migrations() {
  echo ">> 执行数据库迁移..."
  export PATH=$PATH:/usr/local/go/bin
  pushd "${REPO_DIR}/apps/api" >/dev/null
  go mod tidy
  go install github.com/pressly/goose/v3/cmd/goose@latest
  goose -dir internal/database/migrations postgres "postgres://nodeimage:${DB_PASSWORD}@127.0.0.1:5432/nodeimage?sslmode=disable" up
  popd >/dev/null
}

function build_backend() {
  echo ">> 构建 API 与 Worker..."
  export PATH=$PATH:/usr/local/go/bin
  mkdir -p /opt/nodeimage/bin

  pushd "${REPO_DIR}/apps/api" >/dev/null
  go build -o /opt/nodeimage/bin/nodeimage-api ./cmd/api
  popd >/dev/null

  pushd "${REPO_DIR}/apps/worker" >/dev/null
  go build -o /opt/nodeimage/bin/nodeimage-worker ./cmd/worker
  popd >/dev/null
}

function build_frontend() {
  echo ">> 构建前端..."
  export PATH=$PATH:/usr/local/go/bin
  pushd "${REPO_DIR}/apps/web" >/dev/null
  npm install
  npm run build
  popd >/dev/null
}

function setup_systemd() {
  echo ">> 配置 systemd 服务..."

  cat >/etc/systemd/system/nodeimage-api.service <<EOF
[Unit]
Description=NodeImage API
After=network.target postgresql.service redis-server.service minio.service

[Service]
WorkingDirectory=${REPO_DIR}/apps/api
Environment="NODEIMAGE_CONFIG=${REPO_DIR}/config/config.yaml"
ExecStart=/opt/nodeimage/bin/nodeimage-api
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  cat >/etc/systemd/system/nodeimage-worker.service <<EOF
[Unit]
Description=NodeImage Worker
After=redis-server.service minio.service

[Service]
WorkingDirectory=${REPO_DIR}/apps/worker
Environment="NODEIMAGE_WORKER_CONFIG=${REPO_DIR}/config/worker.yaml"
ExecStart=/opt/nodeimage/bin/nodeimage-worker
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  cat >/etc/systemd/system/nodeimage-web.service <<EOF
[Unit]
Description=NodeImage SvelteKit Preview
After=network.target

[Service]
WorkingDirectory=${REPO_DIR}/apps/web
Environment="HOST=127.0.0.1"
Environment="PORT=${FRONTEND_PORT}"
ExecStart=/usr/bin/npm run preview -- --host 127.0.0.1 --port ${FRONTEND_PORT}
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable --now nodeimage-api nodeimage-worker nodeimage-web
}

function setup_nginx() {
  if [[ -z "${DOMAIN}" ]]; then
    echo ">> DOMAIN 未设置，跳过 Nginx/HTTPS 配置。"
    return
  fi

  echo ">> 配置 Nginx..."
  cat >/etc/nginx/sites-available/nodeimage.conf <<EOF
server {
    listen 80;
    server_name ${DOMAIN} www.${DOMAIN};

    location /api/ {
        proxy_pass http://127.0.0.1:${API_PORT}/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }

    location /media/ {
        proxy_pass http://127.0.0.1:${API_PORT}/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }

    location / {
        proxy_pass http://127.0.0.1:${FRONTEND_PORT};
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }
}
EOF

  ln -sf /etc/nginx/sites-available/nodeimage.conf /etc/nginx/sites-enabled/nodeimage.conf
  nginx -t
  systemctl restart nginx

  if [[ "${ENABLE_HTTPS}" == "yes" ]]; then
    if [[ -z "${LETSENCRYPT_EMAIL}" ]]; then
      echo ">> 未提供 LETSENCRYPT_EMAIL，跳过证书申请。"
    else
      echo ">> 申请 Let's Encrypt 证书..."
      apt-get install -y python3-certbot-nginx
      certbot --non-interactive --nginx -d "${DOMAIN}" -d "www.${DOMAIN}" -m "${LETSENCRYPT_EMAIL}" --agree-tos --redirect || \
        echo "证书申请失败，请检查域名解析后手动运行 certbot。"
    fi
  fi
}

function configure_firewall() {
  echo ">> 配置防火墙..."
  ufw allow OpenSSH
  ufw allow 'Nginx Full'
  yes | ufw enable || true
}

require_root
load_config
apt_install
install_go
install_node
configure_postgres
configure_redis
install_minio
configure_minio_buckets
generate_configs
run_migrations
build_backend
build_frontend
setup_systemd
setup_nginx
configure_firewall

if [[ -n "${DOMAIN}" ]]; then
  local scheme="http"
  [[ "${ENABLE_HTTPS}" == "yes" ]] && scheme="https"
  echo "部署完成！请访问 ${scheme}://${DOMAIN} 查看站点。"
else
  echo "部署完成！前端监听端口：${FRONTEND_PORT}，API 端口：${API_PORT}。"
  echo "可通过 http://<服务器IP>:${FRONTEND_PORT} 访问前端界面，或自行配置反向代理。"
fi
echo "若需查看日志：journalctl -u nodeimage-api -f"
