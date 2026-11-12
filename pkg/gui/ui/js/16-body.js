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
    ['pistol', 'Pistolas'], ['smg', 'SMG'], ['heavy', 'Escopetas y pesadas']
];

// Shapes, drawn around x = 180 in a 360 × 430 box. Right-side limbs (viewer's left)
// are drawn once; the left ones are the same path mirrored.
const BODY_MIRROR = 'translate(360,0) scale(-1,1)';
const BODY_SHAPES = {
    head: '<ellipse cx="180" cy="58" rx="23" ry="27"/>',
    neck: '<path d="M170 82 L190 82 L193 99 L167 99 Z"/>',
    chest: '<path d="M140 102 Q180 92 220 102 L224 176 L136 176 Z"/>',
    stomach: '<path d="M136 176 L224 176 L220 228 Q180 238 140 228 Z"/>',
    rightArm: '<path d="M137 104 Q118 106 112 128 L102 196 L94 262 Q98 274 110 270 L118 206 L133 150 Z"/><circle cx="101" cy="279" r="9"/>',
    rightLeg: '<path d="M140 228 Q160 238 178 236 L175 330 L172 402 L146 402 L144 330 Z"/>'
};
BODY_SHAPES.leftArm = '<g transform="' + BODY_MIRROR + '">' + BODY_SHAPES.rightArm + '</g>';
BODY_SHAPES.leftLeg = '<g transform="' + BODY_MIRROR + '">' + BODY_SHAPES.rightLeg + '</g>';

// Where each percentage is written and the point its line points to.
const BODY_LABELS = {
    head: { at: [203, 48], text: [272, 48], end: false },
    neck: { at: [169, 91], text: [88, 91], end: true },
    chest: { at: [214, 134], text: [272, 134], end: false },
    stomach: { at: [146, 206], text: [88, 206], end: true },
    rightArm: { at: [101, 238], text: [62, 262], end: true },
    leftArm: { at: [259, 238], text: [298, 262], end: false },
    rightLeg: { at: [152, 338], text: [88, 338], end: true },
    leftLeg: { at: [208, 338], text: [272, 338], end: false }
};

// Gear, drawn over the zones without hiding their color.
function bodyGear(team) {
    const line = 'fill="none" stroke="var(--text)" stroke-opacity="0.32" stroke-width="1.6" stroke-linejoin="round"';
    const solid = 'fill="#000" fill-opacity="0.28" stroke="var(--text)" stroke-opacity="0.32" stroke-width="1.4"';
    const boots = '<path d="M146 402 L172 402 L174 418 L136 418 Q136 408 146 402 Z" ' + solid + '/>';
    const gloves = '<circle cx="101" cy="279" r="9" ' + solid + '/>';
    const both = '<g>' + boots + gloves + '</g><g transform="' + BODY_MIRROR + '">' + boots + gloves + '</g>';
    if (team === 't') {
        return both +
            // Balaclava: covers the head, with an opening for the eyes
            '<path d="M157 62 Q157 30 180 30 Q203 30 203 62 Q203 86 180 86 Q157 86 157 62 Z" ' + solid + '/>' +
            '<rect x="165" y="50" width="30" height="9" rx="4.5" fill="var(--bg)" fill-opacity="0.85"/>' +
            // Chest rig: crossed straps and a row of pouches
            '<path d="M146 104 L214 176 M214 104 L146 176" ' + line + '/>' +
            '<rect x="146" y="182" width="20" height="20" rx="2" ' + line + '/>' +
            '<rect x="170" y="182" width="20" height="20" rx="2" ' + line + '/>' +
            '<rect x="194" y="182" width="20" height="20" rx="2" ' + line + '/>';
    }
    return both +
        // Helmet with a brim
        '<path d="M155 58 Q155 26 180 26 Q205 26 205 58 L211 62 L149 62 Z" ' + solid + '/>' +
        // Plate carrier: vest outline, plates and a belt
        '<path d="M143 105 Q180 97 217 105 L219 214 Q180 222 141 214 Z" ' + line + '/>' +
        '<path d="M156 120 L204 120 L204 166 L156 166 Z" ' + line + '/>' +
        '<path d="M140 222 Q180 232 220 222" ' + line + '/>';
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

// One figure. Shading uses a fixed scale (45% of hits = full color), so two figures
// side by side can be compared by eye.
function bodySvg(c, team, title) {
    const zone = key => {
        const share = c && c.total ? c.pct[key] : 0;
        const op = c && c.total ? Math.min(0.95, 0.08 + 0.87 * share / 45) : 0;
        return '<g class="bz" data-zone="' + key + '">' +
            '<g fill="var(--panel-2)" stroke="var(--border-strong)" stroke-width="1.2">' + BODY_SHAPES[key] + '</g>' +
            '<g fill="var(--accent)" fill-opacity="' + op.toFixed(3) + '" stroke="none">' + BODY_SHAPES[key] + '</g>' +
            '<title>' + BODY_ZONE_LABEL[key] + (c ? ': ' + ((c.zones[key] || {}).hits || 0) + ' impactos (' + fmtPct(share) + ')' : '') + '</title>' +
        '</g>';
    };
    const label = key => {
        const l = BODY_LABELS[key];
        const share = c && c.total ? fmtPct(c.pct[key]) : '–';
        const anchor = l.end ? 'end' : 'start';
        const lineEnd = l.end ? l.text[0] + 4 : l.text[0] - 4;
        return '<line x1="' + l.at[0] + '" y1="' + l.at[1] + '" x2="' + lineEnd + '" y2="' + (l.text[1] - 4) + '" stroke="var(--muted)" stroke-opacity="0.5" stroke-width="1"/>' +
            '<text x="' + l.text[0] + '" y="' + l.text[1] + '" text-anchor="' + anchor + '" class="bz-pct">' + share + '</text>' +
            '<text x="' + l.text[0] + '" y="' + (l.text[1] + 13) + '" text-anchor="' + anchor + '" class="bz-name">' + BODY_ZONE_LABEL[key] + '</text>';
    };
    const order = ['rightLeg', 'leftLeg', 'stomach', 'chest', 'rightArm', 'leftArm', 'neck', 'head'];
    return '<figure class="body-fig">' +
        '<svg viewBox="0 0 360 430" role="img" aria-label="' + esc(title) + '">' +
            order.map(zone).join('') +
            '<g pointer-events="none">' + bodyGear(team) + '</g>' +
            Object.keys(BODY_LABELS).map(label).join('') +
        '</svg>' +
        '<figcaption>' + esc(title) + '<span class="muted">' + (c ? c.total + ' impactos' : 'sin datos') + '</span></figcaption>' +
    '</figure>';
}

function bodyTable(mine, avg, mode) {
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
        const acc = mine.shots ? fmtPct(100 * mine.total / mine.shots) : '–';
        foot = '<div class="muted body-foot">' + mine.total + ' impactos de ' + mine.shots + ' tiros (precisión ' + acc + ').' +
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
        bodyTable(mine, avg, st.mode);

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
