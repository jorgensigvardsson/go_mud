package mudio

import (
	"fmt"
	"testing"
)

/******* Write interface method tests *******/

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

/******* findTelnetCommand function tests *******/

func Test_findTelnetCommand_EmptyInput(t *testing.T) {
	buffer := []byte{}
	index, len, invalid := findTelnetCommand(buffer, 0)

	if index > 0 {
		t.Error("A command was found in empty input!")
	}

	if len > 0 {
		t.Error("A command was found in empty input!")
	}

	if invalid {
		t.Error("An invalid command was found in empty input!")
	}
}

func Test_findTelnetCommand_RegularInput(t *testing.T) {
	buffer := []byte{40, 41, 42, 43}
	index, len, invalid := findTelnetCommand(buffer, 0)

	if index > 0 {
		t.Error("A command was found in regular input!")
	}

	if len > 0 {
		t.Error("A command was found in regular input!")
	}

	if invalid {
		t.Error("An invalid command was found in regular input!")
	}
}

func testIAC_findTelnetCommand(t *testing.T, buffer []byte, expectedIndex int) {
	index, len, invalid := findTelnetCommand(buffer, 0)

	if index != expectedIndex {
		t.Errorf("An IAC escape was not found in the correct position! Expected position: %v, Received position: %v", expectedIndex, index)
	}

	if len != 2 {
		t.Errorf("IAC escape command was not of length 2!")
	}

	if invalid {
		t.Error("IAC escape command was deemed invalid!")
	}
}

func Test_findTelnetCommand_EscapedIAC_StartOfInput(t *testing.T) {
	testIAC_findTelnetCommand(
		t,
		[]byte{IAC, IAC, 41, 42, 43},
		0,
	)
}

func Test_findTelnetCommand_EscapedIAC_MiddleOfInput(t *testing.T) {
	testIAC_findTelnetCommand(
		t,
		[]byte{40, 40, IAC, IAC, 41, 42, 43},
		2,
	)
}

func Test_findTelnetCommand_EscapedIAC_EndOfInput(t *testing.T) {
	testIAC_findTelnetCommand(
		t,
		[]byte{41, 42, 43, IAC, IAC},
		3,
	)
}

type command_test struct {
	buffer    []byte
	cmd_index int
	cmd_len   int
	start_pos int
}

func Test_findTelnetCommand_SimpleCommand(t *testing.T) {
	mk_test_data := func(command byte) []command_test {
		return []command_test{
			{buffer: []byte{41, command, 41, 41, 41}, cmd_index: -1, cmd_len: 0},
			{buffer: []byte{IAC, command, 41, 41, 41}, cmd_index: 0, cmd_len: 2},
			{buffer: []byte{41, IAC, command, 41, 41}, cmd_index: 1, cmd_len: 2},
			{buffer: []byte{41, 41, 41, IAC, command}, cmd_index: 3, cmd_len: 2},
			{buffer: []byte{41, IAC, command, 41, 41}, cmd_index: 1, cmd_len: 2, start_pos: 1},
			{buffer: []byte{41, 41, 41, IAC, command}, cmd_index: 3, cmd_len: 2, start_pos: 1},
		}
	}

	concat := func(args ...[]command_test) []command_test {
		result := []command_test{}

		for _, v := range args {
			result = append(result, v...)
		}

		return result
	}

	command_tests := concat(
		mk_test_data(NOP),
		mk_test_data(DATA_MARK),
		mk_test_data(BREAK),
		mk_test_data(IP),
		mk_test_data(AO),
		mk_test_data(AYT),
		mk_test_data(EC),
		mk_test_data(EL),
		mk_test_data(GA),
	)

	for _, command_test := range command_tests {
		testname := fmt.Sprintf("%v[%v:%v]", command_test.buffer, command_test.cmd_index, command_test.cmd_index+command_test.cmd_len)

		t.Run(testname, func(t *testing.T) {
			cmd_index, cmd_len, cmd_invalid := findTelnetCommand(command_test.buffer, command_test.start_pos)

			if cmd_invalid {
				t.Errorf("%v is deemed invalid when it should not be!", command_test.buffer)
			} else {
				if command_test.cmd_index != cmd_index {
					t.Errorf("Did not find a command at %v, but at %v instead", command_test.cmd_index, cmd_index)
				}

				if command_test.cmd_len != cmd_len {
					t.Errorf("Did not find a command with length %v, but length %v instead", command_test.cmd_len, cmd_len)
				}
			}
		})
	}
}

