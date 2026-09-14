package service

import (
	"testing"
	"time"
)

func TestLoginRateLimitAfterFiveFailures(t *testing.T) {
	service := &AuthService{loginFails: make(map[string]loginFailure)}
	now := time.Now()

	for attempt := 1; attempt <= maxLoginFailures; attempt++ {
		if retryAfter, blocked := service.recordLoginFailure("user|127.0.0.1", now); blocked {
			t.Fatalf("percobaan ke-%d tidak seharusnya diblokir, retry setelah %s", attempt, retryAfter)
		}
	}

	retryAfter, blocked := service.recordLoginFailure("user|127.0.0.1", now)
	if !blocked {
		t.Fatal("percobaan keenam seharusnya diblokir")
	}
	if retryAfter <= 0 {
		t.Fatal("Retry-After harus bernilai positif")
	}
}

func TestResetLoginFailures(t *testing.T) {
	service := &AuthService{loginFails: make(map[string]loginFailure)}
	service.recordLoginFailure("user|127.0.0.1", time.Now())
	service.resetLoginFailures("user|127.0.0.1")

	if _, blocked := service.loginBlocked("user|127.0.0.1", time.Now()); blocked {
		t.Fatal("login sukses seharusnya mereset rate limit")
	}
}
