package util

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func HumanizeDuration(duration time.Duration) string {
	days := duration.Hours() / 24
	hours := math.Mod(duration.Hours(), 24)
	minutes := math.Mod(duration.Minutes(), 60)
	seconds := math.Mod(duration.Seconds(), 60)
	idays := int64(days)
	ihours := int64(hours)
	iminutes := int64(minutes)
	iseconds := int64(seconds)
	sb := strings.Builder{}
	if idays != 0 {
		sb.WriteString(fmt.Sprintf(" %d days", idays))
	}
	if ihours != 0 {
		sb.WriteString(fmt.Sprintf(" %d hours", ihours))
	}
	if iminutes != 0 {
		sb.WriteString(fmt.Sprintf(" %d minutes", iminutes))
	}
	if iseconds != 0 {
		sb.WriteString(fmt.Sprintf(" %d seconds", iseconds))
	}
	return strings.TrimSpace(sb.String())
}
