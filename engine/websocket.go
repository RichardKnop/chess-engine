package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

		msg, err := e.handleMessage(data)
		if err != nil {
			log.Print(err)
		}

		// debug, err := json.Marshal(response)
		// if err == nil {
		// 	log.Printf("Replied with: %s", debug)
		// }

		if msg != nil {
			// Write response message to socket
			conn.WriteJSON(msg.Data)
		}
	}
}

func (e *Engine) handleMessage(data []byte) (*Message, error) {
	data = bytes.TrimSpace(bytes.Replace(data, newline, space, -1))

	// Unmarshal the message
	msg := new(Message)
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, nil
	}

	// Log the received message
	log.Printf("Received message: %s", data)

	// Handle message based on its type
	switch msg.Type {
	case NewGameMessage:
		g, err := e.NewGame(msg.Data.GameID, msg.Data.Orientation, msg.Data.Position)
		if err != nil {
			return nil, err
		}

		msg.Data.GameID = g.ID
		msg.Data.Orientation = g.Orientation
		msg.Data.Position = g.Position
		return msg, nil
	case JoinGameMessage:
		g, err := e.GetGame(msg.Data.GameID)
		if err != nil {
			return nil, err
		}
		g.Join <- &Player{ID: msg.Data.PlayerID}

		msg.Data.GameID = g.ID
		msg.Data.Orientation = g.Orientation
		msg.Data.Position = g.Position
		return msg, nil
	case MakeMoveMessage:
		g, err := e.GetGame(msg.Data.GameID)
		if err != nil {
			return nil, err
		}
		if err := g.MakeMove(msg.Data.Source, msg.Data.Target, msg.Data.Piece); err != nil {
			return nil, err
		}

		msg.Data.GameID = g.ID
		msg.Data.Orientation = g.Orientation
		msg.Data.Position = g.Position
		return msg, nil
	}

	return nil, fmt.Errorf("Unknown message type: %s", msg.Type)
}
