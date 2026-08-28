package parser

import (
	"fmt"
	"strings"
	"time"
)

// ParseFlexibleDate parses dates from various IDX reporting formats
func ParseFlexibleDate(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	formats := []string{
		"2006-01-02",
		"January 02, 2006",
		"January 2, 2006",
		"02 January 2006",
		"2 January 2006",
		"02-01-2006",
		"02/01/2006",
		"September 30, 2006",
		"March 31, 2006",
		"June 30, 2006",
		"December 31, 2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unknown date format: %s", val)
}
