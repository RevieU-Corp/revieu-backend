# RevieU Backend

Backend services for the RevieU platform.

## ⚠️ 首次克隆后必须执行

```bash
./scripts/setup.sh
```

这会自动安装 Git hooks 防止提交未加密的 secrets 和强制 commit message 规范。

---

## 📂 Project Structure

```
revieu-backend/
├── apps/
│   ├── core/              # Core API Service (Go)
│   └── example-service/   # Placeholder for future services
├── scripts/
│   ├── setup.sh          # 🔥 首次运行这个！
│   ├── seal-secrets.sh   # 加密 secrets
│   └── hooks/            # Git hooks
└── lefthook.yml          # Hook 配置
```

## 🛠 Technology Stack

- **Core Service**: Go 1.24+, Gin, GORM, PostgreSQL
- **Infrastructure**: Kubernetes (k3s), ArgoCD, Sealed Secrets
- **CI/CD**: GitHub Actions

---

## 🚀 Quick Start

### 1. Clone and Setup

```bash
git clone git@github.com:RevieU-Corp/revieu-backend.git
cd revieu-backend

# 🔥 重要：安装 Git hooks
./scripts/setup.sh
```

### 2. Configure Secrets

```bash
# Copy example secrets
cp apps/core/configs/secrets.yaml.example apps/core/configs/secrets.yaml

# Edit with your values
vim apps/core/configs/secrets.yaml

# Encrypt before committing
./scripts/seal-secrets.sh
```

### 3. Run Core Service

```bash
cd apps/core
go run cmd/app/main.go
```

---

## 🔐 Secrets Management

**Never commit unencrypted secrets!** The Git hooks will block you.

### Workflow

1. Edit `apps/core/configs/secrets.yaml`
2. Encrypt: `./scripts/seal-secrets.sh`
3. Commit: `git add apps/core/configs/sealed-secrets.yaml`

See [scripts/README.md](scripts/README.md) for details.

---

## 📝 Commit Message Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

<body>

Closes #<issue-id>

Co-Authored-By: Name <email@example.com>
```

**Types**: feat, fix, docs, style, refactor, perf, test, chore, ci, build, revert

Examples:

- `feat(core): add user authentication`
- `fix(api): resolve null pointer exception`
- `docs(readme): update installation instructions`

The commit-msg hook enforces subject format, non-empty body, issue close trailer, and `Co-Authored-By` trailer.

---

## 🧪 Development

### Run Tests

```bash
cd apps/core
go test ./...
```

### Run with Docker

```bash
docker build -t revieu-core -f apps/core/build/package/Dockerfile apps/core
docker run -p 8080:8080 revieu-core
```

---

## 📚 Documentation

- [Scripts README](scripts/README.md) - Secrets management and hooks
- [Core Service](apps/core/) - API documentation
- [Infrastructure Repo](https://github.com/RevieU-Corp/revieu-infra) - K8s configs

---

## 🤝 Contributing

1. Run `./scripts/setup.sh` first
2. Create a feature branch
3. Follow commit message conventions
4. Encrypt secrets before committing
5. Create a PR to `dev` branch

---

## 📄 License

[Add your license here] test