func Test_findTelnetCommand_Option(t *testing.T) {
	mk_test_data := func(option byte) []command_test {
		return []command_test{
			{buffer: []byte{IAC, WILL, option, 41, 41, 41}, cmd_index: 0, cmd_len: 3},
			{buffer: []byte{41, IAC, WILL, option, 41, 41}, cmd_index: 1, cmd_len: 3},
			{buffer: []byte{41, 41, 41, IAC, WILL, option}, cmd_index: 3, cmd_len: 3},
			{buffer: []byte{41, IAC, WILL, option, 41, 41}, cmd_index: 1, cmd_len: 3, start_pos: 1},
			{buffer: []byte{41, 41, 41, IAC, WILL, option}, cmd_index: 3, cmd_len: 3, start_pos: 1},

			{buffer: []byte{IAC, WONT, option, 41, 41, 41}, cmd_index: 0, cmd_len: 3},
			{buffer: []byte{41, IAC, WONT, option, 41, 41}, cmd_index: 1, cmd_len: 3},
			{buffer: []byte{41, 41, 41, IAC, WONT, option}, cmd_index: 3, cmd_len: 3},
			{buffer: []byte{41, IAC, WONT, option, 41, 41}, cmd_index: 1, cmd_len: 3, start_pos: 1},
			{buffer: []byte{41, 41, 41, IAC, WONT, option}, cmd_index: 3, cmd_len: 3, start_pos: 1},

			{buffer: []byte{IAC, DO, option, 41, 41, 41}, cmd_index: 0, cmd_len: 3},
			{buffer: []byte{41, IAC, DO, option, 41, 41}, cmd_index: 1, cmd_len: 3},
			{buffer: []byte{41, 41, 41, IAC, DO, option}, cmd_index: 3, cmd_len: 3},
			{buffer: []byte{41, IAC, DO, option, 41, 41}, cmd_index: 1, cmd_len: 3, start_pos: 1},
			{buffer: []byte{41, 41, 41, IAC, DO, option}, cmd_index: 3, cmd_len: 3, start_pos: 1},

			{buffer: []byte{IAC, DONT, option, 41, 41, 41}, cmd_index: 0, cmd_len: 3},
			{buffer: []byte{41, IAC, DONT, option, 41, 41}, cmd_index: 1, cmd_len: 3},
			{buffer: []byte{41, 41, 41, IAC, DONT, option}, cmd_index: 3, cmd_len: 3},
			{buffer: []byte{41, IAC, DONT, option, 41, 41}, cmd_index: 1, cmd_len: 3, start_pos: 1},
			{buffer: []byte{41, 41, 41, IAC, DONT, option}, cmd_index: 3, cmd_len: 3, start_pos: 1},
		}
	}

	concat := func(args ...[]command_test) []command_test {
		result := []command_test{}

		for _, v := range args {
			result = append(result, v...)
		}

		return result
	}

	command_tests := concat(
		mk_test_data(TRANSMIT_BINARY),
		mk_test_data(ECHO),
		mk_test_data(SUPPRESS_GO_AHEAD),
		mk_test_data(STATUS),
		mk_test_data(TIMING_MARK),
		mk_test_data(NAOCRD),
		mk_test_data(NAOHTS),
		mk_test_data(NAOHTD),
		mk_test_data(NAOVTS),
		mk_test_data(NAOVTD),
		mk_test_data(NAOLFD),
		mk_test_data(EXTEND_ASCII),
		mk_test_data(TERMINAL_TYPE),
		mk_test_data(NAWS),
		mk_test_data(TERMINAL_SPEED),
		mk_test_data(TOGGLE_FLOW_CONTROL),
		mk_test_data(LINE_MODE),
		mk_test_data(AUTH),
	)

	for _, command_test := range command_tests {
		testname := fmt.Sprintf("%v[%v:%v]", command_test.buffer, command_test.cmd_index, command_test.cmd_index+command_test.cmd_len)

		t.Run(testname, func(t *testing.T) {
			cmd_index, cmd_len, cmd_invalid := findTelnetCommand(command_test.buffer, command_test.start_pos)

			if cmd_invalid {
				t.Errorf("%v is deemed invalid when it should not be!", command_test.buffer)
			} else {
				if command_test.cmd_index != cmd_index {
					t.Errorf("Did not find a command at %v, but at %v instead", command_test.cmd_index, cmd_index)
				}

				if command_test.cmd_len != cmd_len {
					t.Errorf("Did not find a command with length %v, but length %v instead", command_test.cmd_len, cmd_len)
				}
			}
		})
	}
}

