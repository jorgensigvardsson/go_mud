package mudio

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

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

/*** readByte tests ***/

func Test_readByte_SingleDataByte(t *testing.T) {
	readBuffer := []byte{1}
	conn := &implTelnetConnection{
		reader: bufio.NewReader(bytes.NewReader(readBuffer)),
	}

	b, err := conn.readByte()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if b != readBuffer[0] {
		t.Errorf("Unexpected data byte: %v", b)
	}
}

func Test_readByte_EscapedIAC(t *testing.T) {
	readBuffer := []byte{IAC, IAC}
	conn := &implTelnetConnection{
		reader: bufio.NewReader(bytes.NewReader(readBuffer)),
	}

	b, err := conn.readByte()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if b != IAC {
		t.Errorf("Unexpected data byte: %v", b)
	}
}

func Test_readByte_IACIsLastByte(t *testing.T) {
	readBuffer := []byte{IAC}
	conn := &implTelnetConnection{
		reader: bufio.NewReader(bytes.NewReader(readBuffer)),
	}

	_, err := conn.readByte()

	if err != io.EOF {
		t.Errorf("Unexpected error: %v", err)
	}
}

func Test_readByte_CommandFirstThenByte(t *testing.T) {
	spyingObserver := SpyingTelnetObserver{}
	readBuffer := []byte{IAC, IP, 1}
	conn := &implTelnetConnection{
		reader:   bufio.NewReader(bytes.NewReader(readBuffer)),
		observer: &spyingObserver,
	}

	b, err := conn.readByte()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if b != 1 {
		t.Errorf("Unexpected data byte: %v", b)
	}

	if len(spyingObserver.commandsSeen) != 1 {
		t.Errorf("Observer expected to have seen one command, but saw %v", len(spyingObserver.commandsSeen))
	} else if !bytes.Equal(spyingObserver.commandsSeen[0], []byte{IAC, IP}) {
		t.Errorf("Observer expected to have seen command [IAC, IP], but saw %v", spyingObserver.commandsSeen[0])
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_readByte_OptionCommandFirstThenByte(t *testing.T) {
	spyingObserver := SpyingTelnetObserver{}
	readBuffer := []byte{IAC, WILL, IP, 1}
	conn := &implTelnetConnection{
		reader:   bufio.NewReader(bytes.NewReader(readBuffer)),
		observer: &spyingObserver,
	}

	b, err := conn.readByte()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if b != 1 {
		t.Errorf("Unexpected data byte: %v", b)
	}

	if len(spyingObserver.commandsSeen) != 1 {
		t.Errorf("Observer expected to have seen one command, but saw %v", len(spyingObserver.commandsSeen))
	} else if !bytes.Equal(spyingObserver.commandsSeen[0], []byte{IAC, WILL, IP}) {
		t.Errorf("Observer expected to have seen command [IAC, WILL, IP], but saw %v", spyingObserver.commandsSeen[0])
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_readByte_SubnegotiationCommandFirstThenByte(t *testing.T) {
	spyingObserver := SpyingTelnetObserver{}
	readBuffer := []byte{IAC, SB, TERMINAL_TYPE, 'V', 'T', '1', '0', '0', IAC, SE, 1}
	conn := &implTelnetConnection{
		reader:   bufio.NewReader(bytes.NewReader(readBuffer)),
		observer: &spyingObserver,
	}

	b, err := conn.readByte()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if b != 1 {
		t.Errorf("Unexpected data byte: %v", b)
	}

	if len(spyingObserver.commandsSeen) != 1 {
		t.Errorf("Observer expected to have seen one command, but saw %v", len(spyingObserver.commandsSeen))
	} else if !bytes.Equal(spyingObserver.commandsSeen[0], []byte{IAC, SB, TERMINAL_TYPE, 'V', 'T', '1', '0', '0', IAC, SE}) {
		t.Errorf("Observer expected to have seen command [IAC, SB, TERMINAL_TYPE, 'V', 'T', '1', '0', '0', IAC, SE], but saw %v", spyingObserver.commandsSeen[0])
	}

	if len(spyingObserver.invalidCommandsSeen) != 0 {
		t.Errorf("Observer not expected to have seen any invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	}
}

func Test_readByte_InvalidCommandThenDataByte(t *testing.T) {
	spyingObserver := SpyingTelnetObserver{}
	readBuffer := []byte{IAC, 1, 1}
	conn := &implTelnetConnection{
		reader:   bufio.NewReader(bytes.NewReader(readBuffer)),
		observer: &spyingObserver,
	}

	b, err := conn.readByte()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if b != 1 {
		t.Errorf("Unexpected data byte: %v", b)
	}

	if len(spyingObserver.commandsSeen) != 0 {
		t.Errorf("Observer expected to have seen one command, but saw %v", len(spyingObserver.commandsSeen))
	}

	if len(spyingObserver.invalidCommandsSeen) != 1 {
		t.Errorf("Observer not expected to have seen any invalid command, but saw %v", len(spyingObserver.invalidCommandsSeen))
	} else if !bytes.Equal(spyingObserver.invalidCommandsSeen[0], []byte{IAC, 1}) {
		t.Errorf("Observer expected to have seen invalid command [IAC, 1], but saw %v", spyingObserver.invalidCommandsSeen[0])
	}
}

/*** writeByte tests ***/
func Test_writeByte_OneDataByte(t *testing.T) {
	writeBuffer := bytes.NewBuffer([]byte{})
	conn := &implTelnetConnection{
		writer: bufio.NewWriter(writeBuffer),
	}

	err := conn.writeByte(1)
	conn.writer.Flush() // Force write to the underlying buffer

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !bytes.Equal([]byte{1}, writeBuffer.Bytes()) {
		t.Errorf("Unexpected write buffer %v", writeBuffer.Bytes())
	}
}

func Test_writeByte_IACData(t *testing.T) {
	writeBuffer := bytes.NewBuffer([]byte{})
	conn := &implTelnetConnection{
		writer: bufio.NewWriter(writeBuffer),
	}

	err := conn.writeByte(IAC)
	conn.writer.Flush() // Force write to the underlying buffer

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !bytes.Equal([]byte{IAC, IAC}, writeBuffer.Bytes()) {
		t.Errorf("Unexpected write buffer %v", writeBuffer.Bytes())
	}
}

/*** ReadLine tests ***/
func Test_ReadLine_ReadsUntilNewLine(t *testing.T) {
	readBuffer := []byte{'H', 'e', 'l', 'l', 'o', '\r', '\n', 'W', 'o', 'r', 'l', 'd', '\r', '\n'}
	conn := &implTelnetConnection{
		reader: bufio.NewReader(bytes.NewReader(readBuffer)),
	}

	line, err := conn.ReadLine()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if line != "Hello" {
		t.Errorf("Unexpected line: %v", line)
	}
}

func Test_ReadLine_ReadsUntilError(t *testing.T) {
	readBuffer := []byte{'H', 'e', 'l', 'l', 'o'}
	conn := &implTelnetConnection{
		reader: bufio.NewReader(bytes.NewReader(readBuffer)),
	}

	line, err := conn.ReadLine()

	if err != io.EOF {
		t.Errorf("Unexpected error: %v", err)
	}

	if line != "Hello" {
		t.Errorf("Unexpected line: %v", line)
	}
}

/*** WriteLine tests ***/
func Test_WriteLine_AppendsCrAndLf(t *testing.T) {
	writeBuffer := bytes.NewBuffer([]byte{})
	conn := &implTelnetConnection{
		writer: bufio.NewWriter(writeBuffer),
	}

	err := conn.WriteLine("Hello")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !bytes.Equal([]byte{'H', 'e', 'l', 'l', 'o', '\r', '\n'}, writeBuffer.Bytes()) {
		t.Errorf("Unexpected write buffer %v", writeBuffer.Bytes())
	}
}

/*** WriteLine tests ***/
func Test_WriteString(t *testing.T) {
	writeBuffer := bytes.NewBuffer([]byte{})
	conn := &implTelnetConnection{
		writer: bufio.NewWriter(writeBuffer),
	}

	err := conn.WriteString("Hello")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !bytes.Equal([]byte{'H', 'e', 'l', 'l', 'o'}, writeBuffer.Bytes()) {
		t.Errorf("Unexpected write buffer %v", writeBuffer.Bytes())
	}
}

/*** EchoOn tests ***/
func Test_EchoOn(t *testing.T) {
	writeBuffer := bytes.NewBuffer([]byte{})
	conn := &implTelnetConnection{
		writer: bufio.NewWriter(writeBuffer),
	}

	err := conn.EchoOn()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !bytes.Equal([]byte{IAC, WONT, ECHO, 0}, writeBuffer.Bytes()) {
		t.Errorf("Unexpected write buffer %v", writeBuffer.Bytes())
	}
}

/*** EchoOff tests ***/
func Test_EchoOff(t *testing.T) {
	writeBuffer := bytes.NewBuffer([]byte{})
	conn := &implTelnetConnection{
		writer: bufio.NewWriter(writeBuffer),
	}

	err := conn.EchoOff()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !bytes.Equal([]byte{IAC, WILL, ECHO, 0}, writeBuffer.Bytes()) {
		t.Errorf("Unexpected write buffer %v", writeBuffer.Bytes())
	}
}
