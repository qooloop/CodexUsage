package usage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRefreshFetchesOnlineQuotaAndRetainsCacheOnFailure(t *testing.T) {
	t.Setenv("CODEX_HOME", t.TempDir())
	p := NewCodexProvider()
	calls := 0
	fail := false
	p.fetchQuota = func() (*Event, error) {
		calls++
		if fail {
			return nil, errors.New("offline")
		}
		return &Event{Ts: time.Now().UnixMilli(), Primary: &Window{UsedPercent: 42.1, WindowMinutes: 300}}, nil
	}
	if p.Poll().Primary != nil || calls != 0 {
		t.Fatal("Poll should only read cache")
	}
	first := p.Refresh()
	if calls != 1 || first.Primary == nil || DisplayPercent(first.Primary.RemainingPct) != 57 {
		t.Fatalf("real refresh failed: %+v, calls %d", first, calls)
	}
	fail = true
	failed := p.Refresh()
	if failed.Primary == nil || failed.Primary.RemainingPct != first.Primary.RemainingPct || failed.Error == "" || p.Poll().Error == "" {
		t.Fatal("failure must retain cached quota and error state")
	}
	fail = false
	if p.Refresh().Error != "" {
		t.Fatal("successful refresh must clear error")
	}
}

func TestLocalAppendPublishesBeforeNetworkAndWithUnchangedMtime(t *testing.T) {
	t.Setenv("CODEX_HOME", t.TempDir())
	p := NewCodexProvider()
	if err := os.MkdirAll(p.sessionsDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(p.sessionsDir, "rollout-test.jsonl")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().Add(-time.Minute)
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	p.pollOnce()
	if err := os.WriteFile(path, []byte(testQuotaLine+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	var published *Snapshot
	p.onUpdate = func(s *Snapshot) { published = s }
	p.fetchQuota = func() (*Event, error) {
		if published == nil || published.Primary == nil || published.Primary.RemainingPct != 56 || published.QuotaSource != "local" {
			t.Fatal("local snapshot was not published before networking")
		}
		return nil, errors.New("offline")
	}
	s := p.Refresh()
	if s.Primary == nil || s.Primary.RemainingPct != 56 || s.QuotaSource != "local" {
		t.Fatal("network failure blocked local sync")
	}
}

func TestNewerLocalQuotaWinsAfterOnlineFailure(t *testing.T) {
	t.Setenv("CODEX_HOME", t.TempDir())
	p := NewCodexProvider()
	now := time.Now().UnixMilli()
	p.liveQuota = &Event{Ts: now - 60000, Primary: &Window{UsedPercent: 41, WindowMinutes: 300}}
	p.latestQuota = &Event{Ts: now, Primary: &Window{UsedPercent: 44, WindowMinutes: 300}}
	p.fetchQuota = func() (*Event, error) { return nil, errors.New("offline") }
	s := p.Refresh()
	if s.Primary.RemainingPct != 56 || s.Ts != now || s.Error == "" {
		t.Fatalf("stale online snapshot won: %+v", s)
	}
	p.fetchQuota = func() (*Event, error) {
		return &Event{Ts: now + 1000, Primary: &Window{UsedPercent: 45, WindowMinutes: 300}}, nil
	}
	s = p.Refresh()
	if s.Primary.RemainingPct != 55 || s.Error != "" {
		t.Fatalf("recovery failed: %+v", s)
	}
}

func TestExpiredQuotaDoesNotInventReset(t *testing.T) {
	t.Setenv("CODEX_HOME", t.TempDir())
	p := NewCodexProvider()
	expired := time.Now().Add(-time.Minute).Unix()
	p.latestQuota = &Event{Ts: time.Now().UnixMilli(), Primary: &Window{UsedPercent: 95, WindowMinutes: 300, ResetsAt: &expired}}
	s := p.Poll()
	if s.Primary.RemainingPct != 5 || *s.Primary.ResetsAt != expired || s.Source != "stale" {
		t.Fatalf("invented reset: %+v", s.Primary)
	}
}
