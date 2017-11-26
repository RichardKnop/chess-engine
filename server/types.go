package server

const (
	// OrientationBlack means black is on the play facing white
	OrientationBlack = "black"
	// OrientationWhite means white is on the play facing black
	OrientationWhite = "white"

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
	OldPosition string `json:"old_position,omitempty"`
	NewPosition string `json:"new_position,omitempty"`
}
