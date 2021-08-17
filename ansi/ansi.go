package ansi

import (
	"fmt"
	"regexp"
	"strings"
)

//
// ANSI color encoding scheme:
//
// $fg_red$ = foreground red (31)
// $bg_red$ = background red (41)
// $fg_bred$ = foreground bright red (91)
// etc...

var reColorization = regexp.MustCompile(`\$[a-z_]*\$`)

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

type ansiColor struct {
	name    string
	fgValue int
	bgValue int
}

var ansiColors = []ansiColor{
	// Must be sorted on name!
	{"bblack", 90, 100},
	{"bblue", 94, 104},
	{"bcyan", 96, 106},
	{"bgreen", 92, 102},
	{"black", 30, 40},
	{"blue", 34, 44},
	{"bmagenta", 95, 105},
	{"bred", 91, 101},
	{"bwhite", 97, 107},
	{"byellow", 93, 103},
	{"cyan", 36, 46},
	{"green", 32, 42},
	{"magenta", 35, 45},
	{"red", 31, 41},
	{"white", 37, 47},
	{"yellow", 33, 43},
}

func findColor(name string) int {
	startIndex := 0
	endIndex := len(ansiColors) - 1
	midIndex := len(ansiColors) / 2

	for startIndex <= endIndex {
		if ansiColors[midIndex].name == name {
			return midIndex
		}

		if ansiColors[midIndex].name > name {
			endIndex = midIndex - 1
			midIndex = (startIndex + endIndex) / 2
			continue
		}

		startIndex = midIndex + 1
		midIndex = (startIndex + endIndex) / 2
	}

	return -1
}

func transformFunc(f string) string {
	if len(f) < 5 {
		// Can't be a color
		return ""
	}

	fgOrBg := f[1:3]
	colorName := f[4 : len(f)-1]

	i := findColor(colorName)

	if i < 0 {
		return ""
	}

	var ansiIndex int
	if fgOrBg == "fg" {
		ansiIndex = ansiColors[i].fgValue
	} else {
		ansiIndex = ansiColors[i].bgValue
	}

	return fmt.Sprintf("\x1b[%vm", ansiIndex)
}

func Escape(text string) string {
	return strings.ReplaceAll(text, "$", "$$")
}
