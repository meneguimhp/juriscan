package auth

import (
	"errors"
	"testing"
	"time"
)

func TestRequestAndVerifyOTP(t *testing.T) {
	svc := NewService(2*time.Minute, time.Hour, func(email string) (User, error) {
		if email == "admin@juriscan.local" {
			return User{ID: "user-admin", Email: email, Role: "admin"}, nil
		}
		return User{ID: "user-commercial", Email: email, Role: "commercial"}, nil
	})

	code, expires, err := svc.RequestOTP("admin@juriscan.local")
	if err != nil {
		t.Fatalf("RequestOTP failed: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6-digit code, got %q", code)
	}
	if expires <= 0 {
		t.Fatalf("expected positive expires, got %d", expires)
	}

	token, user, err := svc.VerifyOTP("admin@juriscan.local", code)
	if err != nil {
		t.Fatalf("VerifyOTP failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected session token")
	}
	if user.Role != "admin" {
		t.Fatalf("expected admin role, got %q", user.Role)
	}
}

func TestVerifyOTPInvalidCode(t *testing.T) {
	svc := NewService(time.Minute, time.Hour, nil)
	if _, _, err := svc.RequestOTP("user@juriscan.local"); err != nil {
		t.Fatalf("RequestOTP failed: %v", err)
	}
	if _, _, err := svc.VerifyOTP("user@juriscan.local", "000000"); err == nil {
		t.Fatal("expected invalid credentials error")
	}
}

func TestValidateSession(t *testing.T) {
	svc := NewService(time.Minute, time.Hour, nil)
	code, _, err := svc.RequestOTP("user@juriscan.local")
	if err != nil {
		t.Fatalf("RequestOTP failed: %v", err)
	}
	token, _, err := svc.VerifyOTP("user@juriscan.local", code)
	if err != nil {
		t.Fatalf("VerifyOTP failed: %v", err)
	}

	if _, err := svc.ValidateSession(token); err != nil {
		t.Fatalf("ValidateSession failed: %v", err)
	}

	svc.RevokeSession(token)
	if _, err := svc.ValidateSession(token); err == nil {
		t.Fatal("expected revoked session to be invalid")
	}
}

func TestRequestOTPBlockedWhenResolverRejectsUser(t *testing.T) {
	svc := NewService(time.Minute, time.Hour, func(string) (User, error) {
		return User{}, errors.New("user not allowed")
	})
	if _, _, err := svc.RequestOTP("blocked@juriscan.local"); err == nil {
		t.Fatal("expected resolver to block otp request")
	}
}
