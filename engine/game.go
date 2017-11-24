package engine

import (
	"errors"
	"log"
)

// NewPlayer creates a new player object
func NewPlayer(playerID string) (*Player, error) {
	return &Player{
		ID:    playerID,
		Moves: make(chan *Move),
	}, nil
}

// NewGame creates a new game of chess
func NewGame(gameID, orientation, position string) (*Game, error) {
	if orientation != Black && orientation != White {
		return nil, errors.New("Orientation can only be either black or white")
	}
	if position == "" {
		position = InitialPosition
	}

	b := &Game{
		ID:          gameID,
		Broadcast:   make(chan *Move),
		Join:        make(chan *Player),
		Leave:       make(chan *Player),
		Orientation: orientation,
		Position:    position,
		Moves:       make([]*Move, 0),
	}

	log.Printf("New game created: %s", gameID)

	return b, nil
}

// MakeMove moves a piece
func (g *Game) MakeMove(source, target, piece string) error {
	m := &Move{
		Target: target,
		Source: source,
		Piece:  piece,
	}
	g.Moves = append(g.Moves, m)

	// TODO - update position

	// Notify players of the move
	g.White.Moves <- m
	g.Black.Moves <- m

	return nil
}

// ToMap returns a map representation of the game
func (g *Game) ToMap() map[string]string {
	return map[string]string{
		"id":          g.ID,
		"orientation": g.Orientation,
		"position":    g.Position,
	}
}

// Run starts the game to players can join
func (g *Game) Run() error {
	for {
		select {
		case player := <-g.Join:
			// Game already has both players
			if g.White != nil && g.Black != nil {
				continue
			}

			log.Printf("Player %s joined game %s", player.ID, g.ID)

			switch g.Orientation {
			case White:
				if g.White == nil {
					g.White = player
				} else {
					g.Black = player
				}
			case Black:
				if g.Black == nil {
					g.Black = player
				} else {
					g.White = player
				}
			default:
				break
			}
		case player := <-g.Leave:
			log.Printf("Player %s left game %s", player.ID, g.ID)

			if player.ID == g.White.ID {
				g.White = nil
			}
			if player.ID == g.Black.ID {
				g.Black = nil
			}
			close(player.Moves)
		case move := <-g.Broadcast:
			if g.White != nil {
				g.White.Moves <- move
			}
			if g.Black != nil {
				g.Black.Moves <- move
			}
		}
	}
}
