package io

import (
	"bufio"
	"fmt"
	"net"

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

type TelnetConnectionObserver interface {
	// TODO: Extend this interface
	CommandReceived(command []byte)
	InvalidCommand(data []byte)
}

const (
	STATE_NOTHING        = 0
	STATE_IAC            = 1
	STATE_OPTION_COMMAND = 2
	STATE_SUBNEG         = 3
)

type TelnetConnection interface {
	ReadLine() (line string, err error)
	WriteLine(line string) error
	WriteLinef(line string, args ...interface{}) error
	WriteString(text string) error
	WriteStringf(text string, args ...interface{}) error
	EchoOff() error
	EchoOn() error
	Close() error
}

type implTelnetConnection struct {
	connection net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	observer   TelnetConnectionObserver
	logger     logging.Logger
}

func NewTelnetConnection(connection net.Conn, observer TelnetConnectionObserver, logger logging.Logger) TelnetConnection {
	return &implTelnetConnection{
		connection: connection,
		reader:     bufio.NewReader(connection),
		writer:     bufio.NewWriter(connection),
		observer:   observer,
		logger:     logger,
	}
}

func (tconn *implTelnetConnection) readByte() (b byte, err error) {
	state := STATE_NOTHING
	buf := make([]byte, 0, 3)

	for {
		b, err := tconn.reader.ReadByte()
		if err != nil {
			return b, err
		}

		switch state {
		case STATE_NOTHING:
			switch b {
			case IAC:
				buf = append(buf, b)
				state = STATE_IAC
			default:
				return b, nil // Not in a state, so just return byte
			}
		case STATE_IAC:
			switch {
			case b == IAC:
				return b, nil // IAC + IAC -> IAC! An escaped IAC code
			case b == SB:
				state = STATE_SUBNEG
				buf = append(buf, b)
			case b >= FIRST_COMMAND && b <= LAST_COMMAND:
				state = STATE_NOTHING
				// Let observer know we have a command!
				tconn.observer.CommandReceived(append(buf, b))
				buf = buf[:0]
			case b >= FIRST_OPTION_COMMAND && b <= LAST_OPTION_COMMAND:
				state = STATE_OPTION_COMMAND // Wait for the option
				buf = append(buf, b)
			default:
				// ERROR!
				tconn.observer.InvalidCommand(append(buf, b))
				buf = buf[:0]
				state = STATE_NOTHING
			}
		case STATE_OPTION_COMMAND:
			// Option command is IAC <command> <option> <value>, and b = <value>
			tconn.observer.CommandReceived(append(buf, b))
			buf = buf[:0]
			state = STATE_NOTHING
		case STATE_SUBNEG:
			// Swallow everything until we reach SE (this is probably wrong, we
			// should most likely look for IAC SE really... this may be a source of bugs!)
			switch b {
			case SE:
				tconn.observer.CommandReceived(append(buf, b))
				state = STATE_NOTHING
			default:
				buf = append(buf, b)
			}
		}
	}
}

func (tconn *implTelnetConnection) writeByte(b byte) error {
	if b == IAC {
		err := tconn.writer.WriteByte(IAC)
		if err != nil {
			return err
		}
	}

	return tconn.writer.WriteByte(b)
}

/* net.Conn, io.Reader and io.Writer implementations for TelnetConnection */
func (tconn *implTelnetConnection) ReadLine() (line string, err error) {
	buf := make([]byte, 0, 50)
	done := false

	for !done {
		b, err := tconn.readByte()

		if err != nil {
			return string(buf), err // TODO: Do we need proper UTF-8 parsing?
		}

		if b != '\r' && b != '\n' {
			buf = append(buf, b)
		} else if b == '\r' {
			continue // Don't store \r in string
		} else {
			done = true // Must be \n, so we're done!
		}
	}

	return string(buf), nil
}

func (tconn *implTelnetConnection) WriteLine(line string) error {
	for _, b := range []byte(line) {
		err := tconn.writeByte(b)
		if err != nil {
			return err
		}
	}

	err := tconn.writeByte('\r')
	if err != nil {
		return err
	}

	err = tconn.writeByte('\n')
	if err != nil {
		return err
	}

	return tconn.writer.Flush()
}

func (tconn *implTelnetConnection) WriteLinef(line string, args ...interface{}) error {
	if len(args) == 0 {
		return tconn.WriteLine(line)
	}

	return tconn.WriteLine(fmt.Sprintf(line, args...))
}

func (tconn *implTelnetConnection) WriteString(text string) error {
	for _, b := range []byte(text) {
		err := tconn.writeByte(b)
		if err != nil {
			return err
		}
	}

	return tconn.writer.Flush()
}

func (tconn *implTelnetConnection) WriteStringf(text string, args ...interface{}) error {
	if len(args) == 0 {
		return tconn.WriteString(text)
	}

	return tconn.WriteString(fmt.Sprintf(text, args...))
}

func (tconn *implTelnetConnection) Close() error {
	return tconn.connection.Close()
}

func (tconn *implTelnetConnection) LocalAddr() net.Addr {
	return tconn.connection.LocalAddr()
}

func (tconn *implTelnetConnection) RemoteAddr() net.Addr {
	return tconn.connection.RemoteAddr()
}

func (tconn *implTelnetConnection) EchoOn() error {
	_, err := tconn.writer.Write([]byte{IAC, WONT, ECHO, 0})
	if err == nil {
		err = tconn.writer.Flush()
	}
	return err
}

func (tconn *implTelnetConnection) EchoOff() error {
	_, err := tconn.writer.Write([]byte{IAC, WILL, ECHO, 0})
	if err == nil {
		err = tconn.writer.Flush()
	}
	return err
}
