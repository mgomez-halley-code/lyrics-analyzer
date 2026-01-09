package service

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseTimestamp converts timestamp string (mm:ss.xx) to seconds
func (p *Parser) ParseTimestamp(timestamp string) (float64, error) {
	parts := strings.Split(timestamp, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid timestamp format: %s", timestamp)
	}

	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %w", err)
	}

	seconds, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid seconds: %w", err)
	}

	return float64(minutes)*60 + seconds, nil
}
