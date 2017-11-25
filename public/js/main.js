var conn, chessBoard, gameID, playerID,
    newGameBtn = document.getElementById("new-game-btn"),
    log = document.getElementById("console"),
    cfg = {
        draggable: true,
        position: "start",
        onDrop: function(source, target, piece, newPos, oldPos, orientation) {
            // console.log("Source: " + source);
            // console.log("Target: " + target);
            // console.log("Piece: " + piece);
            // console.log("New position: " + ChessBoard.objToFen(newPos));
            // console.log("Old position: " + ChessBoard.objToFen(oldPos));
            // console.log("Orientation: " + orientation);
            // console.log("--------------------");

            conn.send(JSON.stringify({
                type: "make_move",
                data: {
                    "game_id": gameID,
                    "player_id": playerID,
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

            console.log(msg);

            gameID = msg.data["game_id"];


            switch (msg.type) {
                case "game_started":
                    console.log("Game has started");
                    break;
                case "player_left":
                    console.log("Opponent left");
                    break;
                case "move_made":
                    break;
            }
        }
    }
    socket.onclose = function() {
        conn.send(JSON.stringify({
            type: "leave_game",
            data: {
                "player_id": playerID,
                "game_id": gameID,
            },
        }));

        appendLog("Socket closed\n");
    }

    return socket;
}

newGameBtn.addEventListener("click", function(evt) {
    var radios = document.getElementsByName('orientation'),
        orientation;
    for (var i = 0, length = radios.length; i < length; i++) {
        if (radios[i].checked) {
            orientation = radios[i].value;
            break;
        }
    }

    // Choose orientation randomly if not specified by user
    if (orientation === undefined) {
        if (Math.floor(Math.random() * 2) > 0) {
            orientation = "black";
        } else {
            orientation = "white";
        }
    }
    cfg["orientation"] = orientation;
    chessBoard = ChessBoard('board', cfg);

    conn.send(JSON.stringify({
        type: "find_game",
        data: {
            "orientation": cfg["orientation"],
            "player_id": playerID,
        },
    }));

    return false;
});

function generateUUID() { // Public Domain/MIT
    var d = new Date().getTime();
    if (typeof performance !== 'undefined' && typeof performance.now === 'function') {
        d += performance.now(); //use high-precision timer if available
    }
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = (d + Math.random() * 16) % 16 | 0;
        d = Math.floor(d / 16);
        return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
    });
}

function appendLog(msg) {
    console.log(msg);
    log.innerHTML += "" + msg + "\n";
}

window.onload = function() {
    if (window.WebSocket === undefined) {
        appendLog("Your browser does not support WebSockets.");
        return;
    }

    playerID = generateUUID();
    conn = initWebsocket();
};