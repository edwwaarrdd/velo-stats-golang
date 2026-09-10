package httpapi_test

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"velostats/internal/httpapi"
	"velostats/internal/rides"
	"velostats/internal/stations"
	"velostats/internal/testsupport"
)

func get(t *testing.T, db *sql.DB, path string) (int, string) {
	t.Helper()

	handler := httpapi.Router(
		rides.NewRepository(db),
		rides.NewSummaryCalculator(db),
		rides.NewCostCalculator(rides.NewRepository(db)),
		stations.NewRepository(db),
		[]string{"http://localhost:5173"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

	return recorder.Code, recorder.Body.String()
}

func assertJSON(t *testing.T, body, want string) {
	t.Helper()

	var got, expected any

	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode response %s: %v", body, err)
	}

	if err := json.Unmarshal([]byte(want), &expected); err != nil {
		t.Fatalf("decode expectation: %v", err)
	}

	gotNormalised, _ := json.Marshal(got)
	expectedNormalised, _ := json.Marshal(expected)

	if string(gotNormalised) != string(expectedNormalised) {
		t.Errorf("response body\n got: %s\nwant: %s", gotNormalised, expectedNormalised)
	}
}

func TestHealthcheckReportsThatTheApplicationIsUp(t *testing.T) {
	status, body := get(t, testsupport.NewDatabase(t), "/_healthcheck")

	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}

	if body != `{"message":"ok"}` {
		t.Errorf("body = %s", body)
	}
}

func TestEveryEndpointAnswersWithATrailingSlash(t *testing.T) {
	db := testsupport.NewDatabase(t)

	paths := map[string]string{
		"/_healthcheck/":  "/_healthcheck",
		"/rides/":         "/rides",
		"/rides//":        "/rides",
		"/rides/summary/": "/rides/summary",
		"/rides/cost/":    "/rides/cost",
		"/stations/":      "/stations",
	}

	for path, canonical := range paths {
		t.Run(path, func(t *testing.T) {
			status, body := get(t, db, path)

			if status != http.StatusOK {
				t.Fatalf("status = %d, want 200", status)
			}

			wantStatus, wantBody := get(t, db, canonical)

			if status != wantStatus || body != wantBody {
				t.Errorf("%s returned %d %s, want the same as %s: %d %s", path, status, body, canonical, wantStatus, wantBody)
			}
		})
	}
}

func TestAnUnknownPathIsStillNotFound(t *testing.T) {
	for _, path := range []string{"/nope", "/rides/nope", "/rides/summary/nope"} {
		if status, _ := get(t, testsupport.NewDatabase(t), path); status != http.StatusNotFound {
			t.Errorf("%s returned %d, want 404", path, status)
		}
	}
}
