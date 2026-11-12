// ── SHIFT log: walking near enemies the player can't see ───────────────
// One row per case (pkg/core/shift.go): every shift press while the same enemies
// stay close is grouped in it. Informative only: it never changes a player's level.

let shiftPlayer = '';   // filter: SteamID, or '' for everyone
let shiftShown = 25;

function shiftLog() { return (currentReport && currentReport.shiftLog) || []; }

function setShiftPlayer(steamId) {
    shiftPlayer = steamId;
    shiftShown = 25;
    renderShift();
}

function showMoreShift() {
    shiftShown += 25;
    renderShift();
}

// Opens the log filtered to one player and scrolls to it (from the player detail).
function showShiftOf(steamId) {
    closePlayer();
    shiftPlayer = String(steamId);
    shiftShown = 25;
    renderShift();
    document.getElementById('shiftCard').scrollIntoView({ behavior: 'smooth', block: 'start' });
}

function renderShift() {
    const all = shiftLog();

    // Player filter: everyone who appears in the log, most cases first
    const per = {};
    all.forEach(c => {
        const p = per[c.steamId] || (per[c.steamId] = { name: c.player, n: 0 });
        p.n++;
    });
    const ids = Object.keys(per).sort((a, b) => per[b].n - per[a].n || per[a].name.localeCompare(per[b].name));
    if (shiftPlayer && !per[shiftPlayer]) shiftPlayer = '';
    document.getElementById('shiftPlayer').innerHTML =
        '<option value="">Todos los jugadores</option>' +
        ids.map(id => '<option value="' + id + '"' + (id === shiftPlayer ? ' selected' : '') + '>' + esc(per[id].name) + ' (' + per[id].n + ')</option>').join('');

    const rows = shiftPlayer ? all.filter(c => c.steamId === shiftPlayer) : all;
    document.getElementById('shiftCount').innerText = all.length ? (rows.length === all.length ? all.length : rows.length + ' de ' + all.length) : '';
    document.getElementById('shiftNote').innerText = all.length
        ? 'Cada fila es un momento en que el jugador caminó con shift teniendo cerca un enemigo que no veía. Las apretadas seguidas con el mismo enemigo cerca van juntas. Es normal en el juego: no cambia el resultado, es para revisar.'
        : '';

    const body = document.getElementById('shiftBody');
    if (!all.length) {
        body.innerHTML = emptyRow(8, currentReport && currentReport.shiftLog === undefined
            ? 'Esta demo se analizó con una versión anterior. Re-analizala para ver el registro.'
            : 'Nadie caminó con shift cerca de un enemigo sin verlo.');
        document.getElementById('shiftMore').style.display = 'none';
        return;
    }
    const shown = rows.slice(0, shiftShown);
    body.innerHTML = shown.map(c =>
        '<tr>' +
            '<td class="r muted">' + c.round + '</td>' +
            '<td class="player">' + esc(c.player) + '</td>' +
            '<td class="wrap">' + (c.enemies || []).map(esc).join(', ') + '</td>' +
            '<td class="r num">' + c.presses + '</td>' +
            '<td class="r num">' + c.heldSecs.toLocaleString('es-AR') + ' s</td>' +
            '<td class="r num">' + c.minDist + ' u</td>' +
            '<td>' + cmdButton(c.gotoCmd) + '</td>' +
            '<td class="r">' + playButton(c.tick) + '</td>' +
        '</tr>'
    ).join('');
    document.getElementById('shiftMore').style.display = shown.length < rows.length ? 'block' : 'none';
}
