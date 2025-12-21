# test/db_test.py
from flask import Flask
from flask_sqlalchemy import SQLAlchemy
from sqlalchemy import text
import uuid
import sys
import os
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))
from app.models import User
from app.extensions import db

# Flask app
app = Flask(__name__)

# 配置 MySQL 数据库 URI
app.config['SQLALCHEMY_DATABASE_URI'] = "mysql+pymysql://root:123456@172.17.245.161:3306/USCRE?charset=utf8mb4"
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

# 初始化 SQLAlchemy
db.init_app(app)

if __name__ == "__main__":
    with app.app_context():
        try:
            # 测试连接
            result = db.session.execute(text("SELECT 1"))
            print("数据库连接成功！结果:", result.scalar())

            # ==== CREATE ====
            new_user = User(
                id=str(uuid.uuid4()),
                username="test_user",
                email="test_user111@example.com",
                password_hash = "pass_test"
            )
            db.session.add(new_user)
            db.session.commit()
            print("新增用户成功:", new_user.id, new_user.username)

            # ==== READ ====
            users = User.query.all()
            print(f"当前 users 表记录数: {len(users)}")
            for u in users:
                print(f"id={u.id}, username={u.username}, email={u.email}")

            # ==== UPDATE ====
            user_to_update = User.query.filter_by(username="test_user").first()
            if user_to_update:
                user_to_update.username = "updated_user"
                db.session.commit()
                print("更新用户成功:", user_to_update.id, user_to_update.username)

            # ==== DELETE ====
            user_to_delete = User.query.filter_by(username="updated_user").first()
            if user_to_delete:
                db.session.delete(user_to_delete)
                db.session.commit()
                print("删除用户成功:", user_to_delete.id)

        except Exception as e:
            print("数据库操作失败:", e)
