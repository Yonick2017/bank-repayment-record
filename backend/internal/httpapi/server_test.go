package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"bank-repayment-record/backend/internal/httpapi"
	"bank-repayment-record/backend/internal/repayment"
	"bank-repayment-record/backend/internal/storage"
)

func TestCreateRepaymentCompatibilityFields(t *testing.T) {
	handler, store, cleanup := newTestHandler(t)
	defer cleanup()

	body := `{"card":"BOCHK Visa","currency":"RMB","amount":"100.00","repaymentTime":"2026-06-01T10:30"}`
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/repayments", bytes.NewBufferString(body))
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201 using alias fields, got %d: %s", res.Code, res.Body.String())
	}

	var createResp struct {
		Data struct {
			CardName    string `json:"cardName"`
			RepaymentAt string `json:"repaymentAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	if createResp.Data.CardName != "BOCHK Visa" {
		t.Fatalf("expected canonical cardName in response, got %q", createResp.Data.CardName)
	}
	if createResp.Data.RepaymentAt == "" {
		t.Fatalf("expected canonical repaymentAt in response")
	}

	records, err := store.ListRepayments(req.Context(), repayment.Filters{})
	if err != nil {
		t.Fatalf("list repayments: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].AmountCents != 10000 {
		t.Fatalf("expected 10000 cents, got %d", records[0].AmountCents)
	}
}

func TestCreateRepaymentValidationAndMinuteFormats(t *testing.T) {
	handler, _, cleanup := newTestHandler(t)
	defer cleanup()

	tests := []struct {
		name       string
		body       string
		statusCode int
	}{
		{
			name:       "invalid_card",
			body:       `{"cardName":"Unknown","currency":"RMB","amount":"100.00","repaymentAt":"2026-06-01T10:30"}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid_decimal_precision",
			body:       `{"cardName":"BOCHK Visa","currency":"RMB","amount":"12.345","repaymentAt":"2026-06-01T10:30"}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "minute_format",
			body:       `{"cardName":"BOCHK Visa","currency":"RMB","amount":"12.34","repaymentAt":"2026-06-01T10:30"}`,
			statusCode: http.StatusCreated,
		},
		{
			name:       "seconds_format_with_zero_seconds",
			body:       `{"cardName":"BOCHK Visa","currency":"RMB","amount":"12.34","repaymentAt":"2026-06-01T10:30:00"}`,
			statusCode: http.StatusCreated,
		},
		{
			name:       "rfc3339_with_zero_seconds",
			body:       `{"cardName":"BOCHK Visa","currency":"RMB","amount":"12.34","repaymentAt":"2026-06-01T10:30:00+08:00"}`,
			statusCode: http.StatusCreated,
		},
		{
			name:       "non_minute_second_precision",
			body:       `{"cardName":"BOCHK Visa","currency":"RMB","amount":"12.34","repaymentAt":"2026-06-01T10:30:45+08:00"}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "non_minute_fractional_precision",
			body:       `{"cardName":"BOCHK Visa","currency":"RMB","amount":"12.34","repaymentAt":"2026-06-01T10:30:00.123+08:00"}`,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/repayments", bytes.NewBufferString(tt.body))
			handler.ServeHTTP(res, req)
			if res.Code != tt.statusCode {
				t.Fatalf("expected %d, got %d: %s", tt.statusCode, res.Code, res.Body.String())
			}
		})
	}
}

func TestHistoryFiltersDeleteAndAlias(t *testing.T) {
	handler, store, cleanup := newTestHandler(t)
	defer cleanup()

	first := mustCreateRecord(t, store, repayment.Record{
		CardName:    "BOCHK Visa",
		Currency:    "RMB",
		AmountCents: 10000,
		RepaymentAt: mustParseTime(t, "2026-05-10T10:15:00+08:00"),
	})
	mustCreateRecord(t, store, repayment.Record{
		CardName:    "HSBC Pulse",
		Currency:    "HKD",
		AmountCents: 20000,
		RepaymentAt: mustParseTime(t, "2026-06-10T10:15:00+08:00"),
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/repayments/history?card=BOCHK+Visa", nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	var historyResp struct {
		Months []struct {
			Month   string `json:"month"`
			Records []struct {
				ID int64 `json:"id"`
			} `json:"records"`
		} `json:"months"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &historyResp); err != nil {
		t.Fatalf("unmarshal history: %v", err)
	}
	if len(historyResp.Months) != 1 {
		t.Fatalf("expected 1 month, got %d", len(historyResp.Months))
	}
	if len(historyResp.Months[0].Records) != 1 || historyResp.Months[0].Records[0].ID != first.ID {
		t.Fatalf("history result did not match filtered card")
	}

	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/repayments/"+itoa(first.ID), nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204 delete, got %d", res.Code)
	}

	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/repayments/not-an-id", nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid delete id, got %d", res.Code)
	}

	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/repayments/999999", nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for not found delete id, got %d", res.Code)
	}
}

func TestMonthlyStatsFormulaCurrencySplitAndAliasRoute(t *testing.T) {
	handler, store, cleanup := newTestHandler(t)
	defer cleanup()

	mustCreateRecord(t, store, repayment.Record{
		CardName:    "BOCHK Visa",
		Currency:    "RMB",
		AmountCents: 10000,
		RepaymentAt: mustParseTime(t, "2026-01-05T10:00:00+08:00"),
	})
	mustCreateRecord(t, store, repayment.Record{
		CardName:    "BOCHK Visa",
		Currency:    "RMB",
		AmountCents: 30000,
		RepaymentAt: mustParseTime(t, "2026-02-05T10:00:00+08:00"),
	})
	mustCreateRecord(t, store, repayment.Record{
		CardName:    "HSBC Pulse",
		Currency:    "HKD",
		AmountCents: 20000,
		RepaymentAt: mustParseTime(t, "2026-02-06T10:00:00+08:00"),
	})

	for _, route := range []string{"/api/stats/monthly", "/api/repayments/stats"} {
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, route, nil)
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200 on %s, got %d", route, res.Code)
		}

		var response struct {
			Currencies []struct {
				Currency                string `json:"currency"`
				AverageMonthlyRepayment string `json:"averageMonthlyRepayment"`
			} `json:"currencies"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
			t.Fatalf("unmarshal monthly stats on %s: %v", route, err)
		}
		avgByCurrency := map[string]string{}
		for _, item := range response.Currencies {
			avgByCurrency[item.Currency] = item.AverageMonthlyRepayment
		}

		if avgByCurrency["RMB"] != "200.00" {
			t.Fatalf("expected RMB average 200.00 on %s, got %q", route, avgByCurrency["RMB"])
		}
		if avgByCurrency["HKD"] != "200.00" {
			t.Fatalf("expected HKD average 200.00 on %s, got %q", route, avgByCurrency["HKD"])
		}
	}
}

