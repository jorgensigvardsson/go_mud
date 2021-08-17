package io

import (
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/ansi"
	"github.com/jorgensigvardsson/gomud/logging"
	"github.com/jorgensigvardsson/gomud/mudio"
)

func HandleConnections(listener net.Listener, logger logging.Logger, commandChannel chan<- *PlayerInput, listenerErrorChannel chan<- error, connectionsStopChannel <-chan interface{}, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		conn, err := listener.Accept()

		if err != nil {
			listenerErrorChannel <- err
			return
		}

		go handleConnection(conn, logger, commandChannel, connectionsStopChannel, wg)
	}
}

func handleConnection(tcpConnection net.Conn, logger logging.Logger, commandChannel chan<- *PlayerInput, connectionsStopChannel <-chan interface{}, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// TODO: Check if connection is allowed to connect (IP blocks, etc), before wasting too many CPU cycles
	player := absmachine.NewPlayer()
	errorReturnChannel := make(chan error, 1)
	lineInputChannel := make(chan LineInput, 1)
	outputChannel := make(chan *PlayerOutput, 10)
	telnetConnectionObserver := playerTelnetConnectionObserver{logger: logger, player: player}

	// Whip up a TELNET connection (along with an observer)
	connection := NewTelnetConnection(
		tcpConnection,
		&telnetConnectionObserver,
		logger,
	)

	// Kick off a terminal query!
	connection.QueryTerminal()

	wgLineReader := sync.WaitGroup{}
	defer func() {
		connection.Close()
		// Wait for line reader to exit. When we fall out of scope, the line input channel is closed, which
		// may cause the line reader to panic!
		wgLineReader.Wait()
		close(errorReturnChannel)
		close(outputChannel)
		close(lineInputChannel)
	}()

	// The bootstrapping command: Login!
	loginCmd, _ := mudio.NewCommandLogin([]string{})
	commandChannel <- NewCommandPlayerInput(
		loginCmd,
		player,
		errorReturnChannel,
		outputChannel,
	)

	// Kick off line reader
	go readLine(connection, lineInputChannel, &wgLineReader)

	finished := false
	stopped := false
	for !finished && !stopped {
		select {
		case lineInput := <-lineInputChannel:
			if lineInput.err != nil {
				if errors.Is(lineInput.err, net.ErrClosed) {
					logger.Println("Disconnecting client")
				} else {
					logger.Printlnf("Error reading data from player connection: %v", lineInput.err)
				}
				finished = true
			} else {
				commandChannel <- NewTextPlayerInput(
					strings.TrimSpace(lineInput.line),
					player,
					errorReturnChannel,
					outputChannel,
				)
			}
		case output := <-outputChannel:
			if output.text != "" {
				outputString := output.text

				if output.raw {
					// Do nothing - outputString has the raw text!
				} else if !telnetConnectionObserver.isAnsiCapable {
					// No ANSI capabilities? Then strip away the ANSI markup
					outputString = ansi.Strip(outputString)
				} else {
					// We're not outputing raw text, and the client terminal can do ANSI, so encode the ANSI escape codes!
					outputString = ansi.Encode(outputString)
				}
				connection.WriteString(outputString)

				if !output.keepAnsiColorState {
					// Reset color state unless explicitly stated not to!
					connection.WriteString(ansi.Encode("$fg_white$$bg_black$"))
				}
			}

			switch output.echoState {
			case ES_On:
				connection.EchoOn()
			case ES_Off:
				connection.EchoOff()
			}
		case err := <-errorReturnChannel:
			switch err {
			case ErrPlayerQuit:
				// Do nothing in particular (this is kind of expected)
				finished = true
			case ErrTooManyPlayers:
				connection.WriteLine("Too many players connected, please try again later.")
				finished = true
			case ErrTooMuchInput:
				connection.WriteLine("Input limit reached, please back off with commands for a while.")
			default:
				logger.Printlnf("Aborting player connection for %v due to error: %v", player.Name, err.Error())
				finished = true
			}
		case _, isOpen := <-connectionsStopChannel:
			// We've been stopped!
			connection.WriteLine("Shutting down server...")
			stopped = !isOpen
		}
	}

	commandChannel <- NewEventPlayerInput(
		PE_Exited,
		player,
		errorReturnChannel,
		outputChannel,
	)
}

type LineInput struct {
	line string
	err  error
}

func readLine(connection TelnetConnection, lineInputChannel chan<- LineInput, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		line, err := connection.ReadLine()

		if err != net.ErrClosed {
			lineInputChannel <- LineInput{
				line: line,
				err:  err,
			}
		}

		if err != nil {
			return
		}
	}
}

type playerTelnetConnectionObserver struct {
	player        *absmachine.Player
	logger        logging.Logger
	isAnsiCapable bool
}

func (observer *playerTelnetConnectionObserver) CommandReceived(command []byte) {
	if len(command) > 4 && command[0] == IAC && command[1] == SB && command[2] == TERMINAL_TYPE && command[3] == TRANSMIT_BINARY {
		// We got a terminal type! Let's figure out what it is!
		// Grab string from position 4 (after IAC SB TERMINAL-TYPE BINARY) but before IAC SE
		tt := strings.ToLower(string(command[4 : len(command)-2]))

		if strings.Contains(tt, "xterm") || strings.Contains(tt, "ansi") {
			observer.isAnsiCapable = true
		}
	}
}

func (observer *playerTelnetConnectionObserver) InvalidCommand(data []byte) {
	observer.logger.Printlnf("Invalid TELNET command received: %v", data)
}
