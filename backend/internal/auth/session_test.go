package auth

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestPasswordHashMatches(t *testing.T) {
	hash := "e2186dbdb1bb4193608605e84f33208765b5693b55edd4f730a719a100eeea6f"
	if !PasswordHashMatches(hash, hash) {
		t.Fatalf("expected matching hashes")
	}
	if !PasswordHashMatches(hash, "E2186DBDB1BB4193608605E84F33208765B5693B55EDD4F730A719A100EEEA6F") {
		t.Fatalf("expected case-insensitive match")
	}
	if PasswordHashMatches(hash, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatalf("expected mismatch")
	}
	if PasswordHashMatches(hash, "short") {
		t.Fatalf("expected short hash to fail")
	}
}

func TestIssueAndValidateToken(t *testing.T) {
	secret := "test-session-secret"
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	token, expires, err := IssueToken(secret, now, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if !expires.After(now) {
		t.Fatalf("expected expires after now")
	}
	if !ValidateToken(secret, token, now) {
		t.Fatalf("expected token valid at issue time")
	}
	if !ValidateToken(secret, token, now.Add(29*24*time.Hour)) {
		t.Fatalf("expected token valid before expiry")
	}
	if ValidateToken(secret, token, expires) {
		t.Fatalf("expected token invalid at exact expiry")
	}
	if ValidateToken("other-secret", token, now) {
		t.Fatalf("expected token invalid with wrong secret")
	}
	if ValidateToken(secret, "not-a-token", now) {
		t.Fatalf("expected malformed token invalid")
	}
}

func TestSessionCookieRoundTrip(t *testing.T) {
	rec := httptest.NewRecorder()
	expires := time.Now().UTC().Add(time.Hour)
	SetSessionCookie(rec, "payload.sig", expires, false)
	req := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	value, ok := SessionFromRequest(req)
	if !ok || value != "payload.sig" {
		t.Fatalf("expected session cookie value, got %q ok=%v", value, ok)
	}

	clearRec := httptest.NewRecorder()
	ClearSessionCookie(clearRec)
	cleared := false
	for _, cookie := range clearRec.Result().Cookies() {
		if cookie.Name == CookieName && cookie.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatalf("expected cleared session cookie")
	}
}

func TestValidatePasswordHashAndSecret(t *testing.T) {
	if err := ValidatePasswordHash("e2186dbdb1bb4193608605e84f33208765b5693b55edd4f730a719a100eeea6f"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ValidatePasswordHash("not-hex"); err == nil {
		t.Fatalf("expected error for invalid hash")
	}
	if err := ValidateSessionSecret("short"); err == nil {
		t.Fatalf("expected error for short secret")
	}
	if err := ValidateSessionSecret("long-enough-secret"); err != nil {
		t.Fatalf("unexpected secret error: %v", err)
	}
}
