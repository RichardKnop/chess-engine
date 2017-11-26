package server

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

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	engine *Engine
}

// ReadPump pumps messages from the websocket connection to the engine/hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() error {
	for {
		// Read the message from the websocket
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Debug websocket error: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				return fmt.Errorf("Unexpected close error: %v", err)
			}
		}

		data = bytes.TrimSpace(bytes.Replace(data, newline, space, -1))

		// Unmarshal the message
		msg := new(Message)
		if err := json.Unmarshal(data, msg); err != nil {
			continue
		}

		// Log the received message
		log.Printf("Received message: %s", data)

		if err := c.handleMessage(msg); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
}

// WritePump pumps messages from the engine/hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Print(err)
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg *Message) error {
	handlers := map[string]func(msg *Message) error{
		"find_game":  c.findGame,
		"leave_game": c.leaveGame,
		"make_move":  c.makeMove,
	}

	// Handle message based on its type
	handler, ok := handlers[msg.Type]
	if !ok {
		return NewUnknownMessageType(msg.Type)
	}

	return handler(msg)
}

func (c *Client) findGame(msg *Message) error {
	g, err := c.engine.FindGame(msg.Data.Orientation)
	if err != nil {
		return err
	}
	p, err := c.engine.NewPlayer(c, msg.Data.PlayerID, msg.Data.Orientation)
	if err != nil {
		return err
	}
	return g.Join(p)
}

func (c *Client) leaveGame(msg *Message) error {
	g, err := c.engine.GetGame(msg.Data.GameID)
	if err != nil {
		return err
	}
	p, err := c.engine.NewPlayer(c, msg.Data.PlayerID, msg.Data.Orientation)
	if err != nil {
		return err
	}
	return g.Leave(p)
}

func (c *Client) makeMove(msg *Message) error {
	g, err := c.engine.GetGame(msg.Data.GameID)
	if err != nil {
		return err
	}
	return g.MakeMove(
		msg.Data.PlayerID,
		msg.Data.Source,
		msg.Data.Target,
		msg.Data.Piece,
		msg.Data.OldPosition,
		msg.Data.NewPosition,
	)
}
