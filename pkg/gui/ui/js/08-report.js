// ── Players ───────────────────────────────────────────────────────
function setPlayerFilter(f, el) {
    playerFilter = f;
    document.querySelectorAll('#playerFilter button').forEach(b => b.classList.toggle('on', b === el));
    renderPlayers();
}

function sortPlayers(key) {
    if (sortKey === key) sortAsc = !sortAsc;
    else { sortKey = key; sortAsc = key === 'name'; }
    renderPlayers();
}

function renderPlayers() {
    if (!currentReport) return;
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

    players.sort((a, b) => {
        const va = a[sortKey], vb = b[sortKey];
        if (typeof va === 'string') return sortAsc ? va.localeCompare(vb) : vb.localeCompare(va);
        return sortAsc ? va - vb : vb - va;
    });

    document.getElementById('playersCount').innerText = players.length;
    const body = document.getElementById('playersBody');
    if (players.length === 0) {
        body.innerHTML = emptyRow(8, 'Ningún jugador coincide.');
        return;
    }
    body.innerHTML = '';
    players.forEach(p => {
        const tr = document.createElement('tr');
        tr.className = 'click' + rowMark(p);
        tr.onclick = () => openPlayer(p);
        const perfCls = p.perfRatio >= 75 ? 't-red' : (p.perfRatio >= 55 ? 't-yellow' : '');
        tr.innerHTML =
            '<td class="player">' + esc(p.name) + '</td>' +
            '<td class="r">' + p.totalJumps + '</td>' +
            '<td class="r ' + perfCls + '">' + p.perfRatio.toFixed(1) + '% <span class="muted">' + p.perfectHops + '</span></td>' +
            '<td class="r">' + p.oneTickRatio.toFixed(1) + '%</td>' +
            '<td class="r ' + (p.maxStreak >= 5 ? 't-red' : '') + '">' + p.maxStreak + '</td>' +
            '<td class="r ' + (p.flags && p.flags.includes('AIMLOCK') ? 't-red' : '') + '">' + p.snapCount + '</td>' +
            '<td class="r">' + p.maxSpeed + ' <span class="muted">u/s</span></td>' +
            '<td class="r">' + resultCell(p) + '</td>';
        body.appendChild(tr);
    });
}

function renderDetections() {
    // Everyone worth a look except bhop-only cases, which the players table already shows
    const rows = currentReport.players.filter(p => flagged(p) &&
        (confirmationOf(p) || p.flags.some(f => f !== 'AUTOBUNNY' && f !== 'SOSPECHOSO-BHOP')));
    const body = document.getElementById('detectBody');
    if (rows.length === 0) {
        body.innerHTML = emptyRow(8, 'Nadie superó los umbrales de triggerbot, aimlock, silent aim, macro, wallhack, prefire, kills cegado ni humo.');
        return;
    }
    body.innerHTML = '';
    rows.forEach(p => {
        const tr = document.createElement('tr');
        tr.className = 'click' + rowMark(p);
        tr.onclick = () => openPlayer(p);
        const hasTrig = p.triggerAvgMs !== null && p.triggerAvgMs !== undefined;
        tr.innerHTML =
            '<td class="player">' + esc(p.name) + '</td>' +
            '<td class="r ' + (hasTrig && p.triggerAvgMs < 80 ? 't-red' : '') + '">' + (hasTrig ? p.triggerAvgMs + ' ms <span class="muted">±' + (p.triggerSigma || 0).toFixed(0) + '</span>' : '<span class="muted">–</span>') + '</td>' +
            '<td class="r">' + p.snapCount + ' / ' + p.silentCount + '</td>' +
            '<td class="r">' + (p.macroHits || '<span class="muted">–</span>') + '</td>' +
            '<td class="r">' + p.flashKills + ' / ' + p.smokeKills + '</td>' +
            '<td class="r">' + (p.wallTrackCount || '<span class="muted">–</span>') + '</td>' +
            '<td class="r' + ((p.flags || []).includes('PREFIRE') ? ' t-red' : '') + '">' + (p.peekEngagements ? p.prefireRatio + '%' : '<span class="muted">–</span>') + '</td>' +
            '<td class="r">' + resultCell(p) + '</td>';
        body.appendChild(tr);
    });
}

function renderTimeline() {
    const all = currentReport.timeline || [];
    const card = document.getElementById('timelineCard');
    card.style.display = all.length ? 'block' : 'none';
    if (!all.length) return;

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
