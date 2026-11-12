// ── Body figure: where a player's bullets land ────────────────────────
// A 2D soldier (own drawing, CT or T gear) with each body zone shaded by its share
// of the hits, plus a table. Data comes from core.HitProfile (pkg/core/hits.go).
//
//   renderBodyPanel(host, { mine, mineLabel, avg, avgLabel, mode })
//
// mine: the player's profile (one demo or all of them), avg: everyone's. Modes:
// 'mine', 'avg' or 'compare' (both figures side by side); only the ones with data
// are offered.

// Zones in table order. Figure sides are the soldier's own: his left arm is drawn on
// the viewer's right, as when facing someone.
const BODY_ZONES = [
    { key: 'head', label: 'Cabeza' },
    { key: 'neck', label: 'Cuello' },
    { key: 'chest', label: 'Pecho' },
    { key: 'stomach', label: 'Estómago' },
    { key: 'leftArm', label: 'Brazo izq.' },
    { key: 'rightArm', label: 'Brazo der.' },
    { key: 'leftLeg', label: 'Pierna izq.' },
    { key: 'rightLeg', label: 'Pierna der.' },
    { key: 'other', label: 'Sin zona' }
];
const BODY_ZONE_LABEL = Object.fromEntries(BODY_ZONES.map(z => [z.key, z.label]));

const BODY_WEAPONS = [
    ['all', 'Todas las armas'], ['rifle', 'Rifles'], ['sniper', 'Snipers'],
    ['pistol', 'Pistolas'], ['smg', 'SMG'], ['shotgun', 'Escopetas'], ['heavy', 'Ametralladoras']
];

// Shapes, drawn around x = 180 in a 360 × 450 box with real proportions (about eight
// heads tall). Right-side limbs (viewer's left) are drawn once; the left ones are the
// same path mirrored. Zones are separate pieces with a thin gap between them, like a
// hitbox chart.
const BODY_MIRROR = 'translate(360,0) scale(-1,1)';
const BODY_SHAPES = {
    head: '<ellipse cx="180" cy="42" rx="20" ry="25"/>',
    neck: '<path d="M171 64 L189 64 L191 83 L169 83 Z"/>',
    chest: '<path d="M150 82 C135 84 126 93 126 108 L139 166 L221 166 L234 108 C234 93 225 84 210 82 C196 79 164 79 150 82 Z"/>',
    stomach: '<path d="M139 166 L221 166 L215 198 C219 208 222 220 218 232 C198 241 162 241 142 232 C138 220 141 208 145 198 Z"/>',
    rightArm: '<path d="M126 90 C113 92 105 102 104 117 L101 160 C99 168 97 177 96 187 L93 229 C93 238 106 239 107 231 L112 190 C113 181 116 173 119 165 L125 126 Z"/>' +
        '<ellipse cx="99" cy="244" rx="8" ry="11.5"/>',
    rightLeg: '<path d="M144 233 C156 239 168 241 178 240 L176 300 C175 312 172 322 171 332 L169 396 L150 396 L148 332 C146 320 144 306 143 292 Z"/>'
};
BODY_SHAPES.leftArm = '<g transform="' + BODY_MIRROR + '">' + BODY_SHAPES.rightArm + '</g>';
BODY_SHAPES.leftLeg = '<g transform="' + BODY_MIRROR + '">' + BODY_SHAPES.rightLeg + '</g>';

// Where each percentage is written (text) and the point of the zone its line starts
// from (at). Left column ends at x = 84, right column starts at x = 276.
const BODY_LABELS = {
    head: { at: [198, 34], text: [276, 40], end: false },
    neck: { at: [172, 74], text: [84, 76], end: true },
    chest: { at: [214, 118], text: [276, 120], end: false },
    stomach: { at: [146, 204], text: [84, 170], end: true },
    rightArm: { at: [103, 214], text: [84, 240], end: true },
    leftArm: { at: [257, 214], text: [276, 240], end: false },
    rightLeg: { at: [160, 330], text: [84, 332], end: true },
    leftLeg: { at: [200, 330], text: [276, 332], end: false }
};

