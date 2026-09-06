package identity

import (
	"context"
	"crypto/subtle"
	"errors"
)

const UserIDHeader = "X-User-ID"

var ErrInvalidCredentials = errors.New("invalid credentials")

type User struct {
	ID   string
	Name string
}

type Authenticator interface {
	Authenticate(ctx context.Context, username, password string) (User, error)
}

// Static is a single demo user for the template.
// TODO(template): replace with an identity provider or user store.
type Static struct {
	Username, Password string
	User               User
}

func (s Static) Authenticate(_ context.Context, username, password string) (User, error) {
	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(s.Username)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(s.Password)) == 1
	if len(username) != len(s.Username) {
		userOK = false
	}
	if len(password) != len(s.Password) {
		passOK = false
	}
	if !userOK || !passOK {
		return User{}, ErrInvalidCredentials
	}
	return s.User, nil
}
