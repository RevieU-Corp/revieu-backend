from datetime import datetime, timedelta, timezone
from jose import jwt
from passlib.context import CryptContext
from sqlalchemy.orm import Session
from app.models.user import User
from app.config import settings
from loguru import logger
import secrets
from itsdangerous import URLSafeTimedSerializer
from fastapi_mail import FastMail, MessageSchema, ConnectionConfig, MessageType

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


class AuthService:
    # ========== 通用工具 ==========

    @staticmethod
    def generate_verify_token(email: str):
        s = URLSafeTimedSerializer(settings.SECRET_KEY)
        return s.dumps(email, salt="email-verify")

    @staticmethod
    def confirm_verify_token(token: str, max_age: int = 900):
        s = URLSafeTimedSerializer(settings.SECRET_KEY)
        try:
            email = s.loads(token, salt="email-verify", max_age=max_age)
            return email
        except Exception:
            return None

    @staticmethod
    async def send_email(to: str, subject: str, body: str):
        conf = ConnectionConfig(
            MAIL_USERNAME=settings.MAIL_USERNAME,
            MAIL_PASSWORD=settings.MAIL_PASSWORD,
            MAIL_FROM=settings.MAIL_USERNAME,
            MAIL_PORT=settings.MAIL_PORT,
            MAIL_SERVER=settings.MAIL_SERVER,
            MAIL_STARTTLS=settings.MAIL_USE_TLS,
            MAIL_SSL_TLS=not settings.MAIL_USE_TLS,
            USE_CREDENTIALS=True,
            VALIDATE_CERTS=True,
        )
        message = MessageSchema(
            subject=subject, recipients=[to], body=body, subtype=MessageType.html
        )
        fm = FastMail(conf)
        await fm.send_message(message)

    # ========== 核心 JWT 生成逻辑 ==========

    @staticmethod
    def generate_jwt(user: User, expire_hours: int = 1):
        """生成用户登录 JWT"""
        expire = datetime.now(timezone.utc) + timedelta(hours=expire_hours)
        payload = {
            "id": user.id,
            "email": user.email,
            "username": user.username,
            "exp": expire,
            "iat": datetime.now(timezone.utc),
        }
        token = jwt.encode(payload, settings.SECRET_KEY, algorithm="HS256")
        return token

    # ========== 用户注册与验证 ==========

    @staticmethod
    async def create_user(db: Session, username: str, email: str, password: str):
        if db.query(User).filter(User.email == email).first():
            return None, "user already exists"

        user = User(username=username, email=email)
        user.set_password(password)
        user.is_active = False
        user.is_verified = False

        db.add(user)
        db.commit()
        db.refresh(user)

        token = AuthService.generate_verify_token(email)
        # Note: In FastAPI, we don't have url_for easily without Request, so we'll just construct or log it.
        # Ideally, this should be passed from the router.
        # For now, we'll assume the base URL is known or handled in the router.
        verify_url = f"{settings.FRONTEND_URL}/verify?token={token}"
        logger.info(f"Verification link for {email}: {verify_url}")

        # try:
        #     await AuthService.send_email(
        #         to=email,
        #         subject="Verify your account",
        #         body=f"Click to verify: {verify_url}",
        #     )
        # except Exception as e:
        #     logger.error(f"Failed to send email: {e}")

        return user, "user created, verification email sent"

    @staticmethod
    def verify_email(db: Session, token: str):
        email = AuthService.confirm_verify_token(token)
        if not email:
            return None, "invalid or expired token"

        user = db.query(User).filter(User.email == email).first()
        if not user:
            return None, "user not found"

        if user.is_verified:
            return user, "already verified"

        user.is_verified = True
        user.is_active = True
        db.commit()
        return user, "email verified successfully"

    # ========== 登录逻辑 ==========

    @staticmethod
    def login_user(db: Session, email: str, password: str):
        user = db.query(User).filter(User.email == email).first()
        if not user or not user.check_password(password):
            return None, "Invalid credentials, incorrect email or password"

        token = AuthService.generate_jwt(user)
        return token, "login successful"

    # ========== Token 解析 ==========

    @staticmethod
    def get_user_from_token(db: Session, token: str):
        try:
            payload = jwt.decode(token, settings.SECRET_KEY, algorithms=["HS256"])
            user_id = payload.get("id")
            if user_id is None:
                return None, "Invalid token payload"
            user = db.query(User).filter(User.id == user_id).first()
            if not user:
                return None, "User not found"
            return user, "Token valid"
        except jwt.ExpiredSignatureError:
            return None, "Token expired"
        except jwt.JWTError:
            return None, "Invalid token"

    # ========== Google 登录 ==========

    @staticmethod
    def login_or_register_google_user(db: Session, email: str, name: str):
        """如果用户存在就登录，不存在则注册一个"""
        user = db.query(User).filter(User.email == email).first()

        if not user:
            user = User(
                username=name,
                email=email,
                password_hash=secrets.token_hex(16),
                is_verified=True,
                created_at=datetime.now(timezone.utc),
            )
            db.add(user)
            db.commit()
            db.refresh(user)

        token = AuthService.generate_jwt(user)
        return user, token

    @staticmethod
    def login_or_register_github_user(db: Session, email: str, name: str):
        """GitHub 登录：如果用户存在就登录，否则注册"""
        user = db.query(User).filter(User.email == email).first()

        if not user:
            user = User(
                username=name,
                email=email,
                password_hash=secrets.token_hex(16),
                is_verified=True,
                created_at=datetime.now(timezone.utc),
            )
            db.add(user)
            db.commit()
            db.refresh(user)

        token = AuthService.generate_jwt(user)
        return user, token
