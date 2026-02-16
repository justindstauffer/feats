package services

import (
	"errors"
	"testing"
)

func TestRedeemInviteRejectsInvalidCodeLength(t *testing.T) {
	service := NewGroupService(nil, nil)

	_, err := service.RedeemInvite("ABC-123", "user-1")
	if !errors.Is(err, ErrInvalidInviteCode) {
		t.Fatalf("expected ErrInvalidInviteCode, got %v", err)
	}
}

