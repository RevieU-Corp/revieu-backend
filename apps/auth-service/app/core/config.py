
from typing import Optional
from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    # App
    PROJECT_NAME: str = "Auth Service"
    VERSION: str = "0.1.0"
    ADDRESS: str = "0.0.0.0"
    PORT: int = 8082
    FRONTEND_URL: str = "http://localhost:3000"
    ENV: str = "development"

    # Database
    SQLALCHEMY_DATABASE_URI: str
    
    # Security
    SECRET_KEY: str = Field(validation_alias="JWT_SECRET_KEY")

    # Mail
    MAIL_SERVER: str = "smtp.gmail.com"
    MAIL_PORT: int = 465
    MAIL_USE_TLS: bool = False
    MAIL_USE_SSL: bool = True
    MAIL_USERNAME: Optional[str] = None
    MAIL_PASSWORD: Optional[str] = None
    MAIL_SENDER_NAME: Optional[str] = "USCRE"
    MAIL_SENDER_EMAIL: Optional[str] = None

    # OAuth - GitHub
    GITHUB_CLIENT_ID: Optional[str] = None
    GITHUB_CLIENT_SECRET: Optional[str] = None
    GITHUB_REDIRECT_URI: Optional[str] = None
    
    # OAuth - Google
    GOOGLE_CLIENT_ID: Optional[str] = None
    GOOGLE_CLIENT_SECRET: Optional[str] = None
    GOOGLE_REDIRECT_URI: Optional[str] = None

    model_config = SettingsConfigDict(
        env_file=".env", case_sensitive=True, extra="ignore"
    )

    @property
    def MAIL_DEFAULT_SENDER(self):
        if self.MAIL_SENDER_EMAIL:
            return (self.MAIL_SENDER_NAME, self.MAIL_SENDER_EMAIL)
        return (self.MAIL_SENDER_NAME, self.MAIL_USERNAME)


settings = Settings()
