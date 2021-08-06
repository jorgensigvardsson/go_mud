package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/logging"
	"github.com/jorgensigvardsson/gomud/mudio"
)

const TICK = 100 * time.Millisecond

func main() {
	// Make sure interpreter/commands is initialized
	currentTick := uint64(0)
	world := absmachine.NewWorld()
	logger := logging.NewConsoleLogger()
	inputQueue := NewInputQueue()
	subPrompts := make(map[*absmachine.Player]mudio.CommandSubPrompter)

	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		panic("Failed to open TCP port 5000")
	}

	go handleConnections(listener, world, logger, &currentTick, inputQueue)

	// The game loop!
	for {
		time.Sleep(TICK)
		handleCommands(inputQueue, currentTick, subPrompts)
		atomic.AddUint64(&currentTick, 1)
	}
}

func handleCommands(inputQueue *InputQueue, currentTick uint64, subPrompts map[*absmachine.Player]mudio.CommandSubPrompter) {
	inputQueue.ForEachUntilTick(
		currentTick,
		func(playerInput *PlayerInput) {
			// This is what commands and subprompts expect to work with, so let's
			// whip it up so we can serve it further down
			commandContext := mudio.CommandContext{
				Player:     playerInput.player,
				Connection: playerInput.connection,
			}

			var nextSubprompt mudio.CommandSubPrompter = nil

			// If there is no input, don't do anything with it
			if playerInput.input != "" {
				activeSubprompt, hasActiveSubprompt := subPrompts[playerInput.player]

				if hasActiveSubprompt {
					// We have an active subprompt, so let it execute
					subPrompt, error := activeSubprompt.Execute(playerInput.input, &commandContext)
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
					playerInput.connection.WriteString(nextSubprompt.Prompt())

					// Make sure we remember it!
					subPrompts[playerInput.player] = nextSubprompt
				}
			}
		},
	)
}

func handleConnections(listener net.Listener, world *absmachine.World, logger logging.Logger, currentTick *uint64, inputQueue *InputQueue) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			// TODO: Better error handling!
			panic(fmt.Sprintf("Error accepting new TCP connections: %v", err))
		}

		go handleConnection(conn, world, logger, currentTick, inputQueue)
	}
}

func handleConnection(tcpConnection net.Conn, world *absmachine.World, logger logging.Logger, currentTick *uint64, inputQueue *InputQueue) {
	player, connection, err := handleLogin(tcpConnection, world, logger)
	if err != nil {
		return
	}

	defer absmachine.DestroyPlayer(player)

	// Begin with queuing
	inputQueue.Prepend(&PlayerInput{
		tick:       0, // Now!
		player:     player,
		connection: connection,
		input:      "", // Empty input means "no command", and will force the command queue loop to print a prompt!
	})

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

		inputQueue.Append(&PlayerInput{
			tick:       atomic.LoadUint64(currentTick),
			player:     player,
			input:      strings.TrimSpace(line), // Make sure extraneous whitespaces are removed
			connection: connection,
		})
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

func handleLogin(connection net.Conn, world *absmachine.World, logger logging.Logger) (*absmachine.Player, mudio.TelnetConnection, error) {
	playerTelnetConnectionObserver := &PlayerTelnetConnectionObserver{logger: logger}
	telnetConnection := mudio.NewTelnetConnection(connection, playerTelnetConnectionObserver, logger)

	showMotd(telnetConnection)
	player, err := promptLogin(telnetConnection, world)
	if err != nil {
		return nil, nil, err
	}

	// Connect connection observer with player, in case we need to know the player when we process
	// TELNET commands (could be useful for terminal type negotiation, etc)
	playerTelnetConnectionObserver.player = player
	return player, telnetConnection, nil
}

func showMotd(telnetConnection mudio.TelnetConnection) {
	// TODO: Read MOTD from file
	telnetConnection.WriteLine("Welcome to GO mud!")
}

func promptLogin(telnetConnection mudio.TelnetConnection, world *absmachine.World) (*absmachine.Player, error) {
	// TODO: Extend to allow for new registrations of users
	telnetConnection.WriteString("Username: ")

	username, err := telnetConnection.ReadLine()
	if err != nil {
		return nil, err
	}

	fmt.Println("Got user", username)

	telnetConnection.WriteString("Password: ")
	telnetConnection.EchoOff()
	password, err := telnetConnection.ReadLine()
	telnetConnection.EchoOn()

	// Need to emit a new line, because echo off will eat the new line on the client end
	telnetConnection.WriteLine("")

	fmt.Println("Got password", password)
	if err != nil {
		return nil, err
	}

	// TODO: Validate username and password

	player := absmachine.NewPlayer(world)
	player.Name = username
	return player, nil
}
