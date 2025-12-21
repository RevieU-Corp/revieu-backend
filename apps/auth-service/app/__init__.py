from flask import Flask
from .extensions import db, migrate
from app.logger import log
from flask_mail import Mail
from app.config import Config
from flask_cors import CORS
import os

mail = Mail()


def create_app():
    app = Flask(__name__)
    config_name = os.getenv("FLASK_ENV", "default")
    app.config.from_object(Config)

    db.init_app(app)
    migrate.init_app(app, db)
    mail.init_app(app)

    # 延迟导入蓝图
    from .auth.routes import auth_bp

    app.register_blueprint(auth_bp, url_prefix="/api/v1/auth")

    CORS(app)

    log.info(f"✅ Flask app starting in {config_name} mode")
    log.debug("Debug logs are enabled")

    return app
