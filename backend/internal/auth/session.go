package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	CookieName          = "brr_session"
	DefaultSessionDays  = 30
	MinSessionSecretLen = 16
	passwordHashLen     = 64
)

// Config holds shared-password gate settings loaded from YAML.
type Config struct {
	PasswordHash  string
	SessionSecret string
	SessionDays   int
}

func (c Config) SessionTTL() time.Duration {
	days := c.SessionDays
	if days <= 0 {
		days = DefaultSessionDays
	}
	return time.Duration(days) * 24 * time.Hour
}

// ValidatePasswordHash reports whether value is a 64-char hex SHA-256 digest.
func ValidatePasswordHash(value string) error {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) != passwordHashLen {
		return fmt.Errorf("auth.password_hash must be a %d-character hex SHA-256 digest", passwordHashLen)
	}
	if _, err := hex.DecodeString(trimmed); err != nil {
		return fmt.Errorf("auth.password_hash must be hex: %w", err)
	}
	return nil
}

// ValidateSessionSecret reports whether the signing secret meets the minimum length.
func ValidateSessionSecret(value string) error {
	if len(strings.TrimSpace(value)) < MinSessionSecretLen {
		return fmt.Errorf("auth.session_secret must be at least %d characters", MinSessionSecretLen)
	}
	return nil
}

// PasswordHashMatches compares expected and provided hex digests in constant time.
func PasswordHashMatches(expected, provided string) bool {
	expectedNorm := strings.ToLower(strings.TrimSpace(expected))
	providedNorm := strings.ToLower(strings.TrimSpace(provided))
	if len(expectedNorm) != passwordHashLen || len(providedNorm) != passwordHashLen {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expectedNorm), []byte(providedNorm)) == 1
}

// IssueToken creates a signed session token that expires after ttl.
func IssueToken(secret string, now time.Time, ttl time.Duration) (token string, expires time.Time, err error) {
	if ttl <= 0 {
		return "", time.Time{}, fmt.Errorf("session ttl must be positive")
	}
	expires = now.UTC().Add(ttl)
	expUnix := expires.Unix()
	payload := strconv.FormatInt(expUnix, 10)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, expires, nil
}

// ValidateToken reports whether token is a valid, unexpired session for secret.
func ValidateToken(secret, token string, now time.Time) bool {
	payload, sig, ok := strings.Cut(strings.TrimSpace(token), ".")
	if !ok || payload == "" || sig == "" {
		return false
	}
	expUnix, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return false
	}
	if !now.UTC().Before(time.Unix(expUnix, 0).UTC()) {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(strings.ToLower(sig)), []byte(expected)) == 1
}

// SetSessionCookie writes the HttpOnly session cookie.
func SetSessionCookie(w http.ResponseWriter, token string, expires time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

// ClearSessionCookie removes the session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// SessionFromRequest returns the session cookie value when present.
func SessionFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie == nil || strings.TrimSpace(cookie.Value) == "" {
		return "", false
	}
	return cookie.Value, true
}
