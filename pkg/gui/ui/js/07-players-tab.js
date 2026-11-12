// ── Players across all demos ──────────────────────────────────────
let playerRows = [];
let activePlayer = null;
let searchTimer = null;

function setAsideTab(tab) {
    asideTab = tab;
    document.getElementById('tabDemos').classList.toggle('on', tab === 'demos');
    document.getElementById('tabPlayers').classList.toggle('on', tab === 'players');
    document.getElementById('paneDemos').style.display = tab === 'demos' ? 'flex' : 'none';
    document.getElementById('panePlayers').style.display = tab === 'players' ? 'flex' : 'none';
    if (tab === 'players') {
        searchPlayers();
        if (activePlayer) openPlayerPage(activePlayer); else showOverview();
    } else {
        restoreIdle();
    }
}

function searchPlayers() {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
        if (!window.listPlayers) return;
        const q = document.getElementById('playerSearch').value;
        window.listPlayers(q).then(res => {
            playerRows = res.rows || [];
            document.getElementById('playerHint').innerText = res.hint ||
                (res.steamId && !playerRows.length ? 'Ese SteamID no aparece en ninguna demo analizada.' : '');
            renderPlayerList();
            if (res.steamId && playerRows.length === 1) openPlayerPage(playerRows[0].steamId);
        });
    }, 150);
}

function renderPlayerList() {
    const list = document.getElementById('playerList');
    if (!playerRows.length) {
        list.innerHTML = '<div class="hist-empty">' + (historyItems.length ? 'Ningún jugador coincide.' : 'Analizá demos para ver jugadores acá.') + '</div>';
        return;
    }
    list.innerHTML = playerRows.slice(0, 300).map(r => {
        const conf = confirmedMap[r.steamId];
        const tags = (conf ? confirmedTag() : '') + (r.flagTypes || []).map(flagTag).join('');
        const mark = conf || r.level === 'cheat' ? ' flagged' : (r.level === 'sospechoso' ? ' suspect' : '');
        return '<div class="hist-item' + (r.steamId === activePlayer ? ' active' : '') + mark + '" tabindex="0" onclick="openPlayerPage(\'' + r.steamId + '\')" onkeydown="if(event.key===\'Enter\')openPlayerPage(\'' + r.steamId + '\')">' +
            '<div class="hist-name">' + esc(r.name) + '</div>' +
            '<div class="hist-sub"><span>' + r.demos + (r.demos === 1 ? ' demo' : ' demos') + '</span>' +
            (r.flaggedDemos ? (r.level === 'cheat' ? '<span class="hist-flag">cheat en ' : '<span class="t-yellow">sospechoso en ') + r.flaggedDemos + '</span>' : (r.level === 'conducta' ? '<span>conducta</span>' : '')) + '</div>' +
            (tags ? '<div class="hist-tags">' + tags + '</div>' : '') +
            '</div>';
    }).join('');
}

function fmtDate(ts) { return (ts || '').slice(0, 10).split('-').reverse().join('/'); }
function fmtNum(v, suffix) { return v === null || v === undefined ? '–' : v + (suffix || ''); }

// One row of "this player vs everyone": the player's value turns red when it is
// well above the average of all players (twice as high, or +15 points of perfs).
function compareRow(label, mine, avg, suffix, higherIsWorse, perfLike) {
    let cls = '';
    if (mine !== null && mine !== undefined && avg !== null && avg !== undefined && higherIsWorse) {
        const worse = perfLike ? mine - avg >= 15 : (avg > 0 ? mine >= avg * 2 && mine - avg >= 1 : mine >= 1);
        if (worse) cls = 't-red';
    }
    return '<tr><td>' + label + '</td><td class="r ' + cls + '">' + fmtNum(mine, suffix) + '</td><td class="r muted">' + fmtNum(avg, suffix) + '</td></tr>';
}

