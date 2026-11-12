// ── Minimap ───────────────────────────────────────────────────────────
// Draws the match on the radar of the installed CS:GO (Go reads and converts it,
// pkg/radar): kills and deaths, paths, a heatmap of where players spent their time,
// incidents and SHIFT log cases. Positions come from report.map (pkg/core/mapdata.go).
// Click a marker to see what happened and open it in the game.

const MAP_TEAM = { 2: '#e0b040', 3: '#5b9bd5' };   // T, CT
const MAP_LAYERS = [
    ['kills', 'Kills y muertes'],
    ['paths', 'Recorrido'],
    ['heat', 'Mapa de calor'],
    ['incidents', 'Incidentes'],
    ['shift', 'SHIFT']
];

const mapState = { radar: null, error: '', images: {}, floor: 0, player: '', round: '', layers: { kills: true, incidents: true }, markers: [], loadingFor: '' };

function mapData() { return (currentReport && currentReport.map) || null; }

// Called for every report shown
function renderMapCard() {
    const md = mapData();
    mapState.player = '';
    mapState.round = '';
    mapState.floor = 0;
    mapState.markers = [];
    document.getElementById('mapInfo').innerHTML = '<span class="muted">Hacé click en un punto del mapa para ver qué pasó y abrirlo en el juego.</span>';
    renderMapFilters();
    if (!md) {
        mapState.radar = null;
        mapState.error = 'Esta demo se analizó con una versión anterior. Re-analizala para ver el mapa.';
        drawMap();
        return;
    }
    const mapName = currentReport.header.mapName;
    if (mapState.radar && mapState.radar.map === mapName) { drawMap(); return; }
    mapState.radar = null;
    mapState.error = '';
    mapState.loadingFor = mapName;
    drawMap();
    if (!window.getRadar) return;
    window.getRadar(mapName).then(res => {
        if (mapState.loadingFor !== mapName) return; // another report was opened meanwhile
        if (res.error) { mapState.error = res.error; drawMap(); return; }
        mapState.radar = res.radar;
        mapState.images = {};
        let pending = res.radar.sections.length;
        res.radar.sections.forEach((s, i) => {
            const img = new Image();
            img.onload = () => { if (--pending === 0) { renderMapFilters(); drawMap(); } };
            img.onerror = img.onload;
            img.src = s.image;
            mapState.images[i] = img;
        });
    });
}

function renderMapFilters() {
    const md = mapData();
    const players = md ? md.paths.map(p => [p.steamId, p.name]) : [];
    document.getElementById('mapPlayer').innerHTML = '<option value="">Todos los jugadores</option>' +
        players.map(p => '<option value="' + p[0] + '"' + (p[0] === mapState.player ? ' selected' : '') + '>' + esc(p[1]) + '</option>').join('');
    const rounds = (currentReport && currentReport.header.roundStarts || []).length;
    let opts = '<option value="">Todas las rondas</option>';
    for (let r = 1; r <= rounds; r++) opts += '<option value="' + r + '"' + (String(r) === mapState.round ? ' selected' : '') + '>Ronda ' + r + '</option>';
    document.getElementById('mapRound').innerHTML = opts;

    const secs = mapState.radar ? mapState.radar.sections : [];
    document.getElementById('mapFloor').innerHTML = secs.length > 1
        ? secs.map((s, i) => '<button class="' + (i === mapState.floor ? 'on' : '') + '" onclick="setMapFilter(\'floor\', ' + i + ')">' + (i === 0 ? 'Arriba' : 'Abajo' + (secs.length > 2 ? ' ' + i : '')) + '</button>').join('')
        : '';
    document.getElementById('mapLayers').innerHTML = MAP_LAYERS.map(l =>
        '<label class="check"><input type="checkbox"' + (mapState.layers[l[0]] ? ' checked' : '') + ' onchange="setMapLayer(\'' + l[0] + '\', this.checked)">' + l[1] + '</label>').join('');
}

function setMapFilter(kind, value) {
    mapState[kind] = kind === 'floor' ? Number(value) : value;
    if (kind === 'floor') renderMapFilters();
    drawMap();
}

function setMapLayer(layer, on) {
    mapState.layers[layer] = on;
    drawMap();
}

// ── Geometry and filters ──

function mapPx(x, y) {
    const r = mapState.radar;
    const k = 1024 / r.size;
    return [(x - r.posX) / r.scale * k, (r.posY - y) / r.scale * k];
}

