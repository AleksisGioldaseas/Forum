package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type FileSize int64

func (f *FileSize) UnmarshalJSON(data []byte) error {
	var input string
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("fileSize must be a string: %w", err)
	}

	parsed, err := ParseMaxSize(input)
	if err != nil {
		return err
	}

	*f = FileSize(parsed)
	return nil
}

// ParseMaxSize parses a human-readable string like "20MB" into bytes.
func ParseMaxSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	var multiplier int

	switch {
	case strings.HasSuffix(s, "KB"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1 << 20
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "GB"):
		multiplier = 1 << 30
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "B"):
		multiplier = 1
		s = strings.TrimSuffix(s, "B")
	default:
		return 0, fmt.Errorf("unrecognized size unit in %q", s)
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number in size: %w", err)
	}

	return int64(value * float64(multiplier)), nil
}

func (f *FileSize) ToInt64() int64 {
	return int64(*f)
}
