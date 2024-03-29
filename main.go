package main

import (
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/io"
	"github.com/jorgensigvardsson/gomud/logging"
)

const TICK = 100 * time.Millisecond
const MAX_USER_LIMIT = 100
const MAX_PLAYER_INPUT_QUEUE_LIMIT = 20

func buildWorld() *absmachine.World {
	world := absmachine.NewWorld()

	entryRoom := absmachine.NewRoom()
	entryRoom.Title = "The entry room"
	entryRoom.Description = "You are in the starting room of this MUD.\r\nThere are creepy spiders and insects everywhere! RUN!"

	peacefulRoom := absmachine.NewRoom()
	peacefulRoom.Title = "The peaceful room"
	peacefulRoom.Description = "A peaceful room. Cows and elephants are roaming the vast grassfield that continues to the north."

	err := entryRoom.ConnectDuplex(peacefulRoom, absmachine.DIR_NORTH)
	if err != nil {
		panic(err)
	}

	mob1 := absmachine.NewMob()
	mob1.Name = "Angry Spider"
	mob1.Description = "The hairy 8 legged beast is angry!"
	mob1.RoomDescription = "An angry spider is looking straight at you with all of its eyes!"
	mob1.Actions = append(
		mob1.Actions,
		absmachine.MobAction{
			PeriodLength: 20,
			Probability:  0.1,
			Function: &absmachine.SimpleVerbMobAction{
				Verb:        "jumping",
				Preposition: "around",
			},
		},
	)

	world.AddMobs([]*absmachine.Mob{mob1})
	world.AddRooms([]*absmachine.Room{entryRoom, peacefulRoom})

	mob1.RelocateToRoom(entryRoom)
	world.StartRoom = entryRoom

	return world
}

func main() {
	// Make sure we seed the RNG
	rand.Seed(time.Now().UnixNano())

	world := buildWorld()
	logger := logging.NewTimestampLoggerDecorator(
		logging.NewSynchronizingLoggerDecorator(
			logging.NewConsoleLogger(),
			50,
		),
	)
	inputQueue := io.NewInputQueue(MAX_USER_LIMIT, MAX_PLAYER_INPUT_QUEUE_LIMIT, logger)
	commandChannel := make(chan *io.PlayerInput, MAX_USER_LIMIT*MAX_PLAYER_INPUT_QUEUE_LIMIT)
	sigtermChannel := make(chan os.Signal)
	connectionsStopChannel := make(chan interface{})
	listenerErrorChannel := make(chan error, 1)
	workGroup := sync.WaitGroup{}

	defer close(commandChannel)
	defer close(sigtermChannel)
	defer logger.Close()

	logger.Println("Starting up Go MUD on port 5000...")
	listener, err := net.Listen("tcp", ":5000")

	if err != nil {
		panic("Failed to open TCP port 5000")
	}

	// Setup SIGTERM handler
	signal.Notify(sigtermChannel, os.Interrupt, syscall.SIGTERM)
	logger.Println("Stop server with Ctrl+C (SIGTERM)")

	// Spin off in a go routine to handle connections
	go io.HandleConnections(listener, logger, commandChannel, listenerErrorChannel, connectionsStopChannel, &workGroup)

	// The game loop!
	run := true
	tick := 0
	for run {
		// Measure how long time we spent processing the commands
		handleCommandsT0 := time.Now().UTC()
		inputQueue.Execute(world, tick)
		handleCommandsT1 := time.Now().UTC()

		// Remove the delta from the TICK length
		timeToSleep := TICK - handleCommandsT0.Sub(handleCommandsT1)
		timeToWakeup := handleCommandsT1.Add(timeToSleep)

		// Pump input from command channel onto input queue
		for timeToSleep > 0 {
			select {
			case inputOrCommand := <-commandChannel:
				inputQueue.Append(inputOrCommand)
			case <-time.After(timeToSleep):
				// Do nothing on purpose!
			case <-sigtermChannel:
				logger.Println("Shutting down...")
				run = false
			case err := <-listenerErrorChannel:
				logger.Printlnf("Accepting TCP connections failed: %v", err.Error())
				run = false
			}

			// Figure out if we need to sleep more!
			timeToSleep = timeToWakeup.Sub(time.Now().UTC())
		}

		tick++
	}

	// Shut everything down!
	listener.Close()

	// Terminate all connected
	close(connectionsStopChannel)

	// Wait for all go routines to stop
	workGroup.Wait()

	// Now we're no longer accepting new connections, and all existing sessions have been closed

	// TODO: Serialize current state of world!

	logger.Println("Go MUD successfully shut down.")
}