function showOverview() {
    activePlayer = null;
    renderPlayerList();
    if (!window.getOverview) return;
    window.getOverview().then(ov => {
        const el = document.getElementById('overviewView');
        const a = ov.averages || {};
        const flags = Object.keys(ov.flagCounts || {}).sort((x, y) => ov.flagCounts[y] - ov.flagCounts[x]);
        el.innerHTML =
            '<div class="page-head"><div><div class="page-title">Todas las demos</div>' +
            '<div class="page-sub">Promedios de cada jugador en cada demo del historial. Elegí un jugador a la izquierda o buscalo por SteamID.</div></div></div>' +
            '<div class="summary">' +
                '<div class="sum-cell"><div class="sum-label">Demos</div><div class="sum-value">' + ov.demos + '</div><div class="sum-sub">' + ov.cheatDemos + ' con cheat · ' + ov.suspectDemos + ' sospechosas</div></div>' +
                '<div class="sum-cell"><div class="sum-label">Jugadores distintos</div><div class="sum-value">' + ov.players + '</div><div class="sum-sub">' +
                    ov.cheatPlayers + ' cheat · ' + ov.suspectPlayers + ' sospechosos · ' + Object.keys(confirmedMap).length + ' confirmados</div></div>' +
                '<div class="sum-cell"><div class="sum-label">Perfs promedio</div><div class="sum-value">' + fmtNum(a.perfRatio, '%') + '</div><div class="sum-sub">jugadores con 5 saltos o más</div></div>' +
                '<div class="sum-cell"><div class="sum-label">Reacción promedio</div><div class="sum-value">' + fmtNum(a.reactionMs, ' ms') + '</div><div class="sum-sub">' + fmtNum(a.snapsPerDemo) + ' snaps por demo</div></div>' +
            '</div>' +
            '<section class="panel"><div class="panel-head"><div class="panel-title">Detecciones por categoría</div></div>' +
            '<div class="table-wrap"><table><thead><tr><th>Categoría</th><th class="r">Veces</th></tr></thead><tbody>' +
            (flags.length ? flags.map(f => '<tr><td>' + flagTag(f) + '</td><td class="r">' + ov.flagCounts[f] + '</td></tr>').join('')
                          : emptyRow(2, 'Todavía no hay detecciones en el historial.')) +
            '</tbody></table></div></section>' +
            '<section class="panel"><div class="panel-head"><div class="panel-title">Dónde pegan los jugadores</div><span class="muted">Impactos a enemigos en todas las demos</span></div><div class="panel-body" id="ovFigure"></div></section>';
        renderBodyPanel(document.getElementById('ovFigure'), { avg: a.hits });
        showView('overview');
    });
}

