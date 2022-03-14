package time

import "time"

func NowUTC() time.Time {
	return time.Now().In(time.UTC)
}
