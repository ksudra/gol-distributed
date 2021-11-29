package stubs

import (
	"uk.ac.bris.cs/gameoflife/util"
)

var RunGame = "GameOfLife.GOL"
var AliveCells = "GameOfLife.getNumAlive"

type GameReq struct {
	Width   int
	Height  int
	Threads int
	Turns   int
	World   [][]uint8
}

type GameRes struct {
	Alive          []util.Cell
	CompletedTurns int
	World          [][]uint8
}

type NextStateReq struct {
	Width   int
	Height  int
	Threads int
	Turns   int
	World   [][]uint8
}

type NextStateRes struct {
	World [][]uint8
}

type StateReq struct{}

type StateRes struct {
	Turn  int
	world [][]uint8
}

type AliveReq struct{}

type AliveRes struct {
	Turn  int
	Alive int
}

type CloseReq struct{}

type CloseRes struct{}
