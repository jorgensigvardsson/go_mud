package main

import (
	"fmt"

	t "github.com/jorgensigvardsson/gomud/absmachine"
)

func main() {
	foo := t.Room{Description: "Hello"}
	fmt.Printf("%+v", foo)
}