// Gear, drawn over the zones: dark and thin enough to keep their color readable.
function bodyGear(team) {
    const line = 'fill="none" stroke="#000" stroke-opacity="0.35" stroke-width="1.4" stroke-linejoin="round" stroke-linecap="round"';
    const solid = 'fill="#000" fill-opacity="0.38"';
    const boot = '<path d="M150 396 L169 396 L172 412 L139 412 C139 404 144 398 150 396 Z" fill="#1a1b1d" stroke="var(--border-strong)" stroke-width="1"/>';
    const glove = '<ellipse cx="99" cy="244" rx="8" ry="11.5" ' + solid + '/>';
    const both = boot + glove + '<g transform="' + BODY_MIRROR + '">' + boot + glove + '</g>';
    if (team === 't') {
        return both +
            // Balaclava over the whole head, with the eye opening
            '<ellipse cx="180" cy="42" rx="20" ry="25" ' + solid + '/>' +
            '<rect x="166" y="34" width="28" height="8" rx="4" fill="var(--bg)" fill-opacity="0.9"/>' +
            // Chest rig: crossed straps and a row of pouches
            '<path d="M146 90 L214 162 M214 90 L146 162" ' + line + '/>' +
            '<path d="M148 176 h18 v18 h-18 Z M171 176 h18 v18 h-18 Z M194 176 h18 v18 h-18 Z" ' + line + '/>';
    }
    return both +
        // Helmet with a brim
        '<path d="M160 40 C160 14 200 14 200 40 L205 44 L155 44 Z" fill="#1d2228" fill-opacity="0.85" stroke="#000" stroke-opacity="0.4" stroke-width="1"/>' +
        // Plate carrier and belt
        '<path d="M138 92 C160 87 200 87 222 92 L218 196 C200 204 160 204 142 196 Z" ' + line + '/>' +
        '<path d="M160 104 h40 v42 h-40 Z" ' + line + '/>' +
        '<path d="M139 214 C160 222 200 222 221 214" ' + line + '/>';
}

// Hits per zone for a weapon class; null when the profile has nothing for it.
function bodyCounts(h, weapon) {
    if (!h) return null;
    const src = weapon === 'all' ? { shots: h.shots, zones: h.zones } : (h.byClass || {})[weapon];
    if (!src || !src.zones) return null;
    const total = BODY_ZONES.reduce((s, z) => s + ((src.zones[z.key] || {}).hits || 0), 0);
    const pct = {};
    BODY_ZONES.forEach(z => { pct[z.key] = total ? 100 * ((src.zones[z.key] || {}).hits || 0) / total : 0; });
    return { zones: src.zones, shots: src.shots || 0, total: total, pct: pct };
}

function fmtPct(v) { return (Math.round(v * 10) / 10).toLocaleString('es-AR') + '%'; }

// Heat scale: dark grey (no hits) → deep orange → accent orange → pale yellow. Fixed,
// with 45% of the hits as the top, so two figures side by side compare by eye.
const BODY_HEAT_MAX = 45;
const BODY_HEAT_STOPS = [[0, [52, 55, 60]], [0.3, [122, 61, 23]], [0.65, [224, 123, 42]], [1, [255, 214, 150]]];
function heatColor(share) {
    const t = Math.max(0, Math.min(1, share / BODY_HEAT_MAX));
    for (let i = 1; i < BODY_HEAT_STOPS.length; i++) {
        const [t1, c1] = BODY_HEAT_STOPS[i];
        if (t <= t1) {
            const [t0, c0] = BODY_HEAT_STOPS[i - 1];
            const k = (t - t0) / (t1 - t0);
            return 'rgb(' + c0.map((v, j) => Math.round(v + (c1[j] - v) * k)).join(',') + ')';
        }
    }
    return 'rgb(' + BODY_HEAT_STOPS[BODY_HEAT_STOPS.length - 1][1].join(',') + ')';
}

let bodyFigSeq = 0;

