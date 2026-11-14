// ── Players ───────────────────────────────────────────────────────
function setPlayerFilter(f, el) {
    playerFilter = f;
    document.querySelectorAll('#playerFilter button').forEach(b => b.classList.toggle('on', b === el));
    renderPlayers();
}

// Sort value of each column. Missing values (no reaction measured, too few hits…)
// always go last, whatever the direction.
const LEVEL_RANK = { confirmado: 5, cheat: 4, sospechoso: 3, conducta: 2, limpio: 1 };
const PLAYER_SORT = {
    name: p => p.name.toLowerCase(),
    level: p => (confirmationOf(p) ? LEVEL_RANK.confirmado : LEVEL_RANK[levelOf(p)] || 0) * 1000 + (p.flags || []).length,
    perfRatio: p => p.totalJumps >= 5 ? p.perfRatio : null,
    reaction: p => p.triggerAvgMs === null || p.triggerAvgMs === undefined ? null : p.triggerAvgMs,
    prefire: p => p.peekEngagements ? p.prefireRatio : null,
    head: p => { const t = hitTotals(p.hits); return t ? t.headPct : null; },
    snapCount: p => p.snapCount,
    shiftCases: p => p.shiftCases || 0,
    ping: p => p.pingAvg || null
};

function sortPlayers(key) {
    if (sortKey === key) sortAsc = !sortAsc;
    // Names and reaction times read best ascending (fastest first); the rest descending
    else { sortKey = key; sortAsc = key === 'name' || key === 'reaction'; }
    renderPlayers();
}

function renderPlayers() {
    if (!currentReport) return;
    if (!PLAYER_SORT[sortKey]) sortKey = 'level';
    document.querySelectorAll('#playersCard th.sortable').forEach(th => {
        const on = th.getAttribute('data-key') === sortKey;
        th.classList.toggle('sorted', on);
        th.classList.toggle('asc', on && sortAsc);
    });

    const q = document.getElementById('searchPlayer').value.toLowerCase();
    let players = currentReport.players.slice();
    if (q) players = players.filter(p => p.name.toLowerCase().includes(q));
    if (playerFilter === 'cheat') players = players.filter(p => confirmationOf(p) || levelOf(p) === 'cheat');
    if (playerFilter === 'suspect') players = players.filter(p => !confirmationOf(p) && levelOf(p) === 'sospechoso');
    if (playerFilter === 'clean') players = players.filter(p => !flagged(p));

    const get = PLAYER_SORT[sortKey];
    players.sort((a, b) => {
        const va = get(a), vb = get(b);
        if (va === null && vb === null) return a.name.localeCompare(b.name);
        if (va === null) return 1;
        if (vb === null) return -1;
        const c = typeof va === 'string' ? va.localeCompare(vb) : va - vb;
        return (sortAsc ? c : -c) || a.name.localeCompare(b.name);
    });

    document.getElementById('playersCount').innerText = players.length;
    const body = document.getElementById('playersBody');
    if (players.length === 0) {
        body.innerHTML = emptyRow(9, 'Ningún jugador coincide.');
        return;
    }
    const dash = '<span class="muted">–</span>';
    body.innerHTML = '';
    players.forEach(p => {
        const tr = document.createElement('tr');
        tr.className = 'click' + rowMark(p);
        tr.onclick = () => openPlayer(p);
        const flags = p.flags || [];
        const hasTrig = p.triggerAvgMs !== null && p.triggerAvgMs !== undefined;
        const t = hitTotals(p.hits);
        const perfCls = p.perfRatio >= 75 ? 't-red' : (p.perfRatio >= 55 ? 't-yellow' : '');
        tr.innerHTML =
            '<td class="player">' + playerLabel(p.name) + '</td>' +
            '<td>' + resultCell(p) + '</td>' +
            '<td class="r ' + (p.totalJumps >= 5 ? perfCls : '') + '">' + (p.totalJumps >= 5 ? p.perfRatio.toFixed(1) + '% <span class="muted">de ' + p.totalJumps + '</span>' : dash) + '</td>' +
            '<td class="r ' + (hasTrig && p.triggerAvgMs < 80 ? 't-red' : '') + '">' + (hasTrig ? p.triggerAvgMs + ' ms' : dash) + '</td>' +
            '<td class="r' + (flags.includes('PREFIRE') ? ' t-red' : '') + '">' + (p.peekEngagements ? p.prefireRatio + '%' : dash) + '</td>' +
            '<td class="r">' + (t && t.headPct !== null ? t.headPct + '%' : dash) + '</td>' +
            '<td class="r ' + (flags.includes('AIMLOCK') ? 't-red' : (flags.includes('SOSPECHOSO-AIM') || flags.includes('POSIBLE-LAG')) ? 't-yellow' : '') + '">' + (p.snapCount || dash) + '</td>' +
            '<td class="r">' + (p.shiftCases || dash) + '</td>' +
            '<td class="r muted">' + (p.pingAvg ? p.pingAvg + ' ms' : '–') + '</td>';
        body.appendChild(tr);
    });
}

