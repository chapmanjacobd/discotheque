package shellquote

import (
	"regexp"
	"strings"
)

var (
	toEsc      = regexp.MustCompile(`[^\w!%+,\-./:=@^]`)
	simplifyRe = regexp.MustCompile(`(?:'\\''){2,}`)
)

// ShellQuote returns a shell-escaped version of the string
func ShellQuote(s string) string {
	if s == "" {
		return "''"
	}

	if !toEsc.MatchString(s) {
		return s
	}

	y := strings.ReplaceAll(s, "'", "'\\''")
	y = simplifyRe.ReplaceAllStringFunc(y, func(str string) string {
		var inner strings.Builder
		for i := 0; i < len(str)/4; i++ {
			inner.WriteString("'")
		}
		return `'"` + inner.String() + `"'`
	})

	y = "'" + y + "'"
	y = strings.TrimPrefix(y, "''")
	y = strings.TrimSuffix(y, "''")

	return y
}
