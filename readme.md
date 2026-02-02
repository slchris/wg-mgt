# WG MGT

一个简洁高效的 WireGuard 多节点管理工具，帮助你轻松管理分布在多台 VPS 上的 WireGuard 网络。

## 功能特性

### 多节点管理

- 集中管理多个 VPS 上的 WireGuard 实例
- 通过 SSH 远程管理节点上的 WireGuard 配置
- 实时监控各节点运行状态
- 一键同步配置到服务器

### 配置管理
- 可视化生成 WireGuard 配置文件
- 支持配置下载（二维码/文件）
- 批量创建和删除配置
- 自动分配 IP 地址和密钥对

### 网络功能

- 打通多地局域网，实现异地组网
- 支持 Peer-to-Peer 直连
- 灵活的路由配置

### 安全设计

- 支持本地访问模式，无需担心暴露风险
- 数据本地存储，完全私有化部署
- JWT 身份认证

## 技术栈

| 层级 | 技术选型 |
|------|----------|
| 前端 | React 18 + TypeScript + Tailwind CSS |
| 后端 | Go 1.24 + Chi Router + GORM |
| 数据库 | SQLite（零配置、便携） |
| 部署 | Docker / 单文件二进制 |

## 快速开始

### Docker 部署（推荐）

```bash
# 克隆项目
git clone https://github.com/slchris/wg-mgt.git
cd wg-mgt

# 启动服务
docker compose up -d

# 访问 http://localhost:8080
# 默认账号: admin / admin123
```

### 二进制部署

```bash
# 下载最新 Release
# https://github.com/slchris/wg-mgt/releases

# 运行
./wg-mgt

# 访问 http://localhost:8080
```

### 源码编译

```bash
# 克隆项目
git clone https://github.com/slchris/wg-mgt.git
cd wg-mgt

# 构建前端
cd web && npm ci && npm run build && cd ..

# 构建后端
go build -o wg-mgt ./cmd/wg-mgt

# 运行
./wg-mgt
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `WG_MGT_PORT` | `8080` | 服务端口 |
| `WG_MGT_HOST` | `127.0.0.1` | 监听地址 |
| `WG_MGT_LOCAL_ONLY` | `true` | 仅允许本地访问 |
| `WG_MGT_DB_PATH` | `wg-mgt.db` | SQLite 数据库路径 |
| `WG_MGT_JWT_SECRET` | - | JWT 密钥（生产环境必须设置） |
| `WG_MGT_ADMIN_USER` | - | 默认管理员用户名 |
| `WG_MGT_ADMIN_PASS` | - | 默认管理员密码 |

## 项目结构

```
wg-mgt/
├── cmd/wg-mgt/     # 程序入口
├── internal/       # 内部逻辑
│   ├── app/        # 应用初始化
│   ├── config/     # 配置管理
│   ├── domain/     # 领域模型
│   ├── handler/    # HTTP 处理器
│   ├── middleware/ # 中间件
│   ├── pkg/        # 内部工具包
│   ├── repository/ # 数据访问层
│   ├── router/     # 路由配置
│   └── service/    # 业务逻辑层
├── web/            # 前端源码 (React)
└── .github/        # GitHub Actions
```

## 开发

```bash
# 安装依赖
go mod download
cd web && npm ci

# 开发模式（后端）
go run ./cmd/wg-mgt

# 开发模式（前端）
cd web && npm run dev

# 运行测试
go test ./...

# 代码检查
golangci-lint run
cd web && npm run lint
```

## 许可证

MIT License
