// ── Settings & updates ────────────────────────────────────────────
function openSettings() {
    if (window.getConfig) {
        window.getConfig().then(s => {
            try {
                const cfg = JSON.parse(s);
                if (cfg.csgoPath) document.getElementById('cfgCSGOPath').value = cfg.csgoPath;
                if (cfg.prerollSeconds) document.getElementById('cfgPreroll').value = cfg.prerollSeconds.toFixed(1);
                document.getElementById('cfgShiftRadius').value = String(cfg.shiftRadius || 1000);
                document.getElementById('cfgHotkeys').value = cfg.hotkeys || 'ctrl-alt-arrows';
            } catch (e) {}
        });
    }
}

// Settings is a page now; "closing" it goes home
function closeSettings() { goSection('home'); }

function browseCSGO() {
    if (!window.browseCSGOPath) return;
    window.browseCSGOPath().then(p => { if (p) document.getElementById('cfgCSGOPath').value = p; });
}

function saveSettings() {
    const p = document.getElementById('cfgCSGOPath').value.trim();
    const preroll = parseFloat(document.getElementById('cfgPreroll').value) || 2.5;
    const shiftRadius = parseFloat(document.getElementById('cfgShiftRadius').value) || 1000;
    if (!window.saveConfig) { closeSettings(); return; }
    window.saveConfig(p, preroll, shiftRadius)
        .then(saved => window.saveHotkeys ? window.saveHotkeys(document.getElementById('cfgHotkeys').value).then(() => saved) : saved)
        .then(saved => {
            if (saved) document.getElementById('cfgCSGOPath').value = saved;
            showToast(saved ? 'Guardado. Carpeta del juego: ' + saved : 'Ajustes guardados');
        })
        .catch(err => showToast('No se pudo guardar: ' + (err && err.message ? err.message : err)));
}

function openRepo() { if (appInfo.repoUrl) openLink(appInfo.repoUrl); }

function checkUpdates(manual) {
    if (!window.checkUpdate) return;
    window.checkUpdate().then(info => {
        if (info.error) {
            if (manual) showToast('No se pudo consultar GitHub: ' + info.error);
            return;
        }
        if (!info.available) {
            if (manual) showToast(info.current === 'dev' ? 'Versión de desarrollo: las actualizaciones están desactivadas' : 'Ya tenés la última versión');
            return;
        }
        updateInfo = info;
        document.getElementById('updateText').innerText = 'Hay una versión nueva: ' + info.version + ' (tenés ' + info.current + ').';
        document.getElementById('updateApply').style.display = info.canApply ? 'inline-flex' : 'none';
        document.getElementById('updateBar').style.display = 'flex';
        if (manual) closeSettings();
    });
}

function openUpdatePage() { if (updateInfo) openLink(updateInfo.pageUrl); }

function startUpdate() {
    if (!window.applyUpdate) return;
    document.getElementById('updateApply').disabled = true;
    window.applyUpdate();
}

function dismissUpdate() { document.getElementById('updateBar').style.display = 'none'; }

// ── Automatic archive (pkg/archive) ──
const ARC_MODES = ['Competitivo', 'AutoMIX', 'Retake', 'Casual', 'DeathMatch'];

function fillArchive(a) {
    const servers = a.servers || [], modes = a.modes || (a.enabled ? [] : ['Competitivo']);
    document.getElementById('arcEnabled').checked = !!a.enabled;
    document.getElementById('arcDir').value = a.dir || '';
    document.getElementById('arcMaps').value = (a.maps || []).join(', ');
    document.getElementById('arcEvery').value = String(a.everyMin || 60);
    document.getElementById('arcURL').value = a.url || '';
    document.getElementById('arcServers').innerHTML = [1, 2, 3, 4, 5, 6, 7, 8].map(n =>
        '<label class="check"><input type="checkbox" value="' + n + '"' + (servers.includes(n) ? ' checked' : '') + '>#' + n + '</label>').join('');
    document.getElementById('arcModes').innerHTML = ARC_MODES.map(m =>
        '<label class="check"><input type="checkbox" value="' + m + '"' + (modes.some(x => x.toLowerCase() === m.toLowerCase()) ? ' checked' : '') + '>' + m + '</label>').join('');
    toggleArchiveOptions();
}

function toggleArchiveOptions() {
    document.getElementById('arcOptions').style.display = document.getElementById('arcEnabled').checked ? 'block' : 'none';
}
document.addEventListener('change', e => { if (e.target.id === 'arcEnabled') toggleArchiveOptions(); });

