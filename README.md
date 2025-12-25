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

### 🚀 Getting Started

#### 1. Setup Environment

Navigate to the service directory:

```bash
cd apps/auth-service
```

Create a `.env` file based on your configuration. Key variables include:

```ini
# Database
POSTGRES_USER=postgres
POSTGRES_PASSWORD=yourpassword
POSTGRES_DB=USCRE
SQLALCHEMY_DATABASE_URI=postgresql+psycopg2://postgres:yourpassword@localhost:5432/USCRE

# Application
PORT=8082
DOMAIN=http://localhost:8082
FRONTEND_URL=http://localhost:3000

# Security
JWT_SECRET_KEY=your_secret_key

# OAuth (Optional)
GITHUB_CLIENT_ID=...
GOOGLE_CLIENT_ID=...
```

#### 2. Database Setup

Ensure PostgreSQL is running and create the database:

```bash
# Create database
psql -h localhost -U postgres -c 'CREATE DATABASE "USCRE";'

# Initialize schema
psql -h localhost -U postgres -d USCRE -f sample.sql
```

#### 3. Install Dependencies & Run

Using `uv`:

```bash
# Install dependencies and sync environment
uv sync

# Run the development server
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
