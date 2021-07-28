package mudio

import (
	"bufio"
	"net"
	"time"
)

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

type TelnetConnectionObserver interface {
	// TODO: Extend this interface
	CommandReceived(command []byte)
	IncompleteCommand(data []byte)
}

type TelnetConnection struct {
	connection   net.Conn
	reader       *bufio.Reader
	writer       *bufio.Writer
	telnetBuffer []byte
	observer     TelnetConnectionObserver
}

func NewTelnetConnection(connection net.Conn, observer TelnetConnectionObserver) *TelnetConnection {
	return &TelnetConnection{
		connection: connection,
		reader:     bufio.NewReader(connection),
		writer:     bufio.NewWriter(connection),
		observer:   observer,
	}
}

func findTelnetCommand(buf []byte, start int) (index int, len int) {

}

func findTelnetCommandRest(oldBuf []byte, contBuf []byte) int {

}

func copyData(dst []byte, dstIndex int, src []byte, srcIndex int, len int) {
	for i := 0; i < len; i++ {
		dst[i+dstIndex] = src[i+srcIndex]
	}
}

/* net.Conn, io.Reader and io.Writer implementations for TelnetConnection */
func (tconn *TelnetConnection) Read(b []byte) (n int, err error) {
	// Create our own buffer for telnet manipulation
	pbuf := make([]byte, len(b), cap(b))

	// Read data from connection
	n, err = tconn.connection.Read(pbuf)
	if n < 1 || err != nil {
		return
	}

	// Now look for telnet commands in the data stream
	// Check if we're reading an unfinished telnet command sequence from
	// the last call to Read()

	curr_input := 0
	if len(tconn.telnetBuffer) > 0 {
		rest_len := findTelnetCommandRest(tconn.telnetBuffer, pbuf)

		if rest_len == 0 {
			// This is odd! This means that we have found the start of a
			// telnet command (but not a complete one) in a previous read
			// but the consecutive read does not contain the rest!
			// Let's skip it and continue on (after notifying observer)!
			tconn.observer.IncompleteCommand(tconn.telnetBuffer)
			tconn.telnetBuffer = tconn.telnetBuffer[:0] // Clear slice (retain memory)
		} else if rest_len > 0 { // This means we found the complete rest of the command!
			// Let observer know we got a command
			tconn.observer.CommandReceived(append(tconn.telnetBuffer, pbuf[0:rest_len]...))
			tconn.telnetBuffer = tconn.telnetBuffer[:0] // Clear slice (retain memory)
			curr_input += rest_len
		} else { // rest_len is negative, which means it got some data, but not all! command is still incomplete
			tconn.telnetBuffer = append(tconn.telnetBuffer, pbuf[0:-rest_len]...)
			curr_input += -rest_len
		}
	}

	// Now look for the "real" telnet commands (and write data to read buffer)
	datalen := 0

	for curr_input < len(pbuf) {
		cmd_index, cmd_len := findTelnetCommand(pbuf, curr_input)

		if cmd_index < 0 {
			// No more telnet command data. Copy data from start up until the end to b
			copyData(b, datalen, pbuf, curr_input, len(pbuf)-curr_input)
			datalen += len(pbuf) - curr_input
			curr_input += len(pbuf) - curr_input
		} else {
			// A command was found!
			// Separate the data found in front of the command into b (if any)
			if curr_input < cmd_index {
				copyData(b, datalen, pbuf, curr_input, cmd_index-curr_input)
				datalen += cmd_index - curr_input
			}

			// Let observer know we got a command
			tconn.observer.CommandReceived(pbuf[cmd_index : cmd_index+cmd_len])

			// Advance current input
			curr_input = cmd_index + cmd_len
		}
	}

	return datalen, nil
}

func (tconn *TelnetConnection) Write(b []byte) (n int, err error) {
	return tconn.writer.Write(b)
}

func (tconn *TelnetConnection) Close() error {
	return tconn.connection.Close()
}

func (tconn *TelnetConnection) LocalAddr() net.Addr {
	return tconn.connection.LocalAddr()
}

func (tconn *TelnetConnection) RemoteAddr() net.Addr {
	return tconn.connection.RemoteAddr()
}

func (tconn *TelnetConnection) SetDeadline(t time.Time) error {
	return tconn.connection.SetDeadline(t)
}

func (tconn *TelnetConnection) SetReadDeadline(t time.Time) error {
	return tconn.connection.SetReadDeadline(t)
}

func (tconn *TelnetConnection) SetWriteDeadline(t time.Time) error {
	return tconn.connection.SetWriteDeadline(t)
}
