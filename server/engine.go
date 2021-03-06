package server

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

// Engine maintains the set of active clients and broadcasts messages
type Engine struct {
	hub *Hub

	// Active games
	games map[string]*Game
}

// NewEngine creates a new instance of Engine
func NewEngine() *Engine {
	return &Engine{
		hub:   NewHub(),
		games: make(map[string]*Game, 0),
	}
}

// Run starts the hub
func (e *Engine) Run() {
	e.hub.Run()
}

// NewClient creates a new instance of Client
func (e *Engine) NewClient(conn *websocket.Conn) *Client {
	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		engine: e,
	}

	e.hub.register <- client

	return client
}

// FindGame returns in memory game state
func (e *Engine) FindGame(orientation string) (*Game, error) {
	log.Printf("Finding a game for a player with %s pieces", orientation)

	for _, game := range e.games {
		if game.White != nil && game.Black != nil {
			continue
		}

		switch orientation {
		case OrientationWhite:
			if game.White != nil {
				continue
			}
		case OrientationBlack:
			if game.Black != nil {
				continue
			}
		}

		return game, nil
	}

	log.Print("Suitable game not found, creating a new game")

	return e.newGame(InitialPosition)
}

// ClientDisconnected is called when a client disconnects
func (e *Engine) ClientDisconnected(c *Client) error {
	// Remove the client from all game instances
	for gameID, g := range e.games {
		g.Leave(c)

		if len(g.GetPlayers()) == 0 {
			log.Printf("Deleting game %s", gameID)
			delete(e.games, gameID)
		}
	}

	c.engine.hub.unregister <- c

	return nil
}

// GetGame returns in memory game state
func (e *Engine) GetGame(gameID string) (*Game, error) {
	g, ok := e.games[gameID]
	if !ok {
		return nil, NewGameNotFoundError(gameID)
	}
	return g, nil
}

// newGame creates a new game with blank state
func (e *Engine) newGame(position string) (*Game, error) {
	gameID := uuid.NewV4().String()
	_, ok := e.games[gameID]

	// This should never happen (UUIDs should be unique) but just in case
	if ok {
		return nil, NewGameAlreadyExistsError(gameID)
	}

	// Create a new game
	g, err := NewGame(gameID, position)
	if err != nil {
		return nil, err
	}
	e.games[gameID] = g

	return g, nil
}
