package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/jorgensigvardsson/gomud/absmachine"
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
	player := absmachine.NewPlayer(world, connection)
	defer absmachine.DestroyPlayer(player)

	reader := bufio.NewReader(connection)

	for {
		bytes, err := reader.ReadBytes('\n')
		fmt.Println("Handling connection here... (TODO)", bytes, err)

		// TODO: Parse and dispatch commands here. Dispatch to a command queue, and have a timer execute commands...?
		// TODO: reject commands when/if command queue depth is too large to avoid DOS attacks
		time.Sleep(1000 * time.Hour)
	}
}
