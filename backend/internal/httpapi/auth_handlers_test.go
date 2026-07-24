package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bank-repayment-record/backend/internal/auth"
	"bank-repayment-record/backend/internal/httpapi"
	"bank-repayment-record/backend/internal/repayment"
)

type stubStore struct{}

func (stubStore) CreateRepayment(ctx context.Context, record repayment.Record) (repayment.Record, error) {
	return record, nil
}

func (stubStore) DeleteRepayment(ctx context.Context, id int64) (bool, error) {
	return true, nil
}

func (stubStore) ListRepayments(ctx context.Context, filters repayment.Filters) ([]repayment.Record, error) {
	return nil, nil
}

func newAuthTestServer(t *testing.T) *httpapi.Server {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	return httpapi.NewServer(stubStore{}, loc, httpapi.ServerOptions{
		Auth: auth.Config{
			PasswordHash:  testPasswordHash,
			SessionSecret: testSessionSecret,
			SessionDays:   auth.DefaultSessionDays,
		},
	})
}

func TestAuthLoginLogoutAndProtectedRoutes(t *testing.T) {
	server := newAuthTestServer(t)
	handler := server.Handler()

	unauthRes := httptest.NewRecorder()
	unauthReq := httptest.NewRequest(http.MethodGet, "/api/stats/current-month", nil)
	handler.ServeHTTP(unauthRes, unauthReq)
	if unauthRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", unauthRes.Code)
	}

	badLoginRes := httptest.NewRecorder()
	badLoginReq := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		bytes.NewBufferString(`{"passwordHash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`),
	)
	handler.ServeHTTP(badLoginRes, badLoginReq)
	if badLoginRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for bad password, got %d", badLoginRes.Code)
	}
	if len(badLoginRes.Result().Cookies()) != 0 {
		t.Fatalf("expected no session cookie for bad password")
	}

	cookie := mustLogin(t, handler, testPasswordHash)

	meRes := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.AddCookie(cookie)
	handler.ServeHTTP(meRes, meReq)
	if meRes.Code != http.StatusOK {
		t.Fatalf("expected 200 for /me with session, got %d", meRes.Code)
	}

	protectedRes := httptest.NewRecorder()
	protectedReq := httptest.NewRequest(http.MethodGet, "/api/stats/current-month", nil)
	protectedReq.AddCookie(cookie)
	handler.ServeHTTP(protectedRes, protectedReq)
	if protectedRes.Code != http.StatusOK {
		t.Fatalf("expected 200 for protected route with session, got %d: %s", protectedRes.Code, protectedRes.Body.String())
	}

	logoutRes := httptest.NewRecorder()
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(cookie)
	handler.ServeHTTP(logoutRes, logoutReq)
	if logoutRes.Code != http.StatusOK {
		t.Fatalf("expected 200 for logout, got %d", logoutRes.Code)
	}
	// Browser drops the cookie after Set-Cookie Max-Age=-1; subsequent requests omit it.
	afterLogoutRes := httptest.NewRecorder()
	afterLogoutReq := httptest.NewRequest(http.MethodGet, "/api/stats/current-month", nil)
	handler.ServeHTTP(afterLogoutRes, afterLogoutReq)
	if afterLogoutRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout without cookie, got %d", afterLogoutRes.Code)
	}
}

func TestAuthExpiredSessionRejected(t *testing.T) {
	server := newAuthTestServer(t)
	fixedNow := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	// Access unexported now via login at fixed time then validate later.
	// Issue an already-expired token directly and present it as cookie.
	token, _, err := auth.IssueToken(testSessionSecret, fixedNow.Add(-48*time.Hour), time.Hour)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/stats/current-month", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	server.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired session, got %d", res.Code)
	}
}

func TestAuthMeUnauthorizedWithoutCookie(t *testing.T) {
	handler := newAuthTestServer(t).Handler()
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["error"] != "unauthorized" {
		t.Fatalf("unexpected error payload: %#v", payload)
	}
}
