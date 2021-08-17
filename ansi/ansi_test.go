package ansi

import (
	"fmt"
	"testing"
)

func Test_Encode_EmptyString(t *testing.T) {
	result := Encode("")

	if result != "" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_NoEscapes(t *testing.T) {
	result := Encode("this is just text")

	if result != "this is just text" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_SingleEscapedDollar(t *testing.T) {
	result := Encode("Total price: $$50")

	if result != "Total price: $50" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_ForegroundIndexedAnsi(t *testing.T) {
	result := Encode("This is $fg_red$Red")

	if result != "This is \x1b[31mRed" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_BackgroundIndexedAnsi(t *testing.T) {
	result := Encode("This is $bg_bred$Red")

	if result != "This is \x1b[101mRed" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_ForegroundIndexedAnsiComplex(t *testing.T) {
	result := Encode("This $$is $fg_red$Red$fg_bblue$Bright blue")

	if result != "This $is \x1b[31mRed\x1b[94mBright blue" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_BackgroundIndexedAnsiComplex(t *testing.T) {
	result := Encode("This $$is $bg_bred$Red$bg_bblue$Bright blue")

	if result != "This $is \x1b[101mRed\x1b[104mBright blue" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_UnknownEscapesAreStripped(t *testing.T) {
	result := Encode("$meh$$foo$")

	if result != "" {
		t.Errorf("Unexpected result: %v", result)
	}
}

type nameColorIndex struct {
	name               string
	expectedColorIndex int
}

var nameColorIndices = []nameColorIndex{
	{"fg_black", 30},
	{"fg_red", 31},
	{"fg_green", 32},
	{"fg_yellow", 33},
	{"fg_blue", 34},
	{"fg_magenta", 35},
	{"fg_cyan", 36},
	{"fg_white", 37},

	{"fg_bblack", 90},
	{"fg_bred", 91},
	{"fg_bgreen", 92},
	{"fg_byellow", 93},
	{"fg_bblue", 94},
	{"fg_bmagenta", 95},
	{"fg_bcyan", 96},
	{"fg_bwhite", 97},

	{"bg_black", 40},
	{"bg_red", 41},
	{"bg_green", 42},
	{"bg_yellow", 43},
	{"bg_blue", 44},
	{"bg_magenta", 45},
	{"bg_cyan", 46},
	{"bg_white", 47},

	{"bg_bblack", 100},
	{"bg_bred", 101},
	{"bg_bgreen", 102},
	{"bg_byellow", 103},
	{"bg_bblue", 104},
	{"bg_bmagenta", 105},
	{"bg_bcyan", 106},
	{"bg_bwhite", 107},
}

func Test_Encode_NamesAreMappedToCorrectANSIEscapeCode(t *testing.T) {
	for _, testCase := range nameColorIndices {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				result := Encode(fmt.Sprintf("$%v$", testCase.name))

				// Result should be on the following form:
				//   \x1b[<index>m

				if result != fmt.Sprintf("\x1b[%vm", testCase.expectedColorIndex) {
					t.Errorf("Unexpected result: %v", result)
				}
			},
		)
	}
}

func Test_Encode_UndonesEscape(t *testing.T) {
	tetsCases := []string{
		"$bg_red$",
		"$$",
		"$fg_byellow",
	}

	for _, testCase := range tetsCases {
		t.Run(
			testCase,
			func(t *testing.T) {
				result := Encode(Escape(testCase))
				if result != testCase {
					t.Errorf("Unexpected result: %v", result)
				}
			},
		)
	}
}
