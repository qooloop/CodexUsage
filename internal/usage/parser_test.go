package usage

import (
	"os"
	"path/filepath"
	"testing"
)

const testQuotaLine = `{"timestamp":"2026-09-10T01:00:00Z","type":"event_msg","payload":{"type":"token_count","rate_limits":{"primary":{"used_percent":44,"window_minutes":300}}}}`

func TestPartialJSONLWaitsAtExactLineStart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rollout-test.jsonl")
	half := len(testQuotaLine) / 2
	if err := os.WriteFile(path, []byte(testQuotaLine[:half]), 0600); err != nil {
		t.Fatal(err)
	}
	events, offset := ReadNewEvents(path, 0)
	if len(events) != 0 || offset != 0 {
		t.Fatalf("partial first line advanced offset: %d", offset)
	}
	if err := os.WriteFile(path, []byte(testQuotaLine+"\n"+testQuotaLine[:half]), 0600); err != nil {
		t.Fatal(err)
	}
	events, offset = ReadNewEvents(path, offset)
	if len(events) != 1 || offset != int64(len(testQuotaLine)+1) {
		t.Fatalf("partial tail offset=%d events=%d", offset, len(events))
	}
	if err := os.WriteFile(path, []byte(testQuotaLine+"\n"+testQuotaLine+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	events, offset = ReadNewEvents(path, offset)
	if len(events) != 1 || events[0].Primary.UsedPercent != 44 || offset != int64(2*(len(testQuotaLine)+1)) {
		t.Fatal("completed line lost or duplicated")
	}
	if events, next := ReadNewEvents(path, offset); len(events) != 0 || next != offset {
		t.Fatal("reread duplicated events")
	}
}
