package types

const (
	North              = 0
	South              = 1
	East               = 2
	West               = 3
	Up                 = 4
	Down               = 5
	NumberOfDirections = 6
)

type Room struct {
	Description   string
	Players       []*Player
	Mobs          []*Mob
	AdjacentRooms [NumberOfDirections]*Room
}

type Player struct {
	Name        string
	Description string
	Room        *Room
}

type Mob struct {
	Name        string
	Description string
	Room        *Room
}

type World struct {
	Rooms   []*Room
	Players []*Player
	Mobs    []*Mob
}
