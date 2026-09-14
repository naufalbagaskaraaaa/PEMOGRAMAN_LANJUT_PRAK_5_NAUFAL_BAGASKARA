package helper

import (
	"errors"
	"testing"
	"time"

	"latihan-fiber/app/model"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTExpiredTokenIsDistinguishable(t *testing.T) {
	manager := NewJWTManager("test-secret", -time.Minute)
	token, err := manager.Issue(model.AuthUser{UserID: 1, Username: "budi", Role: "user"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	_, err = manager.Parse(token)
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("error = %v, want ErrTokenExpired", err)
	}
}
