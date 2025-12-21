# models/user.py
from datetime import datetime, timezone
from app.extensions import db
from werkzeug.security import generate_password_hash, check_password_hash
import uuid


class User(db.Model):
    __tablename__ = "tb_users"

    # 主键 UUID
    id = db.Column(db.String(36), primary_key=True, default=lambda: str(uuid.uuid4()))

    # 用户信息
    username = db.Column(db.String(50), unique=True, nullable=False)
    email = db.Column(db.String(120), unique=True, nullable=False)
    password_hash = db.Column(db.String(255), nullable=False)

    # 时间字段
    created_at = db.Column(db.DateTime, default=lambda: datetime.now(timezone.utc))
    updated_at = db.Column(
        db.DateTime,
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc),
    )

    # 安全字段
    is_active = db.Column(db.Boolean, default=True)
    is_verified = db.Column(db.Boolean, default=False)

    # 密码操作方法
    def set_password(self, password: str):
        """生成密码哈希"""
        self.password_hash = generate_password_hash(password)
        print(self.password_hash)

    def check_password(self, password: str) -> bool:
        """验证密码"""
        return check_password_hash(self.password_hash, password)

    def __repr__(self):
        return f"<User uuid={self.id} username={self.username} email={self.email}>"
