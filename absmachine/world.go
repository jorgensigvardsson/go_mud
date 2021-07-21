package absmachine

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
}

type World struct {
	Rooms   []*Room
	Players []*Player
	Mobs    []*Mob
}

type Object struct {
	Name        string
	Description string
	Room        *Room
}