// One figure, with the color scale underneath.
function bodySvg(c, team, title) {
    const gid = 'bodyHeat' + (++bodyFigSeq);
    const zone = key => {
        const share = c && c.total ? c.pct[key] : 0;
        const fill = c && c.total ? heatColor(share) : '#2a2c30';
        return '<g class="bz" data-zone="' + key + '" fill="' + fill + '" stroke="var(--bg)" stroke-width="2.5" stroke-linejoin="round">' +
            BODY_SHAPES[key] +
            '<title>' + BODY_ZONE_LABEL[key] + (c ? ': ' + ((c.zones[key] || {}).hits || 0) + ' impactos (' + fmtPct(share) + ')' : '') + '</title>' +
        '</g>';
    };
    const label = key => {
        const l = BODY_LABELS[key];
        const share = c && c.total ? c.pct[key] : null;
        const anchor = l.end ? 'end' : 'start';
        const lineEnd = l.end ? l.text[0] + 5 : l.text[0] - 5;
        const color = share === null || share < 1 ? 'var(--muted)' : heatColor(Math.max(share, 15));
        return '<circle cx="' + l.at[0] + '" cy="' + l.at[1] + '" r="2" fill="var(--text)" fill-opacity="0.7"/>' +
            '<polyline points="' + l.at.join(',') + ' ' + lineEnd + ',' + (l.text[1] - 5) + '" fill="none" stroke="var(--text)" stroke-opacity="0.28" stroke-width="1"/>' +
            '<text x="' + l.text[0] + '" y="' + l.text[1] + '" text-anchor="' + anchor + '" class="bz-pct" fill="' + color + '">' +
                (share !== null ? fmtPct(share) : '–') + '</text>' +
            '<text x="' + l.text[0] + '" y="' + (l.text[1] + 13) + '" text-anchor="' + anchor + '" class="bz-name">' + BODY_ZONE_LABEL[key] + '</text>';
    };
    const order = ['rightLeg', 'leftLeg', 'stomach', 'chest', 'rightArm', 'leftArm', 'neck', 'head'];
    const legend =
        '<defs><linearGradient id="' + gid + '">' + BODY_HEAT_STOPS.map(s => '<stop offset="' + s[0] + '" stop-color="rgb(' + s[1].join(',') + ')"/>').join('') + '</linearGradient></defs>' +
        '<rect x="120" y="432" width="120" height="6" rx="3" fill="url(#' + gid + ')"/>' +
        '<text x="114" y="439" text-anchor="end" class="bz-name">0%</text>' +
        '<text x="246" y="439" class="bz-name">' + BODY_HEAT_MAX + '% o más</text>';
    return '<figure class="body-fig">' +
        '<svg viewBox="0 0 360 450" role="img" aria-label="' + esc(title) + '">' +
            order.map(zone).join('') +
            '<g pointer-events="none">' + bodyGear(team) + '</g>' +
            Object.keys(BODY_LABELS).map(label).join('') +
            legend +
        '</svg>' +
        '<figcaption>' + esc(title) + '<span class="muted">' + (c ? c.total + ' impactos' : 'sin datos') + '</span></figcaption>' +
    '</figure>';
}

function bodyTable(mine, avg, mode, st) {
    const showMine = mode !== 'avg' && mine, showAvg = mode !== 'mine' && avg;
    const zones = BODY_ZONES.filter(z => z.key !== 'other' ||
        (showMine && (mine.zones.other || {}).hits) || (showAvg && (avg.zones.other || {}).hits));
    let head = '<th>Zona</th>';
    if (showMine) head += '<th class="r">Impactos</th><th class="r">Daño</th><th class="r">%</th>';
    if (showAvg) head += '<th class="r">' + (showMine ? '% promedio' : '%') + '</th>';
    if (showMine && showAvg) head += '<th class="r">Diferencia</th>';
    const rows = zones.map(z => {
        let r = '<td>' + z.label + '</td>';
        if (showMine) {
            const zc = mine.zones[z.key] || { hits: 0, damage: 0 };
            r += '<td class="r num">' + zc.hits + '</td><td class="r num">' + zc.damage + '</td><td class="r num">' + fmtPct(mine.pct[z.key]) + '</td>';
        }
        if (showAvg) r += '<td class="r num muted">' + fmtPct(avg.pct[z.key]) + '</td>';
        if (showMine && showAvg) {
            const d = Math.round((mine.pct[z.key] - avg.pct[z.key]) * 10) / 10;
            // Only an unusually high head share is worth pointing out
            const cls = (z.key === 'head' || z.key === 'neck') && d >= 10 ? ' t-yellow' : ' muted';
            r += '<td class="r num' + cls + '">' + (d > 0 ? '+' : '') + d.toLocaleString('es-AR') + ' pts</td>';
        }
        return '<tr>' + r + '</tr>';
    }).join('');
    let foot = '';
    if (showMine) {
        // Shotgun pellets count one hit each: accuracy means nothing for them
        const acc = st.weapon === 'shotgun' ? '' : st.weapon === 'all'
            ? fmtNum(accuracyOf(st.mine), '%')
            : (mine.shots ? fmtPct(100 * mine.total / mine.shots) : '–');
        foot = '<div class="muted body-foot">' + mine.total + ' impactos' + (acc ? ' · precisión ' + acc + (st.weapon === 'all' ? ' (sin escopetas)' : '') : '') + '.' +
            (mine.total < MIN_HITS_FOR_PCT ? ' Son pocos impactos: los porcentajes todavía no dicen mucho.' : '') + '</div>';
    }
    return '<div class="table-wrap"><table class="compare"><thead><tr>' + head + '</tr></thead><tbody>' + rows + '</tbody></table></div>' + foot;
}

