from fastapi import APIRouter, Depends, HTTPException, Query, Request, status
from fastapi.responses import RedirectResponse
from sqlalchemy.orm import Session
from app.services.auth_service import AuthService
from app.api.deps import get_db, get_current_user
from app.core.config import settings
from app.schemas.user import UserCreate, UserLogin, UserResponse, UserUpdate
from app.schemas.common import MessageResponse
from app.models.user import User
import requests
from urllib.parse import urlencode
import structlog

logger = structlog.get_logger()

router = APIRouter()


# Step 1: 跳转到 GitHub 登录授权页
@router.get("/login/github")
def github_login(request: Request):
    # Dynamically generate redirect_uri
    redirect_uri = str(request.url_for("github_callback"))

    if not settings.GITHUB_CLIENT_ID:
        raise HTTPException(status_code=500, detail="GitHub login not configured")

    params = {
        "client_id": settings.GITHUB_CLIENT_ID,
        "redirect_uri": redirect_uri,
        "scope": "read:user user:email",
        "allow_signup": "true",
    }
    logger.debug(f"GitHub login redirect_uri: {redirect_uri}")
    github_auth_url = f"https://github.com/login/oauth/authorize?{urlencode(params)}"
    return RedirectResponse(github_auth_url)


# Step 2: GitHub 回调
@router.get("/callback/github")
def github_callback(request: Request, code: str, db: Session = Depends(get_db)):
    if not code:
        return {"code": 1, "message": "Missing code"}

    redirect_uri = str(request.url_for("github_callback"))

    # 用 code 换取 access token
    token_url = "https://github.com/login/oauth/access_token"
    token_data = {
        "client_id": settings.GITHUB_CLIENT_ID,
        "client_secret": settings.GITHUB_CLIENT_SECRET,
        "code": code,
        "redirect_uri": redirect_uri,
    }

    headers = {"Accept": "application/json"}
    token_resp = requests.post(token_url, data=token_data, headers=headers).json()
    access_token = token_resp.get("access_token")

    if not access_token:
        return {"code": 1, "message": "Failed to get access token"}

    # 获取用户信息
    user_info_resp = requests.get(
        "https://api.github.com/user",
        headers={"Authorization": f"Bearer {access_token}"},
    ).json()

    email_resp = requests.get(
        "https://api.github.com/user/emails",
        headers={"Authorization": f"Bearer {access_token}"},
    ).json()

    email = None
    if isinstance(email_resp, list):
        primary_emails = [
            e["email"] for e in email_resp if e.get("primary") and e.get("verified")
        ]
        if primary_emails:
            email = primary_emails[0]

    if not email:
        return {"code": 1, "message": "No verified email found in GitHub account"}

    name = user_info_resp.get("name") or user_info_resp.get("login")

    user, token = AuthService.login_or_register_oauth_user(
        db, email, name, provider="github", avatar=user_info_resp.get("avatar_url")
    )

    return RedirectResponse(f"{settings.FRONTEND_URL}/oauth-callback?token={token}")


# Step 1: 跳转到 Google 登录
@router.get("/login/google")
def google_login(request: Request):
    redirect_uri = str(request.url_for("google_callback"))
    logger.debug(f"Google login redirect_uri: {redirect_uri}")

    if not settings.GOOGLE_CLIENT_ID:
        raise HTTPException(status_code=500, detail="Google login not configured")

    params = {
        "client_id": settings.GOOGLE_CLIENT_ID,
        "redirect_uri": redirect_uri,
        "response_type": "code",
        "scope": "openid email profile",
        "access_type": "offline",
    }
    google_auth_url = (
        f"https://accounts.google.com/o/oauth2/v2/auth?{urlencode(params)}"
    )
    return RedirectResponse(google_auth_url)


# Step 2: Google 回调
@router.get("/callback/google")
def google_callback(request: Request, code: str, db: Session = Depends(get_db)):
    if not code:
        return {"code": 1, "message": "Missing code"}

    redirect_uri = str(request.url_for("google_callback"))

    # 用 code 换取 access token
    token_url = "https://oauth2.googleapis.com/token"
    token_data = {
        "code": code,
        "client_id": settings.GOOGLE_CLIENT_ID,
        "client_secret": settings.GOOGLE_CLIENT_SECRET,
        "redirect_uri": redirect_uri,
        "grant_type": "authorization_code",
    }

    token_resp = requests.post(token_url, data=token_data).json()
    access_token = token_resp.get("access_token")

    if not access_token:
        return {"code": 1, "message": "Failed to get access token"}

    # 获取用户信息
    user_info_resp = requests.get(
        "https://www.googleapis.com/oauth2/v3/userinfo",
        headers={"Authorization": f"Bearer {access_token}"},
    ).json()

    email = user_info_resp.get("email")
    name = user_info_resp.get("name")

    if not email:
        return {"code": 1, "message": "No email found in Google account"}

    user, token = AuthService.login_or_register_oauth_user(
        db, email, name, provider="google", avatar=user_info_resp.get("picture")
    )

    return RedirectResponse(f"{settings.FRONTEND_URL}/oauth-callback?token={token}")


# 注册
@router.post(
    "/register", response_model=MessageResponse, status_code=status.HTTP_201_CREATED
)
async def register(
    request: Request, user_in: UserCreate, db: Session = Depends(get_db)
):
    # Get base URL for verification link
    base_url = str(request.base_url).rstrip("/")
    user, message = await AuthService.register_user(
        db, user_in.username, user_in.email, user_in.password, base_url=base_url
    )

    if not user:
        raise HTTPException(status_code=400, detail={"code": 1, "message": message})

    return {
        "code": 0,
        "message": message,
        "data": {
            "id": user.id,
            "email": user.email,
            "username": user.username,
            "created_at": user.created_at.isoformat(),
        },
    }


# 注册验证
@router.get("/verify")
def verify_email(token: str = Query(...), db: Session = Depends(get_db)):
    user, message = AuthService.verify_email(db, token)

    if not user:
        code = 2 if message == "user not found" else 1
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": code, "message": message},
        )

    code = 3 if message == "already verified" else 0
    return {"code": code, "message": message}


# 登录
@router.post("/login/revieu")
def login(login_in: UserLogin, request: Request, db: Session = Depends(get_db)):
    ip_address = request.client.host if request.client else None
    token, message = AuthService.login_user(
        db, login_in.email, login_in.password, ip_address=ip_address
    )

    if not token:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"code": 1, "message": message},
        )

    # 生产标准：API 应该只返回数据，由前端决定如何跳转
    return {"code": 0, "message": "Login successful", "data": {"token": token}}


# 获取用户信息（受保护接口）
@router.get("/profile", response_model=MessageResponse)
def profile(current_user: User = Depends(get_current_user)):
    return {
        "code": 0,
        "message": "User profile fetched successfully",
        "data": UserResponse.model_validate(current_user).model_dump(),
    }


# 更新用户信息（受保护接口）
@router.put("/profile", response_model=MessageResponse)
def update_profile(
    user_update: UserUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    if user_update.nickname is not None:
        current_user.nickname = user_update.nickname
    if user_update.avatar is not None:
        current_user.avatar = user_update.avatar
    if user_update.bio is not None:
        current_user.bio = user_update.bio

    db.commit()
    db.refresh(current_user)

    return {
        "code": 0,
        "message": "User profile updated successfully",
        "data": UserResponse.model_validate(current_user).model_dump(),
    }
