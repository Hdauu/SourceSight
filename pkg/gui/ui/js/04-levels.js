// ── Levels & staff confirmations ──────────────────────────────────
// Levels come from Go (core/level.go); "confirmed" is a staff decision.
let FLAG_LEVELS = {};
let confirmedMap = {};
const LEVEL_LABEL = { cheat: 'Cheat detectado', sospechoso: 'Sospechoso', conducta: 'Conducta', limpio: 'Limpio' };
const LEVEL_TAG = { cheat: 'tag-red', sospechoso: 'tag-yellow', conducta: 'tag-plain' };

function levelOf(p) { return p.level || 'limpio'; }
function confirmationOf(p) { return confirmedMap[String(p.steamId)] || null; }
// flagged: cheat or suspicious (worth a look); conduct-only players are not flagged
function flagged(p) { const l = levelOf(p); return l === 'cheat' || l === 'sospechoso' || !!confirmationOf(p); }

function flagTag(f) {
    return '<span class="tag ' + (LEVEL_TAG[FLAG_LEVELS[f] || 'sospechoso'] || 'tag-plain') + '">' + esc(f) + '</span>';
}

function flagTags(p) {
    return '<span class="tags">' + (p.flags || []).map(flagTag).join('') + '</span>';
}

function levelTag(level) {
    if (!level || level === 'limpio') return '<span class="muted">Limpio</span>';
    return '<span class="tag ' + LEVEL_TAG[level] + '">' + LEVEL_LABEL[level] + '</span>';
}

function confirmedTag() { return '<span class="tag tag-confirmed">Confirmado</span>'; }

// Compact cell for tables: confirmation or level, then the detections
function resultCell(p) {
    const conf = confirmationOf(p);
    const head = conf ? confirmedTag() : (levelOf(p) === 'limpio' ? '' : levelTag(levelOf(p)));
    if (!head && !(p.flags || []).length) return '<span class="muted">Limpio</span>';
    return '<span class="tags">' + head + (p.flags || []).map(flagTag).join('') + '</span>';
}

function rowMark(p) {
    if (confirmationOf(p) || levelOf(p) === 'cheat') return ' mark-red';
    if (levelOf(p) === 'sospechoso') return ' mark-yellow';
    return '';
}

function refreshConfirmed(then) {
    if (!window.listConfirmed) { if (then) then(); return; }
    window.listConfirmed().then(m => {
        confirmedMap = m || {};
        if (then) then();
    });
}

// Re-render whatever is on screen after a confirmation changes
function rerenderAll() {
    if (currentReport) { renderPlayers(); renderDetections(); renderSummary(); }
    renderPlayerList();
    if (document.getElementById('playerView').style.display === 'flex' && activePlayer) openPlayerPage(activePlayer);
}

// Block shown in the player modal and on the player page
function confirmBlock(steamId, name, level) {
    const conf = confirmedMap[steamId];
    if (conf) {
        return '<div class="confirm-box confirmed">' +
            '<div><strong>Confirmado como cheater</strong> por el staff el ' + fmtDate(conf.date) +
            (conf.note ? '<div class="muted" style="margin-top:2px;">' + esc(conf.note) + '</div>' : '') + '</div>' +
            '<button class="btn" onclick="setConfirmed(\'' + steamId + '\', false)">Quitar confirmación</button></div>';
    }
    const hint = level === 'cheat' || level === 'sospechoso'
        ? 'La app solo marca indicios. Si viste la demo y es cheat, confirmalo: queda guardado para este SteamID.'
        : 'Si por otra vía sabés que usa cheats, podés confirmarlo igual.';
    return '<div class="confirm-box" id="confirmBox-' + steamId + '">' +
        '<div class="muted">' + hint + '</div>' +
        '<div class="confirm-form">' +
            '<input class="input" id="confirmNote-' + steamId + '" placeholder="Nota (opcional): qué viste, en qué ronda" style="flex:1;">' +
            '<button class="btn btn-danger" onclick="setConfirmed(\'' + steamId + '\', true, \'' + esc(String(name).replace(/['\\]/g, '')) + '\')">Confirmar cheater</button>' +
        '</div></div>';
}

function setConfirmed(steamId, on, name) {
    if (!window.setConfirmed) return;
    const noteEl = document.getElementById('confirmNote-' + steamId);
    const note = on && noteEl ? noteEl.value.trim() : '';
    window.setConfirmed(steamId, name || '', note, on).then(() => {
        refreshConfirmed(() => {
            rerenderAll();
            const modalOpen = document.getElementById('playerModal').style.display === 'flex';
            if (modalOpen && currentReport) {
                const p = currentReport.players.find(x => String(x.steamId) === steamId);
                if (p) openPlayer(p);
            }
            showToast(on ? 'Marcado como cheater confirmado' : 'Confirmación quitada');
        });
    });
}

function emptyRow(cols, text) {
    return '<tr class="empty"><td colspan="' + cols + '">' + text + '</td></tr>';
}
