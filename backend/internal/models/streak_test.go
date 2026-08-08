package models

import (
	"testing"
	"time"
)

// The family's zone; day boundaries must follow this, not UTC.
func eastern(t *testing.T) *time.Location {
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("tz data unavailable: %v", err)
	}
	return tz
}

func TestStreak_FirstActivity(t *testing.T) {
	tz := eastern(t)
	s := &Streak{}
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, tz)
	s.updateForNewActivity(now, now, tz)

	if s.CurrentStreak != 1 || s.LongestStreak != 1 {
		t.Fatalf("first activity: got current=%d longest=%d, want 1/1", s.CurrentStreak, s.LongestStreak)
	}
}

func TestStreak_SecondPostSameLocalDay_NoIncrement(t *testing.T) {
	tz := eastern(t)
	s := &Streak{}
	// 9am ET then 11pm ET the SAME local day. 11pm ET = 03:00 UTC next day, so
	// the old UTC-truncation logic wrongly treated these as two different days.
	morning := time.Date(2026, 8, 10, 9, 0, 0, 0, tz)
	night := time.Date(2026, 8, 10, 23, 0, 0, 0, tz)

	s.updateForNewActivity(morning, morning, tz)
	s.updateForNewActivity(night, night, tz)

	if s.CurrentStreak != 1 {
		t.Fatalf("two posts same local day should stay at 1, got %d", s.CurrentStreak)
	}
}

func TestStreak_ConsecutiveLocalDays_EveningPosts(t *testing.T) {
	tz := eastern(t)
	s := &Streak{}
	// Monday 9am, then Tuesday 9pm ET. Tuesday 9pm ET crosses UTC midnight, but
	// it is still the next *local* calendar day, so the streak must extend.
	mon := time.Date(2026, 8, 10, 9, 0, 0, 0, tz)
	tue := time.Date(2026, 8, 11, 21, 0, 0, 0, tz)

	s.updateForNewActivity(mon, mon, tz)
	s.updateForNewActivity(tue, tue, tz)

	if s.CurrentStreak != 2 {
		t.Fatalf("consecutive local days should give streak 2, got %d", s.CurrentStreak)
	}
}

func TestStreak_MissedDay_Resets(t *testing.T) {
	tz := eastern(t)
	s := &Streak{}
	mon := time.Date(2026, 8, 10, 9, 0, 0, 0, tz)
	wed := time.Date(2026, 8, 12, 9, 0, 0, 0, tz) // skipped Tuesday

	s.updateForNewActivity(mon, mon, tz)
	s.updateForNewActivity(wed, wed, tz)

	if s.CurrentStreak != 1 {
		t.Fatalf("gap should reset to 1, got %d", s.CurrentStreak)
	}
}

func TestStreak_LongestPreservedAfterReset(t *testing.T) {
	tz := eastern(t)
	s := &Streak{}
	d1 := time.Date(2026, 8, 10, 9, 0, 0, 0, tz)
	d2 := time.Date(2026, 8, 11, 9, 0, 0, 0, tz)
	d3 := time.Date(2026, 8, 13, 9, 0, 0, 0, tz) // gap after 2

	s.updateForNewActivity(d1, d1, tz)
	s.updateForNewActivity(d2, d2, tz)
	s.updateForNewActivity(d3, d3, tz)

	if s.CurrentStreak != 1 {
		t.Fatalf("current after reset should be 1, got %d", s.CurrentStreak)
	}
	if s.LongestStreak != 2 {
		t.Fatalf("longest should be preserved at 2, got %d", s.LongestStreak)
	}
}

func TestStreak_CheckReset(t *testing.T) {
	tz := eastern(t)
	base := time.Date(2026, 8, 12, 9, 0, 0, 0, tz)
	day := calendarDay(base, tz)

	cases := []struct {
		name      string
		last      time.Time
		now       time.Time
		wantReset bool
	}{
		{"active today", day, base, false},
		{"active yesterday", day.AddDate(0, 0, -1), base, false},
		{"two days ago", day.AddDate(0, 0, -2), base, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			last := c.last
			s := &Streak{CurrentStreak: 5, LastActivityDate: &last}
			changed := s.checkAndResetIfNeeded(c.now, tz)
			if changed != c.wantReset {
				t.Fatalf("reset=%v, want %v", changed, c.wantReset)
			}
			if c.wantReset && s.CurrentStreak != 0 {
				t.Fatalf("expected streak reset to 0, got %d", s.CurrentStreak)
			}
			if !c.wantReset && s.CurrentStreak != 5 {
				t.Fatalf("expected streak preserved at 5, got %d", s.CurrentStreak)
			}
		})
	}
}
