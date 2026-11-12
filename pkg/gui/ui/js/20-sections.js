// ── Sections: side rail and home page ─────────────────────────────────
// The rail on the left picks a section: Inicio, Demos, Jugadores, Archivo, Ajustes.
// Every main view belongs to one (VIEW_SECTION); showView() calls setChrome() so the
// rail, the left list and the top bar always match what is on screen.

let currentSection = 'home';

const VIEW_SECTION = {
    drop: 'demos', loading: 'demos', batch: 'demos', audit: 'demos',
    overview: 'players', player: 'players',
    home: 'home', archive: 'archive', settings: 'settings'
};
const SECTION_NAME = { home: 'Inicio', demos: 'Demos', players: 'Jugadores', archive: 'Archivo automático', settings: 'Ajustes' };

// Rail highlight, left list (only Demos and Jugadores have one) and top bar
function setChrome(section, view) {
    currentSection = section;
    document.querySelectorAll('.rail-item').forEach(b => b.classList.toggle('on', b.dataset.section === section));
    const pane = section === 'demos' || section === 'players' ? section : '';
    document.getElementById('aside').style.display = pane ? 'flex' : 'none';
    document.getElementById('paneDemos').style.display = pane === 'demos' ? 'flex' : 'none';
    document.getElementById('panePlayers').style.display = pane === 'players' ? 'flex' : 'none';
    // Top bar: the demo's name and its actions when a report is open, else the section
    const report = view === 'audit';
    // Pages with their own title (Inicio, Archivo, Ajustes, the player pages) don't repeat it here
    document.getElementById('pageName').style.display = view === 'drop' ? '' : 'none';
    document.getElementById('pageName').innerText = SECTION_NAME[section] || '';
    document.getElementById('btnReanalyze').style.display = report ? 'inline-flex' : 'none';
    document.getElementById('exportMenu').style.display = report ? 'inline-flex' : 'none';
    if (!report) closeExportMenu();
}

function goSection(section) {
    switch (section) {
        case 'home':
            showView('home');
            renderHome();
            break;
        case 'demos':
            if (typeof batchRunning !== 'undefined' && batchRunning) showView('batch');
            else restoreIdle();
            break;
        case 'players':
            searchPlayers();
            if (activePlayer) openPlayerPage(activePlayer); else showOverview();
            break;
        case 'archive':
            showView('archive');
            renderArchivePage();
            break;
        case 'settings':
            showView('settings');
            openSettings();
            break;
    }
}

// ── Export menu (top bar) ──
function toggleExportMenu(e) {
    e.stopPropagation();
    document.getElementById('exportMenuList').classList.toggle('open');
}
function closeExportMenu() {
    const m = document.getElementById('exportMenuList');
    if (m) m.classList.remove('open');
}
document.addEventListener('click', closeExportMenu);