func TestCurrentMonthStats(t *testing.T) {
	handler, store, cleanup := newTestHandler(t)
	defer cleanup()

	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc).Truncate(time.Minute)
	currentMonthTime := time.Date(now.Year(), now.Month(), 5, 10, 0, 0, 0, loc)
	previousMonthTime := currentMonthTime.AddDate(0, -1, 0)

	mustCreateRecord(t, store, repayment.Record{
		CardName:    "BOCHK Visa",
		Currency:    "RMB",
		AmountCents: 5000,
		RepaymentAt: currentMonthTime,
	})
	mustCreateRecord(t, store, repayment.Record{
		CardName:    "HSBC Pulse",
		Currency:    "HKD",
		AmountCents: -2000,
		RepaymentAt: currentMonthTime,
	})
	mustCreateRecord(t, store, repayment.Record{
		CardName:    "BOCHK Visa",
		Currency:    "RMB",
		AmountCents: 9900,
		RepaymentAt: previousMonthTime,
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/stats/current-month", nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	var response struct {
		Totals map[string]string `json:"totals"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal current month stats: %v", err)
	}
	if response.Totals["RMB"] != "50.00" {
		t.Fatalf("expected RMB 50.00, got %q", response.Totals["RMB"])
	}
	if response.Totals["HKD"] != "-20.00" {
		t.Fatalf("expected HKD -20.00, got %q", response.Totals["HKD"])
	}
}

func TestCORSPreflightAndAllowList(t *testing.T) {
	handler, _, cleanup := newTestHandlerWithOrigins(t, []string{"http://localhost:5173"})
	defer cleanup()

	preflightRes := httptest.NewRecorder()
	preflightReq := httptest.NewRequest(http.MethodOptions, "/api/repayments", nil)
	preflightReq.Header.Set("Origin", "http://localhost:5173")
	preflightReq.Header.Set("Access-Control-Request-Method", "POST")
	handler.ServeHTTP(preflightRes, preflightReq)

	if preflightRes.Code != http.StatusNoContent {
		t.Fatalf("expected 204 preflight, got %d", preflightRes.Code)
	}
	if preflightRes.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("expected allow origin for configured host")
	}

	blockedRes := httptest.NewRecorder()
	blockedReq := httptest.NewRequest(http.MethodOptions, "/api/repayments", nil)
	blockedReq.Header.Set("Origin", "https://not-allowed.example")
	blockedReq.Header.Set("Access-Control-Request-Method", "POST")
	handler.ServeHTTP(blockedRes, blockedReq)

	if blockedRes.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected no allow origin header for blocked host")
	}
}

func newTestHandler(t *testing.T) (http.Handler, *storage.SQLiteStore, func()) {
	return newTestHandlerWithOrigins(t, nil)
}

func newTestHandlerWithOrigins(t *testing.T, origins []string) (http.Handler, *storage.SQLiteStore, func()) {
	t.Helper()

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.OpenSQLite(dbPath, loc)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	server := httpapi.NewServer(store, loc, origins...)
	cleanup := func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close sqlite: %v", err)
		}
	}
	return server.Handler(), store, cleanup
}

func mustCreateRecord(t *testing.T, store *storage.SQLiteStore, record repayment.Record) repayment.Record {
	t.Helper()
	created, err := store.CreateRepayment(contextBackground(), record)
	if err != nil {
		t.Fatalf("create record: %v", err)
	}
	return created
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}

func contextBackground() context.Context {
	return context.Background()
}
