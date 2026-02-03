package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc         Service
	oauthCfg    config.OAuthConfig
	frontendURL string
	apiBasePath string
}

func NewHandler(jwtCfg config.JWTConfig, oauthCfg config.OAuthConfig, smtpCfg config.SMTPConfig, frontendURL string, apiBasePath string) *Handler {
	return &Handler{
		svc:         NewService(nil, jwtCfg, smtpCfg),
		oauthCfg:    oauthCfg,
		frontendURL: frontendURL,
		apiBasePath: apiBasePath,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with username, email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register Request"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + c.Request.Host

	user, err := h.svc.Register(c.Request.Context(), req.Username, req.Email, req.Password, baseURL)
	if err != nil {
		logger.Error(c.Request.Context(), "Registration failed",
			"error", err.Error(),
			"event", "user_register_failed",
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		Message: "User created successfully. Please check your email for verification link (printed in server logs for now).",
		UserID:  user.ID,
	})
}

// Login godoc
// @Summary Login user
// @Description Login with email and password to get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password, ipAddress)
	if err != nil {
		logger.Warn(c.Request.Context(), "Login failed",
			"error", err.Error(),
			"email", req.Email,
			"event", "user_login_failed",
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token, Type: "Bearer"})
}

// GoogleLogin godoc
// @Summary Redirect to Google OAuth
// @Description Redirects user to Google OAuth authorization page
// @Tags auth
// @Success 302 "Redirect to Google OAuth"
// @Router /auth/login/google [get]
func (h *Handler) GoogleLogin(c *gin.Context) {
	clientID := h.oauthCfg.Google.ClientID
	if clientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Google OAuth not configured"})
		return
	}

	frontendURL := h.frontendURL
	if frontendURL == "" {
		if referer := c.GetHeader("Referer"); referer != "" {
			if parsedURL, err := url.Parse(referer); err == nil {
				frontendURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
			}
		} else if origin := c.GetHeader("Origin"); origin != "" {
			frontendURL = origin
		} else {
			frontendURL = "http://localhost:3000"
		}
	}

	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
		scheme = "https"
	} else if c.Request.TLS != nil {
		scheme = "https"
	}
	redirectURI := fmt.Sprintf("%s://%s%s/auth/callback/google", scheme, c.Request.Host, h.apiBasePath)

	state := url.QueryEscape(frontendURL)

	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&state=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape("openid email profile"),
		state,
	)

	c.Redirect(http.StatusFound, authURL)
}

// GoogleCallback godoc
// @Summary Handle Google OAuth callback
// @Description Handles Google OAuth callback, creates/logs in user, redirects to frontend with token
// @Tags auth
// @Param code query string true "Authorization code from Google"
// @Success 302 "Redirect to frontend with token"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/callback/google [get]
func (h *Handler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	state := c.Query("state")
	frontendURL := h.frontendURL
	if frontendURL == "" {
		if state != "" {
			if decodedURL, err := url.QueryUnescape(state); err == nil && decodedURL != "" {
				frontendURL = decodedURL
			} else {
				frontendURL = "http://localhost:3000"
			}
		} else {
			frontendURL = "http://localhost:3000"
		}
	}

	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
		scheme = "https"
	} else if c.Request.TLS != nil {
		scheme = "https"
	}
	redirectURI := fmt.Sprintf("%s://%s%s/auth/callback/google", scheme, c.Request.Host, h.apiBasePath)

	tokenResp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code":          {code},
		"client_id":     {h.oauthCfg.Google.ClientID},
		"client_secret": {h.oauthCfg.Google.ClientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to exchange code for token",
			"error", err.Error(),
			"event", "google_oauth_token_exchange_failed",
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange authorization code"})
		return
	}
	defer tokenResp.Body.Close()

	var tokenData struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		logger.Error(c.Request.Context(), "Failed to decode token response",
			"error", err.Error(),
			"event", "google_oauth_token_decode_failed",
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode token response"})
		return
	}

	if tokenData.Error != "" {
		logger.Error(c.Request.Context(), "Google OAuth error",
			"error", tokenData.Error,
			"event", "google_oauth_error",
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": tokenData.Error})
		return
	}

	userInfoResp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tokenData.AccessToken)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to get user info from Google",
			"error", err.Error(),
			"event", "google_oauth_userinfo_failed",
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}
	defer userInfoResp.Body.Close()

	body, err := io.ReadAll(userInfoResp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read user info"})
		return
	}

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		logger.Error(c.Request.Context(), "Failed to decode user info",
			"error", err.Error(),
			"event", "google_oauth_userinfo_decode_failed",
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode user info"})
		return
	}

	token, err := h.svc.LoginOrRegisterOAuthUser(c.Request.Context(), userInfo.Email, userInfo.Name, "google", userInfo.Picture)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to login/register OAuth user",
			"error", err.Error(),
			"event", "google_oauth_login_failed",
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process login"})
		return
	}

	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, url.QueryEscape(token))
	c.Redirect(http.StatusFound, redirectURL)
}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verify user email using the token sent to their email
// @Tags auth
// @Param token query string true "Verification token"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/verify [get]
func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing verification token"})
		return
	}

	if err := h.svc.VerifyEmail(c.Request.Context(), token); err != nil {
		logger.Warn(c.Request.Context(), "Email verification failed",
			"error", err.Error(),
			"event", "email_verification_failed",
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	frontendURL := h.frontendURL
	if frontendURL == "" {
		if referer := c.GetHeader("Referer"); referer != "" {
			if parsedURL, err := url.Parse(referer); err == nil {
				frontendURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
			}
		} else if origin := c.GetHeader("Origin"); origin != "" {
			frontendURL = origin
		} else {
			frontendURL = "http://localhost:3000"
		}
	}
	redirectURL := fmt.Sprintf("%s/auth/verified", frontendURL)
	c.Redirect(http.StatusFound, redirectURL)
}

// Me godoc
// @Summary Get current user info
// @Description Get the current authenticated user's information (protected route)
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserInfoResponse
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("user_email")
	role, _ := c.Get("user_role")

	c.JSON(http.StatusOK, UserInfoResponse{
		UserID:  userID,
		Email:   email,
		Role:    role,
		Message: "Token is valid!",
	})
}
