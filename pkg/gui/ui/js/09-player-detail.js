// ── Player detail ─────────────────────────────────────────────────
function stat(label, value, cls) {
    return '<div class="stat"><div class="stat-label">' + label + '</div><div class="stat-value ' + (cls || '') + '">' + value + '</div></div>';
}

function openPlayer(p) {
    const isCheat = !!confirmationOf(p) || levelOf(p) === 'cheat';
    const hasSteam = !!p.steamId && p.steamId !== "0";
    const sid = String(p.steamId);

    document.getElementById('pmName').innerText = p.name;
    document.getElementById('pmSub').innerText = hasSteam ? 'SteamID64 ' + sid : 'Sin SteamID (bot o demo sin datos)';

    let actions = '';
    if (hasSteam) actions += '<button class="btn" onclick="openSteam(\'' + sid + '\')">' + icon('external', 'ic-sm') + 'Steam</button>';
    if (hasSteam) actions += '<button class="btn" onclick="exportEvidence(\'' + sid + '\')">' + icon('file', 'ic-sm') + 'Exportar evidencia</button>';
    if (hasSteam && isCheat) actions += '<button class="btn btn-danger" onclick="copyBan(\'' + sid + '\')">' + icon('ban', 'ic-sm') + 'Copiar ban</button>';
    actions += '<button class="btn btn-plain btn-icon" onclick="closePlayer()" title="Cerrar">' + icon('x') + '</button>';
    document.getElementById('pmActions').innerHTML = actions;

    const hasTrig = p.triggerAvgMs !== null && p.triggerAvgMs !== undefined;
    let html = '';

    if (p.previousInfractions > 0) {
        html += '<div class="callout callout-red">Tiene ' + p.previousInfractions + (p.previousInfractions === 1 ? ' detección anterior' : ' detecciones anteriores') + ' en demos que analizaste antes.</div>';
    }

    html += '<div><div class="block-title">Resultado</div>' + resultCell(p) +
        (levelOf(p) !== 'limpio' && !confirmationOf(p)
            ? '<div class="muted" style="margin-top:6px; font-size:12px;">' + ({
                cheat: 'Superó un umbral fuerte de detección. Revisá la demo antes de sancionar.',
                sospechoso: 'Hay indicios, pero no alcanzan para decir que es cheat. Conviene mirar la demo.',
                conducta: 'No es cheat: es algo de comportamiento para el staff (AFK, flashes al equipo).'
            })[levelOf(p)] + '</div>'
            : '') +
        '</div>';
    if (hasSteam) html += '<div>' + confirmBlock(sid, p.name, levelOf(p)) + '</div>';
    html += notesBlock(p.notes);
    if (currentReport.map && (currentReport.map.paths || []).some(x => x.steamId === sid)) {
        html += '<div class="callout"><span class="grow">Su recorrido, sus kills y muertes sobre el radar del mapa.</span>' +
            '<button class="btn" onclick="showOnMap(\'' + sid + '\')">Ver en el mapa</button></div>';
    }
    if (p.shiftCases) {
        html += '<div class="callout"><span class="grow">Caminó con shift cerca de un enemigo que no veía en ' + p.shiftCases + (p.shiftCases === 1 ? ' momento' : ' momentos') + ' de la partida.</span>' +
            '<button class="btn" onclick="showShiftOf(\'' + sid + '\')">Ver en el registro</button></div>';
    }

    html += '<div><div class="block-title">Métricas</div><div class="stats">' +
        stat('Saltos', p.totalJumps) +
        stat('Perfs', p.perfRatio.toFixed(1) + '%', p.perfRatio >= 75 ? 't-red' : (p.perfRatio >= 55 ? 't-yellow' : '')) +
        stat('Presión de 1 tick', p.oneTickRatio.toFixed(1) + '%') +
        stat('Racha de perfs', p.maxStreak, p.maxStreak >= 5 ? 't-red' : '') +
        stat('Vel. máxima', p.maxSpeed + ' u/s') +
        stat('Reacción', hasTrig ? p.triggerAvgMs + ' ms' : '–', hasTrig && p.triggerAvgMs < 80 ? 't-red' : '') +
        stat('Snaps', p.snapCount, p.flags && p.flags.includes('AIMLOCK') ? 't-red' : '') +
        stat('Silent aim', p.silentCount, p.silentCount >= 3 ? 't-red' : '') +
        stat('Wallhack', p.wallTrackCount || 0, p.wallTrackCount >= 3 ? 't-red' : '') +
        stat('Macro', p.macroHits || 0, p.macroHits > 0 ? 't-yellow' : '') +
        stat('Kills cegado', p.flashKills || 0, p.flashKills > 0 ? 't-yellow' : '') +
        stat('Kills por humo', p.smokeKills || 0, p.smokeKills > 0 ? 't-yellow' : '') +
        stat('Teamflash', (p.teamflashSecs || 0).toFixed(1) + ' s', p.teamflashSecs >= 35 ? 't-yellow' : '') +
        stat('AFK', p.isAfk ? 'Sí' : 'No', p.isAfk ? 't-yellow' : '') +
        stat('Prefire', p.peekEngagements ? p.prefireRatio + '% de ' + p.peekEngagements : '–', (p.flags || []).includes('PREFIRE') ? 't-red' : (p.prefireRatio >= 30 && p.peekEngagements >= 10 ? 't-yellow' : '')) +
        hitStats(p) +
        '</div></div>';

    if (p.hits) {
        html += '<div><div class="block-title">Dónde pega</div><div id="pmFigure"></div></div>';
    }

    const incs = p.incidents || [];
    if (incs.length && (currentReport.header.roundStarts || []).length) {
        html += '<div><div class="block-title">Incidentes por ronda</div><div class="chart" id="roundChart"></div></div>';
    }

    html += '<div><div class="block-title">Incidentes <span class="count">' + incs.length + '</span></div>';
    if (incs.length) {
        html += '<div class="list">' + incs.slice().sort((a, b) => a.tick - b.tick).map(inc =>
            '<div class="list-row">' +
                '<span class="muted num" style="width:46px; flex-shrink:0;">R' + roundOf(inc.tick) + '</span>' +
                typeTag(inc.type) +
                '<span class="grow" title="' + esc(inc.note) + '">' + esc(inc.note) + '</span>' +
                cmdButton(inc.gotoCmd) + playButton(inc.tick) +
            '</div>'
        ).join('') + '</div>';
    } else {
        html += '<div class="muted">Sin incidentes.</div>';
    }
    html += '</div>';

    html += '<div id="pmProfile"></div>';

    if (p.chatMessages && p.chatMessages.length) {
        html += '<div><div class="block-title">Chat <span class="count">' + p.chatMessages.length + '</span></div><div class="list" style="max-height:180px;">' +
            p.chatMessages.map(m =>
                '<div class="list-row">' +
                    '<span class="muted num" style="width:46px; flex-shrink:0;">R' + m.round + '</span>' +
                    (m.isToxicity ? '<span class="tag tag-red">Tóxico</span>' : '') +
                    '<span class="grow" style="color:var(--text); user-select:text;">' + esc(m.text) + '</span>' +
                    playButton(m.tick) +
                '</div>'
            ).join('') + '</div></div>';
    }

    document.getElementById('pmBody').innerHTML = html;
    document.getElementById('playerModal').style.display = 'flex';
    document.getElementById('pmBody').scrollTop = 0;

    if (document.getElementById('roundChart')) drawRoundChart(document.getElementById('roundChart'), incs);
    const fig = document.getElementById('pmFigure');
    if (fig) {
        renderBodyPanel(fig, { mine: p.hits, mineLabel: 'En esta demo', mode: 'mine' });
        // The general average arrives later; the figure gains the other two modes then
        if (window.getOverview) window.getOverview().then(ov => {
            if (ov && ov.averages && ov.averages.hits) renderBodyPanel(fig, { avg: ov.averages.hits });
        });
    }
    if (hasSteam && window.getPlayerDetail) {
        window.getPlayerDetail(sid).then(d => renderProfile(d));
    }
}