// Floor of a height: the first extra floor whose range contains it, else the main one
function floorOf(z) {
    const secs = mapState.radar.sections;
    for (let i = 1; i < secs.length; i++) if (z >= secs[i].altMin && z < secs[i].altMax) return i;
    return 0;
}

function roundRange() {
    if (!mapState.round) return [-Infinity, Infinity];
    const starts = currentReport.header.roundStarts || [];
    const r = Number(mapState.round);
    return [starts[r - 1], r < starts.length ? starts[r] : Infinity];
}

// Position of a player at a tick, from their once-per-second path (null if they
// weren't alive or sampled within 2 s of it).
function positionAt(steamId, tick) {
    const path = mapData().paths.find(p => p.steamId === String(steamId));
    if (!path) return null;
    const pts = path.points;
    let best = null, bestD = Infinity;
    for (let i = 0; i < pts.length; i += 4) {
        const d = Math.abs(pts[i] - tick);
        if (d < bestD) { bestD = d; best = i; }
        if (pts[i] > tick && d > bestD) break;
    }
    if (best === null || bestD > 2 * (mapData().every || 64)) return null;
    return [pts[best + 1], pts[best + 2], pts[best + 3]];
}

// ── Drawing ──

function drawMap() {
    const canvas = document.getElementById('mapCanvas');
    const ctx = canvas.getContext('2d');
    const empty = document.getElementById('mapEmpty');
    ctx.clearRect(0, 0, 1024, 1024);
    mapState.markers = [];
    const legend = document.getElementById('mapLegend');

    if (!mapState.radar) {
        empty.innerText = mapState.error || 'Cargando el radar del mapa…';
        empty.style.display = 'flex';
        legend.innerHTML = '';
        return;
    }
    empty.style.display = 'none';
    const img = mapState.images[mapState.floor];
    if (img && img.complete) ctx.drawImage(img, 0, 0, 1024, 1024);
    // Dim the radar a little so markers stand out
    ctx.fillStyle = 'rgba(0,0,0,0.25)';
    ctx.fillRect(0, 0, 1024, 1024);

    const md = mapData();
    const [t0, t1] = roundRange();
    const inRound = t => t >= t0 && t < t1;
    const sel = mapState.player;
    const onFloor = z => floorOf(z) === mapState.floor;

    if (mapState.layers.heat) drawHeat(ctx, md, sel, inRound, onFloor);
    if (mapState.layers.paths) drawPaths(ctx, md, sel, inRound, onFloor);
    if (mapState.layers.kills) drawKills(ctx, md, sel, inRound, onFloor);
    if (mapState.layers.shift) drawShift(ctx, sel, inRound, onFloor);
    if (mapState.layers.incidents) drawIncidents(ctx, sel, inRound, onFloor);

    legend.innerHTML =
        '<div><span class="dot" style="background:' + MAP_TEAM[3] + '"></span>CT <span class="dot" style="background:' + MAP_TEAM[2] + '; margin-left:10px;"></span>TT</div>' +
        (mapState.layers.kills ? '<div><span class="dot ring"></span>quién mató · <b class="x">×</b> dónde murió</div>' : '') +
        (mapState.layers.incidents ? '<div><span class="diamond"></span>incidente</div>' : '') +
        (mapState.layers.shift ? '<div><span class="tri"></span>caminó con SHIFT · línea al enemigo que no veía</div>' : '');
}

function drawPaths(ctx, md, sel, inRound, onFloor) {
    ctx.lineWidth = sel ? 3 : 1.5;
    md.paths.forEach(p => {
        if (sel && p.steamId !== sel) return;
        ctx.strokeStyle = sel ? 'rgba(255,255,255,0.85)' : 'rgba(255,255,255,0.22)';
        ctx.beginPath();
        const gap = 2 * (md.every || 64);
        let prev = null; // tick of the last drawn sample; null breaks the line
        for (let i = 0; i < p.points.length; i += 4) {
            const tick = p.points[i];
            if (!inRound(tick) || !onFloor(p.points[i + 3])) { prev = null; continue; }
            const [x, y] = mapPx(p.points[i + 1], p.points[i + 2]);
            // A gap means death, a new round or another floor: start a new line
            if (prev === null || tick - prev > gap) ctx.moveTo(x, y); else ctx.lineTo(x, y);
            prev = tick;
        }
        ctx.stroke();
    });
}

