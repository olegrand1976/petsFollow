package kernel

var AllowedHeartRateDurations = []int{15, 30, 60}

func IsAllowedHeartRateDuration(sec int) bool {
	for _, d := range AllowedHeartRateDurations {
		if d == sec {
			return true
		}
	}
	return false
}

func NormalizeHeartRateDurations(durations []int) []int {
	seen := make(map[int]bool)
	var out []int
	for _, d := range AllowedHeartRateDurations {
		for _, v := range durations {
			if v == d && !seen[d] {
				seen[d] = true
				out = append(out, d)
				break
			}
		}
	}
	if len(out) == 0 {
		return []int{60}
	}
	return out
}
