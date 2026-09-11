package usage

import (
	"math"
	"testing"
)

func TestDisplayPercent(t *testing.T) {
	for _, c := range []struct {
		value float64
		want  int
	}{
		{57.9, 57}, {57.5, 57}, {57.01, 57}, {57, 57}, {51.99, 51}, {0.99, 0}, {0, 0}, {100, 100}, {101, 100}, {-1, 0}, {math.NaN(), 0}, {math.Inf(1), 0},
	} {
		if got := DisplayPercent(c.value); got != c.want {
			t.Errorf("DisplayPercent(%v)=%d, want %d", c.value, got, c.want)
		}
	}
}
func TestFractionalQuotaKeepsRawPrecision(t *testing.T) {
	q := windowToQuotaWindow(&Window{UsedPercent: 42.1, WindowMinutes: 300}, 0)
	if DisplayPercent(q.RemainingPct) != 57 {
		t.Fatalf("remaining=%v", q.RemainingPct)
	}
	if math.Abs(q.RemainingPct-57.9) > 1e-9 {
		t.Fatal("raw progress data was truncated")
	}
}