function renderProfile(d) {
    const el = document.getElementById('pmProfile');
    if (!el || !d) return;
    const a = d.averages || {};
    const flagKeys = Object.keys(d.flagCounts || {});
    el.innerHTML = '<div class="block-title">En todas tus demos <span class="count">' + d.demos + '</span></div>' +
        '<dl class="kv">' +
            '<dt>Marcado en</dt><dd>' + (d.flaggedDemos ? '<span class="t-red">' + d.flaggedDemos + ' de ' + d.demos + '</span> · ' + flagKeys.map(f => esc(f) + ' ×' + d.flagCounts[f]).join(', ') : 'ninguna') + '</dd>' +
            '<dt>Perfs promedio</dt><dd>' + fmtNum(a.perfRatio, '%') + '</dd>' +
            '<dt>Reacción promedio</dt><dd>' + fmtNum(a.reactionMs, ' ms') + '</dd>' +
            '<dt>Snaps por demo</dt><dd>' + fmtNum(a.snapsPerDemo) + '</dd>' +
            (d.names.length > 1 ? '<dt>Otros nombres</dt><dd>' + d.names.filter(n => n !== d.name).map(esc).join(', ') + '</dd>' : '') +
        '</dl>' +
        '<div style="margin-top:10px;"><button class="btn" onclick="openPlayerPage(\'' + d.steamId + '\')">Ver ficha completa</button></div>';
}

