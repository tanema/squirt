package lang

import (
	"bytes"
	"strings"
)

func Interpolate(in string, fn func(string) (string, error)) (string, error) {
	if !strings.Contains(in, "${") {
		return in, nil
	}
	str := []rune(in)
	var found bool
	var start int
	var buf bytes.Buffer
	parts := []string{}
	for i, rn := range str {
		switch rn {
		case '$':
			if i+1 < len(str) && str[i+1] == '{' {
				found = true
				continue
			}
		case '{':
			if found {
				parts = append(parts, buf.String())
				buf.Reset()
				start = i - 1
				continue
			}
		case '}':
			if start >= 0 {
				evaled, err := fn(buf.String())
				if err != nil {
					return "", err
				}
				parts = append(parts, evaled)
				buf.Reset()
				start = -1
				continue
			}
		}
		buf.WriteRune(rn)
		found = false
	}
	parts = append(parts, buf.String())
	buf.Reset()
	return strings.Join(parts, ""), nil
}
