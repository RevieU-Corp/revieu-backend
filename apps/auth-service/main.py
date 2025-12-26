import uvicorn
from app.main import create_app
from app.core.config import settings

app = create_app()

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=settings.ADDRESS,
        port=settings.PORT,
        reload=settings.ENV == "development",
    )
