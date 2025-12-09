package tools

import (
	"fmt"
	"time"
)

// parseTime parses time strings in RFC3339 format or relative format like "now-1h".
func parseTime(s string) (time.Time, error) {
	if s == "" || s == "now" {
		return time.Now(), nil
	}

	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try relative time format (now-1h, now-15m, etc.)
	if len(s) > 3 && s[:3] == "now" {
		offset := s[3:]
		d, err := time.ParseDuration(offset)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid relative time format: %s", s)
		}
		return time.Now().Add(d), nil
	}

	// Try date-time format without timezone
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t, nil
	}

	// Try date only format
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unrecognized time format: %s", s)
}
