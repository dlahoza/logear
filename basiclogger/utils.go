package basiclogger

import "time"

const TIMEFORMAT = "2006-01-02T15:04:05.999999"

func ConvertTimestamp(timestamp_format, timestamp string) string {
	t, err := time.Parse(timestamp_format, timestamp)
	if err == nil {
		return t.Format(TIMEFORMAT)
	}
	return ""
}

func GString(n string, m map[string]interface{}) string {
	if raw, ok := m[n]; ok {
		if v, ok := raw.(string); ok {
			return v
		}
	}
	return ""
}

func GInt(n string, m map[string]interface{}) int {
	if raw, ok := m[n]; ok {
		if v, ok := raw.(int64); ok {
			return int(v)
		}
	}
	return 0
}

func GBool(n string, m map[string]interface{}) bool {
	if raw, ok := m[n]; ok {
		if v, ok := raw.(bool); ok {
			return v
		}
	}
	return false
}

func GArrString(n string, m map[string]interface{}) []string {
	var arr []string
	if raw, ok := m[n]; ok {
		if v, ok := raw.([]interface{}); ok {
			for _, path := range v {
				arr = append(arr, path.(string))
			}
		}
	}
	return arr
}
