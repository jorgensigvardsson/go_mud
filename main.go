package main

import (
	"container/list"
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

type playerInput struct {
	tick       uint64
	player     *absmachine.Player
	input      string
	connection mudio.TelnetConnection
}

func main() {
	// Make sure interpreter/commands is initialized
	currentTick := uint64(0)
	world := absmachine.NewWorld()
	logger := logging.NewConsoleLogger()
	commandQueue := list.New()
	subPrompts := make(map[*absmachine.Player]mudio.CommandSubPrompter)

	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		panic("Failed to open TCP port 5000")
	}

	go handleConnections(listener, world, logger, &currentTick, commandQueue)

	// The game loop!
	for {
		time.Sleep(TICK)
		handleCommands(commandQueue, currentTick, subPrompts)
		atomic.AddUint64(&currentTick, 1)
	}
}

func handleCommands(commandQueue *list.List, currentTick uint64, subPrompts map[*absmachine.Player]mudio.CommandSubPrompter) {
	iterator := commandQueue.Front()

	if iterator == nil {
		return
	}

	playersProcessed := make(map[*absmachine.Player]bool)
	for iterator != nil {
		playerInput := iterator.Value.(*playerInput)

		// Is the input from before the current tick, or on this tick? If so, then consider it!
		if playerInput.tick <= currentTick {

			// Has this player already been processed in this loop?
			_, isProcessedAlready := playersProcessed[playerInput.player]
			if isProcessedAlready {
				// Move it forward in time so that this input is processed in the next tick
				playerInput.tick = currentTick + 1 // No need to atomically

				// Nothing to do, move on!
				iterator = iterator.Next()
			} else {
				var subPrompt mudio.CommandSubPrompter = nil
				var error error
				subPrompt, hasActiveSubprompt := subPrompts[playerInput.player]

				// If there is no input, don't do anything with it
				if playerInput.input != "" {
					// This is what commands and subprompts expect to work with, so let's
					// whip it up so we can serve it further down
					commandContext := mudio.CommandContext{
						Player:     playerInput.player,
						Connection: playerInput.connection,
					}

					if hasActiveSubprompt {
						// We have an active subprompt, so let it execute
						subPrompt, error = subPrompt.Execute(playerInput.input, &commandContext)
						if error != nil {
							playerInput.connection.WriteLine(error.Error())
						}

						// subPrompt may or may not be nil at this point. We're using the error
						// channel to report things like "invalid input", or whatever, but still
						// keeping the subprompt open.
					} else {
						// We don't have a subprompt, so we'll just have to figure out what command
						// the user typed in.
						command, error := mudio.Parse(playerInput.input)
						if error != nil {
							// Fat fingers -> show it to the user!
							playerInput.connection.WriteLine(error.Error())
						} else {
							// We got a command, so let's execute it. It may optionally return a subprompt!
							subPrompt, error = command.Execute(&commandContext)
							if error != nil {
								playerInput.connection.WriteLine(error.Error())
							}
						}
					}

					if commandContext.TerminationRequested {
						// Terminate player
						playerInput.player.State.SetFlag(absmachine.PS_TERMINATING)
						playerInput.connection.Close()
					}
				}

				// Mark as processed already (to make sure no other queued message is touched, even if we didn't process any input!)
				playersProcessed[playerInput.player] = true

				// Unlink this input and move on to next
				tempIterator := iterator.Next()
				commandQueue.Remove(iterator)
				iterator = tempIterator

				if subPrompt == nil {
					// Show default prompt if we don't have a subprompt
					playerInput.connection.WriteStringf("[H:%v] [M:%v] > ", playerInput.player.Health, playerInput.player.Mana)

					// Make sure we don't have any dangling subprompt
					delete(subPrompts, playerInput.player)
				} else {
					// Show subprompt
					playerInput.connection.WriteString(subPrompt.Prompt())

					// Make sure we remember it!
					subPrompts[playerInput.player] = subPrompt
				}
			}
		} else {
			// Nothing to do, check next!
			iterator = iterator.Next()
		}
	}
}

func handleConnections(listener net.Listener, world *absmachine.World, logger logging.Logger, currentTick *uint64, commandQueue *list.List) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			// TODO: Better error handling!
			panic(fmt.Sprintf("Error accepting new TCP connections: %v", err))
		}

		go handleConnection(conn, world, logger, currentTick, commandQueue)
	}
}

func handleConnection(tcpConnection net.Conn, world *absmachine.World, logger logging.Logger, currentTick *uint64, commandQueue *list.List) {
	player, connection, err := handleLogin(tcpConnection, world, logger)
	if err != nil {
		return
	}

	defer absmachine.DestroyPlayer(player)

	commandQueue.PushFront(&playerInput{
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

		commandQueue.PushBack(&playerInput{
			tick:       atomic.LoadUint64(currentTick),
			player:     player,
			input:      strings.TrimSpace(line), // Make sure extraneous whitespaces are removed
			connection: connection,
		})
	}
}

type PlayerTelnetConnectionObserver struct {
	player *absmachine.Player
}

func (observer *PlayerTelnetConnectionObserver) CommandReceived(command []byte) {
	// TODO: Do something with the telnet command!
}

func (observer *PlayerTelnetConnectionObserver) InvalidCommand(data []byte) {
	// TODO: Do something with the invalid telnet command!
}

func handleLogin(connection net.Conn, world *absmachine.World, logger logging.Logger) (*absmachine.Player, mudio.TelnetConnection, error) {
	playerTelnetConnectionObserver := &PlayerTelnetConnectionObserver{}
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
