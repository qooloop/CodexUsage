package usage

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// Opt-in integration check. Never print credentials, headers or response bodies.
func TestLiveQuota(t *testing.T) {
	if os.Getenv("CODEX_LIVE_CHECK") != "1" {
		t.Skip("set CODEX_LIVE_CHECK=1 for authenticated integration check")
	}
	tokens, err := readAuthTokens(codexAuthPath())
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	status, body, err := whamGet(tokens)
	t.Logf("HTTP status=%d elapsed=%s error=%v", status, time.Since(start).Round(time.Millisecond), err)
	if err != nil || status != 200 {
		t.Fatal("live request failed")
	}
	var response whamUsageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	event := whamToEvent(&response)
	if event.Primary == nil || event.Secondary == nil {
		t.Fatal("missing quota windows")
	}
	t.Logf("5h used=%.2f remaining=%d; week used=%.2f remaining=%d; observed=%s", event.Primary.UsedPercent, DisplayPercent(100-event.Primary.UsedPercent), event.Secondary.UsedPercent, DisplayPercent(100-event.Secondary.UsedPercent), time.Now().Format(time.RFC3339))
}

func TestLiveLocalQuota(t *testing.T) {
	if os.Getenv("CODEX_LIVE_CHECK") != "1" {
		t.Skip("opt-in local account data check")
	}
	p := NewCodexProvider()
	p.pollOnce()
	s := p.Poll()
	if s.QuotaSource != "local" || s.Primary == nil || s.Secondary == nil {
		t.Fatal("local quota unavailable")
	}
	t.Logf("local 5h=%d%% week=%d%% recorded=%s", DisplayPercent(s.Primary.RemainingPct), DisplayPercent(s.Secondary.RemainingPct), time.UnixMilli(s.Ts).Format(time.RFC3339))
}
