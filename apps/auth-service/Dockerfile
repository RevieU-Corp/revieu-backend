FROM python:3.12-slim

# 安装 curl 并安装 uv
RUN apt-get update && apt-get install -y curl \
    && curl -LsSf https://astral.sh/uv/install.sh | sh \
    && rm -rf /var/lib/apt/lists/*

# 把 uv 的路径加到全局 PATH
ENV PATH="/root/.local/bin:${PATH}"

WORKDIR /app
COPY . .

# 构建阶段同步 uv 环境
RUN uv sync
RUN uv add gunicorn

# 容器启动时运行你的应用
CMD [".venv/bin/gunicorn", "-w", "5", "-b", "0.0.0.0:8082", "main:app"]
