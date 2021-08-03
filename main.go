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

func handleConnection(connection net.Conn, world *absmachine.World, logger logging.Logger) {
	playerConnection, err := handleLogin(connection, world, logger)
	if err != nil {
		return
	}

	defer absmachine.DestroyPlayer(playerConnection.player)

	for {
		line, err := playerConnection.connection.ReadLine()

		if err != nil {
			fmt.Println("Error reading data from connection ")
		}
		fmt.Println("Handling connection here... (TODO)", line, err)

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

func handleLogin(connection net.Conn, world *absmachine.World, logger logging.Logger) (*PlayerConnection, error) {
	playerTelnetConnectionObserver := &PlayerTelnetConnectionObserver{}
	telnetConnection := mudio.NewTelnetConnection(connection, playerTelnetConnectionObserver, logger)

	showMotd(telnetConnection)
	player, err := promptLogin(telnetConnection, world)
	if err != nil {
		return nil, err
	}

	playerTelnetConnectionObserver.player = player

	return &PlayerConnection{
		player:     player,
		connection: telnetConnection,
	}, nil
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

	fmt.Println("Got password", password, password[0], password[1], password[2], password[3])
	if err != nil {
		return nil, err
	}

	// TODO: Validate username and password

	player := absmachine.NewPlayer(world)
	player.Name = username
	return player, nil
}

type PlayerConnection struct {
	player     *absmachine.Player
	connection mudio.TelnetConnection
	commands   []*absmachine.Command
}
