from datetime import datetime, timezone
import secrets
import structlog
from sqlalchemy.orm import Session


class AuthService:
    @staticmethod
    async def register_user(db: Session, username: str, email: str, password: str):
        if db.query(User).filter(User.email == email).first():
            return None, "user already exists"

        user = User(username=username, email=email, nickname=username)
        user.set_password(password)
        user.is_active = False
        user.is_verified = False

        db.add(user)
        db.commit()
        db.refresh(user)

        token = generate_email_verification_token(email)
        # verify_url = f"{settings.FRONTEND_URL}/verify?token={token}"
        # Use backend domain for verification link
        verify_url = f"{settings.DOMAIN}/api/v1/auth/verify?token={token}"
        logger.info(f"Verification link for {email}: {verify_url}")

        # Send email (async fire and forget or await)
        # In production, use background tasks
        await EmailService.send_email(
            to=email,
            subject="Verify your account",
            body=f"Click to verify: {verify_url}",
        )

        return user, "user created, verification email sent"

    @staticmethod
    def verify_email(db: Session, token: str):
        email = verify_email_token(token)
        if not email:
            return None, "invalid or expired token"

        user = db.query(User).filter(User.email == email).first()
        if not user:
            return None, "user not found"

        if user.is_verified:
            return user, "already verified"

        user.is_verified = True
        user.is_active = True
        db.commit()
        return user, "email verified successfully"

    @staticmethod
    def login_user(db: Session, email: str, password: str, ip_address: str = None):
        user = db.query(User).filter(User.email == email).first()
        if not user or not user.check_password(password):
            return None, "Invalid credentials, incorrect email or password"

        user.last_login_at = datetime.now(timezone.utc)
        if ip_address:
            user.last_login_ip = ip_address
        db.commit()

        token = create_access_token(user.id, user.email, user.username)
        return token, "login successful"

    @staticmethod
    def login_or_register_oauth_user(
        db: Session, email: str, name: str, provider: str = "oauth", avatar: str = None
    ):
        """OAuth login: login if exists, else register"""
        user = db.query(User).filter(User.email == email).first()

        if not user:
            user = User(
                username=name,
                email=email,
                # Random password for oauth users
                password_hash=secrets.token_hex(16),
                is_verified=True,
                created_at=datetime.now(timezone.utc),
            )
            # Since we don't have set_password call here, we should perhaps hash the random password
            # or handle it differently. The original code set password_hash to the hex,
            # which means they can't login with password unless they reset it.
            # But wait, original code did: password_hash=secrets.token_hex(16)
            # This is storing 'plain' random hex in password_hash column?
            # If User.check_password uses verify_password, it will try to verify(plain, hash).
            # If hash is not a hash, passlib might complain or fail.
            # Ideally we should hash it.
            # But following original behavior for now or improved?
            # Let's improve and hash it so internal consistency is kept.

            user.set_password(secrets.token_hex(16))
            user.nickname = name
            user.avatar = avatar

            db.add(user)
            db.commit()
            db.refresh(user)

        # Update login info for oauth user too
        user.last_login_at = datetime.now(timezone.utc)
        # oauth login usually happens via redirect so obtaining IP might be tricky in this method directly
        # or we update it if we have request context. For now just time.
        db.commit()

        token = create_access_token(user.id, user.email, user.username)
        return user, token
