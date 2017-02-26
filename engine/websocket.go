package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/RichardKnop/chess/engine/message"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// ReadFromWebsocket starts reading message from a websocket connection
func (e *Engine) ReadFromWebsocket(conn *websocket.Conn) error {
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// Read the message from the websocket
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				return fmt.Errorf("Unexpected error: %v", err)
			}
		}

		data = bytes.TrimSpace(bytes.Replace(data, newline, space, -1))

		// Log the received message
		log.Printf("Received message: %s", data)

		response, err := e.handleMessage(data)
		if err != nil {
			log.Print(err)
		}

		log.Printf("Sending message: %v", response)

		// Write response message to socket
		conn.WriteJSON(response)
	}
}

func (e *Engine) handleMessage(data []byte) (interface{}, error) {
	// Unmarshal the message
	msg := new(message.Message)
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("Invalid message: %v", err)
	}

	// Handle message based on its type
	switch msg.Type {
	case message.NewGame:
		orient, _ := msg.Data["orientation"]
		pos, _ := msg.Data["position"]
		board, err := e.NewGame(orient, pos)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"type": message.NewGame,
			"data": board.ToMap(),
		}, nil
	case message.JoinGame:
		boardID, _ := msg.Data["board_id"]
		playerID, _ := msg.Data["player_id"]
		board, err := e.GetBoard(boardID)
		if err != nil {
			return nil, err
		}
		board.Join <- &Player{ID: playerID}
		return map[string]interface{}{
			"type": message.JoinGame,
			"data": board.ToMap(),
		}, nil
	case message.MakeMove:
		boardID, _ := msg.Data["board_id"]
		source, _ := msg.Data["source"]
		target, _ := msg.Data["target"]
		piece, _ := msg.Data["piece"]
		board, err := e.GetBoard(boardID)
		if err != nil {
			return nil, err
		}
		if err := board.MakeMove(source, target, piece); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"type": message.MakeMove,
			"data": board.ToMap(),
		}, nil
	}

	return nil, fmt.Errorf("Unknown message type: %s", msg.Type)
}