// Single-series bar chart: incidents per round, with a per-bar hover breakdown.
function drawRoundChart(host, incs) {
    const rounds = currentReport.header.roundStarts.length;
    const buckets = Array.from({ length: rounds + 1 }, () => ({ total: 0, types: {} }));
    incs.forEach(inc => {
        const r = Math.min(roundOf(inc.tick), rounds);
        buckets[r].total++;
        buckets[r].types[inc.type] = (buckets[r].types[inc.type] || 0) + 1;
    });
    const data = buckets.slice(1);
    const max = Math.max(1, ...data.map(d => d.total));

    const W = Math.max(320, host.clientWidth - 24), H = 130, padL = 26, padB = 18, padT = 6;
    const plotW = W - padL, plotH = H - padB - padT;
    const band = plotW / rounds;
    const gap = Math.min(2, band * 0.3);
    const barW = Math.max(1, band - gap);
    const y = v => padT + plotH - (v / max) * plotH;

    let svg = '<svg viewBox="0 0 ' + W + ' ' + H + '" role="img" aria-label="Incidentes por ronda: máximo ' + max + ' en una ronda">';
    svg += '<line class="grid" x1="' + padL + '" x2="' + W + '" y1="' + y(max) + '" y2="' + y(max) + '"/>';
    svg += '<line class="grid" x1="' + padL + '" x2="' + W + '" y1="' + (padT + plotH) + '" y2="' + (padT + plotH) + '"/>';
    svg += '<text class="axis" x="' + (padL - 6) + '" y="' + (y(max) + 4) + '" text-anchor="end">' + max + '</text>';
    svg += '<text class="axis" x="' + (padL - 6) + '" y="' + (padT + plotH + 4) + '" text-anchor="end">0</text>';

    const step = rounds > 24 ? 5 : (rounds > 12 ? 2 : 1);
    data.forEach((d, i) => {
        const x = padL + i * band + gap / 2;
        svg += '<rect class="hit" data-i="' + i + '" x="' + (padL + i * band) + '" y="' + padT + '" width="' + band + '" height="' + plotH + '"/>';
        if (d.total > 0) {
            const top = y(d.total), h = padT + plotH - top, r = Math.min(4, barW / 2, h);
            svg += '<path class="bar" data-i="' + i + '" d="M' + x + ',' + (padT + plotH) + 'V' + (top + r) + 'Q' + x + ',' + top + ' ' + (x + r) + ',' + top +
                'H' + (x + barW - r) + 'Q' + (x + barW) + ',' + top + ' ' + (x + barW) + ',' + (top + r) + 'V' + (padT + plotH) + 'Z"/>';
        }
        if ((i + 1) % step === 0 || i === 0) {
            svg += '<text class="axis" x="' + (padL + i * band + band / 2) + '" y="' + (H - 4) + '" text-anchor="middle">' + (i + 1) + '</text>';
        }
    });
    svg += '</svg><div class="tooltip" id="chartTip"></div>';
    host.innerHTML = svg;

    const tip = host.querySelector('#chartTip');
    host.querySelectorAll('.hit').forEach(hit => {
        hit.addEventListener('mouseenter', () => {
            const i = +hit.getAttribute('data-i');
            const d = data[i];
            host.querySelectorAll('.bar').forEach(b => b.classList.toggle('hl', b.getAttribute('data-i') === String(i)));
            tip.innerHTML = '<div class="row"><b>Ronda ' + (i + 1) + '</b><b class="num">' + d.total + '</b></div>' +
                Object.keys(d.types).sort((a, b) => d.types[b] - d.types[a]).map(t => '<div class="row"><span>' + esc(t) + '</span><span class="num">' + d.types[t] + '</span></div>').join('');
            tip.style.display = 'block';
            const hostBox = host.getBoundingClientRect(), box = hit.getBoundingClientRect();
            let left = box.left - hostBox.left + box.width / 2 - tip.offsetWidth / 2;
            left = Math.max(4, Math.min(left, hostBox.width - tip.offsetWidth - 4));
            tip.style.left = left + 'px';
            tip.style.top = '-' + (tip.offsetHeight - 4) + 'px';
        });
        hit.addEventListener('mouseleave', () => {
            tip.style.display = 'none';
            host.querySelectorAll('.bar.hl').forEach(b => b.classList.remove('hl'));
        });
    });
}

