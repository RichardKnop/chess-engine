package engine

import (
	"fmt"

	"github.com/RichardKnop/chess/engine/position"
	"github.com/pborman/uuid"
)

// Engine ...
type Engine struct {
	boards map[string]*Board
}

// New returns a new Engine instance
func New() (*Engine, error) {
	e := &Engine{
		boards: make(map[string]*Board),
	}
	return e, nil
}

// NewGame creates a new board with blank state
func (e *Engine) NewGame(orient, pos string) (*Board, error) {
	boardID := uuid.New()
	_, ok := e.boards[boardID]
	if ok {
		return nil, fmt.Errorf("Board %s already exists", boardID)
	}
	if pos == "" {
		pos = position.Initial
	}
	board, err := NewBoard(boardID, orient, pos)
	if err != nil {
		return nil, err
	}
	e.boards[boardID] = board
	go board.Run()
	return board, nil
}

// GetBoard returns in memory board state
func (e *Engine) GetBoard(boardID string) (*Board, error) {
	board, ok := e.boards[boardID]
	if !ok {
		return nil, fmt.Errorf("Board %s does not exist", boardID)
	}
	return board, nil
}
