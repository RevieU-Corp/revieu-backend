# RevieU Auth Service

[![Python Version](https://img.shields.io/badge/python-3.13%2B-blue.svg)](https://www.python.org/downloads/)
[![FastAPI](https://img.shields.io/badge/FastAPI-0.115+-009688.svg)](https://fastapi.tiangolo.com/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

RevieU Auth Service 是一个基于 **FastAPI** 构建的高性能用户认证与授权微服务。它提供了完整的用户生命周期管理，包括注册、登录、个人资料管理以及支持 Google OAuth 2.0 的第三方认证，并集成了现代化的 CI/CD 流程。

---

## 🚀 核心特性

- **认证与授权**: 基于 JWT (JSON Web Tokens) 的无状态会话管理。
- **个人资料一体化**: 优化后的数据库模型，支持用户与资料字段的原生合并。
- **第三方登录**: 深度集成 Google OAuth 2.0。
- **安全保障**: 采用 `bcrypt` 算法进行密码哈希，内置邮箱验证逻辑。
- **现代化架构**: 异步 IO 驱动，使用 `structlog` 渲染结构化 JSON 日志。
- **自动化流转**: 严格的 `feat` -> `dev` -> `main` 代码合并与自动部署流程。

---

## 🛠 技术栈

- **框架**: [FastAPI](https://fastapi.tiangolo.com/)
- **包管理**: [uv](https://github.com/astral-sh/uv) (Extremely fast Python package installer)
- **数据库**: [PostgreSQL](https://www.postgresql.org/) + [SQLAlchemy 2.0](https://www.sqlalchemy.org/)
- **迁移工具**: [Alembic](https://alembic.sqlalchemy.org/)
- **容器化**: [Docker](https://www.docker.com/) & [GitHub Packages (GHCR)](https://github.com/features/packages)
- **流水线**: GitHub Actions

---

## 📂 项目结构

```text
apps/auth-service/
├── app/
│   ├── api/          # 路由定义 (v1)
│   ├── core/         # 配置与安全核心逻辑
│   ├── db/           # 数据库连接与基类
│   ├── models/       # SQLAlchemy 数据库模型
│   ├── schemas/      # Pydantic 数据验证模型
│   ├── services/     # 业务逻辑层
│   └── main.py       # 应用入口
├── doc/              # 文档资源
├── test/             # 自动化测试
├── alembic/          # 数据库迁移脚本
└── docker-compose.yml
```

---

## ⚙️ 快速开始

### 前置要求
- Python 3.13+
- PostgreSQL
- [uv](https://github.com/astral-sh/uv)

### 本地运行
1. **安装依赖**:
   ```bash
   uv sync
   ```
2. **环境配置**:
   拷贝 `.env.example` 并重命名为 `.env`，填入必要的数据库和 OAuth 凭证。
3. **运行迁移**:
   ```bash
   uv run alembic upgrade head
   ```
4. **启动服务**:
   ```bash
   uv run uvicorn app.main:app --reload --port 8080
   ```

---

## 🐳 Docker 与联通性

### 宿主机数据库配置
若数据库在宿主机，容器在 Docker，需确保 `/etc/postgresql/16/main/pg_hba.conf` 允许容器网段访问：
```text
host    all             all             0.0.0.0/0               scram-sha-256
```
并将宿主机的 `postgresql.conf` 设置为 `listen_addresses = '*'`。

### 容器启动
```bash
docker compose up -d
```

---

## 🔄 CI/CD 与流程规范

### 分支保护
- ** 禁止直接 Push 到 `main` 和 `dev` 分支。**
- 所有改动必须通过 Pull Request。
- 发往 `main` 的 PR 必须且只能源自 `dev` 分支。

### 部署密钥 (Secrets)
需要在 GitHub Actions 中配置以下加密变量：
- `DEPLOY_HOST`: 生产服务器 IP。
- `DEPLOY_USER`: 部署用户。
- `DEPLOY_KEY`: SSH 私钥。

---

## 📖 API 文档
服务启动后，可以通过以下路径访问交互式文档：
- **Swagger UI**: `http://localhost:8080/docs`
- **ReDoc**: `http://localhost:8080/redoc`

更多详细说明见 [API 文档](./doc/API.md)。
