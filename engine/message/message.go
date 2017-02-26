package message

const (
	// NewGame is a message requesting a new game
	NewGame = "new_game"
	// JoinGame is a message requesting to join a game
	JoinGame = "join_game"
	// MakeMove is a message signalling moving a piece
	MakeMove = "make_move"
)

// Message is a generic message send via websockets
type Message struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}
