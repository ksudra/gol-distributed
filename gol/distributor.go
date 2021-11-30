package gol

import (
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
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

	world := make([][]uint8, p.ImageHeight)
	for y := range world {
		world[y] = make([]uint8, p.ImageWidth)
		for x := range world[y] {
			world[y][x] = <-c.ioInput
			c.events <- CellFlipped{
				CompletedTurns: 0,
				Cell:           util.Cell{X: x, Y: y},
			}
		}
	}

	return world
}

func sendWorld(p Params, c distributorChannels, world [][]uint8, turn int) {
	c.ioCommand <- ioOutput
	c.ioFilename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight), strconv.Itoa(turn)}, "x")
	for y := range world {
		for x := range world[y] {
			c.ioOutput <- world[y][x]
		}
	}
}

func getAliveCells(ticker *time.Ticker, c distributorChannels, client *rpc.Client) {
	for {
		select {
		case <-ticker.C:
			request := stubs.AliveReq{}
			response := new(stubs.AliveRes)
			err := client.Call(stubs.AliveCells, request, response)
			if err != nil {
				fmt.Println(err)
			}
			c.events <- AliveCellsCount{
				CompletedTurns: response.Turn,
				CellsCount:     response.Alive,
			}

		}
	}
}

func changeState(state int, client *rpc.Client, newState State, c distributorChannels) {
	request := stubs.ChangeStateReq{
		State: state,
	}
	response := new(stubs.ChangeStateRes)
	err := client.Call(stubs.ChangeState, request, response)
	if err != nil {
		fmt.Println(err)
	}

	c.events <- StateChange{
		CompletedTurns: response.Turn,
		NewState:       newState,
	}
}

func keyPresses(keyChan <-chan rune, p Params, c distributorChannels, client *rpc.Client) {
	for {
		select {
		case keyPress := <-keyChan:
			switch keyPress {
			case 's':
				request := stubs.BoardReq{}
				response := new(stubs.BoardRes)
				err := client.Call(stubs.GetBoard, request, response)
				if err != nil {
					fmt.Println(err)
				}
				sendWorld(p, c, response.World, response.Turn)
			case 'q':
				changeState(0, client, Quitting, c)
				return
			case 'p':
				changeState(1, client, Paused, c)
				for {
					keyPress = <-keyChan
					for keyPress != 'p' {
						keyPress = <-keyChan
					}
					fmt.Println("Continuing")
					changeState(2, client, Executing, c)
					break
				}
			case 'k':
				request := stubs.BoardReq{}
				response := new(stubs.BoardRes)
				err := client.Call(stubs.GetBoard, request, response)
				if err != nil {
					fmt.Println(err)
				}
				sendWorld(p, c, response.World, response.Turn)

				req := stubs.CloseReq{}
				res := new(stubs.CloseRes)
				err = client.Call(stubs.ShutDown, req, res)
				if err != nil {
					fmt.Println(err)
				}
				return
			}
		default:
			break
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
	err := client.Call(stubs.RunGame, request, response)
	if err != nil {
		fmt.Println(err)
	}

	completedTurns = response.CompletedTurns
	sendWorld(p, c, response.World, response.CompletedTurns)
	c.events <- FinalTurnComplete{
		CompletedTurns: response.CompletedTurns,
		Alive:          response.Alive,
	}
}
