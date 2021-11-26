package main

import (
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type GameOfLife struct{}

var aliveCount int
var turn int

func (g *GameOfLife) GOL(request stubs.GameReq, response *stubs.GameRes) (err error) {
	tempWorld := make([][]byte, len(request.World))
	for i := range request.World {
		tempWorld[i] = make([]byte, len(request.World[i]))
		copy(tempWorld[i], request.World[i])
	}
	for i := 0; i < request.Turns; i++ {
		tempWorld = calculateNextState(request.Width, request.Height, request.Threads, tempWorld)
		turn = i
		aliveCount = len(calculateAliveCells(tempWorld))
	}
	response.World = tempWorld
	response.CompletedTurns = request.Turns
	response.Alive = calculateAliveCells(tempWorld)
	return
}

func calculateNextState(width int, height int, threads int, world [][]byte) [][]byte {
	tempWorld := make([][]byte, len(world))
	for i := range world {
		tempWorld[i] = make([]byte, len(world[i]))
		copy(tempWorld[i], world[i])
	}

	var wg sync.WaitGroup
	var remainder sync.WaitGroup

	for i := 0; i < threads; i++ {
		start := i * (height - height%threads) / threads
		end := start + (height-height%threads)/threads
		wg.Add(1)
		go worker(&wg, width, height, start, end, tempWorld, world)

	}
	wg.Wait()

	if height%threads > 0 {
		start := height - height%threads
		remainder.Add(1)
		go worker(&remainder, width, height, start, height, tempWorld, world)
	}

	remainder.Wait()

	return tempWorld
}

func countNeighbours(width int, height int, x int, y int, world [][]uint8) int {
	neighbours := [8][2]int{
		{-1, -1},
		{-1, 0},
		{-1, 1},
		{0, -1},
		{0, 1},
		{1, -1},
		{1, 0},
		{1, 1},
	}

	count := 0

	for _, r := range neighbours {
		if world[(y+r[0]+height)%height][(x+r[1]+width)%width] == 255 {
			count++
		}
	}

	return count
}

func worker(wg *sync.WaitGroup, width int, height int, start int, end int, newWorld [][]byte, world [][]byte) {
	defer wg.Done()

	for y := start; y < end; y++ {
		for x := range newWorld {
			count := countNeighbours(width, height, x, y, world)

			if world[y][x] == 255 && (count < 2 || count > 3) {
				newWorld[y][x] = 0
			} else if world[y][x] == 0 && count == 3 {
				newWorld[y][x] = 255
			}
		}
	}
}

func calculateAliveCells(world [][]byte) []util.Cell {
	var cells []util.Cell
	for i := range world {
		for j := range world[i] {
			if world[i][j] == 255 {
				cells = append(cells, util.Cell{X: j, Y: i})
			}
		}
	}
	return cells
}

func (g *GameOfLife) getAlive(request stubs.AliveReq, response *stubs.AliveRes) (err error) {
	response.Turn = turn
	response.Alive = aliveCount
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	err := rpc.Register(&GameOfLife{})
	if err != nil {
		return
	}

	listener, _ := net.Listen("tcp", ":"+*pAddr)

	defer listener.Close()
	rpc.Accept(listener)

}
