package engine

import (
	"github.com/satori/go.uuid"
)

// New returns a new Engine instance
func New() (*Engine, error) {
	e := &Engine{
		games: make(map[string]*Game),
	}
	return e, nil
}

// NewGame creates a new game with blank state
func (e *Engine) NewGame(gameID, orientation, position string) (*Game, error) {
	if gameID == "" {
		gameID = uuid.NewV4().String()
	}
	_, ok := e.games[gameID]

	// This should never happen (UUIDs should be unique) but just in case
	if ok {
		return nil, NewGameAlreadyExistsError(gameID)
	}

	// Default to initial position if not specified
	if position == "" {
		position = InitialPosition
	}

	// Create a new game
	g, err := NewGame(gameID, orientation, position)
	if err != nil {
		return nil, err
	}
	e.games[gameID] = g

	// Run the game in a goroutine
	go g.Run()

	return g, nil
}

// GetGame returns in memory game state
func (e *Engine) GetGame(gameID string) (*Game, error) {
	g, ok := e.games[gameID]
	if !ok {
		return nil, NewGameNotFoundError(gameID)
	}
	return g, nil
}
