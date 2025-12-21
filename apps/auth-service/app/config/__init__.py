import os


class Config:
    # Flask 基本配置
    SQLALCHEMY_DATABASE_URI = os.getenv("SQLALCHEMY_DATABASE_URI")
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    SECRET_KEY = os.getenv("JWT_SECRET_KEY")

    # 邮件服务器配置
    MAIL_SERVER = os.getenv("MAIL_SERVER", "smtp.gmail.com")
    MAIL_PORT = int(os.getenv("MAIL_PORT", 465))  # 转成整数
    MAIL_USE_TLS = os.getenv("MAIL_USE_TLS", "False") == "True"  # 字符串转布尔
    MAIL_USE_SSL = os.getenv("MAIL_USE_SSL", "True") == "True"
    MAIL_USERNAME = os.getenv("MAIL_USERNAME")
    MAIL_PASSWORD = os.getenv("MAIL_PASSWORD")
    MAIL_DEFAULT_SENDER = (
        os.getenv("MAIL_SENDER_NAME", "USCRE"),
        os.getenv("MAIL_SENDER_EMAIL", MAIL_USERNAME),
    )
