package handler

import "testing"

func TestUserHandlerConstruction(t *testing.T) {
	h := NewUserHandler(nil, nil, nil, nil)
	if h == nil {
		t.Fatal("expected handler")
	}
}
