package engine

import (
	"errors"
	"log"

	"github.com/RichardKnop/chess/engine/orientation"
	"github.com/RichardKnop/chess/engine/position"
)

// Move represents a single move
type Move struct {
	Source string
	Target string
	Piece  string
}

// Player represents an active player
type Player struct {
	ID string
	// Buffered channel of moves
	Moves chan *Move
}

// NewPlayer creates a new player object
func NewPlayer(playerID string) (*Player, error) {
	return &Player{
		ID:    playerID,
		Moves: make(chan *Move),
	}, nil
}

// Board represents a game of chess
type Board struct {
	ID string
	// Inbound messages from the players
	Broadcast chan *Move
	// Join requests from players
	Join chan *Player
	// Leave requests from players
	Leave chan *Player
	// Player with white pieces
	White *Player
	// Player with black pieces
	Black *Player
	// Orientation of the board (black / white)
	Orientation string
	// FEM position string
	Position string
	// Sequence of all the moves played
	Moves []*Move
}

// NewBoard creates a new board
func NewBoard(boardID, orient, pos string) (*Board, error) {
	if orient != orientation.Black && orient != orientation.White {
		return nil, errors.New("Orientation can only be either black or white")
	}
	if pos == "" {
		pos = position.Initial
	}

	log.Printf("New board created: %s", boardID)

	return &Board{
		ID:          boardID,
		Broadcast:   make(chan *Move),
		Join:        make(chan *Player),
		Leave:       make(chan *Player),
		Orientation: orient,
		Position:    pos,
		Moves:       make([]*Move, 0),
	}, nil
}

// MakeMove moves a piece
func (b *Board) MakeMove(source, target, piece string) error {
	m := &Move{
		Target: target,
		Source: source,
		Piece:  piece,
	}
	b.Moves = append(b.Moves, m)

	// TODO - update position

	b.White.Moves <- m
	b.Black.Moves <- m
	return nil
}

// ToMap returns a map representation of the board
func (b *Board) ToMap() map[string]string {
	return map[string]string{
		"id":          b.ID,
		"orientation": b.Orientation,
		"position":    b.Position,
	}
}

// Run starts the board to players can join
func (b *Board) Run() error {
	for {
		select {
		case player := <-b.Join:
			// Game already has both players
			if b.White != nil && b.Black != nil {
				continue
			}

			log.Printf("Player %s joined board %s", player.ID, b.ID)

			switch b.Orientation {
			case orientation.White:
				if b.White == nil {
					b.White = player
				} else {
					b.Black = player
				}
			case orientation.Black:
				if b.Black == nil {
					b.Black = player
				} else {
					b.White = player
				}
			default:
				break
			}
		case player := <-b.Leave:
			log.Printf("Player %s left board %s", player.ID, b.ID)

			if player.ID == b.White.ID {
				b.White = nil
			}
			if player.ID == b.Black.ID {
				b.Black = nil
			}
			close(player.Moves)
		case move := <-b.Broadcast:
			if b.White != nil {
				b.White.Moves <- move
			}
			if b.Black != nil {
				b.Black.Moves <- move
			}
		}
	}
}
