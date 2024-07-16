package ptah_agent

import (
	"strings"
	"unicode"
)

func deconsolify(text []byte) []string {
	normalized := strings.Map(func(r rune) rune {
		if r == '\n' {
			return r
		}

		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, string(text))

	split := strings.Split(normalized, "\n")

	result := make([]string, 0, len(split))
	for _, line := range split {
		if line == "" {
			continue
		}

		result = append(result, line)
	}

	return result
}
