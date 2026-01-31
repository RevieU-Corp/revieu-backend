package handler

import "testing"

func TestPublicProfileHandler(t *testing.T) {
	h := NewProfileHandler(nil, nil, nil)
	if h == nil {
		t.Fatal("expected handler")
	}
}