// Heatmap: time spent at each place, accumulated on a coarse grid, blurred and
// colored with the same scale as the body figure.
function drawHeat(ctx, md, sel, inRound, onFloor) {
    const N = 256, grid = new Float32Array(N * N);
    md.paths.forEach(p => {
        if (sel && p.steamId !== sel) return;
        for (let i = 0; i < p.points.length; i += 4) {
            if (!inRound(p.points[i]) || !onFloor(p.points[i + 3])) continue;
            const [x, y] = mapPx(p.points[i + 1], p.points[i + 2]);
            const gx = Math.floor(x / 4), gy = Math.floor(y / 4);
            if (gx >= 0 && gy >= 0 && gx < N && gy < N) grid[gy * N + gx] += 1;
        }
    });
    // Two box blurs ≈ a soft gaussian
    const blur = (src) => {
        const out = new Float32Array(N * N), R = 3;
        for (let y = 0; y < N; y++) for (let x = 0; x < N; x++) {
            let s = 0;
            for (let dy = -R; dy <= R; dy++) for (let dx = -R; dx <= R; dx++) {
                const yy = y + dy, xx = x + dx;
                if (yy >= 0 && xx >= 0 && yy < N && xx < N) s += src[yy * N + xx];
            }
            out[y * N + x] = s;
        }
        return out;
    };
    const g = blur(blur(grid));
    let max = 0;
    for (let i = 0; i < g.length; i++) max = Math.max(max, g[i]);
    if (!max) return;
    const tile = document.createElement('canvas');
    tile.width = tile.height = N;
    const tctx = tile.getContext('2d');
    const data = tctx.createImageData(N, N);
    for (let i = 0; i < g.length; i++) {
        const v = Math.sqrt(g[i] / max); // sqrt: show quiet areas too, not only the hottest spot
        if (v < 0.04) continue;
        const c = heatColor(v * BODY_HEAT_MAX).match(/\d+/g);
        data.data[i * 4] = +c[0]; data.data[i * 4 + 1] = +c[1]; data.data[i * 4 + 2] = +c[2];
        data.data[i * 4 + 3] = Math.round(70 + 150 * v);
    }
    tctx.putImageData(data, 0, 0);
    ctx.imageSmoothingEnabled = true;
    ctx.drawImage(tile, 0, 0, 1024, 1024);
}

function drawKills(ctx, md, sel, inRound, onFloor) {
    md.kills.forEach(k => {
        if (!inRound(k.tick)) return;
        if (sel && k.killerId !== sel && k.victimId !== sel) return;
        const kOn = onFloor(k.kz), vOn = onFloor(k.vz);
        if (!kOn && !vOn) return;
        const [kx, ky] = mapPx(k.kx, k.ky), [vx, vy] = mapPx(k.vx, k.vy);
        const mine = sel && k.killerId === sel;
        const died = sel && k.victimId === sel;
        ctx.strokeStyle = died ? 'rgba(229,83,75,0.8)' : 'rgba(255,255,255,0.35)';
        ctx.lineWidth = sel ? 2 : 1;
        if (kOn && vOn) { ctx.beginPath(); ctx.moveTo(kx, ky); ctx.lineTo(vx, vy); ctx.stroke(); }
        if (kOn) {
            ctx.fillStyle = MAP_TEAM[k.killerTeam] || '#ccc';
            ctx.beginPath(); ctx.arc(kx, ky, mine ? 8 : 6, 0, Math.PI * 2); ctx.fill();
            ctx.strokeStyle = '#000'; ctx.lineWidth = 1.5; ctx.stroke();
        }
        if (vOn) {
            ctx.strokeStyle = died ? '#e5534b' : (MAP_TEAM[k.victimTeam] || '#ccc');
            ctx.lineWidth = died ? 4 : 3;
            const s = died ? 9 : 7;
            ctx.beginPath(); ctx.moveTo(vx - s, vy - s); ctx.lineTo(vx + s, vy + s); ctx.moveTo(vx + s, vy - s); ctx.lineTo(vx - s, vy + s); ctx.stroke();
        }
        const info = '<b>' + esc(k.killer) + '</b> mató a <b>' + esc(k.victim) + '</b>' +
            '<div class="muted">' + esc(k.weapon || '') + (k.headshot ? ' · headshot' : '') + ' · ronda ' + roundOf(k.tick) + ' · a ' + Math.round(Math.hypot(k.kx - k.vx, k.ky - k.vy)) + ' u</div>';
        if (kOn) mapState.markers.push({ x: kx, y: ky, r: 10, tick: k.tick, sid: k.killerId, info: info });
        if (vOn) mapState.markers.push({ x: vx, y: vy, r: 10, tick: k.tick, sid: k.victimId, info: info });
    });
}

