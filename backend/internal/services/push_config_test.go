package services

import (
	"testing"

	"github.com/jstauff/feats-api/internal/config"
)

func TestNewPushService_DisabledWhenBundleIDMissing(t *testing.T) {
	cfg := &config.Config{
		APNsKeyPath:  "/tmp/fake-key.p8",
		APNsKeyID:    "KEYID12345",
		APNsTeamID:   "TEAMID1234",
		APNsBundleID: "",
	}

	svc := NewPushService(nil, cfg)
	if svc.enabled {
		t.Fatal("expected push service to be disabled when APNS_BUNDLE_ID is missing")
	}
}
