import logging
import sys
import os
import structlog
from structlog.types import Processor

# 日志文件路径
LOG_DIR = os.getenv("LOG_DIR", "logs")
os.makedirs(LOG_DIR, exist_ok=True)
LOG_FILE_PATH = os.path.join(LOG_DIR, "app.log")


def setup_logging():
    shared_processors: list[Processor] = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.UnicodeDecoder(),
    ]

    # Configure Structlog
    structlog.configure(
        processors=shared_processors
        + [
            # Prepare event dict for stdlib logging (do not render to JSON here)
            structlog.stdlib.ProcessorFormatter.wrap_for_formatter,
        ],
        logger_factory=structlog.stdlib.LoggerFactory(),
        wrapper_class=structlog.stdlib.BoundLogger,
        cache_logger_on_first_use=True,
    )

    # Configure Standard Logging (to capture Uvicorn/FastAPI logs)
    formatter = structlog.stdlib.ProcessorFormatter(
        # These run ONLY on `logging` entries that do NOT come from structlog
        foreign_pre_chain=shared_processors,
        # These run on ALL entries
        processors=[
            structlog.stdlib.ProcessorFormatter.remove_processors_meta,
            structlog.processors.JSONRenderer(),
        ],
    )

    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(formatter)

    # File Handler
    file_handler = logging.FileHandler(LOG_FILE_PATH, encoding="utf-8")
    file_handler.setFormatter(formatter)

    root_logger = logging.getLogger()
    root_logger.addHandler(handler)
    root_logger.addHandler(file_handler)

    # 1. 获取环境变量中的 LOG_LEVEL，如果没有设置，默认使用 INFO
    # 这样你就可以通过环境变量动态控制，而不用改代码
    log_level_name = os.getenv("LOG_LEVEL", "DEBUG").upper()

    # 2. 将字符串转换为 logging 常量 (例如 "DEBUG" -> 10)
    # 也可以简单粗暴用: log_level = getattr(logging, log_level_name, logging.INFO)
    log_level = logging.getLevelName(log_level_name)

    # 注意：getLevelName 在某些 Python 版本返回的是数字，有些是字符串，
    # 比较稳妥的写法是直接用 getattr:
    if not isinstance(log_level, int):
        log_level = getattr(logging, log_level_name, logging.INFO)

    root_logger = logging.getLogger()
    root_logger.addHandler(handler)
    root_logger.addHandler(file_handler)

    # 3. 这里应用动态获取的级别
    root_logger.setLevel(log_level)

    # Intercept Uvicorn logs
    for _log in ["uvicorn", "uvicorn.error", "uvicorn.access"]:
        # Clear existing handlers
        logging.getLogger(_log).handlers = []
        logging.getLogger(_log).propagate = True

    # Get a logger instance to export
    return structlog.get_logger()


# Export a default logger (though best practice is get_logger() in each file)
log = setup_logging()
