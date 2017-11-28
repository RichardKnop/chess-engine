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

	if len(g.GetPlayers()) == 2 {
		msg := &Message{
			Type: "game_started",
			Data: &MessageData{
				GameID:   g.ID,
				Position: g.Position,
				PlayerID: c.PlayerID,
			},
		}
		g.notifyPlayers(msg)
	}

	return nil
}

// Leave is called when a player leaves the game
func (g *Game) Leave(c *Client) error {
	if g.White != nil && g.White.PlayerID == c.PlayerID {
		g.White = nil
		log.Printf("Player %s left game %s", c.PlayerID, g.ID)
	}
	if g.Black != nil && g.Black.PlayerID == c.PlayerID {
		g.Black = nil
		log.Printf("Player %s left game %s", c.PlayerID, g.ID)
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

	msg := &Message{
		Type: "move_made",
		Data: &MessageData{
			GameID:   g.ID,
			Position: g.Position,
			PlayerID: playerID,
			Target:   target,
			Source:   source,
			Piece:    piece,
		},
	}
	return g.notifyPlayers(msg)
}

// NotifyAboutState notifies players about current game state
func (g *Game) NotifyAboutState() error {
	msg := &Message{
		Type: "state_update",
		Data: &MessageData{
			GameID:   g.ID,
			Position: g.Position,
		},
	}
	if activePlayerID := g.getActivePlayerID(); activePlayerID != nil {
		msg.Data.PlayerID = *activePlayerID
	}
	return g.notifyPlayers(msg)
}

// GetPlayers returns slice of players currently connected to the game
func (g *Game) GetPlayers() []*Client {
	var players []*Client
	if g.White != nil {
		players = append(players, g.White)
	}
	if g.Black != nil {
		players = append(players, g.Black)
	}
	return players
}

// getActivePlayerID returns player ID of a player who is on the move currently
func (g *Game) getActivePlayerID() *string {
	if len(g.Moves)/2 == 0 && g.White != nil {
		return &g.White.PlayerID
	}
	if len(g.Moves)/2 == 1 && g.Black != nil {
		return &g.Black.PlayerID
	}
	return nil
}

// notifyPlayers sends a message to all players
func (g *Game) notifyPlayers(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	for _, p := range g.GetPlayers() {
		p.send <- data
	}

	return nil
}

// findOpponent returns opponent to player
func (g *Game) findOpponent(c *Client) *Client {
	if g.White == c {
		return g.Black
	}
	if g.Black == c {
		return g.White
	}

	return nil
}
