package generator

import (
	"strings"
	"unicode"
)

func ToPascalCase(s string) string {
	s = strings.TrimSpace(s)
	capitalize := true
	var n string
	for _, r := range s {
		if unicode.IsNumber(r) {
			n += string(r)
			capitalize = true
		} else if unicode.IsLetter(r) && unicode.IsUpper(r) {
			n += string(r)
			capitalize = false
		} else if unicode.IsLetter(r) && unicode.IsLower(r) {
			if capitalize {
				n += strings.ToUpper(string(r))
				capitalize = false
			} else {
				n += string(r)
			}
		} else {
			capitalize = true
		}
	}
	return n
}
