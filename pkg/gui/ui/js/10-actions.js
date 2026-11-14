// ── Actions ───────────────────────────────────────────────────────
function pickDemos() {
    if (window.selectDemoFile) window.selectDemoFile();
}

function reanalyze() {
    if (window.reanalyzeCurrent) window.reanalyzeCurrent();
}

function playInGame(tick, steamId) {
    if (window.launchInGame) window.launchInGame(tick, steamId || '');
}

function openLink(url) {
    if (window.openURL) window.openURL(url);
}

function copyText(text, msg) {
    const done = () => showToast(msg);
    if (window.copyToClipboard) {
        window.copyToClipboard(text).then(done).catch(() => fallbackCopy(text, msg));
    } else {
        fallbackCopy(text, msg);
    }
}

function fallbackCopy(text, msg) {
    try {
        const ta = document.createElement('textarea');
        ta.value = text;
        ta.style.position = 'fixed';
        ta.style.left = '-9999px';
        document.body.appendChild(ta);
        ta.select();
        const ok = document.execCommand('copy');
        document.body.removeChild(ta);
        showToast(ok ? msg : 'No se pudo copiar');
    } catch (e) {
        showToast('No se pudo copiar');
    }
}

function copyStaffSummary() {
    if (!currentReport) return;
    const h = currentReport.header;
    const q = String.fromCharCode(96);
    let md = '**SourceSight** · ' + q + h.mapName + q + ' · ' + (h.roundStarts || []).length + ' rondas · ' + (stripEmoji(h.serverName) || 'servidor local') + '\n';
    const marked = currentReport.players.filter(flagged);
    if (!marked.length) {
        md += 'Sin detecciones.\n';
    } else {
        marked.forEach(p => {
            const label = confirmationOf(p) ? 'Confirmado' : LEVEL_LABEL[levelOf(p)];
            md += '- **' + p.name + '** (' + p.steamId + ') · ' + label + ((p.flags || []).length ? ': ' + q + p.flags.join(' + ') + q : '') + '\n';
        });
    }
    const tl = currentReport.timeline || [];
    if (tl.length) {
        md += '\nPrimeros incidentes:\n';
        tl.slice(0, 5).forEach(inc => {
            md += '- R' + roundOf(inc.tick) + ' ' + inc.player + ' ' + inc.type + ': ' + q + inc.gotoCmd + q + (inc.ping ? ' · ' + inc.ping + ' ms' + (inc.pingSpike ? ' (pico)' : '') : '') + '\n';
        });
    }
    copyText(md, 'Resumen copiado');
}

let toastTimer = null;
function showToast(msg) {
    const t = document.getElementById('toast');
    t.innerText = msg;
    t.classList.add('show');
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => t.classList.remove('show'), 3200);
}
