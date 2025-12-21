from flask import current_app, url_for
from itsdangerous import URLSafeTimedSerializer
from flask_mail import Message
from datetime import datetime, timedelta, timezone
import jwt
import secrets

from app.extensions import db
from app.models import User
from app import mail
from loguru import logger


class AuthService:
    # ========== 通用工具 ==========

    @staticmethod
    def generate_verify_token(email):
        s = URLSafeTimedSerializer(current_app.config["SECRET_KEY"])
        return s.dumps(email, salt="email-verify")

    @staticmethod
    def confirm_verify_token(token, max_age=900):
        s = URLSafeTimedSerializer(current_app.config["SECRET_KEY"])
        try:
            email = s.loads(token, salt="email-verify", max_age=max_age)
            return email
        except Exception:
            return None

    @staticmethod
    def send_email(to, subject, body):
        msg = Message(subject=subject, recipients=[to], body=body)
        mail.send(msg)

    # ========== 核心 JWT 生成逻辑 ==========

    @staticmethod
    def generate_jwt(user, expire_hours=1):
        """生成用户登录 JWT"""
        payload = {
            "id": user.id,
            "email": user.email,
            "username": user.username,
            "exp": datetime.now(timezone.utc) + timedelta(hours=expire_hours),
            "iat": datetime.now(timezone.utc),
        }
        token = jwt.encode(payload, current_app.config["SECRET_KEY"], algorithm="HS256")
        return token

    # ========== 用户注册与验证 ==========

    @staticmethod
    def create_user(username, email, password):
        if User.query.filter(User.email == email).first():
            return None, "user already exists"

        user = User(username=username, email=email)
        user.set_password(password)
        user.is_active = False
        user.is_verified = False

        db.session.add(user)
        db.session.commit()

        token = AuthService.generate_verify_token(email)
        verify_url = url_for("auth.verify_email", token=token, _external=True)
        logger.info(f"Verification link for {email}: {verify_url}")

        AuthService.send_email(
            to=email,
            subject="Verify your account",
            body=f"Click to verify: {verify_url}",
        )

        return user, "user created, verification email sent"

    @staticmethod
    def verify_email(token):
        email = AuthService.confirm_verify_token(token)
        if not email:
            return None, "invalid or expired token"

        user = User.query.filter_by(email=email).first()
        if not user:
            return None, "user not found"

        if user.is_verified:
            return user, "already verified"

        user.is_verified = True
        user.is_active = True
        db.session.commit()
        return user, "email verified successfully"

    # ========== 登录逻辑 ==========

    @staticmethod
    def login_user(email, password):
        user = User.query.filter_by(email=email).first()
        if not user or not user.check_password(password):
            return None, "Invalid credentials, incorrect email or password"

        token = AuthService.generate_jwt(user)
        return token, "login successful"

    # ========== Token 解析 ==========

    @staticmethod
    def get_user_from_token(token):
        if not token:
            return None, "Missing token"
        if token.startswith("Bearer "):
            token = token[7:]

        try:
            payload = jwt.decode(
                token, current_app.config["SECRET_KEY"], algorithms=["HS256"]
            )
            user = User.query.get(payload["id"])
            if not user:
                return None, "User not found"
            return user, "Token valid"
        except jwt.ExpiredSignatureError:
            return None, "Token expired"
        except jwt.InvalidTokenError:
            return None, "Invalid token"

    # ========== Google 登录 ==========

    @staticmethod
    def login_or_register_google_user(email, name, avatar=None):
        """如果用户存在就登录，不存在则注册一个"""
        user = User.query.filter_by(email=email).first()

        if not user:
            # 创建一个新用户（Google 登录用户没有密码）
            user = User(
                username=name,
                email=email,
                password_hash=secrets.token_hex(16),
                is_verified=True,
                created_at=datetime.now(timezone.utc),
            )
            db.session.add(user)
            db.session.commit()

        token = AuthService.generate_jwt(user)
        return user, token

    @staticmethod
    def login_or_register_github_user(email, name):
        """GitHub 登录：如果用户存在就登录，否则注册"""
        user = User.query.filter_by(email=email).first()

        if not user:
            user = User(
                username=name,
                email=email,
                password_hash=secrets.token_hex(16),
                is_verified=True,
                created_at=datetime.now(timezone.utc),
            )
            db.session.add(user)
            db.session.commit()

        token = AuthService.generate_jwt(user)
        return user, token
