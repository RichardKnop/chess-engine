package server

import (
	"encoding/json"
	"log"
)

// Move represents a single move
type Move struct {
	PlayerID string
	Source   string
	Target   string
	Piece    string
}

// Game represents a game of chess
type Game struct {
	ID string
	// FEM position string
	Position string
	// Sequence of all the moves played
	Moves []*Move
	// Player with white pieces
	White *Client
	// Player with black pieces
	Black *Client
}

// NewGame creates a new game of chess
func NewGame(gameID, position string) (*Game, error) {
	// Default to initial position if not specified
	if position == "" {
		position = InitialPosition
	}

	g := &Game{
		ID:       gameID,
		Position: position,
		Moves:    make([]*Move, 0),
	}

	log.Printf("New game created: %s", g.ID)

	return g, nil
}

// ActivePlayerID returns player ID of a player who is on the move currently
func (g *Game) ActivePlayerID() string {
	if len(g.Moves)/2 == 0 {
		return g.White.PlayerID
	}
	return g.Black.PlayerID
}

// NotifyPlayers sends a message to all players
func (g *Game) NotifyPlayers(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if g.White != nil {
		g.White.send <- data
	}
	if g.Black != nil {
		g.Black.send <- data
	}

	return nil
}

// Join is called when a player joins the game
func (g *Game) Join(c *Client, playerID, orientation string) error {
	c.PlayerID = playerID

	switch orientation {
	case OrientationWhite:
		g.White = c
	case OrientationBlack:
		g.Black = c
	}

	log.Printf("Player %s joined game %s playing with %s pieces", c.PlayerID, g.ID, orientation)

	if g.Black != nil && g.White != nil {
		g.NotifyPlayers(&Message{
			Type: "game_started",
			Data: &MessageData{
				GameID:   g.ID,
				Position: g.Position,
				PlayerID: c.PlayerID,
			},
		})
	}

	return nil
}

// MakeMove moves a piece
func (g *Game) MakeMove(playerID, source, target, piece, oldPosition, newPosition string) error {
	m := &Move{
		PlayerID: playerID,
		Target:   target,
		Source:   source,
		Piece:    piece,
	}

	// TODO - validate move

	g.Position = newPosition
	g.Moves = append(g.Moves, m)

	return g.NotifyPlayers(&Message{
		Type: "move_made",
		Data: &MessageData{
			GameID:   g.ID,
			Position: g.Position,
			PlayerID: playerID,
			Target:   target,
			Source:   source,
			Piece:    piece,
		},
	})
}

// NotifyAboutState notifies players about current game state
func (g *Game) NotifyAboutState() error {
	return g.NotifyPlayers(&Message{
		Type: "state_update",
		Data: &MessageData{
			GameID:   g.ID,
			Position: g.Position,
			PlayerID: g.ActivePlayerID(),
		},
	})
}

func (g *Game) findOpponent(c *Client) *Client {
	if g.White == c {
		return g.Black
	}
	if g.Black == c {
		return g.White
	}

	return nil
}
