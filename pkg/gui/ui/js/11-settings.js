// ── Settings & updates ────────────────────────────────────────────
function openSettings() {
    if (window.getConfig) {
        window.getConfig().then(s => {
            try {
                const cfg = JSON.parse(s);
                if (cfg.csgoPath) document.getElementById('cfgCSGOPath').value = cfg.csgoPath;
                if (cfg.prerollSeconds) document.getElementById('cfgPreroll').value = cfg.prerollSeconds.toFixed(1);
                document.getElementById('cfgShiftRadius').value = String(cfg.shiftRadius || 1000);
            } catch (e) {}
        });
    }
    document.getElementById('settingsModal').style.display = 'flex';
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
