package absmachine

import (
	"sync"
)

type Direction int

const (
	DirectionNorth Direction = iota
	DirectionSouth
	DirectionEast
	DirectionWest
	DirectionUp
	DirectionDown
	NumberOfDirections
)

type Room struct {
	Description   string
	Players       []*Player
	Mobs          []*Mob
	Objects       []*Object
	AdjacentRooms [NumberOfDirections]*Room
	World         *World
}

type Player struct {
	Name        string
	Description string
	Room        *Room
	World       *World
}

type Mob struct {
	Name        string
	Description string
	Room        *Room
	World       *World
}

type World struct {
	Rooms   []*Room
	Players []*Player
	Mobs    []*Mob
	Objects []*Object

	lock sync.Mutex
}

type Object struct {
	Name        string
	Description string
	Room        *Room
	World       *World
}

type RelocatableToRoom interface {
	RelocateToRoom(room *Room) *LowLevelOpsError
}

type DirectionMovable interface {
	Move(direction Direction) *LowLevelOpsError
}
