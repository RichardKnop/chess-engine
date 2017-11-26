package main

import (
	"log"
	"net/http"

	"github.com/RichardKnop/chess-engine/server"
	"github.com/gorilla/websocket"
)

var engine *server.Engine

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	engine = server.NewEngine()

	// Start the engine
	go engine.Run()

	// Web sockets handler
	http.HandleFunc("/ws", wsHandler)

	// Serving static files from public directory
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./client"))))

	log.Print("Websocket running at :8080/ws")

	panic(http.ListenAndServe(":8080", nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Open a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Register the client connection with the engine
	client := engine.NewClient(conn)

	go client.ReadPump()
	go client.WritePump()
}
