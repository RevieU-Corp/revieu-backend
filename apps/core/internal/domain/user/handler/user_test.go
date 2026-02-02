package handler

import "testing"

func TestUserHandlerConstruction(t *testing.T) {
	h := NewUserHandler(nil)
	if h == nil {
		t.Fatal("expected handler")
	}
}
