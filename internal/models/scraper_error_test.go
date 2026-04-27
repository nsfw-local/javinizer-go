package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScraperStatusErrorClassifiesKinds(t *testing.T) {
	tests := []struct {
		status int
		want   ScraperErrorKind
	}{
		{status: 404, want: ScraperErrorKindNotFound},
		{status: 429, want: ScraperErrorKindRateLimited},
		{status: 403, want: ScraperErrorKindBlocked},
		{status: 502, want: ScraperErrorKindUnavailable},
		{status: 418, want: ScraperErrorKindUnknown},
	}

	for _, tt := range tests {
		err := NewScraperStatusError("Test", tt.status, "")
		if err.Kind != tt.want {
			t.Fatalf("status %d -> kind %q, want %q", tt.status, err.Kind, tt.want)
		}
	}
}

func TestAsScraperErrorUnwrap(t *testing.T) {
	inner := NewScraperStatusError("Test", 502, "upstream bad gateway")
	outer := errors.New("wrapped: " + inner.Error())

	if got, ok := AsScraperError(outer); ok || got != nil {
		t.Fatalf("plain wrapped string should not match ScraperError")
	}

	wrapped := errors.Join(errors.New("context"), inner)
	got, ok := AsScraperError(wrapped)
	if !ok || got == nil {
		t.Fatalf("expected AsScraperError to find joined scraper error")
	}
	if got.StatusCode != 502 {
		t.Fatalf("unexpected status code: %d", got.StatusCode)
	}
}

func TestScraperError_ErrorFallbacks(t *testing.T) {
	tests := []struct {
		name string
		err  *ScraperError
		want string
	}{
		{
			name: "nil receiver",
			err:  nil,
			want: "",
		},
		{
			name: "explicit message wins",
			err: &ScraperError{
				Scraper: "DMM",
				Message: "explicit message",
			},
			want: "explicit message",
		},
		{
			name: "status with scraper",
			err: &ScraperError{
				Scraper:    "DMM",
				StatusCode: 503,
			},
			want: "DMM returned status code 503",
		},
		{
			name: "status without scraper",
			err: &ScraperError{
				StatusCode: 404,
			},
			want: "scraper returned status code 404",
		},
		{
			name: "scraper only",
			err: &ScraperError{
				Scraper: "DMM",
			},
			want: "DMM scraper error",
		},
		{
			name: "generic fallback",
			err:  &ScraperError{},
			want: "scraper error",
		},
	}

	for _, tt := range tests {
		got := tt.err.Error()
		if got != tt.want {
			t.Fatalf("%s: Error() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestScraperError_Unwrap(t *testing.T) {
	var nilErr *ScraperError
	if nilErr.Unwrap() != nil {
		t.Fatalf("nil receiver Unwrap() should return nil")
	}

	cause := errors.New("root cause")
	err := &ScraperError{Cause: cause}
	if err.Unwrap() != cause {
		t.Fatalf("Unwrap() did not return cause")
	}
}

func TestScraperError_ConstructorsTrimAndDefaults(t *testing.T) {
	notFound := NewScraperNotFoundError("JavDB", "  no result  ")
	if notFound.Message != "no result" {
		t.Fatalf("not found message should be trimmed, got %q", notFound.Message)
	}

	challengeWithMessage := NewScraperChallengeError("JavDB", "  challenge blocked  ")
	if challengeWithMessage.Message != "challenge blocked" {
		t.Fatalf("challenge message should be trimmed, got %q", challengeWithMessage.Message)
	}

	challengeDefault := NewScraperChallengeError("JavDB", "")
	if challengeDefault.Message == "" {
		t.Fatal("expected default challenge message")
	}
	if challengeDefault.Kind != ScraperErrorKindBlocked {
		t.Fatalf("expected blocked kind, got %q", challengeDefault.Kind)
	}
	if !challengeDefault.Temporary || !challengeDefault.Retryable {
		t.Fatal("expected challenge error to be temporary and retryable")
	}
}

func TestScraperHTTPError_Error(t *testing.T) {
	tests := []struct {
		name         string
		err          *ScraperHTTPError
		wantContains string
	}{
		{"nil returns empty", nil, ""},
		{"message present", &ScraperHTTPError{Scraper: "S", StatusCode: 429, Message: "rate limited"}, "rate limited"},
		{"no message with scraper", &ScraperHTTPError{Scraper: "JavDB", StatusCode: 403, Message: ""}, "JavDB returned HTTP 403"},
		{"no message no scraper", &ScraperHTTPError{Scraper: "", StatusCode: 500, Message: ""}, "HTTP 500"},
		{"no message no status", &ScraperHTTPError{Scraper: "", StatusCode: 0, Message: ""}, "scraper HTTP error"},
		{"whitespace message", &ScraperHTTPError{Scraper: "", StatusCode: 502, Message: "   "}, "HTTP 502"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.err.Error()
			if tc.wantContains == "" {
				assert.Equal(t, "", got)
			} else {
				assert.Equal(t, tc.wantContains, got)
			}
		})
	}
}

func TestNewScraperHTTPError(t *testing.T) {
	err := NewScraperHTTPError("TestScraper", 404, "not found")
	assert.Equal(t, "TestScraper", err.Scraper)
	assert.Equal(t, 404, err.StatusCode)
	assert.Equal(t, "not found", err.Message)
}
