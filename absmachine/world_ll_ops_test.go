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

func Test_MovePlayerToRoom_NotInPreviousRoom(t *testing.T) {
	// Arrange
	room := Room{}
	player := Player{}
	world := World{}
	world.AddRooms([]*Room{&room})
	world.AddPlayers([]*Player{&player})

	if player.Room != nil {
		t.Error("Player is in a room prior to move!")
	}

	// Act
	err := world.MovePlayerToRoom(&player, &room)
	if err != nil {
		t.Errorf("MovePlayerToRoom failed: %+v", *err)
	}

	// Assert
	if player.Room != &room {
		t.Error("Player is not in the room it was moved to!")
	}

	if !isPlayerInRoom(&room, &player) {
		t.Error("Player is not in the room's list")
	}
}

func Test_MovePlayerToRoom_InPreviousRoom(t *testing.T) {
	// Arrange
	room1 := Room{}
	room2 := Room{}
	player := Player{Room: &room1}
	room1.Players = append(room1.Players, &player)
	world := World{}
	world.AddPlayers([]*Player{&player})
	world.AddRooms([]*Room{&room1, &room2})

	if player.Room != &room1 {
		t.Error("Player is not in room1 prior to move!")
	}

	// Act
	err := world.MovePlayerToRoom(&player, &room2)
	if err != nil {
		t.Errorf("MovePlayerToRoom failed: %+v", *err)
	}

	// Assert
	if player.Room != &room2 {
		t.Error("Player is not in room2 after move!")
	}

	if isPlayerInRoom(&room1, &player) {
		t.Error("Player is still in room1!")
	}

	if !isPlayerInRoom(&room2, &player) {
		t.Error("Player is not in room2!")
	}
}

func Test_MovePlayer_NoRoomNorWorld(t *testing.T) {
	// Arrange
	player := &Player{}

	// Act
	err := player.MovePlayer(DirectionDown)

	// Assert
	if err == nil {
		t.Error("Move succeeded when it should not!")
	} else if err.ErrorCode() != ErrorIconsistency {
		t.Errorf("Unexpected error: %+v", *err)
	}
}

func Test_MovePlayer_RoomButNoWorld(t *testing.T) {
	// Arrange
	room := &Room{}
	player := &Player{Room: room}

	// Act
	err := player.MovePlayer(DirectionDown)

	// Assert
	if err == nil {
		t.Error("Move succeeded when it should not!")
	} else if err.ErrorCode() != ErrorIconsistency {
		t.Errorf("Unexpected error: %+v", *err)
	}
}

func Test_MovePlayer_InvalidDirection(t *testing.T) {
	// Arrange
	room := &Room{}
	player := &Player{Room: room}
	world := World{}
	world.AddRooms([]*Room{room})
	world.AddPlayers([]*Player{player})

	// Act
	err := player.MovePlayer(DirectionDown)

	// Assert
	if err == nil {
		t.Error("Move succeeded when it should not!")
	} else if err.ErrorCode() != ErrorInvalidDirection {
		t.Errorf("Unexpected error: %+v", *err)
	}
}

func Test_MovePlayer_ValidDirection(t *testing.T) {
	// Arrange
	northRoom := &Room{}
	room := &Room{}
	room.Connect(northRoom, DirectionNorth)

	player := &Player{Room: room}
	world := World{}
	world.AddRooms([]*Room{room, northRoom})
	world.AddPlayers([]*Player{player})

	// Act
	err := player.MovePlayer(DirectionNorth)

	// Assert

	if err != nil {
		t.Errorf("Unexpected error: %+v", *err)
	}
}
