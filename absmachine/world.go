package absmachine

type Direction int

const (
	DIR_NORTH Direction = iota
	DIR_SOUTH
	DIR_EAST
	DIR_WEST
	DIR_UP
	DIR_DOWN
)

const NUM_DIR = 6

type PlayerState uint32

const (
	PS_STANDING PlayerState = 1 << iota
	PS_LOGGED_IN
)

type Room struct {
	Description   string
	Players       []*Player
	Mobs          []*Mob
	Objects       []*Object
	AdjacentRooms [NUM_DIR]*Room
	World         *World
}

type Player struct {
	Name        string
	Description string
	Room        *Room
	World       *World
	Health      int
	Mana        int
	Level       int
	State       PlayerState
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

func (ps PlayerState) HasFlag(f PlayerState) bool { return f&ps != 0 }
func (ps *PlayerState) SetFlag(f PlayerState)     { *ps |= f }
func (ps *PlayerState) ClearFlag(f PlayerState)   { *ps &= ^f }
func (ps *PlayerState) ToggleFlag(f PlayerState)  { *ps ^= f }
