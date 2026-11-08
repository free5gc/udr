package sbi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/free5gc/udr/internal/logger"
	"github.com/free5gc/udr/pkg/factory"
	logger_util "github.com/free5gc/util/logger"
)

func TestEeGroupSubscriptionsLogForging(t *testing.T) {
	const marker = "FORGED_LOG_LINE_1231"

	// Capture UDR logs and restore the original logger after testing.
	originalOutput := logger.Log.Out
	originalLevel := logger.Log.GetLevel()
	t.Cleanup(func() {
		logger.Log.SetOutput(originalOutput)
		logger.Log.SetLevel(originalLevel)
	})
	logger.Log.SetLevel(logrus.InfoLevel)

	tests := []struct {
		name       string
		method     string
		encoded    string
		wantEscape string
	}{
		{"GET LF injection", http.MethodGet, "%0A", `\n`},
		{"GET CRLF injection", http.MethodGet, "%0D%0A", `\r\n`},
		{"POST LF injection", http.MethodPost, "%0A", `\n`},
		{"POST CRLF injection", http.MethodPost, "%0D%0A", `\r\n`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger.Log.SetOutput(&logs)

			// Use the same Gin logging middleware as the UDR server.
			router := logger_util.NewGinWithLogrus(logger.GinLog)
			path := factory.UdrDrResUriPrefix +
				"/subscription-data/group-data/:ueGroupId/ee-subscriptions"
			router.GET(path, (&Server{}).HandleQueryEeGroupSubscriptions)
			router.POST(path, (&Server{}).HandleCreateEeGroupSubscriptions)
			url := factory.UdrDrResUriPrefix +
				"/subscription-data/group-data/invalid" +
				tt.encoded + marker + "/ee-subscriptions"

			// POST requires a valid JSON body before ueGroupId validation.
			body := ""
			if tt.method == http.MethodPost {
				body = "{}"
			}

			req := httptest.NewRequestWithContext(
				context.Background(), tt.method, url, strings.NewReader(body),
			)

			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}
			rsp := httptest.NewRecorder()
			router.ServeHTTP(rsp, req)

			// Invalid ueGroupId must still return HTTP 400.
			if rsp.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want 400", rsp.Code)
			}

			dataRepoFound := false
			ginFound := false

			// Both logs must keep injected content on the original line.
			for _, line := range strings.Split(logs.String(), "\n") {
				if !strings.Contains(line, marker) {
					continue
				}

				if !strings.Contains(line, tt.wantEscape) ||
					strings.Contains(line, "\r") {
					t.Errorf("unsafe log line: %q", line)
				}

				switch {
				case strings.Contains(line, "Invalid ueGroupId:"):
					dataRepoFound = true
				case strings.Contains(line, "| 400 |"):
					ginFound = true
				default:
					t.Errorf("possible forged log line: %q", line)
				}
			}

			if !dataRepoFound || !ginFound {
				t.Errorf("missing sanitized logs: DataRepo=%v, Gin=%v",
					dataRepoFound, ginFound)
			}

			if strings.Count(logs.String(), marker) != 2 {
				t.Errorf("expected marker once in each log, got:\n%s",
					logs.String())
			}
		})
	}
}
