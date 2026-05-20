package utils

import (
	"fmt"
	"hash/fnv"
	"time"
)

// Hash returns a stable hexadecimal FNV-1a 64-bit hash for s.
// The same input always produces the same identifier string.
func Hash(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s)) // Write never returns error for fnv
	return fmt.Sprintf("%x", h.Sum64())
}

// RelativeTime returns a compact relative time string for t relative to now.
func RelativeTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	now := time.Now()
	diff := now.Sub(t)
	if diff < 0 {
		return "now"
	}

	seconds := int(diff.Seconds())
	if seconds < 60 {
		return "now"
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dmin", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 24
	if days < 30 {
		return fmt.Sprintf("%dd", days)
	}
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%dm", months)
	}
	years := months / 12
	return fmt.Sprintf("%dy", years)
}
