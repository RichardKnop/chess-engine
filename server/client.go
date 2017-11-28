package server

import (
	"bytes"
	"encoding/json"
	"errors"
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
	//maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	PlayerID string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	engine *Engine
}

// Notify sends a message to client
func (c *Client) Notify(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	c.send <- data
	return nil
}

// ReadPump pumps messages from the websocket connection to the engine/hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() (err error) {
	defer func() {
		// Intercept panic from failed websocket connection such as
		// https://github.com/gorilla/websocket/blob/master/conn.go#L959
		// and deregister client
		if e := recover(); e != nil {
			switch e := e.(type) {
			case error:
				err = e
			case string:
				err = errors.New(e)
			}

			c.engine.hub.unregister <- c
		}
	}()

	for {
		// Read the message from the websocket
		_, data, e := c.conn.ReadMessage()
		if e != nil {
			if websocket.IsUnexpectedCloseError(e, websocket.CloseGoingAway) {
				err = fmt.Errorf("Unexpected close error: %v", e)
				return
			}
		}

		data = bytes.TrimSpace(bytes.Replace(data, newline, space, -1))

		// Unmarshal the message
		msg := new(Message)
		if json.Unmarshal(data, msg) != nil {
			continue
		}

		// Log the received message
		log.Printf("Received message: %s", data)

		if e := c.handleMessage(msg); e != nil {
			log.Printf("Error handling message: %v", e)
		}
	}
}

// WritePump pumps messages from the engine/hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() error {
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
				return nil
			}

			w, e := c.conn.NextWriter(websocket.TextMessage)
			if e != nil {
				return e
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if e := w.Close(); e != nil {
				return e
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if e := c.conn.WriteMessage(websocket.PingMessage, []byte{}); e != nil {
				return e
			}
		}
	}
}

func (c *Client) handleMessage(msg *Message) error {
	handlers := map[string]func(msg *Message) error{
		"find_game": c.findGame,
		"get_game":  c.getGame,
		"make_move": c.makeMove,
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
	return g.Join(c, msg.Data.PlayerID, msg.Data.Orientation)
}

func (c *Client) getGame(msg *Message) error {
	g, err := c.engine.GetGame(msg.Data.GameID)
	if err != nil {
		return err
	}
	return g.NotifyAboutState()
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
