package main

import (
	"fmt"

	t "github.com/jorgensigvardsson/gomud/types"
)

func main() {
	foo := t.Room{Description: "Hello"}
	fmt.Printf("%+v", foo)
}
