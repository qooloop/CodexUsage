package usage

import "math"

// DisplayPercent truncates fractional remaining quota instead of rounding up.
// Keep the original RemainingPct for progress bars and data calculations.
func DisplayPercent(remaining float64) int {
	if math.IsNaN(remaining) || math.IsInf(remaining, 0) {
		return 0
	}
	return int(math.Floor(math.Max(0, math.Min(100, remaining))))
}
