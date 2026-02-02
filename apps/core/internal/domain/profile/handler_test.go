package profile

import "testing"

func TestPublicProfileHandler(t *testing.T) {
	h := NewHandler(nil)
	if h == nil {
		t.Fatal("expected handler")
	}
}
