from typing import Optional, Any
from pydantic import Field, PostgresDsn, field_validator, ValidationInfo
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    PROJECT_NAME: str = "Auth Service"
    VERSION: str = "0.1.0"
    # 核心配置：开发环境设为 http://localhost:5173，生产环境设为空字符串即可实现相对路径
    FRONTEND_URL: str = ""

    # Database
    POSTGRES_USER: str = "postgres"
    POSTGRES_PASSWORD: str = "123456"
    POSTGRES_DB: str = "revieu"
    POSTGRES_HOST: str = "localhost"
    POSTGRES_PORT: int = 5432
    SQLALCHEMY_DATABASE_URI: Optional[str] = None

    @field_validator("SQLALCHEMY_DATABASE_URI", mode="before")
    @classmethod
    def assemble_db_connection(cls, v: Optional[str], info: ValidationInfo) -> Any:
        if isinstance(v, str):
            return v
        return str(
            PostgresDsn.build(
                scheme="postgresql+psycopg2",
                username=info.data.get("POSTGRES_USER"),
                password=info.data.get("POSTGRES_PASSWORD"),
                host=info.data.get("POSTGRES_HOST"),
                port=info.data.get("POSTGRES_PORT"),
                path=f"{info.data.get('POSTGRES_DB') or ''}",
            )
        )

    # Security
    JWT_SECRET_KEY: str = "JWT_SECRET_KEY"

    # Mail
    MAIL_SERVER: str = "smtp.gmail.com"
    MAIL_PORT: int = 465
    MAIL_USE_TLS: bool = False
    MAIL_USE_SSL: bool = True
    MAIL_USERNAME: Optional[str] = None
    MAIL_PASSWORD: Optional[str] = None
    MAIL_SENDER_NAME: Optional[str] = "RevieU"
    MAIL_SENDER_EMAIL: Optional[str] = None

    # OAuth
    GITHUB_CLIENT_ID: Optional[str] = None
    GITHUB_CLIENT_SECRET: Optional[str] = None
    GOOGLE_CLIENT_ID: Optional[str] = None
    GOOGLE_CLIENT_SECRET: Optional[str] = None

    model_config = SettingsConfigDict(
        env_file=".env", case_sensitive=True, extra="ignore"
    )

    @property
    def MAIL_DEFAULT_SENDER(self):
        if self.MAIL_SENDER_EMAIL:
            return (self.MAIL_SENDER_NAME, self.MAIL_SENDER_EMAIL)
        return (self.MAIL_SENDER_NAME, self.MAIL_USERNAME)


settings = Settings()
