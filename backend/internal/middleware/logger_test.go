package middleware

import (
	"net/url"
	"strings"
	"testing"
)

func TestRedactSensitiveQuery(t *testing.T) {
	q := url.Values{}
	q.Set("token", "abc123")
	q.Set("access_token", "def456")
	q.Set("refresh_token", "ghi789")
	q.Set("authorization", "Bearer xyz")
	q.Set("page", "2")

	got := redactSensitiveQuery(q)
	if strings.Contains(got, "abc123") || strings.Contains(got, "def456") || strings.Contains(got, "ghi789") || strings.Contains(got, "Bearer+xyz") {
		t.Fatalf("sensitive values leaked in query output: %s", got)
	}

	if !strings.Contains(got, "page=2") {
		t.Fatalf("non-sensitive key unexpectedly removed: %s", got)
	}

	if !strings.Contains(got, "token=%5BREDACTED%5D") {
		t.Fatalf("token was not redacted: %s", got)
	}
}
