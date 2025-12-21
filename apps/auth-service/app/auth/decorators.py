from functools import wraps
from flask import request, g, jsonify
from .services import AuthService

def token_required(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        auth_header = request.headers.get("Authorization")
        user, message = AuthService.get_user_from_token(auth_header)

        if not user:
            return jsonify({"error": message}), 401
        
        g.current_user = user
        return f(*args, **kwargs)
    return decorated_function
