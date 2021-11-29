package gol

import (
	"net/rpc"
	"strconv"
	"strings"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

type GameOfLife struct{}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	world := buildWorld(p, c)
	ticker := time.NewTicker(2 * time.Second)
	// TODO: Create a 2D slice to store the world.

	turn := 0
	server := "127.0.0.1:8030"
	client, _ := rpc.Dial("tcp", server)

	defer client.Close()

	go getAliveCells(ticker, c, client)
	makeCall(client, p, c, world, turn)
	ticker.Stop()

	// TODO: Execute all turns of the Game of Life.

	// TODO: Report the final state using FinalTurnCompleteEvent.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func buildWorld(p Params, c distributorChannels) [][]uint8 {
	c.ioCommand <- ioInput
	c.ioFilename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight)}, "x")

	world := make([][]byte, p.ImageHeight)
	for y := range world {
		world[y] = make([]byte, p.ImageWidth)
		for x := range world[y] {
			world[y][x] = <-c.ioInput
		}
	}

	return world
}

func getAliveCells(ticker *time.Ticker, c distributorChannels, client *rpc.Client) {
	for {
		select {
		case <-ticker.C:
			request := stubs.AliveReq{}
			response := new(stubs.AliveRes)
			client.Call(stubs.AliveCells, request, response)
			c.events <- AliveCellsCount{
				CompletedTurns: response.Turn,
				CellsCount:     response.Alive,
			}

		}
	}
}

func makeCall(client *rpc.Client, p Params, c distributorChannels, world [][]uint8, completedTurns int) {
	request := stubs.GameReq{
		Width:   p.ImageWidth,
		Height:  p.ImageHeight,
		Threads: p.Threads,
		Turns:   p.Turns,
		World:   world,
	}

	response := new(stubs.GameRes)
	client.Call(stubs.RunGame, request, response)

	completedTurns = response.CompletedTurns
	c.events <- FinalTurnComplete{
		CompletedTurns: response.CompletedTurns,
		Alive:          response.Alive,
	}
}
