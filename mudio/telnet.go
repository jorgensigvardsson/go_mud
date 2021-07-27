package mudio

import "bufio"

const (
	// Telnet commands - see telnet protocol
	IAC  = 255
	DONT = 254
	DO   = 253
	WONT = 252
	WILL = 251

	// Telnet options - see telnet protocol
	ECHO = 1
)

func EchoOn(writer *bufio.Writer) error {
	_, err := writer.Write([]byte{IAC, WONT, ECHO, 0})
	if err == nil {
		err = writer.Flush()
	}
	return err
}

func EchoOff(writer *bufio.Writer) error {
	_, err := writer.Write([]byte{IAC, WILL, ECHO, 0})
	if err == nil {
		err = writer.Flush()
	}
	return err
}
