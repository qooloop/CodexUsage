package desktop

import (
	"codex-monitor/internal/usage"
	"os"
	"testing"
	"time"
)

type refreshTestProvider struct{ calls int }

func (p *refreshTestProvider) ID() string                        { return "test" }
func (p *refreshTestProvider) Name() string                      { return "test" }
func (p *refreshTestProvider) ShortName() string                 { return "test" }
func (p *refreshTestProvider) Color() string                     { return "" }
func (p *refreshTestProvider) Emoji() string                     { return "" }
func (p *refreshTestProvider) Start(func(*usage.Snapshot)) error { return nil }
func (p *refreshTestProvider) Stop() error                       { return nil }
func (p *refreshTestProvider) Poll() *usage.Snapshot {
	panic("refresh must collect data, not read cached Poll")
}

// Exercises the same real collection/publication path as both refresh actions.
func TestLiveManualRefresh(t *testing.T) {
	if os.Getenv("CODEX_LIVE_CHECK") != "1" {
		t.Skip("opt-in live account integration")
	}
	oldReg, oldSnaps, oldTray := providerReg, snapshots, trayLatest
	defer func() { providerReg, snapshots, trayLatest = oldReg, oldSnaps, oldTray }()
	providerReg = usage.NewProviderRegistry()
	snapshots = make(map[string]*usage.Snapshot)
	p := usage.NewCodexProvider()
	providerReg.Register(p)
	// Two consecutive clicks must each finish real collection and publish it.
	var previous int64
	for i := 0; i < 2; i++ {
		start := time.Now()
		out := NewApp().Refresh()
		if len(out) != 1 || out[0].Error != "" || out[0].Primary == nil || out[0].Secondary == nil {
			t.Fatal("live refresh did not publish valid quota")
		}
		s := out[0]
		if s.Ts <= previous || s.Ts < start.UnixMilli() {
			t.Fatal("refresh returned old quota")
		}
		if trayLatest != s {
			t.Fatal("tray and frontend received different snapshots")
		}
		state := trayStateFromSnapshot(s)
		if state.ShortQuotaPercent != usage.DisplayPercent(s.Primary.RemainingPct) || refreshRunning.Load() {
			t.Fatal("tray number or refresh completion incorrect")
		}
		t.Logf("refresh %d: elapsed=%s 5h=%d%% week=%d%% tray=%d frontend=published observed=%s", i+1, time.Since(start).Round(time.Millisecond), usage.DisplayPercent(s.Primary.RemainingPct), usage.DisplayPercent(s.Secondary.RemainingPct), state.ShortQuotaPercent, time.UnixMilli(s.Ts).Format(time.RFC3339))
		previous = s.Ts
	}
}
func (p *refreshTestProvider) Refresh() *usage.Snapshot {
	p.calls++
	return &usage.Snapshot{ID: "test", Primary: &usage.QuotaWindow{RemainingPct: 57.9}}
}
func TestManualRefreshPublishesCollectedSnapshot(t *testing.T) {
	oldReg, oldSnaps := providerReg, snapshots
	defer func() { providerReg, snapshots = oldReg, oldSnaps }()
	providerReg = usage.NewProviderRegistry()
	snapshots = make(map[string]*usage.Snapshot)
	p := &refreshTestProvider{}
	providerReg.Register(p)
	forceRefresh()
	if p.calls != 1 || snapshots["test"] == nil || snapshots["test"].Primary.RemainingPct != 57.9 {
		t.Fatal("collected quota was not published")
	}
	if refreshRunning.Load() {
		t.Fatal("refresh status did not reset")
	}
}
