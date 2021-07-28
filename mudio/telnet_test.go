package mudio

import "testing"

type SpyingWriter struct {
	buf []byte
}

func (writer *SpyingWriter) Write(buf []byte) (n int, err error) {
	writer.buf = append(writer.buf, buf...)
	return len(buf), nil
}

func doTestWrite(t *testing.T, input []byte, expectedOutput []byte) {
	spyingWriter := &SpyingWriter{}
	telnetConnection := &TelnetConnection{writer: spyingWriter}

	nn, err := telnetConnection.Write(input)

	if err != nil {
		t.Error("Write returned an error")
	} else {
		if nn != len(expectedOutput) {
			t.Errorf("Write did not return expected length. Expected %v but got %v (output = %v)", len(expectedOutput), nn, expectedOutput)
		}
	}
}

func Test_TelnetConnection_Write_EmptySlice(t *testing.T) {
	doTestWrite(t, []byte{}, []byte{})
}

func Test_TelnetConnection_Write_SliceWoIAC(t *testing.T) {
	doTestWrite(t, []byte{1, 2, 3}, []byte{1, 2, 3})
}

func Test_TelnetConnection_Write_SliceWIACAtStart(t *testing.T) {
	doTestWrite(t, []byte{IAC, 1, 2, 3}, []byte{IAC, IAC, 1, 2, 3})
}

func Test_TelnetConnection_Write_SliceWIACAtEnd(t *testing.T) {
	doTestWrite(t, []byte{1, 2, 3, IAC}, []byte{1, 2, 3, IAC, IAC})
}

func Test_TelnetConnection_Write_SliceWIACInMiddle(t *testing.T) {
	doTestWrite(t, []byte{1, IAC, 2, IAC, 3}, []byte{1, IAC, IAC, 2, IAC, IAC, 3})
}

func Test_TelnetConnection_Write_SliceWConsecutiveIACs(t *testing.T) {
	doTestWrite(t, []byte{1, IAC, IAC, 2}, []byte{1, IAC, IAC, IAC, IAC, 2})
}
