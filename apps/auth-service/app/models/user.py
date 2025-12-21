# models/user.py
from datetime import datetime, timezone
from sqlalchemy import Column, String, DateTime, Boolean
from app.db.base import Base
from app.core.security import get_password_hash, verify_password
import uuid


class User(Base):
    __tablename__ = "tb_users"

    # 主键 UUID
    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))

    # 用户信息
    username = Column(String(50), unique=True, nullable=False)
    email = Column(String(120), unique=True, nullable=False)
    password_hash = Column(String(255), nullable=False)

    # 时间字段
    created_at = Column(DateTime, default=lambda: datetime.now(timezone.utc))
    updated_at = Column(
        DateTime,
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc),
    )

    # 安全字段
    is_active = Column(Boolean, default=True)
    is_verified = Column(Boolean, default=False)

    # 密码操作方法
    def set_password(self, password: str):
        """生成密码哈希"""
        self.password_hash = get_password_hash(password)

    def check_password(self, password: str) -> bool:
        """验证密码"""
        return verify_password(password, self.password_hash)

    def __repr__(self):
        return f"<User uuid={self.id} username={self.username} email={self.email}>"