function closePlayer() { document.getElementById('playerModal').style.display = 'none'; }

function findPlayer(sid) { return currentReport.players.find(p => String(p.steamId) === sid); }

function openSteam(sid) { openLink('https://steamcommunity.com/profiles/' + sid); }

function exportEvidence(sid) { if (window.exportPlayerEvidence) window.exportPlayerEvidence(sid); }

function copyBan(sid) {
    const p = findPlayer(sid);
    if (!p) return;
    const reason = (p.flags || []).length ? p.flags.join(' + ') : 'confirmado por staff';
    const cmd = 'sm_ban #' + sid + ' 0 "SourceSight: ' + reason + '"';
    copyText(cmd, 'Comando de ban copiado');
}

// Hit and ping cells of the metrics grid.
function hitStats(p) {
    const t = hitTotals(p.hits);
    let html = '';
    if (t) {
        html += stat('A la cabeza', t.headPct !== null ? t.headPct + '% de ' + t.total : (t.total ? t.head + ' de ' + t.total : '–')) +
            stat('Precisión', t.accuracy !== null ? t.accuracy + '%' : '–') +
            stat('Kills con HS', p.hits.kills ? p.hits.headshotKills + ' de ' + p.hits.kills : '–');
    }
    if (p.pingAvg) {
        html += stat('Ping', p.pingAvg + ' ms' + (p.pingMax > p.pingAvg + 50 ? ' (pico ' + p.pingMax + ')' : ''));
    }
    if (p.lagDiscarded) {
        html += stat('Reacciones con lag', p.lagDiscarded + ' descartadas');
    }
    return html;
}
