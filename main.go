package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		panic("Failed to open TCP port 5000")
	}

	for {
		go handleConnections(listener)
	}
}

func handleConnections(listener net.Listener) error {
	for {
		conn, err := listener.Accept()

		if err != nil {
			return err
		}

		fmt.Println("TCP accepted", conn)
	}
	return nil
}