function drawIncidents(ctx, sel, inRound, onFloor) {
    (currentReport.players || []).forEach(p => {
        const sid = String(p.steamId);
        if (sel && sid !== sel) return;
        (p.incidents || []).forEach(inc => {
            if (!inRound(inc.tick)) return;
            const pos = positionAt(sid, inc.tick);
            if (!pos || !onFloor(pos[2])) return;
            const [x, y] = mapPx(pos[0], pos[1]);
            ctx.fillStyle = RED_TYPES.test(inc.type) ? '#e5534b' : '#d4b23f';
            ctx.strokeStyle = '#000'; ctx.lineWidth = 1.5;
            ctx.beginPath(); ctx.moveTo(x, y - 10); ctx.lineTo(x + 10, y); ctx.lineTo(x, y + 10); ctx.lineTo(x - 10, y); ctx.closePath();
            ctx.fill(); ctx.stroke();
            mapState.markers.push({ x: x, y: y, r: 12, tick: inc.tick, sid: sid,
                info: typeTag(inc.type) + ' <b>' + esc(p.name) + '</b><div class="muted">ronda ' + roundOf(inc.tick) + '</div><div style="margin-top:4px;">' + esc(inc.note) + '</div>' });
        });
    });
}

function drawShift(ctx, sel, inRound, onFloor) {
    (currentReport.shiftLog || []).forEach(c => {
        if (sel && c.steamId !== sel) return;
        if (!inRound(c.tick)) return;
        const pos = positionAt(c.steamId, c.tick);
        if (!pos || !onFloor(pos[2])) return;
        const [x, y] = mapPx(pos[0], pos[1]);
        // Dashed line to each enemy they couldn't see, where that enemy was
        ctx.setLineDash([6, 5]);
        ctx.strokeStyle = 'rgba(212,178,63,0.7)';
        ctx.lineWidth = 1.5;
        (c.enemies || []).forEach(name => {
            const path = mapData().paths.find(pp => pp.name === name);
            const ep = path && positionAt(path.steamId, c.tick);
            if (!ep) return;
            const [ex, ey] = mapPx(ep[0], ep[1]);
            ctx.beginPath(); ctx.moveTo(x, y); ctx.lineTo(ex, ey); ctx.stroke();
        });
        ctx.setLineDash([]);
        ctx.fillStyle = '#d4b23f'; ctx.strokeStyle = '#000'; ctx.lineWidth = 1.5;
        ctx.beginPath(); ctx.moveTo(x, y - 9); ctx.lineTo(x + 8, y + 7); ctx.lineTo(x - 8, y + 7); ctx.closePath();
        ctx.fill(); ctx.stroke();
        mapState.markers.push({ x: x, y: y, r: 11, tick: c.tick, sid: c.steamId,
            info: '<span class="tag tag-yellow">SHIFT</span> <b>' + esc(c.player) + '</b><div class="muted">ronda ' + c.round + ' · ' + c.presses + (c.presses === 1 ? ' apretada' : ' apretadas') + ' · ' + c.heldSecs.toLocaleString('es-AR') + ' s</div>' +
                '<div style="margin-top:4px;">Cerca sin ver: ' + (c.enemies || []).map(esc).join(', ') + ' (a ' + c.minDist + ' u)</div>' });
    });
}

// Click: the closest marker under the pointer
document.addEventListener('click', e => {
    if (e.target.id !== 'mapCanvas') return;
    const rect = e.target.getBoundingClientRect();
    const x = (e.clientX - rect.left) * 1024 / rect.width;
    const y = (e.clientY - rect.top) * 1024 / rect.height;
    let best = null, bestD = Infinity;
    mapState.markers.forEach(m => {
        const d = Math.hypot(m.x - x, m.y - y);
        if (d <= m.r * 1.6 && d < bestD) { best = m; bestD = d; }
    });
    if (!best) return;
    document.getElementById('mapInfo').innerHTML = best.info + '<div style="margin-top:8px;">' + playButton(best.tick, best.sid) + '</div>';
});

// Opens the minimap filtered to one player, with their path (from the player detail).
function showOnMap(steamId) {
    closePlayer();
    mapState.player = String(steamId);
    mapState.round = '';
    mapState.layers.paths = true;
    renderMapFilters();
    drawMap();
    document.getElementById('mapCard').scrollIntoView({ behavior: 'smooth', block: 'start' });
}
