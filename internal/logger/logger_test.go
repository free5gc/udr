package logger

import (
	"testing"

	"github.com/sirupsen/logrus"

	logger_util "github.com/free5gc/util/logger"
)

func TestGinLogSanitizer(t *testing.T) {
	tests := []struct {
		name     string
		category string
		input    string
		want     string
	}{
		{
			name:     "newline injection",
			category: "GIN",
			input:    "invalid\n[INFO] forged",
			want:     `invalid\n[INFO] forged`,
		},
		{
			name:     "CRLF injection",
			category: "GIN",
			input:    "invalid\r\n[INFO] forged",
			want:     `invalid\r\n[INFO] forged`,
		},
		{
			name:     "normal request",
			category: "GIN",
			input:    "/nudr-dr/v2/subscription-data",
			want:     "/nudr-dr/v2/subscription-data",
		},
		{
			name:     "other log category",
			category: "DataRepo",
			input:    "message\nsecond line",
			want:     "message\nsecond line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &logrus.Entry{
				Data: logrus.Fields{
					logger_util.FieldCategory: tt.category,
				},
				Message: tt.input,
			}

			if err := (ginLogSanitizer{}).Fire(entry); err != nil {
				t.Fatalf("sanitizer failed: %v", err)
			}

			if entry.Message != tt.want {
				t.Errorf("got %q, want %q", entry.Message, tt.want)
			}
		})
	}
}
