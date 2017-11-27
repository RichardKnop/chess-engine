package server

import (
	"encoding/json"
	"log"
)

// Player represents an active player
type Player struct {
	ID          string
	Orientation string

	client *Client
}

// NewPlayer creates a new player object
func NewPlayer(client *Client, playerID, orientation string) (*Player, error) {
	return &Player{
		ID:          playerID,
		Orientation: orientation,
		client:      client,
	}, nil
}

// Notify sends a message to player
func (p *Player) Notify(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
	}
	p.client.send <- data
}

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
	White *Player
	// Player with black pieces
	Black *Player
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

// NotifyPlayers sends a message to all players
func (g *Game) NotifyPlayers(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if g.White != nil {
		g.White.client.send <- data
	}
	if g.Black != nil {
		g.Black.client.send <- data
	}

	return nil
}

// Join is called when a player joins the game
func (g *Game) Join(p *Player) error {
	switch p.Orientation {
	case OrientationWhite:
		g.White = p
	case OrientationBlack:
		g.Black = p
	}

	log.Printf("Player %s joined game %s playing with %s pieces", p.ID, g.ID, p.Orientation)

	if g.Black != nil && g.White != nil {
		g.NotifyPlayers(&Message{
			Type: "game_started",
			Data: &MessageData{
				GameID:   g.ID,
				Position: g.Position,
				PlayerID: p.ID,
			},
		})
	}

	return nil
}

// Leave is called when a player leaves the game
func (g *Game) Leave(p *Player) error {
	if g.White != nil && p.ID == g.White.ID {
		g.White = nil
	}
	if g.Black != nil && p.ID == g.Black.ID {
		g.Black = nil
	}

	log.Printf("Player %s left game %s", p.ID, g.ID)

	if opponent := g.findOpponent(p.ID); opponent != nil {
		opponent.Notify(&Message{
			Type: "player_left",
			Data: &MessageData{
				GameID:   g.ID,
				Position: g.Position,
				PlayerID: p.ID,
			},
		})
	}

	return nil
}

// MakeMove moves a piece
func (g *Game) MakeMove(playerID, source, target, piece, oldPosition, newPosition string) error {
	// Validate move

	m := &Move{
		PlayerID: playerID,
		Target:   target,
		Source:   source,
		Piece:    piece,
	}
	g.Position = newPosition
	g.Moves = append(g.Moves, m)

	g.NotifyPlayers(&Message{
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

	return nil
}

func (g *Game) findOpponent(playerID string) *Player {
	if g.White != nil && playerID == g.White.ID {
		return g.Black
	}
	if g.Black != nil && playerID == g.Black.ID {
		return g.White
	}

	return nil
}
