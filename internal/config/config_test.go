package config

import "testing"

func TestRefreshIntervals(t *testing.T) {
	for _, v := range []int{5, 10, 30, 60, 300} {
		if NormalizeInterval(v) != v {
			t.Fatalf("unsupported interval %d", v)
		}
	}
	for _, v := range []int{0, -1, 1, 120, 301} {
		if NormalizeInterval(v) != 60 {
			t.Fatalf("invalid interval %d not migrated", v)
		}
	}
}
