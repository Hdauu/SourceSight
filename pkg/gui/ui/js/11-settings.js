// ── Settings & updates ────────────────────────────────────────────
function openSettings() {
    if (window.getConfig) {
        window.getConfig().then(s => {
            try {
                const cfg = JSON.parse(s);
                if (cfg.csgoPath) document.getElementById('cfgCSGOPath').value = cfg.csgoPath;
                if (cfg.prerollSeconds) document.getElementById('cfgPreroll').value = cfg.prerollSeconds.toFixed(1);
                document.getElementById('cfgShiftRadius').value = String(cfg.shiftRadius || 1000);
                fillArchive(cfg.archive || {});
                document.getElementById('cfgHotkeys').value = cfg.hotkeys || 'ctrl-alt-arrows';
            } catch (e) {}
        });
    }
    document.getElementById('settingsModal').style.display = 'flex';
    refreshArchiveStatus();
}

function closeSettings() { document.getElementById('settingsModal').style.display = 'none'; }

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
        .then(saved => window.saveArchiveConfig ? window.saveArchiveConfig(readArchive()).then(() => saved) : saved)
        .then(saved => window.saveHotkeys ? window.saveHotkeys(document.getElementById('cfgHotkeys').value).then(() => saved) : saved)
        .then(saved => {
            if (saved) document.getElementById('cfgCSGOPath').value = saved;
            showToast(saved ? 'Guardado. Carpeta del juego: ' + saved : 'Configuración guardada');
            closeSettings();
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
        if (document.getElementById('settingsModal').style.display === 'flex') setTimeout(refreshArchiveStatus, 3000);
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
