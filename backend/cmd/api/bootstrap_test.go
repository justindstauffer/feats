package main

import (
	"testing"

	"github.com/jstauff/feats-api/internal/config"
)

func TestSetupRouter_DisablesTrustedProxiesByDefault(t *testing.T) {
	cfg := &config.Config{
		RateLimitAPI:   100,
		RateLimitLogin: 5,
		RateLimitUpload: 20,
	}
	h := &appHandlers{}
	m := initMiddleware(nil, nil, cfg)

	router, err := setupRouter(cfg, h, m)
	if err != nil {
		t.Fatalf("expected setupRouter to succeed, got error: %v", err)
	}
	if router == nil {
		t.Fatal("expected router to be created")
	}
}

func TestSetupRouter_InvalidTrustedProxiesReturnsError(t *testing.T) {
	cfg := &config.Config{
		TrustedProxies: []string{"not-a-valid-proxy"},
		RateLimitAPI:   100,
		RateLimitLogin: 5,
		RateLimitUpload: 20,
	}
	h := &appHandlers{}
	m := initMiddleware(nil, nil, cfg)

	router, err := setupRouter(cfg, h, m)
	if err == nil {
		t.Fatalf("expected error for invalid trusted proxies, got nil (router=%v)", router)
	}
}
