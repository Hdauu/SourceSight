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

function playButton(tick) {
    return '<button class="play" onclick="event.stopPropagation(); playInGame(' + tick + ')" title="Abrir en CS:GO">' + icon('play') + 'Ver</button>';
}

const RED_TYPES = /TRIGGER|SILENT|WALL/;
function typeTag(type) {
    const cls = RED_TYPES.test(type) ? 'tag-red' : 'tag-yellow';
    return '<span class="tag ' + cls + '">' + esc(type) + '</span>';
}

// ── Hits per body zone (pkg/core/hits.go) ──
// Fewer hits than this say nothing about where someone aims (same as the Go side).
const MIN_HITS_FOR_PCT = 10;

function hitTotals(h) {
    if (!h || !h.zones) return null;
    const z = h.zones;
    const total = Object.keys(z).reduce((s, k) => s + z[k].hits, 0);
    return {
        total: total,
        head: z.head.hits + z.neck.hits,
        headPct: total >= MIN_HITS_FOR_PCT ? Math.round(1000 * (z.head.hits + z.neck.hits) / total) / 10 : null,
        accuracy: h.shots > 0 ? Math.round(1000 * total / h.shots) / 10 : null
    };
}
