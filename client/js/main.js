var conn,
    board,
    newGameBtn = document.getElementById('new-game-btn'),
    log = document.getElementById('console'),
    player = {
        ID: generateUUID(),
    },
    game = {
        ID: '', // will be populated later
        started: false,
        myTurn: false,
    },
    cfg = {
        draggable: true,
        onDrop: function(source, target, piece, newPos, oldPos, orientation) {
            if (!game.started || !game.myTurn) {
                // http://chessboardjs.com/docs#config:onDrop
                return 'snapback';
            }
            game.myTurn = !game.myTurn;
            conn.send(JSON.stringify({
                type: 'make_move',
                data: {
                    'game_id': game.ID,
                    'player_id': player.ID,
                    'source': source,
                    'target': target,
                    'piece': piece,
                    'old_position': ChessBoard.objToFen(oldPos),
                    'new_position': ChessBoard.objToFen(newPos),
                },
            }));
        },
    };

function initWebsocket() {
    var wsHost = 'localhost:8080',
        socket = new WebSocket('ws://' + wsHost + '/ws');

    socket.onopen = function() {
        appendLog('Socket is open');
    };
    socket.onmessage = function(evt) {
        var messages = evt.data.split('\n');
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

            game.ID = msg.data['game_id'];

            switch (msg.type) {
                case 'game_started':
                    appendLog('Game started.')
                    board.position('start');
                    game.started = true;
                    if (cfg.orientation === 'white') {
                        game.myTurn = true;
                    }
                    break;
                case 'player_left':
                    appendLog('Opponent left');
                    break;
                case 'move_made':
                    board.position(msg.data['position']);
                    game.myTurn = !game.myTurn;
                    break;
            }
        }
    }
    socket.onclose = function() {
        conn.send(JSON.stringify({
            type: 'leave_game',
            data: {
                'player_id': player['ID'],
                'game_id': gameID,
            },
        }));

        appendLog('Socket closed\n');
    }

    return socket;
}

newGameBtn.addEventListener('click', function(evt) {
    var orientation = getOrientation();

    // Choose orientation randomly if not specified by user
    if (orientation === undefined) {
        if (Math.floor(Math.random() * 2) > 0) {
            orientation = 'black';
        } else {
            orientation = 'white';
        }
    }
    setOrientation(orientation);
    cfg.orientation = orientation;

    board = ChessBoard('board', cfg);

    conn.send(JSON.stringify({
        type: 'find_game',
        data: {
            'orientation': cfg.orientation,
            'player_id': player.ID,
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
    log.innerHTML += '' + msg + '\n';
}

function getOrientation() {
    var radios = document.getElementsByName('orientation')
    for (var i = 0, length = radios.length; i < length; i++) {
        if (radios[i].checked) {
            return radios[i].value;
        }
    }
    return undefined;
}

function setOrientation(orientation) {
    var radios = document.getElementsByName('orientation')
    for (var i = 0, length = radios.length; i < length; i++) {
        if (radios[i].value === orientation && !radios[i].checked) {
            radios[i].checked = true;
            break;
        }
    }
}

window.onload = function() {
    if (window.WebSocket === undefined) {
        appendLog('Your browser does not support WebSockets.');
        return;
    }

    conn = initWebsocket();
};