// ── Home ──
// What needs attention first: demos with cheat or suspects, totals of the history,
// the automatic archive and demos analyzed by an older version.
function renderHome() {
    const el = document.getElementById('homeView');
    const flagged = historyItems.filter(e => e.level === 'cheat' || e.level === 'sospechoso').slice(0, 8);
    const analysisVer = Number(appInfo.analysisVersion || 0);
    const outdated = analysisVer ? historyItems.filter(e => (e.analysisVersion || 0) < analysisVer).length : 0;

    el.innerHTML =
        '<div class="page-head"><div><div class="page-title">Inicio</div>' +
        '<div class="page-sub">Lo que conviene revisar primero. Para analizar, soltá demos en cualquier parte de la ventana.</div></div>' +
        '</div>' +
        (outdated ? '<div class="notice"><span class="grow">' + outdated + (outdated === 1 ? ' demo se analizó' : ' demos se analizaron') +
            ' con una versión anterior y no tienen las detecciones y datos nuevos.</span><button class="btn" onclick="runReanalyzeOutdated()">Re-analizar</button></div>' : '') +
        '<div class="summary" id="homeTotals"><div class="sum-cell"><div class="sum-label">Historial</div><div class="sum-value">' + historyItems.length + '</div><div class="sum-sub">demos analizadas</div></div></div>' +
        '<div class="home-grid">' +
            '<section class="panel"><div class="panel-head"><div class="panel-title">Para revisar</div>' +
                '<a class="link" onclick="goSection(\'demos\')">Ver todas las demos</a></div>' +
                (flagged.length
                    ? '<div class="table-wrap"><table><tbody>' + flagged.map(e =>
                        '<tr class="click" onclick="openHistoryItem(\'' + esc(e.id) + '\')">' +
                            '<td>' + levelTag(e.level) + '</td>' +
                            '<td class="player">' + esc(e.demoName) + '<div class="muted" style="font-size:12px;">' + esc(e.mapName) + ' · ' + fmtDate(e.analyzedAt) + '</div></td>' +
                            '<td class="wrap"><span class="tags">' + (e.flagTypes || []).map(flagTag).join('') + '</span></td>' +
                        '</tr>').join('') + '</tbody></table></div>'
                    : '<div class="panel-body muted">Ninguna demo del historial tiene cheats ni sospechosos.</div>') +
            '</section>' +
            '<section class="panel" id="homeArchive"><div class="panel-head"><div class="panel-title">Archivo automático</div>' +
                '<a class="link" onclick="goSection(\'archive\')">Abrir</a></div><div class="panel-body muted">Cargando…</div></section>' +
        '</div>';

    if (window.getOverview) window.getOverview().then(ov => {
        if (!ov || currentSection !== 'home') return;
        document.getElementById('homeTotals').innerHTML =
            '<div class="sum-cell"><div class="sum-label">Demos</div><div class="sum-value">' + ov.demos + '</div><div class="sum-sub">' + ov.cheatDemos + ' con cheat · ' + ov.suspectDemos + ' sospechosas</div></div>' +
            '<div class="sum-cell"><div class="sum-label">Jugadores</div><div class="sum-value">' + ov.players + '</div><div class="sum-sub">distintos en todas tus demos</div></div>' +
            '<div class="sum-cell"><div class="sum-label">Con cheat</div><div class="sum-value' + (ov.cheatPlayers ? ' t-red' : '') + '">' + ov.cheatPlayers + '</div><div class="sum-sub">' + ov.suspectPlayers + ' sospechosos</div></div>' +
            '<div class="sum-cell"><div class="sum-label">Confirmados</div><div class="sum-value">' + Object.keys(confirmedMap).length + '</div><div class="sum-sub">por el staff</div></div>';
    });
    renderHomeArchive();
}

function renderHomeArchive() {
    if (!window.getConfig || !window.getArchiveStatus) return;
    Promise.all([window.getConfig(), window.getArchiveStatus()]).then(([cfgText, st]) => {
        const box = document.querySelector('#homeArchive .panel-body');
        if (!box) return;
        let a = {};
        try { a = JSON.parse(cfgText).archive || {}; } catch (e) { /* keep defaults */ }
        box.className = 'panel-body';
        box.innerHTML = !a.enabled
            ? '<span class="muted">Apagado. Bajá automáticamente las demos del servidor y guardá solo las que tengan algo para revisar.</span>' +
              '<div style="margin-top:10px;"><button class="btn" onclick="goSection(\'archive\')">Configurar</button></div>'
            : '<dl class="kv">' +
                '<dt>Estado</dt><dd>' + (st.running ? 'Revisando la página de demos…' : 'Activo, revisa cada ' + (a.everyMin || 60) + ' min') + '</dd>' +
                '<dt>Última revisión</dt><dd>' + (st.lastRun || 'todavía no revisó en esta sesión') + '</dd>' +
                '<dt>En esta sesión</dt><dd>' + (st.analyzed || 0) + ' analizadas · ' + (st.kept || 0) + ' guardadas</dd>' +
                '<dt>Carpeta</dt><dd class="mono">' + esc(a.dir || '') + '</dd>' +
              '</dl>';
    });
}
