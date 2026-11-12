// ── History ───────────────────────────────────────────────────────
function refreshHistory() {
    if (!window.listHistory) return;
    window.listHistory().then(items => {
        historyItems = items || [];
        renderHistory();
    });
}

let histFilter = 'all';
let asideTab = 'demos';

function setHistFilter(f, el) {
    histFilter = f;
    document.querySelectorAll('#histFilter button').forEach(b => b.classList.toggle('on', b === el));
    renderHistory();
}

function renderHistory() {
    const list = document.getElementById('histList');
    const q = document.getElementById('histSearch').value.toLowerCase();
    const catSel = document.getElementById('histCategory');
    document.getElementById('histCount').innerText = historyItems.length || '';

    // Demos analyzed by older detectors miss newer detections (e.g. prefire)
    const analysisVer = Number(appInfo.analysisVersion || 0);
    const outdated = analysisVer ? historyItems.filter(e => (e.analysisVersion || 0) < analysisVer).length : 0;
    document.getElementById('outdatedNotice').style.display = outdated ? 'flex' : 'none';
    document.getElementById('outdatedText').innerText = outdated + (outdated === 1 ? ' demo se analizó' : ' demos se analizaron') + ' con una versión anterior y no tienen las detecciones nuevas.';

    // Category options come from what is actually in the history
    const cats = [...new Set(historyItems.flatMap(e => e.flagTypes || []))].sort();
    const current = catSel.value;
    catSel.innerHTML = '<option value="">Todas las categorías</option>' +
        cats.map(c => '<option value="' + esc(c) + '"' + (c === current ? ' selected' : '') + '>' + esc(c) + '</option>').join('');
    const cat = cats.includes(current) ? current : '';

    const items = historyItems.filter(e => {
        const lvl = e.level || 'limpio';
        if (histFilter === 'cheat' && lvl !== 'cheat') return false;
        if (histFilter === 'suspect' && lvl !== 'sospechoso') return false;
        if (histFilter === 'clean' && (lvl === 'cheat' || lvl === 'sospechoso')) return false;
        if (cat && !(e.flagTypes || []).includes(cat)) return false;
        return !q || (e.demoName + ' ' + e.mapName + ' ' + (e.flagged || []).join(' ')).toLowerCase().includes(q);
    });

    if (historyItems.length === 0) {
        list.innerHTML = '<div class="hist-empty">Las demos que analices van a aparecer acá.</div>';
        return;
    }
    if (items.length === 0) {
        list.innerHTML = '<div class="hist-empty">Ninguna demo coincide con el filtro.</div>';
        return;
    }

    const activeId = currentMeta && currentMeta.historyId;
    list.innerHTML = items.map(e => {
        const lvl = e.level || 'limpio';
        const date = (e.analyzedAt || '').slice(0, 10).split('-').reverse().join('/');
        const tags = (e.flagTypes || []).map(flagTag).join('');
        const levelText = { cheat: '<span class="hist-flag">cheat</span>', sospechoso: '<span class="t-yellow">sospechosa</span>', conducta: '<span>conducta</span>', limpio: '<span>limpia</span>' }[lvl];
        return '<div class="hist-item' + (e.id === activeId ? ' active' : '') + (lvl === 'cheat' ? ' flagged' : (lvl === 'sospechoso' ? ' suspect' : '')) + '" tabindex="0" onclick="openHistoryItem(\'' + e.id + '\')" onkeydown="if(event.key===\'Enter\')openHistoryItem(\'' + e.id + '\')" title="' + esc(e.demoPath) + '">' +
            '<div class="hist-name">' + esc(e.demoName) + '</div>' +
            '<div class="hist-sub"><span>' + esc(e.mapName) + '</span><span>' + date + '</span>' + levelText + '</div>' +
            (tags ? '<div class="hist-tags">' + tags + '</div>' : '') +
            '<button class="hist-remove" onclick="event.stopPropagation(); removeHistoryItem(\'' + e.id + '\')" title="Quitar del historial (no borra la demo)">' + icon('x', 'ic-sm') + '</button>' +
            '</div>';
    }).join('');
}

function runReanalyzeOutdated() {
    if (window.reanalyzeOutdated) window.reanalyzeOutdated();
}

function openHistoryItem(id) {
    if (window.openHistory) window.openHistory(id);
}

function removeHistoryItem(id) {
    if (window.removeHistory) window.removeHistory(id).then(() => {
        refreshHistory();
        if (asideTab === 'players') { searchPlayers(); showOverview(); }
    });
}