func Test_findTelnetCommand_SubNegOption(t *testing.T) {
	mk_test_data := func() []command_test {
		return []command_test{
			{buffer: []byte{IAC, SB, 41, 41, 41, IAC, SE}, cmd_index: 0, cmd_len: 7},
			{buffer: []byte{1, 2, 3, IAC, SB, 41, 41, 41, IAC, SE}, cmd_index: 3, cmd_len: 7},
			{buffer: []byte{IAC, SB, 41, 41, 41, IAC, SE, 1, 2, 3}, cmd_index: 0, cmd_len: 7},
			{buffer: []byte{1, 2, 3, IAC, SB, 41, 41, 41, IAC, SE, 1, 2, 3}, cmd_index: 3, cmd_len: 7},
			{buffer: []byte{1, 2, 3, IAC, SB, 41, 41, 41, IAC}, cmd_index: 3, cmd_len: -6},
			{buffer: []byte{1, 2, 3, IAC, SB}, cmd_index: 3, cmd_len: -2},
		}
	}

	command_tests := mk_test_data()

	for _, command_test := range command_tests {
		testname := fmt.Sprintf("%v[%v:%v]", command_test.buffer, command_test.cmd_index, command_test.cmd_index+command_test.cmd_len)

		t.Run(testname, func(t *testing.T) {
			cmd_index, cmd_len, cmd_invalid := findTelnetCommand(command_test.buffer, command_test.start_pos)

			if cmd_invalid {
				t.Errorf("%v is deemed invalid when it should not be!", command_test.buffer)
			} else {
				if command_test.cmd_index != cmd_index {
					t.Errorf("Did not find a command at %v, but at %v instead", command_test.cmd_index, cmd_index)
				}

				if command_test.cmd_len != cmd_len {
					t.Errorf("Did not find a command with length %v, but length %v instead", command_test.cmd_len, cmd_len)
				}
			}
		})
	}
}

type iac_test struct {
	buffer         []byte
	is_escaped_iac bool
}

func Test_isEscapedIAC(t *testing.T) {
	test_buffers := []iac_test{
		{buffer: []byte{1, 2, 3}, is_escaped_iac: false},
		{buffer: []byte{}, is_escaped_iac: false},
		{buffer: []byte{IAC, IAC}, is_escaped_iac: true},
		{buffer: []byte{IAC}, is_escaped_iac: false},
		{buffer: []byte{1, IAC}, is_escaped_iac: false},
		{buffer: []byte{IAC, 1}, is_escaped_iac: false},
	}

	for _, test_buffer := range test_buffers {
		testname := fmt.Sprintf("%v", test_buffer.buffer)

		t.Run(testname, func(t *testing.T) {
			if test_buffer.is_escaped_iac != isEscapedIAC(test_buffer.buffer) {
				t.Errorf("Not an escape IAC!")
			}
		})
	}
}

type subneg_buffer struct {
	buffer          []byte
	expected_se_pos int
}

func Test_findSubnegotiationEnd(t *testing.T) {
	subneg_buffers := []subneg_buffer{
		{buffer: []byte{}, expected_se_pos: -1},
		{buffer: []byte{1}, expected_se_pos: -1},
		{buffer: []byte{1, 2, 3}, expected_se_pos: -1},
		{buffer: []byte{SE}, expected_se_pos: 0},
		{buffer: []byte{SE, 2, 3}, expected_se_pos: 0},
		{buffer: []byte{1, SE, 3}, expected_se_pos: 1},
		{buffer: []byte{1, 2, SE}, expected_se_pos: 2},
	}

	for _, subneg_buffer := range subneg_buffers {
		testname := fmt.Sprintf("%v", subneg_buffer.buffer)

		t.Run(testname, func(t *testing.T) {
			actual_se_pos := findSubnegotiationEnd(subneg_buffer.buffer, 0)
			if subneg_buffer.expected_se_pos != actual_se_pos {
				if subneg_buffer.expected_se_pos < 0 && actual_se_pos >= 0 {
					t.Errorf("SE not supposed to be in buffer, but was found at %v", actual_se_pos)
				} else if subneg_buffer.expected_se_pos >= 0 && actual_se_pos < 0 {
					t.Errorf("SE supposed to be in buffer at %v, but was not found", subneg_buffer.expected_se_pos)
				} else {
					t.Errorf("SE supposed to be in buffer at %v, but was found at %v", subneg_buffer.expected_se_pos, actual_se_pos)
				}
			}
		})
	}
}
