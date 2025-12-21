from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    SECRET_KEY: str = "you-will-never-guess"
    DATABASE_URL: str = "sqlite:///app.db"

    # Mail server settings
    MAIL_SERVER: Optional[str] = None
    MAIL_PORT: int = 25
    MAIL_USE_TLS: bool = False
    MAIL_USERNAME: Optional[str] = None
    MAIL_PASSWORD: Optional[str] = None

    # OAuth
    GITHUB_CLIENT_ID: Optional[str] = None
    GITHUB_CLIENT_SECRET: Optional[str] = None
    GITHUB_REDIRECT_URI: Optional[str] = None

    GOOGLE_CLIENT_ID: Optional[str] = None
    GOOGLE_CLIENT_SECRET: Optional[str] = None
    GOOGLE_REDIRECT_URI: Optional[str] = None

    FRONTEND_URL: str = "http://localhost:5173"

    PORT: int = 5000
    ADDRESS: str = "0.0.0.0"

    class Config:
        env_file = ".env"


settings = Settings()
