package mudio

import (
	"bufio"
	"io"
	"net"
	"time"

	"github.com/jorgensigvardsson/gomud/logging"
)

const (
	// Telnet commands - see telnet protocol
	FIRST_COMMAND = 240
	SE            = 240
	NOP           = 241
	DATA_MARK     = 242
	BREAK         = 243
	IP            = 244 // interrupt process
	AO            = 245 // abort output
	AYT           = 246 // are you there
	EC            = 247 // Erase character EC
	EL            = 248 // Erase line EL
	GA            = 249 // Go ahead
	SB            = 250 // Subnegotiation start
	LAST_COMMAND  = 250

	// Option negotiation commands
	FIRST_OPTION_COMMAND = 251
	WILL                 = 251
	WONT                 = 252
	DO                   = 253
	DONT                 = 254
	LAST_OPTION_COMMAND  = 254

	// Command escape code
	IAC = 255

	// Telnet options - see telnet protocol
	TRANSMIT_BINARY     = 0
	ECHO                = 1
	SUPPRESS_GO_AHEAD   = 3
	STATUS              = 5
	TIMING_MARK         = 6
	NAOCRD              = 10 // Output carriage return disposition
	NAOHTS              = 11 // Output horizontal tab stops
	NAOHTD              = 12 // Output horizontal tab stop disposition
	NAOFFD              = 13 // Output formfeed disposition
	NAOVTS              = 14 // Output vertical tabstops
	NAOVTD              = 15 // Output vertical tab disposition
	NAOLFD              = 16 // Output Linefeed disposition
	EXTEND_ASCII        = 17
	TERMINAL_TYPE       = 24
	NAWS                = 31 // Negotiate about window size
	TERMINAL_SPEED      = 32
	TOGGLE_FLOW_CONTROL = 33
	LINE_MODE           = 34
	AUTH                = 37
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
	InvalidCommand(data []byte)
}

type TelnetConnection struct {
	connection   net.Conn
	reader       io.Reader
	writer       io.Writer
	telnetBuffer []byte
	observer     TelnetConnectionObserver
	logger       logging.Logger
}

func NewTelnetConnection(connection net.Conn, observer TelnetConnectionObserver, logger logging.Logger) *TelnetConnection {
	return &TelnetConnection{
		connection: connection,
		reader:     bufio.NewReader(connection),
		writer:     bufio.NewWriter(connection),
		observer:   observer,
		logger:     logger,
	}
}

func findSubnegotiationEnd(buf []byte, start int) int {
	for i := start; i < len(buf); i++ {
		if buf[i] == SE {
			return i
		}
	}

	return -1
}

