package authorization

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/model"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/testutil"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/token"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/pkg/database"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		userID   int64
		role     string
		expected int
	}{
		{name: "admin", userID: 1, role: "admin", expected: http.StatusNoContent},
		{name: "regular-user", userID: 1, role: "user", expected: http.StatusForbidden},
		{name: "missing-principal", role: "admin", expected: http.StatusUnauthorized},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if testCase.userID > 0 {
					c.Set(UserIDKey, testCase.userID)
				}
				c.Set(UserRoleKey, testCase.role)
				c.Next()
			}, RequireRole("admin"))
			router.GET("/admin", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/admin", nil)
			router.ServeHTTP(recorder, request)
			if recorder.Code != testCase.expected {
				t.Fatalf("expected status %d, got %d: %s", testCase.expected, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestJWTAuthRejectsDisabledUserAndAcceptsHttpOnlyCookie(t *testing.T) {
	db := testutil.SetupTestDB(t)
	database.DB = db

	user := model.User{Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	auth := model.UserAuth{UserID: user.ID, IdentityType: "email", Identifier: "middleware@example.com"}
	if err := db.Create(&auth).Error; err != nil {
		t.Fatalf("create auth: %v", err)
	}
	cfg := config.JWTConfig{Secret: "middleware-secret", ExpireHour: 1}
	accessToken, err := token.New(cfg).GenerateToken(&user, &auth)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", JWTAuth(cfg), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	activeReq := httptest.NewRequest(http.MethodGet, "/protected", nil)
	activeReq.AddCookie(&http.Cookie{Name: AccessTokenCookieName, Value: accessToken})
	activeRec := httptest.NewRecorder()
	r.ServeHTTP(activeRec, activeReq)
	if activeRec.Code != http.StatusNoContent {
		t.Fatalf("active cookie request = %d, want 204", activeRec.Code)
	}

	if err := db.Model(&model.User{}).Where("id = ?", user.ID).Update("status", 1).Error; err != nil {
		t.Fatalf("disable user: %v", err)
	}
	disabledReq := httptest.NewRequest(http.MethodGet, "/protected", nil)
	disabledReq.AddCookie(&http.Cookie{Name: AccessTokenCookieName, Value: accessToken})
	disabledRec := httptest.NewRecorder()
	r.ServeHTTP(disabledRec, disabledReq)
	if disabledRec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled cookie request = %d, want 401", disabledRec.Code)
	}
}
