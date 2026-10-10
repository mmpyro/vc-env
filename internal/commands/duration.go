package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDurationShorthand parses a duration string using Go's
// time.ParseDuration semantics, extended with the shorthand units "d" (days)
// and "w" (weeks).  The second return value is true if a non-empty input was
// parsed — convenient for distinguishing "flag not provided" from "flag
// provided with zero".
//
// Supported shorthand forms (case-insensitive):
//
//	30d       → 30 * 24h
//	2w        → 2 * 7 * 24h
//	1w3d      → 1 * 7 * 24h + 3 * 24h
//	12h30m    → regular Go-style duration, unchanged
//
// Mixing shorthand and sub-day units in the same token is not supported
// (e.g. `30d12h`); split them if needed or stick to Go-style hours only.
func ParseDurationShorthand(s string) (time.Duration, bool, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0, false, fmt.Errorf("empty duration")
	}

	// Fast path: pure Go duration (contains no d/w, or already parses).
	if d, err := time.ParseDuration(trimmed); err == nil {
		return d, true, nil
	}

	// Shorthand path: parse a sequence of <number><unit> pairs where unit is
	// one of d/w (days/weeks).  Any other unit is treated as an error so we
	// don't silently accept malformed input like `30x`.
	var total time.Duration
	i := 0
	runes := []rune(strings.ToLower(trimmed))
	parsedAny := false
	for i < len(runes) {
		// Read digits.
		start := i
		for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
			i++
		}
		if start == i {
			return 0, false, fmt.Errorf("expected number at position %d", start)
		}
		n, err := strconv.Atoi(string(runes[start:i]))
		if err != nil {
			return 0, false, fmt.Errorf("invalid number %q", string(runes[start:i]))
		}
		if i >= len(runes) {
			return 0, false, fmt.Errorf("missing unit after number %d", n)
		}
		unit := runes[i]
		i++
		switch unit {
		case 'd':
			total += time.Duration(n) * 24 * time.Hour
		case 'w':
			total += time.Duration(n) * 7 * 24 * time.Hour
		default:
			return 0, false, fmt.Errorf("unknown unit %q (expected d or w, or a Go duration like 30m/24h)", string(unit))
		}
		parsedAny = true
	}
	if !parsedAny {
		return 0, false, fmt.Errorf("could not parse %q as a duration", s)
	}
	return total, true, nil
}
