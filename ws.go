package main

import (
	"log"
	"net/http"

	"github.com/RichardKnop/chess-game/engine"
	"github.com/gorilla/websocket"
)

var (
	eng *engine.Engine
	err error
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	eng, err = engine.New()
	if err != nil {
		log.Fatal(err)
	}

	// Web sockets handler
	http.HandleFunc("/ws", wsHandler)

	// Serving static files from public directory
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./public"))))

	log.Print("Websocket running at :8080/ws")

	panic(http.ListenAndServe(":8080", nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Open a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("FAFASF", err)
		return
	}
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	defer conn.Close()

	// Start the engine
	if err := eng.ReadFromWebsocket(conn); err != nil {
		log.Fatal(err)
	}
}
