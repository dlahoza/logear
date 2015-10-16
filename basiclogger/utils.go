package basiclogger

import "time"

func ConvertTimestamp(timestamp_format, timestamp string) string {
	tf := "2006-01-02T15:04:05.999999Z"
	t, err := time.Parse(timestamp_format, timestamp)
	if err == nil {
		return t.Format(tf)
	}
	return ""
}
