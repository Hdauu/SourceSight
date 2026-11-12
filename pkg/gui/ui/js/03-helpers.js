// ── Helpers ───────────────────────────────────────────────────────
function esc(s) {
    if (s === null || s === undefined) return '';
    return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

// Server names usually carry decorative emoji; drop pictographs and dangling separators
function stripEmoji(s) {
    return String(s || '')
        .replace(/[\p{Extended_Pictographic}\u{FE0F}\u{200D}]/gu, '')
        .replace(/\s*([|»•·])\s*(?=[|»•·]|$)/g, '')
        .replace(/\s{2,}/g, ' ')
        .trim();
}


function roundOf(tick) {
    const starts = (currentReport && currentReport.header.roundStarts) || [];
    let lo = 0, hi = starts.length;
    while (lo < hi) {
        const mid = (lo + hi) >> 1;
        if (starts[mid] > tick) hi = mid; else lo = mid + 1;
    }
    return lo;
}

function cmdButton(cmd, label) {
    return '<button class="cmd" onclick="event.stopPropagation(); copyText(\'' + esc(cmd) + '\', \'Copiado: ' + esc(cmd) + '\')" title="Copiar">' + esc(label || cmd) + icon('copy', 'ic-sm') + '</button>';
}

// steamId (optional): the game opens in first person from that player's eyes.
function playButton(tick, steamId) {
    const sid = steamId ? String(steamId).replace(/\D/g, '') : '';
    return '<button class="play" onclick="event.stopPropagation(); playInGame(' + tick + ', \'' + sid + '\')" title="Abrir en CS:GO' + (sid ? ' desde los ojos del jugador' : '') + '">' + icon('play') + 'Ver</button>';
}

const RED_TYPES = /TRIGGER|SILENT|WALL/;
function typeTag(type) {
    const cls = RED_TYPES.test(type) ? 'tag-red' : 'tag-yellow';
    return '<span class="tag ' + cls + '">' + esc(type) + '</span>';
}

// ── Hits per body zone (pkg/core/hits.go) ──
// Fewer hits than this say nothing about where someone aims (same as the Go side).
const MIN_HITS_FOR_PCT = 10;

// Zones never hit are not in the data (Go leaves them out), so a missing one counts 0.
function zoneHits(zones, key) { return ((zones || {})[key] || {}).hits || 0; }

function hitTotals(h) {
    if (!h || !h.zones) return null;
    const z = h.zones;
    const total = Object.keys(z).reduce((s, k) => s + zoneHits(z, k), 0);
    const head = zoneHits(z, 'head') + zoneHits(z, 'neck');
    return {
        total: total,
        head: head,
        headPct: total >= MIN_HITS_FOR_PCT ? Math.round(1000 * head / total) / 10 : null,
        accuracy: accuracyOf(h)
    };
}

// Observations: numbers that stand out against other players. Informative only.
function notesBlock(notes, title) {
    if (!notes || !notes.length) return '';
    return '<div class="callout notes">' +
        '<div class="block-title" style="margin:0;">' + (title || 'Para mirar') + '</div>' +
        '<ul>' + notes.map(n => '<li>' + esc(n) + '</li>').join('') + '</ul>' +
        '<div class="notes-foot">Solo informativo: no cambia el resultado. Conviene confirmarlo mirando la demo.</div>' +
    '</div>';
}

// Accuracy without shotguns: several pellets per shell can each count as a hit
// (same as HitProfile.Accuracy in Go). Null when there is nothing to measure.
function accuracyOf(h) {
    if (!h || !h.zones) return null;
    const hitsOf = zones => Object.keys(zones || {}).reduce((s, k) => s + zoneHits(zones, k), 0);
    let hits = hitsOf(h.zones), shots = h.shots || 0;
    const sg = (h.byClass || {}).shotgun;
    if (sg) { hits -= hitsOf(sg.zones); shots -= sg.shots || 0; }
    return shots > 0 ? Math.round(1000 * hits / shots) / 10 : null;
}
