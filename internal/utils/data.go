package utils

import "strings"

// MatchAnyPrefix checks if the string starts with any of the prefixes in the given string
func MatchAnyPrefix(str string, prefixString string) bool {
	curPrefix := ""
	for _, prefix := range strings.Split(prefixString, "") {
		curPrefix += prefix
		if strings.HasPrefix(str, curPrefix) {
			return true
		}
	}

	return false
}
