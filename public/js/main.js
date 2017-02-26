var conn;
var chessBoard;
var board;
var log = document.getElementById("log");
var cfg = {
  draggable: true,
  position: "start",
  orientation: "white",
  onDrop: function(source, target, piece, newPos, oldPos, orientation) {
    // console.log("Source: " + source);
    // console.log("Target: " + target);
    // console.log("Piece: " + piece);
    // console.log("New position: " + ChessBoard.objToFen(newPos));
    // console.log("Old position: " + ChessBoard.objToFen(oldPos));
    // console.log("Orientation: " + orientation);
    // console.log("--------------------");

    conn.send(JSON.stringify({
      "type": "make_move",
      "data": {
        "board_id": board["id"],
        "source": source,
        "target": target,
        "piece": piece,
      },
    }));
  },
};

function initWebsocket() {
  var wsHost = "localhost:8080",
      socket = new WebSocket("ws://" + wsHost + "/ws");

  socket.onopen = function() {
    appendLog("Socket is open");

    conn.send(JSON.stringify({
      "type": "new_game",
      "data": {
        "id": generateUUID(),
        "orientation": cfg["orientation"],
        "position": "",
      },
    }));

  };
  socket.onmessage = function(evt) {
    var messages = evt.data.split("\n");
    for (var i = 0; i < messages.length; i++) {
        var msg = messages[i];

        try {
          var msg = JSON.parse(msg);
        } catch (e) {
          return;
        }

        if (msg === null) {
          return;
        }

        appendLog("" + msg);

        switch(msg["type"]) {

          case "new_game":
            board = msg["data"];
            conn.send(JSON.stringify({
              "type": "join_game",
              "data": {
                "board_id": board["id"],
                "player_id": "Richard",
              },
            }));
            break;

          case "join_game":
            break;

          case "make_move":
            break;

          default:
            console.log("Unknown message type")
        }
    }
  }
  socket.onclose = function() {
    appendLog("Socket closed\n");
  }

  return socket;
}

function generateUUID () { // Public Domain/MIT
  var d = new Date().getTime();
  if (typeof performance !== 'undefined' && typeof performance.now === 'function'){
    d += performance.now(); //use high-precision timer if available
  }
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    var r = (d + Math.random() * 16) % 16 | 0;
    d = Math.floor(d / 16);
    return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
  });
}

function appendLog(msg) {
  console.log(msg);
  log.innerHTML += msg + "\n";
}

window.onload = function () {
  if (window.WebSocket === undefined) {
    appendLog("Your browser does not support WebSockets.");
    return;
  } else {
    conn = initWebsocket();
    chessBoard = ChessBoard('board', cfg);
  }
};