function readArchive() {
    const checked = id => Array.from(document.querySelectorAll('#' + id + ' input:checked')).map(i => i.value);
    return {
        enabled: document.getElementById('arcEnabled').checked,
        dir: document.getElementById('arcDir').value.trim(),
        url: document.getElementById('arcURL').value.trim(),
        servers: checked('arcServers').map(Number),
        modes: checked('arcModes'),
        maps: document.getElementById('arcMaps').value.split(',').map(s => s.trim()).filter(Boolean),
        everyMin: Number(document.getElementById('arcEvery').value) || 60
    };
}

function refreshArchiveStatus() {
    if (!window.getArchiveStatus) return;
    window.getArchiveStatus().then(s => {
        document.getElementById('arcStatus').innerText = s.running ? 'Revisando la página de demos…'
            : s.lastRun ? 'Última revisión: ' + s.lastRun + '. ' + (s.lastNote || '') : 'Todavía no revisó en esta sesión.';
        if (currentSection === 'archive') setTimeout(refreshArchiveStatus, 3000);
    });
}

function runArchiveNowClick() {
    if (!window.saveArchiveConfig) return;
    // Save first, so "Revisar ahora" uses what is on screen
    window.saveArchiveConfig(readArchive())
        .then(() => window.runArchiveNow())
        .then(() => { showToast('Revisando la página de demos…'); refreshArchiveStatus(); })
        .catch(err => showToast('No se pudo: ' + (err && err.message ? err.message : err)));
}

// ── Archive page (Archivo in the rail) ──
function renderArchivePage() {
    if (window.getConfig) window.getConfig().then(s => {
        let cfg = {};
        try { cfg = JSON.parse(s); } catch (e) { /* defaults */ }
        fillArchive(cfg.archive || {});
        updateArchiveDot(cfg.archive || {});
    });
    refreshArchiveStatus();
    renderArchiveSummary();
    renderArchived();
}

function renderArchiveSummary() {
    if (!window.getArchiveStatus) return;
    window.getArchiveStatus().then(st => {
        const box = document.getElementById('archiveSummary');
        box.innerHTML =
            '<div class="sum-cell"><div class="sum-label">Estado</div><div class="sum-value" style="font-size:16px;">' + (st.running ? 'Revisando…' : document.getElementById('arcEnabled').checked ? 'Activo' : 'Apagado') + '</div>' +
                '<div class="sum-sub">' + (st.lastRun ? 'última revisión ' + esc(st.lastRun) : 'sin revisiones en esta sesión') + '</div></div>' +
            '<div class="sum-cell"><div class="sum-label">Analizadas</div><div class="sum-value">' + (st.analyzed || 0) + '</div><div class="sum-sub">en esta sesión</div></div>' +
            '<div class="sum-cell"><div class="sum-label">Guardadas</div><div class="sum-value">' + (st.kept || 0) + '</div><div class="sum-sub">en esta sesión, con cheat o sospechosos</div></div>';
    });
}

function renderArchived() {
    if (!window.listArchived) return;
    window.listArchived().then(rows => {
        rows = rows || [];
        document.getElementById('archivedCount').innerText = rows.length || '';
        const body = document.getElementById('archivedBody');
        if (!rows.length) {
            body.innerHTML = emptyRow(5, 'Todavía no se guardó ninguna demo.');
            return;
        }
        body.innerHTML = rows.map(r => {
            const d = r.demo;
            const open = r.historyId ? 'openHistoryItem(\'' + esc(r.historyId) + '\')' : 'openArchivedPath(' + JSON.stringify(d.path).replace(/"/g, '&quot;') + ')';
            return '<tr class="click" onclick="' + open + '">' +
                '<td class="player">' + esc(d.file) + '<div class="muted" style="font-size:12px;">#' + d.server + ' ' + esc(d.mode || '') + '</div></td>' +
                '<td class="muted">' + esc(d.map || '') + '</td>' +
                '<td>' + levelTag(d.level) + ' <span class="tags">' + (d.flags || []).map(flagTag).join('') + '</span></td>' +
                '<td class="r num">' + Math.round(d.sizeMB || 0) + ' MB</td>' +
                '<td class="r muted">' + fmtDate((d.at || '').slice(0, 10)) + '</td>' +
            '</tr>';
        }).join('');
    });
}

function openArchivedPath(path) {
    if (window.openDemoPath) window.openDemoPath(path);
}

function saveArchiveSettings() {
    if (!window.saveArchiveConfig) return;
    const cfg = readArchive();
    window.saveArchiveConfig(cfg)
        .then(saved => {
            fillArchive(saved || cfg);
            updateArchiveDot(saved || cfg);
            renderArchiveSummary();
            showToast(cfg.enabled ? 'Archivo automático activado' : 'Archivo automático apagado');
        })
        .catch(err => showToast('No se pudo guardar: ' + (err && err.message ? err.message : err)));
}

// Small dot on the rail when the archive is on
function updateArchiveDot(a) {
    document.getElementById('archiveDot').style.display = a && a.enabled ? 'block' : 'none';
}
