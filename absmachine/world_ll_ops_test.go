package absmachine

import (
	"testing"
)

func isPlayerInRoom(room *Room, player *Player) bool {
	for _, v := range room.Players {
		if v == player {
			return true
		}
	}

	return false
}

func isMobInRoom(room *Room, mob *Mob) bool {
	for _, v := range room.Mobs {
		if v == mob {
			return true
		}
	}

	return false
}

func isObjectInRoom(room *Room, object *Object) bool {
	for _, v := range room.Objects {
		if v == object {
			return true
		}
	}

	return false
}

func Test_Player_RelocateToRoom_NotInPreviousRoom(t *testing.T) {
	// Arrange
	world := NewWorld()
	player := NewPlayer()
	room := NewRoom()
	world.AddRooms([]*Room{room})
	world.AddPlayers([]*Player{player})

	if player.Room != nil {
		t.Error("Player is in a room prior to move!")
	}

	// Act
	err := player.RelocateToRoom(room)
	if err != nil {
		t.Errorf("RelocateToRoom failed: %+v", *err)
	}

	// Assert
	if player.Room != room {
		t.Error("Player is not in the room it was moved to!")
	}

	if !isPlayerInRoom(room, player) {
		t.Error("Player is not in the room's list")
	}
}

func Test_Player_RelocateToRoom_InPreviousRoom(t *testing.T) {
	// Arrange
	world := NewWorld()
	room1 := NewRoom()
	room2 := NewRoom()
	player := NewPlayer()
	world.AddPlayers([]*Player{player})
	world.AddRooms([]*Room{room1, room2})
	player.RelocateToRoom(room1)

	if player.Room != room1 {
		t.Error("Player is not in room1 prior to move!")
	}

	// Act
	err := player.RelocateToRoom(room2)
	if err != nil {
		t.Errorf("RelocateToRoom failed: %+v", *err)
	}

	// Assert
	if player.Room != room2 {
		t.Error("Player is not in room2 after move!")
	}

	if isPlayerInRoom(room1, player) {
		t.Error("Player is still in room1!")
	}

	if !isPlayerInRoom(room2, player) {
		t.Error("Player is not in room2!")
	}
}

func Test_Mob_RelocateToRoom_NotInPreviousRoom(t *testing.T) {
	// Arrange
	world := NewWorld()
	mob := NewMob()
	room := NewRoom()
	world.AddRooms([]*Room{room})
	world.AddMobs([]*Mob{mob})

	if mob.Room != nil {
		t.Error("Mob is in a room prior to move!")
	}

	// Act
	err := mob.RelocateToRoom(room)
	if err != nil {
		t.Errorf("RelocateToRoom failed: %+v", *err)
	}

	// Assert
	if mob.Room != room {
		t.Error("Mob is not in the room it was moved to!")
	}

	if !isMobInRoom(room, mob) {
		t.Error("Mob is not in the room's list")
	}
}

func Test_Mob_RelocateToRoom_InPreviousRoom(t *testing.T) {
	// Arrange
	world := NewWorld()
	room1 := NewRoom()
	room2 := NewRoom()
	mob := NewMob()
	world.AddMobs([]*Mob{mob})
	world.AddRooms([]*Room{room1, room2})
	mob.RelocateToRoom(room1)

	if mob.Room != room1 {
		t.Error("Mob is not in room1 prior to move!")
	}

	// Act
	err := mob.RelocateToRoom(room2)
	if err != nil {
		t.Errorf("RelocateToRoom failed: %+v", *err)
	}

	// Assert
	if mob.Room != room2 {
		t.Error("Mob is not in room2 after move!")
	}

	if isMobInRoom(room1, mob) {
		t.Error("Mob is still in room1!")
	}

	if !isMobInRoom(room2, mob) {
		t.Error("Mob is not in room2!")
	}
}

func Test_Object_RelocateToRoom_NotInPreviousRoom(t *testing.T) {
	// Arrange
	world := NewWorld()
	object := NewObject()
	room := NewRoom()
	world.AddRooms([]*Room{room})
	world.AddObjects([]*Object{object})

	if object.Room != nil {
		t.Error("Object is in a room prior to move!")
	}

	// Act
	err := object.RelocateToRoom(room)
	if err != nil {
		t.Errorf("RelocateToRoom failed: %+v", *err)
	}

	// Assert
	if object.Room != room {
		t.Error("Object is not in the room it was moved to!")
	}

	if !isObjectInRoom(room, object) {
		t.Error("Object is not in the room's list")
	}
}

