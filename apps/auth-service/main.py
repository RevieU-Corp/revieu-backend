from dotenv import load_dotenv
import os

# 指定加载路径（保证 gunicorn 时也能找到）
print(os.path.join(os.path.dirname(__file__), ".env"))
load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), ".env"))

from app import create_app

app = create_app()

if __name__ == "__main__":
    port = os.getenv("PORT", 5000)
    address = os.getenv("ADDRESS", "0.0.0.0")
    app.run(host=address, port=port, debug=True)