func findTelnetCommand(buf []byte, start int) (index int, nn int, invalid bool) {
	for i := 0; i < len(buf); i++ {
		if buf[i] == IAC {
			// We found the IAC character, so let's scan to the end!
			// General structure of telnet commands:
			// Length 2: IAC IAC - escaped IAC
			// Length 2: IAC CMD - CMD sent
			// Length 3: IAC (WILL|WONT|DO|DONT) OPT - option
			// Length N: IAC SB ... IAC SE - option subnegotiation

			// Let's check if we have a complete command!

			// If we find a valid prefix for a command, but not complete,
			// we return the prefix's length as a negative number!

			length_left := len(buf) - (i + 1)

			if length_left > 0 {
				if buf[i+1] == IAC {
					return i, 2, false // we found an escaped IAC
				}

				if buf[i+1] == SB {
					// Subnegotiation start!
					// Check if it's an IAC SB ... IAC SE sequence
					sub_end := findSubnegotiationEnd(buf, i+1)
					if sub_end < 0 {
						// Did not find the end, it must arrive later!
						return i, -(len(buf) - i), false
					} else {
						// Did find the end, so yay!
						return i, sub_end - i + 1, false
					}
				}

				if buf[i+1] >= FIRST_OPTION_COMMAND && buf[i+1] <= LAST_OPTION_COMMAND {
					if length_left >= 2 {
						return i, 3, false // We found an option command
					} else {
						// It's an incomplete option command, we'll get the rest later!
						return i, -2, false
					}
				}

				if buf[i+1] >= FIRST_COMMAND && buf[i+1] <= LAST_COMMAND {
					// It's a basic command
					return i, 2, false
				}

				// It's an invalid command!
				return i, 2, true
			}

			// Nope, we only got the lonely IAC! The rest will come in the next read
			return i, -1, false
		}
	}

	return -1, 0, false
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func findTelnetCommandRest(oldBuf []byte, contBuf []byte, logger logging.Logger) int {
	newBuf := make([]byte, len(oldBuf)+len(contBuf))
	newBuf = append(append(newBuf, oldBuf...), contBuf...)

	cmd_index, cmd_len, cmd_invalid := findTelnetCommand(newBuf, 0)

	if cmd_invalid {
		// A command was found, but it was invalid
		logger.Printf("Found an invalid TELNET command in subsequent read: %v\r\n", newBuf[0:min(len(newBuf), 20)])
		return 0
	}

	if cmd_index < 0 {
		// No command found
		logger.Printf("No TELNET command in subsequent read: %v\r\n", newBuf[0:min(len(newBuf), 20)])
		return 0
	}

	if cmd_len < 0 {
		// We found a partial, so let's return it
		// Note: the length is negative because it signfies that it was not complete,
		//       so we need to remove the sign!
		// Note: We also used the old buffer as a prefix, so we need to remove the offset length caused by the prefix
		// Note: We need to return a negative value, indicating to the caller that we still haven't found the complete
		//       command sequence
		return -(-cmd_len - len(oldBuf))
	}

	// If we reach here, we must've had a positive length, so let's return it!
	return cmd_len - len(oldBuf) // we found the command with old buf as prefix, so remove the offset length caused by the prefix
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
		rest_len := findTelnetCommandRest(tconn.telnetBuffer, pbuf, tconn.logger)

		if rest_len == 0 {
			// This is odd! This means that we have found the start of a
			// telnet command (but not a complete one) in a previous read
			// but the consecutive read does not contain the rest!
			// Let's skip it and continue on (after notifying observer)!
			tconn.observer.InvalidCommand(tconn.telnetBuffer)
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
		cmd_index, cmd_len, cmd_invalid := findTelnetCommand(pbuf, curr_input)

		if cmd_invalid {
			// the command was invalid! Let's report it, and move on
			tconn.observer.InvalidCommand(pbuf[cmd_index : cmd_index+cmd_len])
			curr_input = cmd_index + cmd_len
		} else if cmd_index < 0 {
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

			if cmd_len < 0 {
				// rest_len cmd_len negative, which means it got some data, but not all! command is still incomplete
				tconn.telnetBuffer = pbuf[0:-cmd_len]

				if -cmd_len+curr_input < len(pbuf) {
					panic("BUG: we've missed data looking for TELNET commands!")
				}
				// We're done!
				break
			} else {
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
	}

	return datalen, nil
}

func (tconn *TelnetConnection) Write(b []byte) (n int, err error) {
	start := 0 // We try to write as many full slices as possible - we only stop at IAC markers
	// The variable `start` tracks the position that comes just after an IAC marker or start.
	// When we find a marker, we write everything from `start` up and including the IAC marker
	// and then we write the IAC marker again (to escape it). Then we adjust `start` to point
	// after the found IAC marker in a loop until everything has been printed!
	totalWritten := 0

	var i int
	for i = 0; i < len(b); i++ {
		if b[i] == IAC {
			// Write all data up to and including the IAC marker
			written, err := tconn.writer.Write(b[start : i+1])
			if err != nil {
				return written, err
			}

			totalWritten += written

			// Write out the IAC marker again to escape it
			written, err = tconn.writer.Write(b[i : i+1])
			if err != nil {
				return written, err
			}

			totalWritten += written

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

		totalWritten += written
	}

	return totalWritten, nil
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
