package ansi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//
// ANSI color encoding scheme:
//
// $fg(#31) = foreground color 31
// $bg(#101) = background color 101
//

var reColorization = regexp.MustCompile(`\$(((?P<fn>[a-z]+)\(#(?P<index>\d{2,3})\))|\$)`)

func Encode(text string) string {
	// Optimization. Without this check, this function is 20 times slower!
	if strings.IndexRune(text, '$') < 0 {
		// Nothing to encode
		return text
	}

	return reColorization.ReplaceAllStringFunc(text, func(s string) string {
		if s == "$$" {
			return "$"
		}

		return transformFunc(s)
	})
}

func Strip(text string) string {
	// Optimization. Without this check, this function is 20 times slower!
	if strings.IndexRune(text, '$') < 0 {
		// Nothing to encode
		return text
	}

	return reColorization.ReplaceAllLiteralString(text, "")
}

func transformFunc(f string) string {
	matches := reColorization.FindStringSubmatch(f)
	funcName := matches[3]
	if index, err := strconv.Atoi(matches[4]); err != nil {
		// Not sure what the index was!?
		return ""
	} else {
		var ansiIndex int
		switch {
		case funcName == "fg" && (index >= 30 && index <= 37 || index >= 90 && index <= 97),
			funcName == "bg" && (index >= 40 && index <= 47 || index >= 100 && index <= 107):
			ansiIndex = index
		default:
			ansiIndex = -1
		}
		if ansiIndex < 0 {
			return ""
		}
		return fmt.Sprintf("\x1b[%vm", ansiIndex)
	}
}

func Escape(text string) string {
	return strings.ReplaceAll(text, "$", "$$")
}
