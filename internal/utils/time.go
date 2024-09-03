package utils

import (
	"fmt"
	"time"
)

func FmtDuration(d time.Duration) string {
	if 0 > d {
		return "0s"
	}

	hour := d / time.Hour
	minute := (d - (hour * time.Hour)) / time.Minute
	second := (d - (hour * time.Hour) - (minute * time.Minute)) / time.Second

	switch {
	case hour == 0 && minute == 0:
		return fmt.Sprintf("%ds", second)
	case hour == 0:
		return fmt.Sprintf("%dm:%ds", minute, second)
	default:
		return fmt.Sprintf("%dh:%dm:%ds", hour, minute, second)
	}
}
