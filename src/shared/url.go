package shared

import (
	"fmt"
	"strings"
)

/*
// Inspired by https://github.com/php/php-src/blob/72d83709d9524945c93012f7bbb222e412df485a/ext/standard/url.c

func UrlEncodeOld(input string, raw bool) string {
	output := ""
	for c := range input {
		if !raw && c == ' ' {
			output = output + "+"
		} else if (c < '0' && c != '-' && c != '.') || (c < 'A' && c > '9') || (c > 'Z' && c < 'a' && c != '_') || (c > 'z' && (!raw || c != '~')) {
			output = output + "%" + fmt.Sprintf("%X", c)
		} else {
			output = output + string(c)
		}
	}
	return output
}
*/

func UrlEncode(input string, raw bool) string {
	var builder strings.Builder

	for _, c := range input {
		if !raw && c == ' ' {
			builder.WriteByte('+')
		} else if shouldEscape(c, raw) {
			// %% escapes %  and then %X is for printing as HEX
			builder.WriteString(fmt.Sprintf("%%%X", c))
		} else {
			builder.WriteRune(c)
		}
	}

	return builder.String()
}

func shouldEscape(c rune, raw bool) bool {
	return !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || (raw && c == '~'))
}
