package main

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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
	logger := logging.NewTimestampLoggerDecorator(
		logging.NewSynchronizingLoggerDecorator(
			logging.NewConsoleLogger(),
			50,
		),
	)
	inputQueue := NewInputQueue(MAX_USER_LIMIT, MAX_PLAYER_INPUT_QUEUE_LIMIT)
	commandChannel := make(chan *PlayerInput, MAX_USER_LIMIT*MAX_PLAYER_INPUT_QUEUE_LIMIT)
	sigtermChannel := make(chan os.Signal)
	connectionsStopChannel := make(chan interface{})
	listenerErrorChannel := make(chan error, 1)
	workGroup := sync.WaitGroup{}

	defer close(commandChannel)
	defer close(sigtermChannel)
	defer logger.Close()

	logger.Println("Starting up Go MUD on port 5000...")
	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		panic("Failed to open TCP port 5000")
	}

	// Setup SIGTERM handler
	signal.Notify(sigtermChannel, os.Interrupt, syscall.SIGTERM)
	logger.Println("Stop server with Ctrl+C (SIGTERM)")

	// Spin off in a go routine to handle connections
	go handleConnections(listener, logger, commandChannel, listenerErrorChannel, connectionsStopChannel, &workGroup)

	// The game loop!
	run := true
	for run {
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
			case <-sigtermChannel:
				logger.Println("Shutting down...")
				run = false
			case err := <-listenerErrorChannel:
				logger.Printlnf("Accepting TCP connections failed: %v", err.Error())
				run = false
			}

			// Figure out if we need to sleep more!
			timeToSleep = timeToWakeup.Sub(time.Now().UTC())
		}
	}

	// Shut everything down!
	listener.Close()

	// Terminate all connected
	close(connectionsStopChannel)

	// Wait for all go routines to stop
	workGroup.Wait()

	// Now we're no longer accepting new connections, and all existing sessions have been closed

	// TODO: Serialize current state of world!

	logger.Println("Go MUD successfully shut down.")
}

func handleConnections(listener net.Listener, logger logging.Logger, commandChannel chan<- *PlayerInput, listenerErrorChannel chan<- error, connectionsStopChannel <-chan interface{}, wg *sync.WaitGroup) {
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

	// The bootstrapping command: Login!
	commandChannel <- &PlayerInput{
		connection:         connection,
		player:             player,
		command:            mudio.NewCommandLogin([]string{}),
		errorReturnChannel: errorReturnChannel,
	}

	// Kick off line reader
	go readLine(connection, lineInputChannel)

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
				logger.Printlnf("Aborting player connection for %v due to error: %v", player.Name, err.Error())
				finished = true
			}
		case _, isOpen := <-connectionsStopChannel:
			// We've been stopped!
			connection.WriteLine("Shutting down server...")
			stopped = !isOpen
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
	observer.logger.Printlnf("Invalid TELNET command received: %v", data)
}
