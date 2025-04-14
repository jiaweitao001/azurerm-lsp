package utils

import "strings"

func MatchAnyPrefix(str string, prefixStrings ...string) bool {
	for _, prefix := range prefixStrings {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}

	return false
}
