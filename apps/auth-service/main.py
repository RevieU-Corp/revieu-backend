from dotenv import load_dotenv
import os
import uvicorn

# 指定加载路径
load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), ".env"))

from app import create_app
from app.config import settings

app = create_app()

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=settings.ADDRESS,
        port=settings.PORT,
        reload=True if os.getenv("ENV") == "development" else False,
    )
