from datetime import datetime, timedelta, timezone
from typing import Optional, Union, Any
from jose import jwt, JWTError
from passlib.context import CryptContext
from itsdangerous import URLSafeTimedSerializer
from app.core.config import settings

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


def verify_password(plain_password: str, hashed_password: str) -> bool:
    return pwd_context.verify(plain_password, hashed_password)


def get_password_hash(password: str) -> str:
    return pwd_context.hash(password)


def create_access_token(subject: Union[str, Any], email: str, username: str, expires_delta: Optional[timedelta] = None) -> str:
    if expires_delta:
        expire = datetime.now(timezone.utc) + expires_delta
    else:
        expire = datetime.now(timezone.utc) + timedelta(minutes=15)
    
    to_encode = {
        "id": str(subject),
        "email": email,
        "username": username,
        "exp": expire,
        "iat": datetime.now(timezone.utc)
    }
    encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm="HS256")
    return encoded_jwt


def generate_email_verification_token(email: str) -> str:
    s = URLSafeTimedSerializer(settings.SECRET_KEY)
    return s.dumps(email, salt="email-verify")


def verify_email_token(token: str, max_age: int = 900) -> Optional[str]:
    s = URLSafeTimedSerializer(settings.SECRET_KEY)
    try:
        email = s.loads(token, salt="email-verify", max_age=max_age)
        return email
    except Exception:
        return None

def decode_access_token(token: str) -> Optional[dict]:
    try:
        payload = jwt.decode(token, settings.SECRET_KEY, algorithms=["HS256"])
        return payload
    except JWTError:
        return None
