package shared

import "fmt"

// Inspired by https://github.com/php/php-src/blob/72d83709d9524945c93012f7bbb222e412df485a/ext/standard/url.c

func UrlEncode(input string, raw bool) string {
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
