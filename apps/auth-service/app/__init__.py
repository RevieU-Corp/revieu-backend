from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.auth.routes import auth_router
from app.logger import log
from app.config import settings


def create_app() -> FastAPI:
    app = FastAPI(
        title="Auth Service", description="Auth Service for RevieU", version="0.1.0"
    )

    # CORS setup
    app.add_middleware(
        CORSMiddleware,
        allow_origins=[settings.FRONTEND_URL],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # Include routers
    app.include_router(auth_router)

    log.info(f"✅ FastAPI app starting on {settings.ADDRESS}:{settings.PORT}")

    return app
