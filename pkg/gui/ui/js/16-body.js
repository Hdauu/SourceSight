// ── Body figure: where a player's bullets land ────────────────────────
// A minimal 2D silhouette with each body zone shaded by its share
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

// Minimal silhouette, drawn around x = 180 in a 360 × 450 box: flat rounded pieces
// with a small gap between them and no outlines, only the color of each zone.
// Right-side limbs (viewer's left) are drawn once; the left ones are mirrored.
const BODY_MIRROR = 'translate(360,0) scale(-1,1)';
const BODY_SHAPES = {
    head: '<circle cx="180" cy="44" r="22"/>',
    neck: '<rect x="172" y="70" width="16" height="12" rx="4"/>',
    chest: '<rect x="134" y="86" width="92" height="78" rx="18"/>',
    stomach: '<rect x="140" y="168" width="80" height="62" rx="14"/>',
    rightArm: '<rect x="104" y="88" width="24" height="152" rx="12" transform="rotate(8 116 88)"/>',
    rightLeg: '<rect x="144" y="234" width="32" height="176" rx="15"/>'
};
BODY_SHAPES.leftArm = '<g transform="' + BODY_MIRROR + '">' + BODY_SHAPES.rightArm + '</g>';
BODY_SHAPES.leftLeg = '<g transform="' + BODY_MIRROR + '">' + BODY_SHAPES.rightLeg + '</g>';

// Where each percentage is written (text) and the point of the zone its line starts
// from (at). Left column ends at x = 84, right column starts at x = 276.
const BODY_LABELS = {
    head: { at: [200, 40], text: [276, 44], end: false },
    neck: { at: [173, 76], text: [84, 78], end: true },
    chest: { at: [214, 120], text: [276, 122], end: false },
    stomach: { at: [146, 200], text: [84, 180], end: true },
    rightArm: { at: [101, 200], text: [84, 244], end: true },
    leftArm: { at: [259, 200], text: [276, 244], end: false },
    rightLeg: { at: [160, 330], text: [84, 332], end: true },
    leftLeg: { at: [200, 330], text: [276, 332], end: false }
};

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
function bodySvg(c, title) {
    const gid = 'bodyHeat' + (++bodyFigSeq);
    const zone = key => {
        const share = c && c.total ? c.pct[key] : 0;
        const fill = c && c.total ? heatColor(share) : '#2a2c30';
        return '<g class="bz" data-zone="' + key + '" fill="' + fill + '">' +
            BODY_SHAPES[key] +
            '<title>' + BODY_ZONE_LABEL[key] + (c ? ': ' + ((c.zones[key] || {}).hits || 0) + ' impactos (' + fmtPct(share) + ')' : '') + '</title>' +
        '</g>';
    };
    const label = key => {
        const l = BODY_LABELS[key];
        const share = c && c.total ? c.pct[key] : null;
        const anchor = l.end ? 'end' : 'start';
        const lineEnd = l.end ? l.text[0] + 5 : l.text[0] - 5;
        return '<circle cx="' + l.at[0] + '" cy="' + l.at[1] + '" r="2" fill="var(--text)" fill-opacity="0.7"/>' +
            '<polyline points="' + l.at.join(',') + ' ' + lineEnd + ',' + (l.text[1] - 5) + '" fill="none" stroke="var(--text)" stroke-opacity="0.25" stroke-width="1"/>' +
            '<text x="' + l.text[0] + '" y="' + l.text[1] + '" text-anchor="' + anchor + '" class="bz-pct">' +
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
    if (showAvg) head += '<th class="r">' + (showMine ? '% ' + esc(avgName(st).toLowerCase()) : '%') + '</th>';
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

function renderBodyPanel(host, opts) {
    if (!host) return;
    const st = host._body = Object.assign(host._body || { weapon: 'all' }, opts);
    const modes = [];
    if (st.mine) modes.push(['mine', st.mineLabel || 'Jugador']);
    if (st.avg) modes.push(['avg', avgName(st)]);
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
    if (st.mode !== 'avg') figs += bodySvg(mine, st.mineLabel || 'Jugador');
    if (st.mode !== 'mine') figs += bodySvg(avg, avgName(st));

    host.innerHTML =
        '<div class="body-controls">' +
            (modes.length > 1 ? seg(modes, st.mode, 'mode') : '') +
            '<select class="input" data-fn="weapon" aria-label="Arma">' + weapons.map(w => '<option value="' + w[0] + '"' + (w[0] === st.weapon ? ' selected' : '') + '>' + w[1] + '</option>').join('') + '</select>' +
        '</div>' +
        '<div class="body-figs' + (st.mode === 'compare' ? ' two' : '') + '">' + figs + '</div>' +
        bodyTable(mine, avg, st.mode, st);

    host.querySelectorAll('[data-fn]').forEach(el => {
        const fn = el.dataset.fn;
        const apply = v => {
            if (fn === 'mode') st.mode = v;
            if (fn === 'weapon') st.weapon = v;
            renderBodyPanel(host, {});
        };
        if (el.tagName === 'SELECT') el.onchange = () => apply(el.value);
        else el.onclick = () => apply(el.dataset.v);
    });
}

// What the player is compared with ("Resto de la partida", "Resto de los jugadores"…).
function avgName(st) { return (st && st.avgLabel) || 'Promedio general'; }

// Sum of several hit profiles (the rest of a match), in the same shape as Go's.
function mergeHits(list) {
    const out = { shots: 0, kills: 0, headshotKills: 0, zones: {}, byClass: {} };
    const addZones = (dst, src) => Object.keys(src || {}).forEach(k => {
        const z = dst[k] || (dst[k] = { hits: 0, damage: 0 });
        z.hits += src[k].hits || 0;
        z.damage += src[k].damage || 0;
    });
    let any = false;
    list.forEach(h => {
        if (!h) return;
        any = true;
        out.shots += h.shots || 0;
        out.kills += h.kills || 0;
        out.headshotKills += h.headshotKills || 0;
        addZones(out.zones, h.zones);
        Object.keys(h.byClass || {}).forEach(c => {
            const w = out.byClass[c] || (out.byClass[c] = { shots: 0, zones: {} });
            w.shots += h.byClass[c].shots || 0;
            addZones(w.zones, h.byClass[c].zones);
        });
    });
    return any ? out : null;
}
