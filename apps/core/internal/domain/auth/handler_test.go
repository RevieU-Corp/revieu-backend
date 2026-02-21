package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/gin-gonic/gin"
)

type stubAuthService struct {
	refreshFn func(context.Context, string) (LoginTokens, error)
}

func (s stubAuthService) Register(context.Context, string, string, string, string) (*model.User, error) {
	return nil, errors.New("not implemented")
}

func (s stubAuthService) Login(context.Context, string, string, string) (LoginTokens, error) {
	return LoginTokens{}, errors.New("not implemented")
}

func (s stubAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (LoginTokens, error) {
	return s.refreshFn(ctx, refreshToken)
}

func (s stubAuthService) LoginOrRegisterOAuthUser(context.Context, string, string, string, string) (string, error) {
	return "", errors.New("not implemented")
}

func (s stubAuthService) VerifyEmail(context.Context, string) error {
	return errors.New("not implemented")
}

func TestRefreshHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{
		svc: stubAuthService{
			refreshFn: func(_ context.Context, refreshToken string) (LoginTokens, error) {
				if refreshToken != "valid-refresh" {
					return LoginTokens{}, errors.New("invalid refresh token")
				}
				return LoginTokens{
					AccessToken:  "new-access",
					RefreshToken: "new-refresh",
				}, nil
			},
		},
	}

	r := gin.New()
	r.POST("/auth/refresh", h.Refresh)

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "valid-refresh"})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp RefreshResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.AccessToken != "new-access" {
		t.Fatalf("expected access token new-access, got %q", resp.AccessToken)
	}
	if resp.RefreshToken != "new-refresh" {
		t.Fatalf("expected refresh token new-refresh, got %q", resp.RefreshToken)
	}
	if resp.Type != "Bearer" {
		t.Fatalf("expected type Bearer, got %q", resp.Type)
	}
}

func TestRefreshHandlerInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{
		svc: stubAuthService{
			refreshFn: func(_ context.Context, refreshToken string) (LoginTokens, error) {
				return LoginTokens{}, errors.New("invalid refresh token")
			},
		},
	}

	r := gin.New()
	r.POST("/auth/refresh", h.Refresh)

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "bad-refresh"})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
