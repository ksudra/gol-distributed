package stubs

import (
	"uk.ac.bris.cs/gameoflife/util"
)

var RunGame = "GameOfLife.GOL"
var AliveCells = "GameOfLife.GetNumAlive"
var ChangeState = "GameOfLife.StateChange"
var GetBoard = "GameOfLife.GetBoard"
var ShutDown = "GameOfLife.ShutDown"

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

type BoardReq struct{}

type BoardRes struct {
	Turn  int
	World [][]uint8
}

type ChangeStateReq struct {
	State int
}

type ChangeStateRes struct {
	Turn int
}

type AliveReq struct{}

type AliveRes struct {
	Turn  int
	Alive int
}

type CellReq struct{}

type CellRes struct {
	Turn int
	Cell util.Cell
}

type CloseReq struct{}

type CloseRes struct{}