function renderTimeline() {
    const all = currentReport.timeline || [];
    if (!all.length) {
        document.getElementById('timelineCount').innerText = '';
        document.getElementById('timelineBody').innerHTML = emptyRow(7, 'Ningún jugador superó un umbral de detección en esta demo.');
        document.getElementById('timelineMore').style.display = 'none';
        return;
    }

    const shown = Math.min(timelineShown, all.length);
    document.getElementById('timelineCount').innerText = all.length;
    document.getElementById('timelineBody').innerHTML = all.slice(0, shown).map(inc =>
        '<tr>' +
            '<td class="r muted">' + roundOf(inc.tick) + '</td>' +
            '<td>' + typeTag(inc.type) + '</td>' +
            '<td class="player">' + esc(inc.player) + '</td>' +
            '<td>' + cmdButton(inc.gotoCmd) + '</td>' +
            '<td>' + playButton(inc.tick, inc.steamId) + '</td>' +
            '<td class="wrap muted" style="min-width:260px;">' + esc(inc.note) + '</td>' +
        '</tr>'
    ).join('');
    document.getElementById('timelineMore').style.display = shown < all.length ? 'block' : 'none';
}

function showMoreTimeline() {
    timelineShown += 25;
    renderTimeline();
}

function renderAdmin() {
    const log = currentReport.adminLog || [];
    document.getElementById('adminCount').innerText = log.length || '';
    const body = document.getElementById('adminBody');
    if (!log.length) {
        body.innerHTML = emptyRow(6, 'No hubo comandos de castigo de admins en esta demo.');
        return;
    }
    body.innerHTML = log.map(ev =>
        '<tr class="' + (ev.isRevenge ? 'mark-red' : '') + '">' +
            '<td class="r muted">' + ev.round + '</td>' +
            '<td class="player">' + esc(ev.adminName) + '</td>' +
            '<td><code>' + esc(ev.command) + '</code></td>' +
            '<td>' + esc(ev.targetName) + '</td>' +
            '<td class="wrap">' + (ev.isRevenge ? '<span class="tag tag-red">Venganza</span> ' : '') + '<span class="muted">' + esc(ev.note || '') + '</span></td>' +
            '<td class="r">' + (ev.tick ? playButton(ev.tick) : '') + '</td>' +
        '</tr>'
    ).join('');
}

function setChatFilter(f, el) {
    chatFilter = f;
    document.querySelectorAll('#chatFilter button').forEach(b => b.classList.toggle('on', b === el));
    renderChat();
}

function renderChat() {
    const all = currentReport.chatLog || [];
    let msgs = all;
    if (chatFilter === 'toxic') msgs = all.filter(m => m.isToxicity);
    if (chatFilter === 'ghosting') msgs = all.filter(m => m.isGhosting);
    if (chatFilter === 'spam') msgs = all.filter(m => m.category === 'SPAM BIND');

    document.getElementById('chatCount').innerText = all.length ? (msgs.length === all.length ? all.length : msgs.length + ' de ' + all.length) : '';
    const body = document.getElementById('chatBody');
    if (!all.length) { body.innerHTML = emptyRow(7, 'La demo no tiene mensajes de chat.'); return; }
    if (!msgs.length) { body.innerHTML = emptyRow(7, 'Ningún mensaje de este tipo.'); return; }

    body.innerHTML = msgs.map(m => {
        let tag = '<span class="muted">–</span>';
        if (m.isToxicity) tag = '<span class="tag tag-red">Tóxico</span>';
        else if (m.isGhosting) tag = '<span class="tag tag-yellow">Ghosting</span>';
        else if (m.category === 'SPAM BIND') tag = '<span class="tag tag-yellow">Spam</span>';
        return '<tr class="' + (m.isToxicity ? 'mark-red' : '') + '">' +
            '<td class="r muted">' + m.round + '</td>' +
            '<td class="player">' + esc(m.senderName) + '</td>' +
            '<td class="muted">' + (m.isTeamOnly ? 'Equipo' : 'Global') + '</td>' +
            '<td class="wrap" style="min-width:240px; color:var(--text); user-select:text;">' + esc(m.text) + '</td>' +
            '<td>' + tag + '</td>' +
            '<td>' + cmdButton(m.gotoCmd, String(m.tick)) + '</td>' +
            '<td class="r">' + playButton(m.tick, m.senderSteamId) + '</td>' +
        '</tr>';
    }).join('');
}
