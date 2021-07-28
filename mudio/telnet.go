package mudio

import (
	"bufio"
	"io"
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
	reader       io.Reader
	writer       io.Writer
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
	// TODO: Implement me
	return -1, 0
}

func findTelnetCommandRest(oldBuf []byte, contBuf []byte) int {
	// TODO: Implement me
	return -1
}

func isEscapedIAC(buf []byte) bool {
	return len(buf) == 2 && buf[0] == IAC && buf[1] == IAC
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
	datalen := 0

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
			possibleCommand := append(tconn.telnetBuffer, pbuf[0:rest_len]...)
			if isEscapedIAC(possibleCommand) {
				// It's an escaped IAC character, so let's just push IAC (255) onto the data buffer
				// and pretend this never happened!
				copyData(b, datalen, possibleCommand, 0, 1)
				datalen += 1
				curr_input += 1
			} else {
				// Nope! It was a command!
				tconn.observer.CommandReceived(possibleCommand)
			}
			tconn.telnetBuffer = tconn.telnetBuffer[:0] // Clear slice (retain memory)
			curr_input += rest_len
		} else { // rest_len is negative, which means it got some data, but not all! command is still incomplete
			tconn.telnetBuffer = append(tconn.telnetBuffer, pbuf[0:-rest_len]...)
			curr_input += -rest_len
		}
	}

	// Now look for the "real" telnet commands (and write data to read buffer)
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
			possibleCommand := pbuf[cmd_index : cmd_index+cmd_len]
			if isEscapedIAC(possibleCommand) {
				// It's an escaped IAC character so let's just push an IAC character to the data buffer
				copyData(b, datalen, possibleCommand, 0, 1)
				datalen += 1
			} else {
				tconn.observer.CommandReceived(possibleCommand)
			}

			// Advance current input
			curr_input = cmd_index + cmd_len
		}
	}

	return datalen, nil
}

func (tconn *TelnetConnection) Write(b []byte) (n int, err error) {
	start := 0 // We try to write as many full slices as possible - we only stop at IAC markers
	// The variable `start` tracks the position that comes just after an IAC marker or start.
	// When we find a marker, we write everything from `start` up and including the IAC marker
	// and then we write the IAC marker again (to escape it). Then we adjust `start` to point
	// after the found IAC marker in a loop until everything has been printed!
	written_total := 0

	var i int
	for i = 0; i < len(b); i++ {
		if b[i] == IAC {
			// Write all data up to and including the IAC marker
			written, err := tconn.writer.Write(b[start : i+1])
			if err != nil {
				return written, err
			}

			written_total += written

			// Write out the IAC marker again to escape it
			written, err = tconn.writer.Write(b[i : i+1])
			if err != nil {
				return written, err
			}

			written_total += written

			// We know the next slice must start right _after_ the IAC marker
			start = i + 1
		}
	}

	// Did we hit the end of the slice but still have something to write out?
	if start < len(b) {
		written, err := tconn.writer.Write(b[start:i])
		if err != nil {
			return written, err
		}

		written_total += written
	}

	return written_total, nil
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
