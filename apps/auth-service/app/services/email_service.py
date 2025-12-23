from app.core.config import settings
from fastapi_mail import FastMail, MessageSchema, ConnectionConfig, MessageType
import structlog


logger = structlog.get_logger()


class EmailService:
    @staticmethod
    async def send_email(to: str, subject: str, body: str):
        if not (settings.MAIL_USERNAME and settings.MAIL_PASSWORD):
            logger.warning("Mail credentials not configured, skipping email send.")
            return

        conf = ConnectionConfig(
            MAIL_USERNAME=settings.MAIL_USERNAME,
            MAIL_PASSWORD=settings.MAIL_PASSWORD,
            MAIL_FROM=settings.MAIL_USERNAME,
            MAIL_PORT=settings.MAIL_PORT,
            MAIL_SERVER=settings.MAIL_SERVER,
            MAIL_STARTTLS=settings.MAIL_USE_TLS,
            MAIL_SSL_TLS=settings.MAIL_USE_SSL,
            USE_CREDENTIALS=True,
            VALIDATE_CERTS=True,
        )
        message = MessageSchema(
            subject=subject, recipients=[to], body=body, subtype=MessageType.html
        )
        fm = FastMail(conf)
        try:
            await fm.send_message(message)
            logger.info(f"Email sent to {to}")
        except Exception as e:
            logger.error(f"Failed to send email: {e}")
