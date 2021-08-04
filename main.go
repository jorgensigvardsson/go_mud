package main

import (
	"fmt"
	"net"
	"time"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/logging"
	"github.com/jorgensigvardsson/gomud/mudio"
)

func main() {
	// Make sure interpreter/commands is initialized
	mudio.InitializeInterpreter()

	world := absmachine.NewWorld()
	logger := logging.NewConsoleLogger()

	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		panic("Failed to open TCP port 5000")
	}

	for {
		go handleConnections(listener, world, logger)
		time.Sleep(1000 * time.Second)
	}
}

func handleConnections(listener net.Listener, world *absmachine.World, logger logging.Logger) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			return // TODO: Handle err
		}

		go handleConnection(conn, world, logger)
	}
}

func handleConnection(tcpConnection net.Conn, world *absmachine.World, logger logging.Logger) {
	player, connection, err := handleLogin(tcpConnection, world, logger)
	if err != nil {
		return
	}

	context := mudio.CommandContext{
		Player:     player,
		Connection: connection,
	}

	defer absmachine.DestroyPlayer(player)

	var commandSubPrompter mudio.CommandSubPrompter = nil

	for {
		// Present prompt
		if commandSubPrompter != nil {
			connection.WriteString(commandSubPrompter.Prompt())

			// Read input from user
			line, err := connection.ReadLine()
			if err != nil {
				fmt.Println("Error reading data from connection ")
				continue
			}

			nextCommandSubPrompter, promptError := commandSubPrompter.GiveInput(line, &context)
			if promptError != nil {
				connection.WriteLine(promptError.Error())
			} else {
				commandSubPrompter = nextCommandSubPrompter
			}
		} else {
			connection.WriteStringf("[H:%v] [M:%v] > ", player.Health, player.Mana)

			// Read input from user
			line, err := connection.ReadLine()

			if err != nil {
				fmt.Println("Error reading data from connection ")
				continue
			}

			command, parseError := mudio.Parse(line)
			if parseError != nil {
				connection.WriteLine(parseError.Error())
			} else {
				nextCommandSubPrompter, commandError := command.Execute(&context)
				if commandError != nil {
					connection.WriteLine(commandError.Error())
				} else {
					commandSubPrompter = nextCommandSubPrompter
				}
				// TODO: dispatch command to command queue
			}
		}

		// TODO: Parse and dispatch commands here. Dispatch to a command queue, and have a timer execute commands...?
		// TODO: reject commands when/if command queue depth is too large to avoid DOS attacks
		//time.Sleep(1000 * time.Hour)
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
