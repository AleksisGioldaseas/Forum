package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// Custom type to unmarshal time.Duration values from json
type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	errorMsg := "unmarshalDuration %w"
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	*d = Duration(duration)
	return nil
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

// Transforms utils.Duration type value to time duration value
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}
