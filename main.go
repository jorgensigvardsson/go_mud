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

func main() {
	// Make sure interpreter/commands is initialized
	world := absmachine.NewWorld()
	logger := logging.NewConsoleLogger()
	inputQueue := NewInputQueue()
	subPrompts := make(map[*absmachine.Player]mudio.CommandSubPrompter)
	commandChannel := make(chan *PlayerInputOrCommand, 500)

	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		panic("Failed to open TCP port 5000")
	}

	go handleConnections(listener, logger, inputQueue, commandChannel)

	// The game loop!
	for {
		// Measure how long time we spent processing the commands
		handleCommandsT0 := time.Now().UTC()
		handleCommands(inputQueue, subPrompts, world)
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

		// We have completed one tick!
		inputQueue.Tick()
	}
}

func handleCommands(inputQueue *InputQueue, subPrompts map[*absmachine.Player]mudio.CommandSubPrompter, world *absmachine.World) {
	inputQueue.ForEachCurrentTick(
		func(playerInput *PlayerInputOrCommand) {
			// This is what commands and subprompts expect to work with, so let's
			// whip it up so we can serve it further down
			commandContext := mudio.CommandContext{
				World:      world,
				Player:     playerInput.player,
				Connection: playerInput.connection,
			}

			var nextSubprompt mudio.CommandSubPrompter = nil

			// If there is no input, don't do anything with it
			if playerInput.command != nil {
				var error error
				// We got a command, so let's execute it. It may optionally return a subprompt!
				nextSubprompt, error = playerInput.command.Execute(&commandContext)
				if error != nil {
					playerInput.connection.WriteLine(error.Error())
				}
			} else if playerInput.input != "" {
				activeSubprompt, hasActiveSubprompt := subPrompts[playerInput.player]

				if hasActiveSubprompt {
					// We have an active subprompt, so let it execute
					subPrompt, error := activeSubprompt.ExecuteSubprompt(playerInput.input, &commandContext)
					if error != nil {
						playerInput.connection.WriteLine(error.Error())
					}

					// subPrompt may or may not be nil at this point. We're using the error
					// channel to report things like "invalid input", or whatever, but still
					// keeping the subprompt open.
					nextSubprompt = subPrompt
				} else {
					// We don't have a subprompt, so we'll just have to figure out what command
					// the user typed in.
					command, error := mudio.Parse(playerInput.input)
					if error != nil {
						// Fat fingers -> show it to the user!
						playerInput.connection.WriteLine(error.Error())
					} else {
						// We got a command, so let's execute it. It may optionally return a subprompt!
						nextSubprompt, error = command.Execute(&commandContext)
						if error != nil {
							playerInput.connection.WriteLine(error.Error())
						}
					}
				}
			}

			if commandContext.TerminationRequested {
				// Terminate player
				playerInput.player.State.SetFlag(absmachine.PS_TERMINATING)
				playerInput.connection.Close()
				delete(subPrompts, playerInput.player)
			} else {
				if nextSubprompt == nil {
					// Show default prompt if we don't have a subprompt
					playerInput.connection.WriteStringf("[H:%v] [M:%v] > ", playerInput.player.Health, playerInput.player.Mana)

					// Make sure we don't have any dangling subprompt
					delete(subPrompts, playerInput.player)
				} else {
					// Show subprompt
					promptText, error := nextSubprompt.Prompt(&commandContext)
					if error != nil {
						playerInput.connection.WriteLine(error.Error())
					}

					if !commandContext.TerminationRequested {
						playerInput.connection.WriteString(promptText)
						// Make sure we remember it!
						subPrompts[playerInput.player] = nextSubprompt
					} else {
						// Make sure we don't have any dangling subprompt
						playerInput.player.State.SetFlag(absmachine.PS_TERMINATING)
						playerInput.connection.Close()
						delete(subPrompts, playerInput.player)
					}
				}
			}
		},
	)
}

func handleConnections(listener net.Listener, logger logging.Logger, inputQueue *InputQueue, commandChannel chan *PlayerInputOrCommand) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			// TODO: Better error handling!
			panic(fmt.Sprintf("Error accepting new TCP connections: %v", err))
		}

		go handleConnection(conn, logger, inputQueue, commandChannel)
	}
}

func handleConnection(tcpConnection net.Conn, logger logging.Logger, inputQueue *InputQueue, commandChannel chan<- *PlayerInputOrCommand) {
	// TODO: Check if connection is allowed to connect (IP blocks, etc), before wasting too many CPU cycles

	player := absmachine.NewPlayer()
	defer absmachine.DestroyPlayer(player)

	// Whip up a TELNET connection (along with an observer)
	connection := mudio.NewTelnetConnection(
		tcpConnection,
		&PlayerTelnetConnectionObserver{logger: logger, player: player},
		logger,
	)

	// Show message of the day to user
	showMotd(connection)

	// The bootstrapping command: Login!
	commandChannel <- &PlayerInputOrCommand{
		connection: connection,
		player:     player,
		command:    mudio.NewCommandLogin(),
	}

	for {
		// Read input from user (must be done asynchronously, because we should print a new prompt if the player is hit!)
		line, err := connection.ReadLine()
		if err != nil {
			if errors.Is(err, net.ErrClosed) && player.State.HasFlag(absmachine.PS_TERMINATING) {
				logger.WriteLine("Disconnecting client")
			} else {
				// TODO: Terminate connection
				logger.WriteLinef("Error reading data from player connection: %v", err)
			}
			return
		}

		commandChannel <- &PlayerInputOrCommand{
			connection: connection,
			player:     player,
			input:      strings.TrimSpace(line),
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
