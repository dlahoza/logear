package basiclogger

import "testing"

func TestConvertTimestamp(t *testing.T) {
	cases := []struct {
		timestamp_format, timestamp, want string
	}{
		{"2006-01-02 15:04:05.999", "2011-10-18 02:24:35.123", "2011-10-18T02:24:35.123"},
		{"2006-01-02 15:04:05.999", "2012-10-18 02:24:35", "2012-10-18T02:24:35"},
		{"2006-01-02 15:56:05.999", "2013-10-18 02:24:35.123", ""},
	}
	for _, c := range cases {
		got := ConvertTimestamp(c.timestamp_format, c.timestamp)
		if got != c.want {
			t.Errorf("ConvertTimestamp(%q, %q) == %q, want %q", c.timestamp_format, c.timestamp, got, c.want)
		}
	}
}

func TestGString(t *testing.T) {
	cases := []struct {
		n    string
		m    map[string]interface{}
		want string
	}{
		{n: "1", m: map[string]interface{}{"1": "a"}, want: "a"},
		{n: "2", m: map[string]interface{}{"1": "a"}, want: ""},
		{n: "3", m: map[string]interface{}{"3": 1}, want: ""},
	}
	for _, c := range cases {
		got := GString(c.n, c.m)
		if got != c.want {
			t.Errorf("GString(%q, %q) == %q, want %q", c.n, c.m, got, c.want)
		}
	}
}

func TestGInt(t *testing.T) {
	cases := []struct {
		n    string
		m    map[string]interface{}
		want int
	}{
		{n: "1", m: map[string]interface{}{"1": int64(1)}, want: 1},
		{n: "2", m: map[string]interface{}{"1": int64(2)}, want: 0},
		{n: "3", m: map[string]interface{}{"3": "a"}, want: 0},
	}
	for _, c := range cases {
		got := GInt(c.n, c.m)
		if got != c.want {
			t.Errorf("GInt(%q, %q) == %q, want %q", c.n, c.m, got, c.want)
		}
	}
}
