package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bank-repayment-record/backend/internal/repayment"
)

var (
	allowedCards = map[string]struct{}{
		"BOCHK Visa":           {},
		"BOCHK Mastercard":     {},
		"HSBC Visa Gold":       {},
		"HSBC Pulse":           {},
		"Hang Seng Travel+":    {},
		"HSBC Visa Signature":  {},
		"Amex US":              {},
		"BEA GOAL":             {},
		"CITIC Motion":         {},
		"Earnmore":             {},
		"SC Smart":             {},
		"ICBC SUP":             {},
		"ICBC 奋斗":            {},
	}
	allowedCurrencies = map[string]struct{}{
		"RMB": {},
		"HKD": {},
	}
)

type Store interface {
	CreateRepayment(ctx context.Context, record repayment.Record) (repayment.Record, error)
	DeleteRepayment(ctx context.Context, id int64) (bool, error)
	ListRepayments(ctx context.Context, filters repayment.Filters) ([]repayment.Record, error)
}

type Server struct {
	store           Store
	loc             *time.Location
	allowedOrigins  map[string]struct{}
	frontendDistDir string
}

func NewServer(store Store, loc *time.Location, frontendDistDir string, corsAllowedOrigins ...string) *Server {
	return &Server{
		store:           store,
		loc:             loc,
		allowedOrigins:  buildAllowedOrigins(corsAllowedOrigins),
		frontendDistDir: strings.TrimSpace(frontendDistDir),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/repayments", s.handleRepayments)
	mux.HandleFunc("/api/repayments/history", s.handleRepaymentHistory)
	mux.HandleFunc("/api/repayments/", s.handleRepaymentByID)
	mux.HandleFunc("/api/repayments/stats", s.handleMonthlyStats)
	mux.HandleFunc("/api/stats/monthly", s.handleMonthlyStats)
	mux.HandleFunc("/api/stats/current-month", s.handleCurrentMonthStats)
	mux.HandleFunc("/", s.handleFrontend)

	return corsMiddleware(mux, s.allowedOrigins)
}

func (s *Server) handleFrontend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}
	if s.frontendDistDir == "" {
		http.NotFound(w, r)
		return
	}

	cleanPath := path.Clean("/" + r.URL.Path)
	if cleanPath != "/" {
		targetFile := filepath.Join(s.frontendDistDir, filepath.FromSlash(strings.TrimPrefix(cleanPath, "/")))
		if isRegularFile(targetFile) {
			http.ServeFile(w, r, targetFile)
			return
		}
	}

	indexFile := filepath.Join(s.frontendDistDir, "index.html")
	if isRegularFile(indexFile) {
		http.ServeFile(w, r, indexFile)
		return
	}

	http.NotFound(w, r)
}

func isRegularFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

type createRepaymentRequest struct {
	CardName      string      `json:"cardName"`
	Card          string      `json:"card"`
	Currency      string      `json:"currency"`
	Amount        interface{} `json:"amount"`
	RepaymentAt   string      `json:"repaymentAt"`
	RepaymentTime string      `json:"repaymentTime"`
}

