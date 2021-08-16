package ansi

import "testing"

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
	result := Encode("This is $fg(#31)Red")

	if result != "This is \x1b[31mRed" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_BackgroundIndexedAnsi(t *testing.T) {
	result := Encode("This is $bg(#101)Red")

	if result != "This is \x1b[101mRed" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_ForegroundIndexedAnsiComplex(t *testing.T) {
	result := Encode("This $$is $fg(#31)Red$fg(#94)Bright blue")

	if result != "This $is \x1b[31mRed\x1b[94mBright blue" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_BackgroundIndexedAnsiComplex(t *testing.T) {
	result := Encode("This $$is $bg(#101)Red$bg(#104)Bright blue")

	if result != "This $is \x1b[101mRed\x1b[104mBright blue" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func Test_Encode_UnknownEscapesAreStripped(t *testing.T) {
	result := Encode("$blah(#123)$g(#666)")

	if result != "" {
		t.Errorf("Unexpected result: %v", result)
	}
}

func BenchmarkFoo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode("this is just text")
	}
}
