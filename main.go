package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/mudio"
)

func main() {
	world := absmachine.NewWorld()

	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		panic("Failed to open TCP port 5000")
	}

	for {
		go handleConnections(listener, world)
		time.Sleep(1000 * time.Second)
	}
}

func handleConnections(listener net.Listener, world *absmachine.World) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			return // TODO: Handle err
		}

		go handleConnection(conn, world)
	}
}

func handleConnection(connection net.Conn, world *absmachine.World) {
	playerConnection, err := handleLogin(connection, world)
	if err != nil {
		return
	}

	defer absmachine.DestroyPlayer(playerConnection.player)

	for {
		bytes, err := playerConnection.reader.ReadString('\n')

		if err != nil {
			fmt.Println("Error reading data from connection ")
		}
		fmt.Println("Handling connection here... (TODO)", bytes, err)

		// TODO: Parse and dispatch commands here. Dispatch to a command queue, and have a timer execute commands...?
		// TODO: reject commands when/if command queue depth is too large to avoid DOS attacks
		time.Sleep(1000 * time.Hour)
	}
}

func handleLogin(connection net.Conn, world *absmachine.World) (*PlayerConnection, error) {
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)

	showMotd(writer)
	player, err := promptLogin(writer, reader, world)
	if err != nil {
		return nil, err
	}

	return &PlayerConnection{
		player: player,
		reader: reader,
		writer: writer,
	}, nil
}

func showMotd(writer *bufio.Writer) {
	// TODO: Read MOTD from file
	writer.WriteString("Welcome to GO mud!\r\n")
	writer.Flush()
}

func promptLogin(writer *bufio.Writer, reader *bufio.Reader, world *absmachine.World) (*absmachine.Player, error) {
	writer.WriteString("Username: ")
	writer.Flush()

	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	fmt.Println("Got user", username)

	writer.WriteString("Password: ")
	writer.Flush()
	mudio.EchoOff(writer)
	password, err := reader.ReadString('\n')
	mudio.EchoOn(writer)

	// Need to emit a new line, because echo off will eat the new line on the client end
	writer.WriteString("\r\n")
	writer.Flush()

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
	player   *absmachine.Player
	reader   *bufio.Reader
	writer   *bufio.Writer
	commands []*absmachine.Command
}

func rawWriteLine(writer *bufio.Writer, text string) error {
	_, err := writer.WriteString(text)
	if err != nil {
		return err
	}

	_, err = writer.WriteString("\r\n")
	if err != nil {
		return err
	}

	return nil
}

func (playerConnection PlayerConnection) WriteLine(text string) error {
	err := rawWriteLine(playerConnection.writer, text)
	if err == nil {
		err = playerConnection.writer.Flush()
	}
	return err
}

func (playerConnection PlayerConnection) WriteLines(textLines []string) error {
	for _, text := range textLines {
		err := rawWriteLine(playerConnection.writer, text)
		if err != nil {
			return err
		}
	}

	return nil
}