let bodyTeam = 'ct';
try { bodyTeam = localStorage.getItem('bodyTeam') === 't' ? 't' : 'ct'; } catch (e) { /* storage unavailable */ }

function renderBodyPanel(host, opts) {
    if (!host) return;
    const st = host._body = Object.assign(host._body || { weapon: 'all' }, opts);
    const modes = [];
    if (st.mine) modes.push(['mine', st.mineLabel || 'Jugador']);
    if (st.avg) modes.push(['avg', 'Promedio general']);
    if (st.mine && st.avg) modes.push(['compare', 'Comparación']);
    if (!modes.length) {
        host.innerHTML = '<div class="muted">Todavía no hay impactos registrados. Las demos analizadas con una versión anterior no los tienen: re-analizalas desde el historial.</div>';
        return;
    }
    if (!modes.some(m => m[0] === st.mode)) st.mode = modes[modes.length - 1][0];

    // Weapon classes present in either profile
    const has = w => w === 'all' || [st.mine, st.avg].some(h => h && h.byClass && h.byClass[w]);
    const weapons = BODY_WEAPONS.filter(w => has(w[0]));
    if (!weapons.some(w => w[0] === st.weapon)) st.weapon = 'all';

    const mine = bodyCounts(st.mine, st.weapon), avg = bodyCounts(st.avg, st.weapon);
    const seg = (items, cur, fn) => '<div class="seg">' + items.map(i =>
        '<button class="' + (i[0] === cur ? 'on' : '') + '" data-v="' + i[0] + '" data-fn="' + fn + '">' + i[1] + '</button>').join('') + '</div>';

    let figs = '';
    if (st.mode !== 'avg') figs += bodySvg(mine, bodyTeam, st.mineLabel || 'Jugador');
    if (st.mode !== 'mine') figs += bodySvg(avg, bodyTeam, 'Promedio general');

    host.innerHTML =
        '<div class="body-controls">' +
            (modes.length > 1 ? seg(modes, st.mode, 'mode') : '') +
            '<select class="input" data-fn="weapon" aria-label="Arma">' + weapons.map(w => '<option value="' + w[0] + '"' + (w[0] === st.weapon ? ' selected' : '') + '>' + w[1] + '</option>').join('') + '</select>' +
            seg([['ct', 'CT'], ['t', 'TT']], bodyTeam, 'team') +
        '</div>' +
        '<div class="body-figs' + (st.mode === 'compare' ? ' two' : '') + '">' + figs + '</div>' +
        bodyTable(mine, avg, st.mode, st);

    host.querySelectorAll('[data-fn]').forEach(el => {
        const fn = el.dataset.fn;
        const apply = v => {
            if (fn === 'mode') st.mode = v;
            if (fn === 'weapon') st.weapon = v;
            if (fn === 'team') {
                bodyTeam = v;
                try { localStorage.setItem('bodyTeam', v); } catch (e) { /* storage unavailable */ }
            }
            renderBodyPanel(host, {});
        };
        if (el.tagName === 'SELECT') el.onchange = () => apply(el.value);
        else el.onclick = () => apply(el.dataset.v);
    });
}
