# app/logger.py
import os
from loguru import logger

# 日志文件路径
LOG_DIR = os.getenv("LOG_DIR", "logs")
os.makedirs(LOG_DIR, exist_ok=True)

LOG_FILE_PATH = os.path.join(LOG_DIR, "app.log")

# 日志格式
LOG_FORMAT = (
    "<green>{time:YYYY-MM-DD HH:mm:ss.SSS}</green> | "
    "<level>{level: <8}</level> | "
    "<cyan>{name}</cyan>:<cyan>{function}</cyan>:<cyan>{line}</cyan> - "
    "<level>{message}</level>"
)

# 移除默认的 handler（loguru 默认只输出到 stdout）
logger.remove()

# 输出到终端
logger.add(
    sink=lambda msg: print(msg, end=""),  # 直接输出到控制台
    format=LOG_FORMAT,
    colorize=True,
    level="INFO",
)

# 输出到文件（每天新文件，保留7天）
logger.add(
    LOG_FILE_PATH,
    rotation="00:00",  # 每天凌晨新文件
    retention="7 days",  # 保留7天
    encoding="utf-8",
    enqueue=True,  # 多线程安全
    level="INFO",
    format=LOG_FORMAT,
)

# 导出 logger 实例
log = logger
