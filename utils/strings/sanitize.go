package strings

import (
    "strings"
    "unicode"
)


// Remove non-printable characters from a string
func Sanitize(str string) string {
    return strings.Map(func(r rune) rune {
        if unicode.IsPrint(r) {
            return r
        }
        return -1
    }, str)
}

