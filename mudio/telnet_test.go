package mudio

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/jorgensigvardsson/gomud/logging"
)

/******* findTelnetCommand and findTelnetCommandRest function tests *******/

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

func Test_findTelnetCommand_NoCommandInBuffer(t *testing.T) {
	buffer := []byte{1, 2, 3, 4}
	cmd_index, _, cmd_invalid := findTelnetCommand(buffer, 0)

	if cmd_invalid {
		t.Errorf("%v is deemed as an invalid command when it should be!", buffer)
	} else {
		if cmd_index != -1 {
			t.Errorf("Found a command at %v in %v, when none should have been found!", cmd_index, buffer)
		}
	}
}

func Test_findTelnetCommand_IACOnlyInBuffer(t *testing.T) {
	buffer := []byte{IAC}
	cmd_index, cmd_len, cmd_invalid := findTelnetCommand(buffer, 0)

	if cmd_invalid {
		t.Errorf("%v is deemed as an invalid command when it should be!", buffer)
	} else if cmd_index != 0 {
		t.Errorf("Did not find an IAC at 0 in %v, when it should have been found!", buffer)
	} else if cmd_len != -1 {
		t.Errorf("Partial length -1 should have been reported, but we got %v instead!", cmd_len)
	}
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

func Test_findTelnetCommandRest_SimpleCommand(t *testing.T) {
	commands := []byte{
		NOP,
		DATA_MARK,
		BREAK,
		IP,
		AO,
		AYT,
		EC,
		EL,
		GA,
	}

	for _, command := range commands {
		alreadyReadBuffer := []byte{IAC}
		incomingBuffer := []byte{command}

		testname := fmt.Sprintf("%v + %v", alreadyReadBuffer, incomingBuffer)

		t.Run(testname, func(t *testing.T) {
			len, invalid := findTelnetCommandRest(alreadyReadBuffer, incomingBuffer, logging.NewNullLogger())

			if invalid {
				t.Errorf("%v + %v does not yield a valid TELNET command!", alreadyReadBuffer, incomingBuffer)
			} else if len == 0 {
				t.Errorf("%v + %v does not yield a TELNET command!", alreadyReadBuffer, incomingBuffer)
			} else if len < 0 {
				t.Errorf("%v + %v yields a partial TELNET command when it shouldn't!", alreadyReadBuffer, incomingBuffer)
			} else if len > 1 {
				t.Errorf("%v + %v yields a command of length > 1!", alreadyReadBuffer, incomingBuffer)
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

func Test_findTelnetCommandRest_Option(t *testing.T) {
	optionTypes := []byte{
		WILL, WONT, DO, DONT,
	}

	options := []byte{
		TRANSMIT_BINARY,
		ECHO,
		SUPPRESS_GO_AHEAD,
		STATUS,
		TIMING_MARK,
		NAOCRD,
		NAOHTS,
		NAOHTD,
		NAOVTS,
		NAOVTD,
		NAOLFD,
		EXTEND_ASCII,
		TERMINAL_TYPE,
		NAWS,
		TERMINAL_SPEED,
		TOGGLE_FLOW_CONTROL,
		LINE_MODE,
		AUTH,
	}

	for _, optionType := range optionTypes {
		for _, option := range options {
			alreadyReadBuffer := []byte{IAC, optionType}
			incomingBuffer := []byte{option}

			testname := fmt.Sprintf("%v + %v", alreadyReadBuffer, incomingBuffer)

			t.Run(testname, func(t *testing.T) {
				len, invalid := findTelnetCommandRest(alreadyReadBuffer, incomingBuffer, logging.NewNullLogger())

				if invalid {
					t.Errorf("%v + %v does not yield a valid TELNET command!", alreadyReadBuffer, incomingBuffer)
				} else if len == 0 {
					t.Errorf("%v + %v does not yield a TELNET command!", alreadyReadBuffer, incomingBuffer)
				} else if len < 0 {
					t.Errorf("%v + %v yields a partial TELNET command when it shouldn't!", alreadyReadBuffer, incomingBuffer)
				} else if len != 1 {
					t.Errorf("%v + %v yields a command of length != 1! (%d was returned)", alreadyReadBuffer, incomingBuffer, len)
				}
			})
		}
	}
}

func Test_findTelnetCommandRest_OptionInvalidRest(t *testing.T) {
	alreadyReadBuffer := []byte{IAC}
	incomingBuffer := []byte{166 /* not a valid option type */}

	_, invalid := findTelnetCommandRest(alreadyReadBuffer, incomingBuffer, logging.NewNullLogger())

	if !invalid {
		t.Errorf("%v + %v did yield a valid TELNET command!", alreadyReadBuffer, incomingBuffer)
	}
}

func Test_findTelnetCommandRest_OptionPartial(t *testing.T) {
	optionTypes := []byte{
		WILL, WONT, DO, DONT,
	}

	for _, optionType := range optionTypes {
		alreadyReadBuffer := []byte{IAC}
		incomingBuffer := []byte{optionType}

		testname := fmt.Sprintf("%v + %v", alreadyReadBuffer, incomingBuffer)

		t.Run(testname, func(t *testing.T) {
			len, invalid := findTelnetCommandRest(alreadyReadBuffer, incomingBuffer, logging.NewNullLogger())

			if invalid {
				t.Errorf("%v + %v does not yield a valid TELNET command!", alreadyReadBuffer, incomingBuffer)
			} else if len == 0 {
				t.Errorf("%v + %v does not yield a TELNET command!", alreadyReadBuffer, incomingBuffer)
			} else if len > 0 {
				t.Errorf("%v + %v yields a complete TELNET command when it shouldn't!", alreadyReadBuffer, incomingBuffer)
			} else if len < -1 {
				t.Errorf("%v + %v yields a command of length < -2! (%d was returned)", alreadyReadBuffer, incomingBuffer, len)
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

/******* isEscapedIAC function tests *******/

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

/******* findSubnegotiationEnd function tests *******/

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

/******* copyData function tests *******/

func Test_copyData_Index0_To_Index0(t *testing.T) {
	source := []byte{1, 2, 3, 4}
	destination := []byte{0, 0, 0}

	copyData(destination, 0, source, 0, 2)

	if !bytes.Equal(destination, []byte{1, 2, 0}) {
		t.Errorf("Data did not copy properly. destination was %v", destination)
	}
}

func Test_copyData_Index1_To_Index0(t *testing.T) {
	source := []byte{1, 2, 3, 4}
	destination := []byte{0, 0, 0}

	copyData(destination, 0, source, 1, 2)

	if !bytes.Equal(destination, []byte{2, 3, 0}) {
		t.Errorf("Data did not copy properly. destination was %v", destination)
	}
}

func Test_copyData_Index0_To_Index1(t *testing.T) {
	source := []byte{1, 2, 3, 4}
	destination := []byte{0, 0, 0}

	copyData(destination, 1, source, 0, 2)

	if !bytes.Equal(destination, []byte{0, 1, 2}) {
		t.Errorf("Data did not copy properly. destination was %v", destination)
	}
}

func Test_copyData_Index1_To_Index1(t *testing.T) {
	source := []byte{1, 2, 3, 4}
	destination := []byte{0, 0, 0}

	copyData(destination, 1, source, 1, 2)

	if !bytes.Equal(destination, []byte{0, 2, 3}) {
		t.Errorf("Data did not copy properly. destination was %v", destination)
	}
}

/******* Write interface method tests *******/

type SpyingWriter struct {
	buf     []byte
	errorAt int
	error   error
	pos     int
}

func (writer *SpyingWriter) Write(buf []byte) (n int, err error) {
	if writer.error != nil {
		// Check if we should error out here
		if len(buf)+writer.pos >= writer.errorAt {
			return writer.errorAt - writer.pos, writer.error
		}
	}
	writer.buf = append(writer.buf, buf...)
	writer.pos += len(buf)
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

func Test_TelnetConnection_Write_IntermediateError(t *testing.T) {
	for i := 0; i < 10; i++ {
		spyingWriter := &SpyingWriter{error: errors.New("Simulated error"), errorAt: i}
		telnetConnection := &TelnetConnection{writer: spyingWriter}
		buffer := []byte{1, IAC, 3, 4, 5, 6, 7, 8, 9, 10}

		testname := fmt.Sprintf("Length %v", i)

		t.Run(testname, func(t *testing.T) {
			n, err := telnetConnection.Write(buffer)
			if err != spyingWriter.error {
				t.Errorf("The expected error did not occur, but %v did!", err)
			}

			if n != i {
				t.Errorf("Foo i = %v, n = %v", i, n)
			}
		})
	}
}

/******* Read interface method tests *******/
type FakeReader struct {
	buf     []byte
	pos     int
	error   error
	errorAt int
}

type SpyingTelnetObserver struct {
	commandsSeen        [][]byte
	invalidCommandsSeen [][]byte
}

func (observer *SpyingTelnetObserver) CommandReceived(command []byte) {
	observer.commandsSeen = append(observer.commandsSeen, command)
}

func (observer *SpyingTelnetObserver) InvalidCommand(data []byte) {
	observer.invalidCommandsSeen = append(observer.invalidCommandsSeen, data)
}

func (reader *FakeReader) Read(buf []byte) (n int, err error) {
	if reader.error != nil {
		// Check if we should error out here
		if len(buf)+reader.pos >= reader.errorAt {
			return reader.errorAt - reader.pos, reader.error
		}
	}

	prevReaderPos := reader.pos
	for i := 0; i < len(buf) && reader.pos < len(reader.buf); i++ {
		buf[i] = reader.buf[reader.pos]
		reader.pos++
	}
	return reader.pos - prevReaderPos, nil
}

func Test_Read_ReaderHasErrors(t *testing.T) {
	buffer := []byte{}
	fakeReader := &FakeReader{buf: buffer, error: errors.New("Fake error"), errorAt: 0}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver}
	outBuffer := make([]byte, 1)

	_, err := telnetConnection.Read(outBuffer)

	if err != fakeReader.error {
		t.Errorf("Read should have failed with error %v, but failed with error %v", fakeReader.error, err)
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_Read_SingleEscapedIAC(t *testing.T) {
	buffer := []byte{IAC, IAC}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver}
	outBuffer := make([]byte, 2)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 1 {
		t.Errorf("Read should have returned 1 byte, but returned %v", n)
	}

	if outBuffer[0] != IAC {
		t.Errorf("Read should have returned [255], but returned [%v]", outBuffer[0])
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_Read_CommandWithTrailingData(t *testing.T) {
	buffer := []byte{IAC, NOP, 1, 2, 3}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 3 {
		t.Errorf("Read should have returned 3 bytes, but returned %v", n)
	}

	if !bytes.Equal(outBuffer[:n], []byte{1, 2, 3}) {
		t.Errorf("Read should have returned [1, 2, 3], but returned %v", outBuffer[:n])
	}

	if len(spyingObserver.commandsSeen) != 1 {
		t.Errorf("Observer expected to have seen one command, but saw %v", len(spyingObserver.commandsSeen))
	} else if !bytes.Equal(spyingObserver.commandsSeen[0], []byte{IAC, NOP}) {
		t.Errorf("Observer expected to have seen command [IAC, NOP], but saw %v", len(spyingObserver.commandsSeen[0]))
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_Read_CommandWithLeadingData(t *testing.T) {
	buffer := []byte{1, 2, 3, IAC, NOP}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 3 {
		t.Errorf("Read should have returned 3 bytes, but returned %v", n)
	}

	if !bytes.Equal(outBuffer[:n], []byte{1, 2, 3}) {
		t.Errorf("Read should have returned [1, 2, 3], but returned %v", outBuffer[:n])
	}

	if len(spyingObserver.commandsSeen) != 1 {
		t.Errorf("Observer expected to have seen one command, but saw %v", len(spyingObserver.commandsSeen))
	} else if !bytes.Equal(spyingObserver.commandsSeen[0], []byte{IAC, NOP}) {
		t.Errorf("Observer expected to have seen command [IAC, NOP], but saw %v", len(spyingObserver.commandsSeen[0]))
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_Read_CommandWithTrailingAndLeadingData(t *testing.T) {
	buffer := []byte{1, 2, 3, IAC, NOP, 4, 5, 6}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 6 {
		t.Errorf("Read should have returned 6 bytes, but returned %v", n)
	}

	if !bytes.Equal(outBuffer[:n], []byte{1, 2, 3, 4, 5, 6}) {
		t.Errorf("Read should have returned [1, 2, 3, 4, 5, 6], but returned %v", outBuffer[:n])
	}

	if len(spyingObserver.commandsSeen) != 1 {
		t.Errorf("Observer expected to have seen one command, but saw %v", len(spyingObserver.commandsSeen))
	} else if !bytes.Equal(spyingObserver.commandsSeen[0], []byte{IAC, NOP}) {
		t.Errorf("Observer expected to have seen command [IAC, NOP], but saw %v", len(spyingObserver.commandsSeen[0]))
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_Read_InvalidCommandWithLeadingData(t *testing.T) {
	buffer := []byte{1, 2, 3, IAC, 1}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 3 {
		t.Errorf("Read should have returned 3 bytes, but returned %v", n)
	}

	if !bytes.Equal(outBuffer[:n], []byte{1, 2, 3}) {
		t.Errorf("Read should have returned [1, 2, 3], but returned %v", outBuffer[:n])
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer expected to have seen no command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 1 {
		t.Errorf("Observer not expected to have seen one invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	} else if !bytes.Equal(spyingObserver.invalidCommandsSeen[0], []byte{IAC, 1}) {
		t.Errorf("Observer expected to have seen invalid command [IAC, 1], but saw %v", len(spyingObserver.invalidCommandsSeen[0]))
	}
}

func Test_Read_InvalidCommandWithTrailingData(t *testing.T) {
	buffer := []byte{IAC, 1, 2, 3}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 2 {
		t.Errorf("Read should have returned 2 bytes, but returned %v", n)
	}

	if !bytes.Equal(outBuffer[:n], []byte{2, 3}) {
		// First byte after IAC has been "consumed" as part of the invalid command
		t.Errorf("Read should have returned [2, 3], but returned %v", outBuffer[:n])
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer expected to have seen no command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 1 {
		t.Errorf("Observer not expected to have seen one invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	} else if !bytes.Equal(spyingObserver.invalidCommandsSeen[0], []byte{IAC, 1}) {
		t.Errorf("Observer expected to have seen invalid command [IAC, 1], but saw %v", len(spyingObserver.invalidCommandsSeen[0]))
	}
}

func Test_Read_InvalidCommandWithLeadingAndTrailingData(t *testing.T) {
	buffer := []byte{1, 2, 3, IAC, 1, 2, 3}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 5 {
		t.Errorf("Read should have returned 5 bytes, but returned %v", n)
	}

	if !bytes.Equal(outBuffer[:n], []byte{1, 2, 3, 2, 3}) {
		// First byte after IAC has been "consumed" as part of the invalid command
		t.Errorf("Read should have returned [1, 2, 3, 2, 3], but returned %v", outBuffer[:n])
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer expected to have seen no command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 1 {
		t.Errorf("Observer not expected to have seen one invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	} else if !bytes.Equal(spyingObserver.invalidCommandsSeen[0], []byte{IAC, 1}) {
		t.Errorf("Observer expected to have seen invalid command [IAC, 1], but saw %v", len(spyingObserver.invalidCommandsSeen[0]))
	}
}

func Test_Read_PartialCommandReturnsZeroBytesButNoError(t *testing.T) {
	buffer := []byte{IAC}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 0 {
		t.Errorf("Read should have returned 0 bytes, but returned %v", n)
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer expected to have seen no command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen no invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_Read_PartialCommandIsCompletedWithSubsequentRead(t *testing.T) {
	buffer := []byte{NOP, 1, 2, 3}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver, telnetBuffer: []byte{IAC}}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 3 {
		t.Errorf("Read should have returned 3 bytes, but returned %v", n)
	}

	if !bytes.Equal(outBuffer[:n], []byte{1, 2, 3}) {
		t.Errorf("Read should have returned [1, 2, 3], but returned %v", outBuffer[:n])
	}

	if len(spyingObserver.commandsSeen) != 1 {
		t.Errorf("Observer expected to have seen one command, but saw %v", len(spyingObserver.commandsSeen))
	} else if !bytes.Equal(spyingObserver.commandsSeen[0], []byte{IAC, NOP}) {
		t.Errorf("Observer expected to have seen command [IAC, NOP], but saw %v", len(spyingObserver.commandsSeen[0]))
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen no invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_Read_PartialCommandWasEscapedIAC(t *testing.T) {
	buffer := []byte{IAC, 1, 2, 3}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver, telnetBuffer: []byte{IAC}}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 4 {
		t.Errorf("Read should have returned 4 bytes, but returned %v", n)
	}

	if !bytes.Equal(outBuffer[:n], []byte{IAC, 1, 2, 3}) {
		t.Errorf("Read should have returned [IAC, 1, 2, 3], but returned %v", outBuffer[:n])
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer expected to have seen no command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen no invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_Read_PartialCommandIsCompletedAsInvalidCommandWithSubsequentRead(t *testing.T) {
	buffer := []byte{1, 2, 3}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver, telnetBuffer: []byte{IAC}, logger: logging.NewNullLogger()}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 2 {
		t.Errorf("Read should have returned 2 bytes, but returned %v", n)
	}

	if !bytes.Equal(outBuffer[:n], []byte{2, 3}) {
		// First data byte consumed as "invalid command" with IAC
		t.Errorf("Read should have returned [2, 3], but returned %v", outBuffer[:n])
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer expected to have seen no command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 1 {
		t.Errorf("Observer not expected to have seen one invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	} else if !bytes.Equal(spyingObserver.invalidCommandsSeen[0], []byte{IAC, 1}) {
		t.Errorf("Observer expected to have seen command [IAC, 1], but saw %v", spyingObserver.invalidCommandsSeen[0])
	}
}

func Test_Read_PartialCommandIsStillNotCompletedAfterSubsequentRead(t *testing.T) {
	buffer := []byte{WILL}
	fakeReader := &FakeReader{buf: buffer}
	spyingObserver := &SpyingTelnetObserver{}
	telnetConnection := &TelnetConnection{reader: fakeReader, observer: spyingObserver, telnetBuffer: []byte{IAC}, logger: logging.NewNullLogger()}
	outBuffer := make([]byte, 10)

	n, err := telnetConnection.Read(outBuffer)

	if err != nil {
		t.Errorf("Read should not have failed, but failed with error %v", err)
	}

	if n != 0 {
		t.Errorf("Read should have returned no bytes, but returned %v", n)
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer expected to have seen no command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen no invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}
