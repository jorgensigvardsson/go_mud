package absmachine

func NewWorld() *World {
	return &World{}
}

func NewPlayer(world *World) *Player {
	player := &Player{
		State: PS_STANDING,
	}
	world.Lock()
	defer world.Unlock()
	world.Players = append(world.Players, player)
	player.World = world
	return player
}

func NewRoom(world *World) *Room {
	room := &Room{}
	world.Lock()
	defer world.Unlock()
	world.Rooms = append(world.Rooms, room)
	room.World = world
	return room
}

func NewMob(world *World) *Mob {
	mob := &Mob{}
	world.Lock()
	defer world.Unlock()
	world.Mobs = append(world.Mobs, mob)
	mob.World = world
	return mob
}

func NewObject(world *World) *Object {
	object := &Object{}
	world.Lock()
	defer world.Unlock()
	world.Objects = append(world.Objects, object)
	object.World = world
	return object
}

func DestroyPlayer(player *Player) {
	if player.World == nil {
		return
	}

	player.World.Lock()
	defer player.World.Unlock()

	if player.Room != nil {
		removePlayerFromRoom(player.Room, player)
	}

	removePlayerFromWorld(player.World, player)
}

// Adds a set of rooms to a world. None of the rooms may be associated with a world already!
func (world *World) AddRooms(rooms []*Room) *LowLevelOpsError {
	// Check for inconsistencies

	// Are any of the supplied rooms already attached to a world?
	for _, room := range rooms {
		if room.World != nil {
			return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "At least one room is already attached to another world!"}
		}
	}

	// Nope, let's go right ahead and add them!
	world.Rooms = append(world.Rooms, rooms...)
	for _, room := range rooms {
		room.World = world
	}
	return nil
}

// Adds a set of players to a world. None of the players may be associated with a world already!
func (world *World) AddPlayers(players []*Player) *LowLevelOpsError {
	// Check for inconsistencies

	// Are any of the supplied players already attached to a world?
	for _, player := range players {
		if player.World != nil {
			return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "At least one player is already attached to another world!"}
		}
	}

	// Nope, let's go right ahead and add them!
	world.Players = append(world.Players, players...)
	for _, player := range players {
		player.World = world
	}

	return nil
}

// Adds a set of mobs to a world. None of the mobs may be associated with a world already!
func (world *World) AddMobs(mobs []*Mob) *LowLevelOpsError {
	// Check for inconsistencies

	// Are any of the supplied mobs already attached to a world?
	for _, mob := range mobs {
		if mob.World != nil {
			return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "At least one mob is already attached to another world!"}
		}
	}

	// Nope, let's go right ahead and add them!
	world.Mobs = append(world.Mobs, mobs...)
	for _, mob := range mobs {
		mob.World = world
	}
	return nil
}

// Adds a set of objects to a world. None of the objects may be associated with a world already!
func (world *World) AddObjects(objects []*Object) *LowLevelOpsError {
	// Check for inconsistencies

	// Are any of the supplied objects already attached to a world?
	for _, object := range objects {
		if object.World != nil {
			return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "At least one object is already attached to another world!"}
		}
	}

	// Nope, let's go right ahead and add them!
	world.Objects = append(world.Objects, objects...)
	for _, object := range objects {
		object.World = world
	}
	return nil
}

func (room *Room) Connect(otherRoom *Room, direction Direction) *LowLevelOpsError {
	if room.AdjacentRooms[direction] != nil {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Room is already connected to another room in specified direction"}
	}

	room.AdjacentRooms[direction] = otherRoom
	return nil
}

// Puts player in a specific room
func (player *Player) RelocateToRoom(room *Room) *LowLevelOpsError {
	if room == player.Room {
		return nil
	}

	if player.Room != nil {
		err := removePlayerFromRoom(player.Room, player)
		if err != nil {
			return err
		}
	}

	room.Players = append(room.Players, player)
	player.Room = room
	return nil
}

// Puts mob in a specific room
func (mob *Mob) RelocateToRoom(room *Room) *LowLevelOpsError {
	if room == mob.Room {
		return nil
	}

	if mob.Room != nil {
		err := removeMobFromRoom(mob.Room, mob)
		if err != nil {
			return err
		}
	}

	room.Mobs = append(room.Mobs, mob)
	mob.Room = room
	return nil
}

// Puts object in a specific room
func (object *Object) RelocateToRoom(room *Room) *LowLevelOpsError {
	if room == object.Room {
		return nil
	}

	if object.Room != nil {
		err := removeObjectFromRoom(object.Room, object)
		if err != nil {
			return err
		}
	}

	room.Objects = append(room.Objects, object)
	object.Room = room
	return nil
}

func (world *World) GetPlayers() []*Player {
	world.Lock()
	defer world.Unlock()

	playersCopy := make([]*Player, len(world.Players))
	copy(playersCopy, world.Players)
	return playersCopy
}

// Moves the player in a specific direction
func (player *Player) Move(direction Direction) *LowLevelOpsError {
	// Sanity checks!
	if player.Room == nil {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Player has no reference to a room!"}
	}

	if player.Room.World == nil {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Player has no reference to World via its referenced room!"}
	}

	// Does a room exist in the direction of the player?
	if player.Room.AdjacentRooms[direction] == nil {
		return &LowLevelOpsError{errorCode: ErrorInvalidDirection, message: "Player cannot move in specified direction!"}
	}

	return player.RelocateToRoom(player.Room.AdjacentRooms[direction])
}

func indexOfWorldPlayer(world *World, player *Player) int {
	for index, v := range world.Players {
		if player == v {
			return index
		}
	}

	return -1
}

func indexOfRoomPlayer(room *Room, player *Player) int {
	for index, v := range room.Players {
		if player == v {
			return index
		}
	}

	return -1
}

func indexOfRoomMob(room *Room, mob *Mob) int {
	for index, v := range room.Mobs {
		if mob == v {
			return index
		}
	}

	return -1
}

func indexOfRoomObject(room *Room, object *Object) int {
	for index, v := range room.Objects {
		if object == v {
			return index
		}
	}

	return -1
}

func removePlayerFromWorld(world *World, player *Player) *LowLevelOpsError {
	index := indexOfWorldPlayer(world, player)
	if index < 0 {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Player was not in world's list of players!"}
	}

	world.Players = append(world.Players[:index], world.Players[index+1:]...)
	player.World = nil
	return nil
}

func removePlayerFromRoom(room *Room, player *Player) *LowLevelOpsError {
	index := indexOfRoomPlayer(room, player)
	if index < 0 {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Player was not in room's list of players!"}
	}

	room.Players = append(room.Players[:index], room.Players[index+1:]...)
	player.Room = nil
	return nil
}

func removeMobFromRoom(room *Room, mob *Mob) *LowLevelOpsError {
	index := indexOfRoomMob(room, mob)
	if index < 0 {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Mob was not in room's list of mobs!"}
	}

	room.Mobs = append(room.Mobs[:index], room.Mobs[index+1:]...)
	mob.Room = nil
	return nil
}

func removeObjectFromRoom(room *Room, object *Object) *LowLevelOpsError {
	index := indexOfRoomObject(room, object)
	if index < 0 {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Object was not in room's list of mobs!"}
	}

	room.Objects = append(room.Objects[:index], room.Objects[index+1:]...)
	object.Room = nil
	return nil
}

func (world *World) Lock() {
	world.lock.Lock()
}

func (world *World) Unlock() {
	world.lock.Unlock()
}