function openPlayerPage(steamId) {
    if (!window.getPlayerDetail) return;
    closePlayer();
    if (asideTab !== 'players') {
        asideTab = 'players';
        document.getElementById('tabDemos').classList.remove('on');
        document.getElementById('tabPlayers').classList.add('on');
        document.getElementById('paneDemos').style.display = 'none';
        document.getElementById('panePlayers').style.display = 'flex';
        searchPlayers();
    }
    activePlayer = steamId;
    renderPlayerList();
    window.getPlayerDetail(steamId).then(d => {
        const el = document.getElementById('playerView');
        if (!d) {
            el.innerHTML = '<div class="muted">Este jugador no aparece en el historial.</div>';
            showView('player');
            return;
        }
        const a = d.averages || {}, g = (d.global && d.global.averages) || {};
        const flagKeys = Object.keys(d.flagCounts || {});
        el.innerHTML =
            '<div class="page-head"><div style="min-width:0;"><div class="page-title">' + esc(d.name) + '</div>' +
            '<div class="page-sub mono">SteamID64 ' + esc(d.steamId) + '</div>' +
            (d.names.length > 1 ? '<div class="page-sub">También jugó como: ' + d.names.filter(n => n !== d.name).map(esc).join(', ') + '</div>' : '') +
            '</div><div class="modal-actions">' +
                '<button class="btn" onclick="copyText(\'' + d.steamId + '\', \'SteamID copiado\')">' + icon('copy', 'ic-sm') + 'SteamID</button>' +
                '<button class="btn" onclick="openSteam(\'' + d.steamId + '\')">' + icon('external', 'ic-sm') + 'Steam</button>' +
            '</div></div>' +
            '<div style="display:flex; flex-direction:column; gap:8px;">' +
                '<div>' + (confirmedMap[d.steamId] ? confirmedTag() : levelTag(d.level)) +
                ' <span class="muted" style="font-size:12px;">peor resultado en todas sus demos</span></div>' +
                confirmBlock(d.steamId, d.name, d.level) +
            '</div>' +
            '<div class="summary">' +
                '<div class="sum-cell"><div class="sum-label">Demos</div><div class="sum-value">' + d.demos + '</div><div class="sum-sub">del ' + fmtDate(d.firstSeen) + ' al ' + fmtDate(d.lastSeen) + '</div></div>' +
                '<div class="sum-cell"><div class="sum-label">Marcado en</div><div class="sum-value' + (d.flaggedDemos ? ' t-red' : '') + '">' + d.flaggedDemos + '</div><div class="sum-sub">' + (flagKeys.length ? flagKeys.map(f => f + ' ×' + d.flagCounts[f]).join(', ') : 'ninguna demo') + '</div></div>' +
                '<div class="sum-cell"><div class="sum-label">Incidentes</div><div class="sum-value">' + d.totals.incidents + '</div><div class="sum-sub">' + fmtNum(a.incidentsPerDemo) + ' por demo</div></div>' +
                '<div class="sum-cell"><div class="sum-label">Kills cegado / humo</div><div class="sum-value">' + d.totals.blind + ' / ' + d.totals.smoke + '</div><div class="sum-sub">macro: ' + d.totals.macro + '</div></div>' +
            '</div>' +
            '<section class="panel"><div class="panel-head"><div class="panel-title">Promedio del jugador vs. todos</div><span class="muted">En rojo, lo que está muy por encima del promedio general</span></div>' +
            '<div class="table-wrap"><table class="compare"><thead><tr><th></th><th class="r">Este jugador</th><th class="r">Promedio general</th></tr></thead><tbody>' +
                compareRow('Perfs de bhop', a.perfRatio, g.perfRatio, '%', true, true) +
                compareRow('Reacción (triggerbot)', a.reactionMs, g.reactionMs, ' ms', false) +
                compareRow('Prefire: daño apenas aparece el rival', a.prefirePct, g.prefirePct, '%', true, true) +
                compareRow('Impactos a la cabeza', a.headshotPct, g.headshotPct, '%', true, true) +
                compareRow('Ping promedio', a.pingAvg, g.pingAvg, ' ms', false) +
                compareRow('Snaps por demo', a.snapsPerDemo, g.snapsPerDemo, '', true) +
                compareRow('Silent aim por demo', a.silentPerDemo, g.silentPerDemo, '', true) +
                compareRow('Wallhack por demo', a.wallPerDemo, g.wallPerDemo, '', true) +
                compareRow('Incidentes por demo', a.incidentsPerDemo, g.incidentsPerDemo, '', true) +
            '</tbody></table></div></section>' +
            '<section class="panel"><div class="panel-head"><div class="panel-title">Dónde pega</div><span class="muted">Impactos a enemigos en todas sus demos</span></div><div class="panel-body" id="ppFigure"></div></section>' +
            '<section class="panel"><div class="panel-head"><div class="panel-title">Demos <span class="count">' + d.demos + '</span></div><span class="muted">Click para abrir la demo</span></div>' +
            '<div class="table-wrap"><table><thead><tr><th>Fecha</th><th>Demo</th><th>Mapa</th><th class="r">Perfs</th><th class="r">Reacción</th><th class="r">Cabeza</th><th class="r">Snaps</th><th class="r">Wallhack</th><th class="r">Resultado</th></tr></thead><tbody>' +
            d.appearances.map(ap =>
                '<tr class="click' + (ap.flags && ap.flags.length ? ' mark-red' : '') + '" onclick="openDemoFromPlayer(\'' + ap.entryId + '\')">' +
                    '<td class="num">' + fmtDate(ap.analyzedAt) + '</td>' +
                    '<td class="player">' + esc(ap.demoName) + '</td>' +
                    '<td class="muted">' + esc(ap.mapName) + '</td>' +
                    '<td class="r">' + (ap.totalJumps >= 5 ? ap.perfRatio.toFixed(1) + '%' : '<span class="muted">–</span>') + '</td>' +
                    '<td class="r">' + fmtNum(ap.reactionMs, ' ms') + '</td>' +
                    '<td class="r">' + fmtNum(ap.headshotPct, '%') + '</td>' +
                    '<td class="r">' + ap.snaps + '</td>' +
                    '<td class="r">' + ap.wallhack + '</td>' +
                    '<td class="r">' + (ap.flags && ap.flags.length ? '<span class="tags">' + ap.flags.map(flagTag).join('') + '</span>' : '<span class="muted">Limpio</span>') + '</td>' +
                '</tr>'
            ).join('') +
            '</tbody></table></div></section>';
        renderBodyPanel(document.getElementById('ppFigure'), { mine: a.hits, mineLabel: 'Este jugador', avg: g.hits, mode: 'compare' });
        showView('player');
    });
}

function openDemoFromPlayer(entryId) {
    asideTab = 'demos';
    document.getElementById('tabDemos').classList.add('on');
    document.getElementById('tabPlayers').classList.remove('on');
    document.getElementById('paneDemos').style.display = 'flex';
    document.getElementById('panePlayers').style.display = 'none';
    openHistoryItem(entryId);
}
