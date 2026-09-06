package identity

import (
	"errors"
	"testing"
)

func TestStaticRejectsWrongPassword(t *testing.T) {
	auth := Static{Username: "demo", Password: "secret", User: User{ID: "u1", Name: "Demo"}}
	if _, err := auth.Authenticate(t.Context(), "demo", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v", err)
	}
	user, err := auth.Authenticate(t.Context(), "demo", "secret")
	if err != nil || user.ID != "u1" || user.Name != "Demo" {
		t.Fatalf("user = %+v, err = %v", user, err)
	}
}
