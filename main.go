package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/logging"
	"github.com/jorgensigvardsson/gomud/mudio"
)

const TICK = 100 * time.Millisecond
const MAX_USER_LIMIT = 100
const MAX_PLAYER_INPUT_QUEUE_LIMIT = 20

func main() {
	// Make sure interpreter/commands is initialized
	world := absmachine.NewWorld()
	logger := logging.NewConsoleLogger()
	inputQueue := NewInputQueue(MAX_USER_LIMIT, MAX_PLAYER_INPUT_QUEUE_LIMIT)
	commandChannel := make(chan *PlayerInput, MAX_USER_LIMIT*MAX_PLAYER_INPUT_QUEUE_LIMIT)
	defer close(commandChannel)

	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		panic("Failed to open TCP port 5000")
	}

	// Spin off in a go routine to handle connections
	go handleConnections(listener, logger, commandChannel)

	// The game loop!
	for {
		// Measure how long time we spent processing the commands
		handleCommandsT0 := time.Now().UTC()
		inputQueue.Execute(world)
		handleCommandsT1 := time.Now().UTC()

		// Remove the delta from the TICK length
		timeToSleep := TICK - handleCommandsT0.Sub(handleCommandsT1)
		timeToWakeup := handleCommandsT1.Add(timeToSleep)

		// Pump input from command channel onto input queue
		for timeToSleep > 0 {
			select {
			case inputOrCommand := <-commandChannel:
				inputQueue.Append(inputOrCommand)
			case <-time.After(timeToSleep):
				// Do nothing on purpose!
			}

			// Figure out if we need to sleep more!
			timeToSleep = timeToWakeup.Sub(time.Now().UTC())
		}
	}
}

func handleConnections(listener net.Listener, logger logging.Logger, commandChannel chan *PlayerInput) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			// TODO: Better error handling!
			panic(fmt.Sprintf("Error accepting new TCP connections: %v", err))
		}

		go handleConnection(conn, logger, commandChannel)
	}
}

func handleConnection(tcpConnection net.Conn, logger logging.Logger, commandChannel chan<- *PlayerInput) {
	// TODO: Check if connection is allowed to connect (IP blocks, etc), before wasting too many CPU cycles
	player := absmachine.NewPlayer()
	errorReturnChannel := make(chan error, 1)
	defer close(errorReturnChannel)

	lineInputChannel := make(chan LineInput, 1)
	defer close(lineInputChannel)

	// Whip up a TELNET connection (along with an observer)
	connection := mudio.NewTelnetConnection(
		tcpConnection,
		&PlayerTelnetConnectionObserver{logger: logger, player: player},
		logger,
	)
	defer connection.Close()

	// Show message of the day to user
	showMotd(connection)

	// The bootstrapping command: Login!
	commandChannel <- &PlayerInput{
		connection:         connection,
		player:             player,
		command:            mudio.NewCommandLogin(),
		errorReturnChannel: errorReturnChannel,
	}

	// Kick off line reader
	go readLine(connection, lineInputChannel)

	finished := false
	for !finished {
		select {
		case lineInput := <-lineInputChannel:
			if lineInput.err != nil {
				if errors.Is(lineInput.err, net.ErrClosed) {
					logger.WriteLine("Disconnecting client")
				} else {
					logger.WriteLinef("Error reading data from player connection: %v", lineInput.err)
				}
				finished = true
			} else {
				commandChannel <- &PlayerInput{
					connection:         connection,
					player:             player,
					text:               strings.TrimSpace(lineInput.line),
					errorReturnChannel: errorReturnChannel,
				}
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
				logger.WriteLinef("Aborting player connection for %v due to error: %v", player.Name, err.Error())
				finished = true
			}
		}
	}

	commandChannel <- &PlayerInput{
		connection:         connection,
		player:             player,
		event:              PE_Exited,
		errorReturnChannel: errorReturnChannel,
	}
}

type LineInput struct {
	line string
	err  error
}

func readLine(connection mudio.TelnetConnection, lineInputChannel chan<- LineInput) {
	for {
		line, err := connection.ReadLine()
		lineInputChannel <- LineInput{
			line: line,
			err:  err,
		}

		if err != nil {
			return
		}
	}
}

type PlayerTelnetConnectionObserver struct {
	player *absmachine.Player
	logger logging.Logger
}

func (observer *PlayerTelnetConnectionObserver) CommandReceived(command []byte) {
	// TODO: Do something with the telnet command!
}

func (observer *PlayerTelnetConnectionObserver) InvalidCommand(data []byte) {
	observer.logger.WriteLinef("Invalid TELNET command received: %v", data)
}

func showMotd(telnetConnection mudio.TelnetConnection) {
	// TODO: Read MOTD from file
	telnetConnection.WriteLine("Welcome to GO mud!")
}
