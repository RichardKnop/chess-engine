package engine

const (
	// NewGameMessage is a message requesting a new game
	NewGameMessage = "new_game"
	// JoinGameMessage is a message requesting to join a game
	JoinGameMessage = "join_game"
	// MakeMoveMessage is a message signalling moving a piece
	MakeMoveMessage = "make_move"

	// Black means black is on the play facing white
	Black = "black"
	// White means white is on the play facing black
	White = "white"

	// Initial is a FEM representation of initial board state
	// See https://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation
	InitialPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
)

// Message is a generic message send via websockets
type Message struct {
	Type string       `json:"type"`
	Data *MessageData `json:"data"`
}

// MessageData ...
type MessageData struct {
	Orientation string `json:"orientation"`
	Position    string `json:"position"`
	GameID      string `json:"game_id"`
	PlayerID    string `json:"player_id,omitempty"`
	Source      string `json:"source,omitempty"`
	Target      string `json:"target,omitempty"`
	Piece       string `json:"piece,omitempty"`
}

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

// Game represents a game of chess
type Game struct {
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

// Engine is a chess engine which stores all games
type Engine struct {
	games map[string]*Game
}
