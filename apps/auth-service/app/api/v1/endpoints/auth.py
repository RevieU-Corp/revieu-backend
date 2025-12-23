from fastapi import APIRouter, Depends, HTTPException, Query, Request, status
from fastapi.responses import RedirectResponse
from sqlalchemy.orm import Session
from app.services.auth_service import AuthService
from app.api.deps import get_db, get_current_user
from app.core.config import settings
from app.schemas.user import UserCreate, UserLogin, UserResponse, UserProfileUpdate
from app.schemas.common import MessageResponse
from app.models.user import User
import requests
from urllib.parse import urlencode
import structlog

logger = structlog.get_logger()

router = APIRouter()


# Step 1: 跳转到 GitHub 登录授权页
@router.get("/github/login")
def github_login():
    if not (settings.GITHUB_CLIENT_ID and settings.GITHUB_REDIRECT_URI):
        raise HTTPException(status_code=500, detail="GitHub login not configured")

    params = {
        "client_id": settings.GITHUB_CLIENT_ID,
        "redirect_uri": settings.GITHUB_REDIRECT_URI,
        "scope": "read:user user:email",
        "allow_signup": "true",
    }
    logger.info(
        f"redirect_uri:{settings.GITHUB_REDIRECT_URI}\nclient_id:{settings.GITHUB_CLIENT_ID}"
    )
    github_auth_url = f"https://github.com/login/oauth/authorize?{urlencode(params)}"
    return RedirectResponse(github_auth_url)


# Step 2: GitHub 回调
@router.get("/github/callback")
def github_callback(code: str, db: Session = Depends(get_db)):
    if not code:
        return {"code": 1, "message": "Missing code"}

    # 用 code 换取 access token
    token_url = "https://github.com/login/oauth/access_token"
    token_data = {
        "client_id": settings.GITHUB_CLIENT_ID,
        "client_secret": settings.GITHUB_CLIENT_SECRET,
        "code": code,
        "redirect_uri": settings.GITHUB_REDIRECT_URI,
    }

    headers = {"Accept": "application/json"}
    token_resp = requests.post(token_url, data=token_data, headers=headers).json()
    access_token = token_resp.get("access_token")

    if not access_token:
        # Avoid returning data directly in production
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

    # GitHub 可能没有公开邮箱，需要取第一个 primary email
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

    # 登录或注册
    user, token = AuthService.login_or_register_oauth_user(
        db, email, name, provider="github", avatar=user_info_resp.get("avatar_url")
    )

    frontend_callback_url = f"{settings.FRONTEND_URL}/oauth-callback?token={token}"
    return RedirectResponse(frontend_callback_url)


# Step 1: 跳转到 Google 登录
@router.get("/google/login")
def google_login():
    if not (settings.GOOGLE_CLIENT_ID and settings.GOOGLE_REDIRECT_URI):
        raise HTTPException(status_code=500, detail="Google login not configured")

    params = {
        "client_id": settings.GOOGLE_CLIENT_ID,
        "redirect_uri": settings.GOOGLE_REDIRECT_URI,
        "response_type": "code",
        "scope": "openid email profile",
        "access_type": "offline",
    }
    google_auth_url = (
        f"https://accounts.google.com/o/oauth2/v2/auth?{urlencode(params)}"
    )
    return RedirectResponse(google_auth_url)


# Step 2: Google 回调
@router.get("/google/callback")
def google_callback(code: str, db: Session = Depends(get_db)):
    if not code:
        return {"code": 1, "message": "Missing code"}

    # 用 code 换取 access token
    token_url = "https://oauth2.googleapis.com/token"
    token_data = {
        "code": code,
        "client_id": settings.GOOGLE_CLIENT_ID,
        "client_secret": settings.GOOGLE_CLIENT_SECRET,
        "redirect_uri": settings.GOOGLE_REDIRECT_URI,
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

    frontend_callback_url = f"{settings.FRONTEND_URL}/oauth-callback?token={token}"
    return RedirectResponse(frontend_callback_url)


# 注册
@router.post(
    "/register", response_model=MessageResponse, status_code=status.HTTP_201_CREATED
)
async def register(user_in: UserCreate, db: Session = Depends(get_db)):
    user, message = await AuthService.register_user(
        db, user_in.username, user_in.email, user_in.password
    )

    if not user:
        # Should probably raise HTTPException depending on message
        # Original code raised 400 with {code:1, message:message}
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
@router.post("/login")
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

    frontend_callback_url = f"{settings.FRONTEND_URL}/oauth-callback?token={token}"
    # Original behavior returned RedirectResponse for POST login?
    # Usually login API returns JSON with token.
    # But original code returned RedirectResponse.
    # That is weird for a JSON API used by frontend, but maybe they want to redirect immediately?
    # I will keep original behavior for compatibility.
    return RedirectResponse(frontend_callback_url)


# 获取用户信息（受保护接口）
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
    profile_in: UserProfileUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    if not current_user.profile:
        from app.models.user import UserProfile

        current_user.profile = UserProfile(user_id=current_user.id)

    if profile_in.nickname is not None:
        current_user.profile.nickname = profile_in.nickname
    if profile_in.avatar is not None:
        current_user.profile.avatar = profile_in.avatar
    if profile_in.bio is not None:
        current_user.profile.bio = profile_in.bio

    db.commit()
    db.refresh(current_user)

    return {
        "code": 0,
        "message": "User profile updated successfully",
        "data": UserResponse.model_validate(current_user).model_dump(),
    }
