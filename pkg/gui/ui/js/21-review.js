// ── Reviewing a player demo by demo ───────────────────────────────────
// Opening a demo from a player's page starts a review: the demo opens on that player's
// detail (incidents with their ticks), and a bar lets you go to their previous or next
// demo, or back to their page, without searching for anything again.

let reviewCtx = null;        // { steamId, name, all: [{entryId, flagged}], only: bool }
let lastPlayerDetail = null; // the player page on screen, set by openPlayerPage()

function reviewList() {
    if (!reviewCtx) return [];
    return reviewCtx.only ? reviewCtx.all.filter(x => x.flagged) : reviewCtx.all;
}

function startReview(entryId) {
    const d = lastPlayerDetail;
    if (!d) { openHistoryItem(entryId); return; }
    reviewCtx = {
        steamId: String(d.steamId),
        name: d.name,
        all: d.appearances.map(ap => ({ entryId: ap.entryId, flagged: !!(ap.flags && ap.flags.length) })),
        only: reviewCtx && reviewCtx.steamId === String(d.steamId) ? reviewCtx.only : false,
        entryId: entryId
    };
    openHistoryItem(entryId);
}

// Called after every report loads. A demo of the review keeps it going (also after
// re-analyzing it); any other demo ends it.
function reviewAfterLoad() {
    const id = currentMeta && currentMeta.historyId;
    if (!reviewCtx || !reviewCtx.all.some(x => x.entryId === id)) {
        reviewCtx = null;
        renderReviewBar();
        return;
    }
    reviewCtx.entryId = id;
    renderReviewBar();
    const p = findPlayer(reviewCtx.steamId);
    if (p) openPlayer(p);
}

// Neighbours of the current demo in the player's order. With "only flagged" on, the
// nearest flagged demo on each side, even when the current one is clean.
function reviewNeighbours() {
    const all = reviewCtx.all, ok = x => !reviewCtx.only || x.flagged;
    const i = all.findIndex(x => x.entryId === reviewCtx.entryId);
    let prev = null, next = null;
    for (let j = i - 1; j >= 0; j--) if (ok(all[j])) { prev = all[j]; break; }
    for (let j = i + 1; j < all.length; j++) if (ok(all[j])) { next = all[j]; break; }
    const list = reviewList();
    return { prev: prev, next: next, pos: list.findIndex(x => x.entryId === reviewCtx.entryId), total: list.length };
}

function reviewStep(dir) {
    if (!reviewCtx) return;
    const n = reviewNeighbours();
    const target = dir > 0 ? n.next : n.prev;
    if (!target) return;
    closePlayer();
    openHistoryItem(target.entryId);
}

function reviewBack() {
    if (!reviewCtx) return;
    const sid = reviewCtx.steamId;
    closePlayer();
    openPlayerPage(sid);
}

function reviewEnd() {
    reviewCtx = null;
    renderReviewBar();
    closePlayer();
}

function setReviewOnlyFlagged(on) {
    if (!reviewCtx) return;
    reviewCtx.only = on;
    renderReviewBar();
    if (document.getElementById('playerModal').style.display === 'flex') {
        const p = findPlayer(reviewCtx.steamId);
        if (p) openPlayer(p);
    }
}

function reviewOpenPlayer() {
    const p = reviewCtx && findPlayer(reviewCtx.steamId);
    if (p) openPlayer(p);
}

// The navigation controls, used in the bar above the report and in the player detail
function reviewNav() {
    const n = reviewNeighbours();
    const pos = n.pos >= 0 ? 'Demo ' + (n.pos + 1) + ' de ' + n.total : n.total + (reviewCtx.only ? ' marcadas' : ' demos');
    return '<button class="btn" onclick="reviewStep(-1)"' + (n.prev ? '' : ' disabled') + ' title="Demo anterior de este jugador">' + String.fromCharCode(8249) + ' Anterior</button>' +
        '<span class="muted num" style="white-space:nowrap;">' + pos + '</span>' +
        '<button class="btn" onclick="reviewStep(1)"' + (n.next ? '' : ' disabled') + ' title="Demo siguiente de este jugador">Siguiente ' + String.fromCharCode(8250) + '</button>';
}

function renderReviewBar() {
    const bar = document.getElementById('reviewBar');
    if (!bar) return;
    if (!reviewCtx) {
        bar.style.display = 'none';
        bar.innerHTML = '';
        return;
    }
    const inDemo = !!findPlayer(reviewCtx.steamId);
    bar.innerHTML =
        '<span class="grow">Revisando a <b>' + esc(reviewCtx.name) + '</b>' +
        (inDemo ? '' : ' <span class="t-yellow">(no aparece en esta demo)</span>') + '</span>' +
        '<label class="muted" style="display:flex; align-items:center; gap:6px; white-space:nowrap; cursor:pointer;">' +
            '<input type="checkbox" onchange="setReviewOnlyFlagged(this.checked)"' + (reviewCtx.only ? ' checked' : '') + '> Solo marcadas</label>' +
        reviewNav() +
        (inDemo ? '<button class="btn" onclick="reviewOpenPlayer()">Ver sus incidentes</button>' : '') +
        '<button class="btn" onclick="reviewBack()">Volver a su ficha</button>' +
        '<button class="btn btn-plain btn-icon" onclick="reviewEnd()" title="Terminar la revisión">' + icon('x') + '</button>';
    bar.style.display = 'flex';
}
