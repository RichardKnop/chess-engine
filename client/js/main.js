var conn,
    board,
    newGameBtn = document.getElementById('new-game-btn'),
    log = document.getElementById('console'),
    player = {
        ID: generateUUID(),
    },
    game = {
        ID: getQueryStringParam('game_id'),
        started: false,
        myTurn: false,
    },
    cfg = {
        draggable: true,
        onDrop: function(source, target, piece, newPos, oldPos, orientation) {
            if (!game.started || !game.myTurn || (newPos === oldPos)) {
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
    console.log("Player ID: ", player.ID);

    var wsHost = 'localhost:8080',
        socket = new ReconnectingWebSocket('ws://' + wsHost + '/ws');

    socket.onopen = function() {
        console.log('Connection established');

        if (game.ID) {
            var orientation = getOrientation();
            setOrientation(orientation);

            conn.send(JSON.stringify({
                type: 'get_game',
                data: {
                    'game_id': game.ID,
                    'player_id': player.ID,
                    'orientation': cfg.orientation,
                },
            }));
        }
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

            switch (msg.type) {
                case 'state_update':
                    if (!board) {
                        board = ChessBoard('board', cfg);
                    }

                    game.ID = msg.data['game_id'];

                    // Append game ID to URL
                    setQueryStringParams({ 'game_id': game.ID, 'orientation': cfg.orientation });

                    // Set game.myTurn
                    game.myTurn = msg.data['player_id'] == player.ID;

                    break;
                case 'game_started':
                    if (!game.started) {
                        game.started = true;

                        // Set board position
                        board.position(msg.data['position']);

                        appendLog('Game started.');
                    }
                    break;
                case 'move_made':
                    if (board.fen() !== msg.data['position']) {
                        board.position(msg.data['position']);
                        game.myTurn = !game.myTurn;
                    }
                    break;
            }
        }
    }
    socket.onclose = function(e) {
        switch (e) {
            case 1000: // CLOSE_NORMAL
                console.log('Connection closed');
                break;
            default: // Abnormal closure
                console.log(e);
                break;
        }
    }

    return socket;
}

newGameBtn.addEventListener('click', function(evt) {
    // Reset game data
    game = {
        ID: null,
        started: false,
        myTurn: false,
    }

    // Reset query string params
    setQueryStringParams({});

    var orientation = getOrientation();
    setOrientation(orientation);

    board = ChessBoard('board', cfg);

    conn.send(JSON.stringify({
        type: 'find_game',
        data: {
            'player_id': player.ID,
            'orientation': cfg.orientation,
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
    // First look in the query string
    var param = getQueryStringParam('orientation');
    if (param) {
        return param;
    }

    // Second, look at radio buttons
    var radios = document.getElementsByName('orientation')
    for (var i = 0, length = radios.length; i < length; i++) {
        if (radios[i].checked) {
            return radios[i].value;
        }
    }

    // Finally, choose orientation randomly if not specified by user
    if (Math.floor(Math.random() * 2) > 0) {
        return 'black';
    }
    return 'white';
}

function setOrientation(orientation) {
    cfg.orientation = orientation;

    var radios = document.getElementsByName('orientation')
    for (var i = 0, length = radios.length; i < length; i++) {
        if (radios[i].value === orientation && !radios[i].checked) {
            radios[i].checked = true;
            break;
        }
    }
}

function getQueryStringParam(key) {
    if ('URLSearchParams' in window) {
        var searchParams = new URLSearchParams((new URL(window.location.href)).search);
        return searchParams.get(key);
    }
    return null;
}

function setQueryStringParams(params) {
    if (history.pushState && 'URLSearchParams' in window) {
        var searchParams = new URLSearchParams();
        for (var key in params) {
            searchParams.set(key, params[key]);
        }
        var newurl = window.location.protocol + "//" + window.location.host + window.location.pathname + '?' + searchParams.toString();
        window.history.pushState({ path: newurl }, '', newurl);
    }
}

window.onload = function() {
    if (window.WebSocket === undefined) {
        appendLog('Your browser does not support WebSockets.');
        return;
    }

    conn = initWebsocket();
};