func Test_Object_RelocateToRoom_InPreviousRoom(t *testing.T) {
	// Arrange
	world := NewWorld()
	room1 := NewRoom()
	room2 := NewRoom()
	object := NewObject()
	world.AddObjects([]*Object{object})
	world.AddRooms([]*Room{room1, room2})
	object.RelocateToRoom(room1)

	if object.Room != room1 {
		t.Error("Object is not in room1 prior to move!")
	}

	// Act
	err := object.RelocateToRoom(room2)
	if err != nil {
		t.Errorf("RelocateToRoom failed: %+v", *err)
	}

	// Assert
	if object.Room != room2 {
		t.Error("Object is not in room2 after move!")
	}

	if isObjectInRoom(room1, object) {
		t.Error("Object is still in room1!")
	}

	if !isObjectInRoom(room2, object) {
		t.Error("Object is not in room2!")
	}
}

func Test_Move_NoRoomNorWorld(t *testing.T) {
	// Arrange
	player := &Player{}

	// Act
	err := player.Move(DIR_DOWN)

	// Assert
	if err == nil {
		t.Error("Move succeeded when it should not!")
	} else if err.ErrorCode() != ErrorIconsistency {
		t.Errorf("Unexpected error: %+v", *err)
	}
}

func Test_Move_RoomButNoWorld(t *testing.T) {
	// Arrange
	room := &Room{}
	player := &Player{Room: room}

	// Act
	err := player.Move(DIR_DOWN)

	// Assert
	if err == nil {
		t.Error("Move succeeded when it should not!")
	} else if err.ErrorCode() != ErrorIconsistency {
		t.Errorf("Unexpected error: %+v", *err)
	}
}

func Test_Move_InvalidDirection(t *testing.T) {
	// Arrange
	world := NewWorld()
	room := NewRoom()
	player := NewPlayer()

	world.AddRooms([]*Room{room})
	world.AddPlayers([]*Player{player})
	player.RelocateToRoom(room)

	// Act
	err := player.Move(DIR_DOWN)

	// Assert
	if err == nil {
		t.Error("Move succeeded when it should not!")
	} else if err.ErrorCode() != ErrorInvalidDirection {
		t.Errorf("Unexpected error: %+v", *err)
	}
}

func Test_Move_ValidDirection(t *testing.T) {
	// Arrange
	world := NewWorld()
	northRoom := NewRoom()
	room := NewRoom()
	player := NewPlayer()

	world.AddRooms([]*Room{room, northRoom})
	world.AddPlayers([]*Player{player})
	room.Connect(northRoom, DIR_NORTH)
	player.RelocateToRoom(room)

	// Act
	err := player.Move(DIR_NORTH)

	// Assert

	if err != nil {
		t.Errorf("Unexpected error: %+v", *err)
	}
}

func Test_Connect_Unidirectional(t *testing.T) {
	// Arrange
	world := NewWorld()
	northRoom := NewRoom()
	room := NewRoom()

	world.AddRooms([]*Room{room, northRoom})

	// Act
	err := room.Connect(northRoom, DIR_NORTH)

	// Assert

	if err != nil {
		t.Errorf("Unexpected error: %+v", *err)
		return
	}

	if room.AdjacentRooms[DIR_NORTH] != northRoom {
		t.Errorf("Failed to connect room to northRoom in the north direction!")
	}

	if northRoom.AdjacentRooms[DIR_SOUTH] != nil {
		t.Errorf("Connection was not unidirectional!")
	}
}

func Test_ConnectDuplex_Bidirectional(t *testing.T) {
	// Arrange
	world := NewWorld()
	northRoom := NewRoom()
	room := NewRoom()

	world.AddRooms([]*Room{room, northRoom})

	// Act
	err := room.ConnectDuplex(northRoom, DIR_NORTH)

	// Assert

	if err != nil {
		t.Errorf("Unexpected error: %+v", *err)
		return
	}

	if room.AdjacentRooms[DIR_NORTH] != northRoom {
		t.Errorf("Failed to connect room to northRoom in the north direction!")
	}

	if northRoom.AdjacentRooms[DIR_SOUTH] != room {
		t.Errorf("Failed to connect northRoom to room in the south direction (not bidirectional)!")
	}
}
