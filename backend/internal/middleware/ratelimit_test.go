package middleware

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func newLoginContext(body string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	c.Request.RemoteAddr = "203.0.113.9:5000"
	return c
}

func TestLoginRateKey_DistinctEmailsGetDistinctKeys(t *testing.T) {
	k1 := loginRateKey(newLoginContext(`{"email":"alice@example.com","password":"x"}`))
	k2 := loginRateKey(newLoginContext(`{"email":"bob@example.com","password":"y"}`))
	if k1 == k2 {
		t.Fatalf("different emails must not share a rate-limit key: %q == %q", k1, k2)
	}
	if !strings.HasPrefix(k1, "login:email:") {
		t.Fatalf("expected email-based key, got %q", k1)
	}
}

func TestLoginRateKey_IsCaseInsensitive(t *testing.T) {
	k1 := loginRateKey(newLoginContext(`{"email":"Alice@Example.com"}`))
	k2 := loginRateKey(newLoginContext(`{"email":"alice@example.com"}`))
	if k1 != k2 {
		t.Fatalf("email key should be case-insensitive: %q != %q", k1, k2)
	}
}

func TestLoginRateKey_FallsBackToIPWithoutEmail(t *testing.T) {
	// e.g. token refresh has no email in the body.
	key := loginRateKey(newLoginContext(`{"refresh_token":"abc"}`))
	if !strings.HasPrefix(key, "login:ip:") {
		t.Fatalf("expected IP fallback key, got %q", key)
	}
}

func TestExtractLoginEmail_RestoresBodyForHandler(t *testing.T) {
	const body = `{"email":"carol@example.com","password":"secret"}`
	c := newLoginContext(body)

	_ = extractLoginEmail(c) // consumes the body internally

	// The handler must still be able to read the full original body.
	got, err := io.ReadAll(c.Request.Body)
	if err != nil {
		t.Fatalf("reading restored body: %v", err)
	}
	if string(got) != body {
		t.Fatalf("body not restored: got %q want %q", string(got), body)
	}
}