type repaymentResponse struct {
	ID          int64  `json:"id"`
	CardName    string `json:"cardName"`
	Currency    string `json:"currency"`
	Amount      string `json:"amount"`
	RepaymentAt string `json:"repaymentAt"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type historyMonth struct {
	Month   string              `json:"month"`
	Records []repaymentResponse `json:"records"`
}

type monthlyCurrencyStat struct {
	Currency                string        `json:"currency"`
	MonthlyTotals           []monthAmount `json:"monthlyTotals"`
	AverageMonthlyRepayment string        `json:"averageMonthlyRepayment"`
}

type monthAmount struct {
	Month string `json:"month"`
	Total string `json:"total"`
}

func (s *Server) handleRepayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req createRepaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	cardName := firstNonEmpty(req.CardName, req.Card)
	cardName = strings.TrimSpace(cardName)
	if _, ok := allowedCards[cardName]; !ok {
		writeError(w, http.StatusBadRequest, "cardName is not supported")
		return
	}

	currency := strings.TrimSpace(req.Currency)
	if _, ok := allowedCurrencies[currency]; !ok {
		writeError(w, http.StatusBadRequest, "currency must be RMB or HKD")
		return
	}

	amountCents, err := parseAmountToCents(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	repaymentAtValue := firstNonEmpty(req.RepaymentAt, req.RepaymentTime)
	repaymentAt, err := parseMinuteLevelTime(repaymentAtValue, s.loc)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	record, err := s.store.CreateRepayment(r.Context(), repayment.Record{
		CardName:    cardName,
		Currency:    currency,
		AmountCents: amountCents,
		RepaymentAt: repaymentAt,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create repayment")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]repaymentResponse{
		"data": toRepaymentResponse(record),
	})
}

func (s *Server) handleRepaymentHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filters, err := parseFilters(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	records, err := s.store.ListRepayments(r.Context(), filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list history")
		return
	}

	grouped := map[string][]repaymentResponse{}
	for _, record := range records {
		month := record.RepaymentAt.In(s.loc).Format("2006-01")
		grouped[month] = append(grouped[month], toRepaymentResponse(record))
	}

	months := make([]string, 0, len(grouped))
	for month := range grouped {
		months = append(months, month)
	}
	sortDesc(months)

	result := make([]historyMonth, 0, len(months))
	for _, month := range months {
		result = append(result, historyMonth{
			Month:   month,
			Records: grouped[month],
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"filters": map[string]string{
			"cardName": filters.CardName,
			"currency": filters.Currency,
		},
		"months": result,
	})
}

func (s *Server) handleRepaymentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idRaw := strings.TrimPrefix(r.URL.Path, "/api/repayments/")
	idRaw = strings.TrimSpace(idRaw)
	if idRaw == "" {
		writeError(w, http.StatusBadRequest, "repayment id is required")
		return
	}
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "repayment id must be a positive integer")
		return
	}

	ok, err := s.store.DeleteRepayment(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete repayment")
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "repayment not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMonthlyStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filters, err := parseFilters(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	records, err := s.store.ListRepayments(r.Context(), filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load stats")
		return
	}

	response := map[string]interface{}{
		"filters": map[string]string{
			"cardName": filters.CardName,
			"currency": filters.Currency,
		},
		"currencies": buildMonthlyCurrencyStats(records, s.loc),
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleCurrentMonthStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	now := time.Now().In(s.loc)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, s.loc)
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	records, err := s.store.ListRepayments(r.Context(), repayment.Filters{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load current month stats")
		return
	}

	totals := map[string]int64{"RMB": 0, "HKD": 0}
	for _, record := range records {
		repaymentAt := record.RepaymentAt.In(s.loc)
		if repaymentAt.Before(monthStart) || !repaymentAt.Before(nextMonthStart) {
			continue
		}
		totals[record.Currency] += record.AmountCents
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"month": monthStart.Format("2006-01"),
		"totals": map[string]string{
			"RMB": centsToDecimalString(totals["RMB"]),
			"HKD": centsToDecimalString(totals["HKD"]),
		},
	})
}

func parseFilters(r *http.Request) (repayment.Filters, error) {
	query := r.URL.Query()
	cardName := strings.TrimSpace(query.Get("cardName"))
	if cardName == "" {
		cardName = strings.TrimSpace(query.Get("card"))
	}
	filters := repayment.Filters{
		CardName: cardName,
		Currency: strings.TrimSpace(query.Get("currency")),
	}

	if filters.CardName != "" {
		if _, ok := allowedCards[filters.CardName]; !ok {
			return repayment.Filters{}, errors.New("cardName filter is not supported")
		}
	}
	if filters.Currency != "" {
		if _, ok := allowedCurrencies[filters.Currency]; !ok {
			return repayment.Filters{}, errors.New("currency filter must be RMB or HKD")
		}
	}

	return filters, nil
}

func parseAmountToCents(raw interface{}) (int64, error) {
	var amountStr string
	switch value := raw.(type) {
	case string:
		amountStr = strings.TrimSpace(value)
	case float64:
		amountStr = strconv.FormatFloat(value, 'f', -1, 64)
	default:
		return 0, errors.New("amount must be a decimal string or number")
	}

	if amountStr == "" {
		return 0, errors.New("amount is required")
	}

	sign := int64(1)
	if strings.HasPrefix(amountStr, "-") {
		sign = -1
		amountStr = strings.TrimPrefix(amountStr, "-")
	}

	parts := strings.Split(amountStr, ".")
	if len(parts) > 2 {
		return 0, errors.New("amount must be a decimal with at most 2 places")
	}
	if parts[0] == "" {
		return 0, errors.New("amount must include integer part")
	}

	intPart, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("amount must be numeric")
	}

	fracPart := int64(0)
	if len(parts) == 2 {
		if len(parts[1]) == 0 || len(parts[1]) > 2 {
			return 0, errors.New("amount must be a decimal with at most 2 places")
		}
		if len(parts[1]) == 1 {
			parts[1] += "0"
		}
		fracPart, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, errors.New("amount must be numeric")
		}
	}

	return sign * ((intPart * 100) + fracPart), nil
}

func parseMinuteLevelTime(raw string, loc *time.Location) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, errors.New("repaymentAt is required")
	}

	if parsed, err := time.ParseInLocation("2006-01-02T15:04", value, loc); err == nil {
		return parsed, nil
	}

	if parsed, err := time.ParseInLocation("2006-01-02T15:04:05", value, loc); err == nil {
		if parsed.Second() != 0 {
			return time.Time{}, errors.New("repaymentAt must be minute-level precision")
		}
		return parsed, nil
	}

	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		parsed = parsed.In(loc)
		if parsed.Second() != 0 || parsed.Nanosecond() != 0 {
			return time.Time{}, errors.New("repaymentAt must be minute-level precision")
		}
		return parsed, nil
	}

	return time.Time{}, errors.New("repaymentAt must be an ISO datetime")
}

func toRepaymentResponse(record repayment.Record) repaymentResponse {
	return repaymentResponse{
		ID:          record.ID,
		CardName:    record.CardName,
		Currency:    record.Currency,
		Amount:      centsToDecimalString(record.AmountCents),
		RepaymentAt: record.RepaymentAt.Format(time.RFC3339),
		CreatedAt:   record.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   record.UpdatedAt.Format(time.RFC3339),
	}
}

func buildMonthlyCurrencyStats(records []repayment.Record, loc *time.Location) []monthlyCurrencyStat {
	monthTotalsByCurrency := map[string]map[string]int64{
		"RMB": {},
		"HKD": {},
	}
	for _, record := range records {
		month := record.RepaymentAt.In(loc).Format("2006-01")
		monthTotalsByCurrency[record.Currency][month] += record.AmountCents
	}

	stats := make([]monthlyCurrencyStat, 0, 2)
	for _, currency := range []string{"RMB", "HKD"} {
		monthMap := monthTotalsByCurrency[currency]
		months := make([]string, 0, len(monthMap))
		var totalAcrossMonths int64
		for month, total := range monthMap {
			months = append(months, month)
			totalAcrossMonths += total
		}
		sortAsc(months)

		monthlyTotals := make([]monthAmount, 0, len(months))
		for _, month := range months {
			monthlyTotals = append(monthlyTotals, monthAmount{
				Month: month,
				Total: centsToDecimalString(monthMap[month]),
			})
		}

		average := "0.00"
		if len(months) > 0 {
			avgCents := int64(math.Round(float64(totalAcrossMonths) / float64(len(months))))
			average = centsToDecimalString(avgCents)
		}

		stats = append(stats, monthlyCurrencyStat{
			Currency:                currency,
			MonthlyTotals:           monthlyTotals,
			AverageMonthlyRepayment: average,
		})
	}
	return stats
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func centsToDecimalString(amountCents int64) string {
	sign := ""
	if amountCents < 0 {
		sign = "-"
		amountCents = -amountCents
	}
	return fmt.Sprintf("%s%d.%02d", sign, amountCents/100, amountCents%100)
}

func sortDesc(items []string) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j] > items[j-1]; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

func sortAsc(items []string) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j] < items[j-1]; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

func corsMiddleware(next http.Handler, allowedOrigins map[string]struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func buildAllowedOrigins(origins []string) map[string]struct{} {
	result := map[string]struct{}{}
	candidates := origins
	if len(candidates) == 0 {
		candidates = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}
	for _, origin := range candidates {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		result[trimmed] = struct{}{}
	}
	if len(result) == 0 {
		result["http://localhost:5173"] = struct{}{}
		result["http://127.0.0.1:5173"] = struct{}{}
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
