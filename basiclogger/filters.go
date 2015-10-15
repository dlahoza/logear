package basiclogger

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type filter struct {
	t        string         //Filter type
	regexp   *regexp.Regexp //For regexp type of filter
	template string         // Template for parsed data
}

var filters map[string]filter

func FilterData(name, data string) (*map[string]interface{}, error) {
	var m map[string]interface{}
	if f, ok := filters[name]; ok {
		switch f.t {
		case "json":
			err := json.Unmarshal([]byte(data), &m)
			return &m, err
		case "regexp":
			if f.regexp != nil && len(f.template) > 0 {
				matches := f.regexp.FindStringSubmatch(data)
				j := f.template
				for i, match := range matches {
					escaped := url.QueryEscape(match)
					j = strings.Replace(j, "$("+strconv.Itoa(i)+")", escaped, -1)
				}
				err := json.Unmarshal([]byte(j), &m)
				return &m, err
			}
		default:
			return nil, errors.New("[" + name + "] unknown filter type \"" + f.t + "\"")
		}
	} else {
		return nil, errors.New("Can't find filter \"" + name + "\"")
	}
	return nil, errors.New("Unknown error")
}
