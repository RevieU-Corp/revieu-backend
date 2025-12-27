# RevieU Backend

Backend services for the RevieU platform, built with Python and FastAPI.

## 📂 Project Structure

The project follows a monorepo-style structure containing individual microservices:

```
revieu-backend/
├── apps/
│   └── auth-service/     # Authentication & User Management Service
└── README.md
```

## 🛠 Technology Stack

- **Language**: Python 3.12+
- **Framework**: FastAPI
- **Database**: PostgreSQL
- **ORM**: SQLAlchemy
- **Package Manager**: [uv](https://github.com/astral-sh/uv)
- **Logging**: Structlog

---

## 🔐 Auth Service

The `auth-service` handles user registration, login (email/password & OAuth), and profile management.

### Prerequisites

- Python 3.12 or higher
- PostgreSQL
- `uv` package manager

### 🚀 快速开始

#### 1. 设置环境与依赖

详细的子服务配置（如数据库、OAuth、部署等）请参阅各子目录下的 `README.md`。

```bash
cd apps/auth-service
uv sync
```

#### 2. 启动服务

```bash
cd apps/auth-service
uv run uvicorn main:app --reload --port 8082
```

The service will start at `http://localhost:8082`.

### 📚 API Documentation

Once the server is running, you can access the interactive API docs at:

- **Swagger UI**: [http://localhost:8082/api/v1/docs](http://localhost:8082/api/v1/docs)
- **ReDoc**: [http://localhost:8082/api/v1/redoc](http://localhost:8082/api/v1/redoc)

### ✅ Verification

You can verify the setup by running the included helper script (if available) or by checking the status via curl:

```bash
curl http://localhost:8082/api/v1/docs
```
