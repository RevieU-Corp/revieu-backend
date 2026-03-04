package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMerchantHandlerDeleteMePlaceholder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/merchant/me", nil)

	h := NewMerchantHandler(nil)
	h.DeleteMe(c)

	if recorder.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d", http.StatusNotImplemented, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "not implemented") {
		t.Fatalf("expected response body to mention not implemented, got %s", recorder.Body.String())
	}
}
