# One API 部署文档

**部署日期：** 2026-04-11
**部署环境：** WSL2 Ubuntu 22.04
**部署路径：** `/mnt/d/aicode/one-api/`

---

## 步骤一：环境准备

### 1.1 安装 Docker（如使用 Docker 部署）

```bash
# 安装 Docker
apt-get update
apt-get install -y docker.io

# 启动 Docker 服务
dockerd &>/tmp/dockerd.log &
sleep 3
docker ps  # 验证 Docker 运行
```

### 1.2 安装 Go 1.22+（如从源码编译）

```bash
# 下载 Go 1.22
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz -O /tmp/go1.22.tar.gz

# 解压安装
rm -rf /usr/local/go
tar -C /usr/local -xzf /tmp/go1.22.tar.gz

# 验证版本
export PATH=/usr/local/go/bin:$PATH
go version  # 应显示 go1.22.0
```

### 1.3 安装 Node.js（如构建前端）

```bash
# npm 已预装，如需安装
npm --version
```

---

## 步骤二：下载与构建

### 方案 A：Docker 部署（推荐生产环境）

```bash
# 拉取镜像
docker pull ghcr.io/songquanpeng/one-api:latest

# 运行容器
docker run --name one-api -d --restart always \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -v ~/data/one-api:/data \
  ghcr.io/songquanpeng/one-api:latest
```

> 注：如 Docker Hub 拉取超时，可使用 GitHub Container Registry：
> `docker pull ghcr.io/songquanpeng/one-api:latest`

### 方案 B：从源码编译（本次测试使用）

```bash
# 克隆仓库
cd /tmp
git clone https://github.com/songquanpeng/one-api.git

# 构建前端（使用国内镜像）
cd one-api/web/default
npm install --registry=https://registry.npmmirror.com
npm run build

# 返回项目根目录
cd ../..

# 下载 Go 依赖
go mod download

# 编译二进制
go build -ldflags "-s -w" -o one-api
```

---

## 步骤三：部署

### 3.1 创建目录

```bash
mkdir -p /mnt/d/aicode/one-api/data
mkdir -p /mnt/d/aicode/one-api/logs
cp one-api /mnt/d/aicode/one-api/
chmod +x /mnt/d/aicode/one-api/one-api
```

### 3.2 启动服务

```bash
cd /mnt/d/aicode/one-api

# 启动（SQLite 模式，数据存储在 ./data）
./one-api --port 3000 --log-dir ./logs &

# 验证启动
curl http://localhost:3000/
```

### 3.3 验证运行

```bash
# 检查进程
ps aux | grep one-api

# 查看日志
tail -f logs/one-api.log
```

---

## 步骤四：初始化配置

### 4.1 登录管理后台

- 访问地址：http://localhost:3000/
- 默认账号：`root`
- 默认密码：`123456`

> ⚠️ 首次登录后请立即修改密码！

### 4.2 配置 MiniMax 渠道

1. 进入 **渠道管理** → **添加渠道**
2. 填写配置：

| 字段 | 值 |
|------|-----|
| 渠道名称 | MiniMax |
| 渠道类型 | MiniMax |
| API Key | 你的 MiniMax API Key |
| 模型列表 | MiniMax-M2.7, MiniMax-M2.5 等 |
| 权重 | 100（默认） |

3. 点击 **保存**

### 4.3 创建 API Token

1. 进入 **令牌管理** → **创建令牌**
2. 设置名称、额度、过期时间
3. 复制生成的 Token

### 4.4 测试调用

```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <你的Token>" \
  -d '{
    "model": "MiniMax-M2.7",
    "messages": [{"role": "user", "content": "say hi"}],
    "stream": true
  }'
```

---

## 步骤五：配置 OpenClaw

在 OpenClaw 的 `openclaw.json` 中添加 One API provider：

```json
"models": {
  "providers": {
    "one-api": {
      "baseUrl": "http://<one-api地址>:3000/v1",
      "apiKey": "<One API Token>",
      "models": [
        {
          "id": "MiniMax-M2.7",
          "name": "MiniMax M2.7"
        }
      ]
    }
  }
}
```

---

## 步骤六：生产环境建议

### 6.1 使用 MySQL 替代 SQLite

```bash
docker run --name one-api -d --restart always \
  -p 3000:3000 \
  -e SQL_DSN="root:password@tcp(localhost:3306)/oneapi" \
  -e TZ=Asia/Shanghai \
  -v ~/data/one-api:/data \
  ghcr.io/songquanpeng/one-api:latest
```

### 6.2 配置 Redis（可选，提升性能）

```bash
-e REDIS_CONN_STRING="redis:6379"
```

### 6.3 配置邮件服务（可选，发送通知）

```bash
-e SMTP_HOST="smtp.example.com"
-e SMTP_PORT=587
-e SMTP_FROM="noreply@example.com"
-e SMTP_USER="user"
-e SMTP_PASS="password"
```

### 6.4 反向代理配置（Nginx）

```nginx
location / {
    proxy_pass http://localhost:3000;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_buffering off;  # 重要：关闭缓冲以支持流式响应
}
```

---

## 常见问题

### Q1: Docker 拉取超时

**原因：** Docker Hub 在部分地区访问受限

**解决方案：**
1. 使用 GitHub Container Registry：`docker pull ghcr.io/songquanpeng/one-api:latest`
2. 或使用国内镜像加速器

### Q2: 编译时 Go 版本过低

**错误：** `package slices is not in GOROOT`

**解决方案：** 升级到 Go 1.22+

```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=/usr/local/go/bin:$PATH
```

### Q3: npm 安装依赖超时

**解决方案：** 使用国内镜像

```bash
npm install --registry=https://registry.npmmirror.com
```

### Q4: 流式响应不工作

**检查项：**
1. Nginx 配置中是否关闭了 `proxy_buffering`
2. 客户端是否正确处理 SSE
3. 查看 One API 日志排查

---

## 相关资源

- GitHub: https://github.com/songquanpeng/one-api
- Docker Hub: https://hub.docker.com/r/justsong/one-api
- 在线演示: https://openai.justsong.cn/
