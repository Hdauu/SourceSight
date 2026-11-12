// ── Called from Go ────────────────────────────────────────────────
function setProgress(pct, label) {
    document.getElementById('loadingBar').style.width = pct + '%';
    document.getElementById('loadingText').innerText = label || ('Analizando... ' + pct + '%');
}

// Only one main view is visible at a time
const VIEWS = { drop: 'dropzone', loading: 'loadingState', batch: 'batchView', audit: 'auditContent', overview: 'overviewView', player: 'playerView' };
function showView(name) {
    Object.keys(VIEWS).forEach(k => {
        document.getElementById(VIEWS[k]).style.display = k === name ? 'flex' : 'none';
    });
    const demoButtons = name === 'audit' ? 'inline-flex' : 'none';
    ['btnReanalyze', 'btnPNG', 'btnSummary'].forEach(id => document.getElementById(id).style.display = demoButtons);
    if (name !== 'audit') document.getElementById('demoMeta').innerHTML = '';
    document.getElementById('main').scrollTop = 0;
}

function showLoading() {
    setProgress(0, 'Abriendo demo...');
    showView('loading');
}

function restoreIdle() {
    if (currentReport) {
        showView('audit');
        renderDemoMeta();
    } else {
        showView('drop');
    }
}

function renderDemoMeta() {
    const h = currentReport.header;
    document.getElementById('demoMeta').innerHTML =
        '<span class="map">' + esc(h.mapName) + '</span>' +
        '<span>' + esc(h.demoFile || '') + '</span>' +
        '<span>' + esc(stripEmoji(h.serverName) || 'Servidor local') + '</span>';
}

function showPrompt(id, title, text, options) {
    document.getElementById('promptTitle').innerText = title;
    document.getElementById('promptText').innerText = text;
    const actions = document.getElementById('promptActions');
    actions.innerHTML = '';
    options.forEach(o => {
        const b = document.createElement('button');
        b.className = 'btn' + (o.primary ? ' btn-primary' : '');
        b.innerText = o.label;
        b.onclick = () => {
            document.getElementById('promptModal').style.display = 'none';
            if (window.answerPrompt) window.answerPrompt(id, o.id);
        };
        actions.appendChild(b);
    });
    document.getElementById('promptModal').style.display = 'flex';
    const primary = actions.querySelector('.btn-primary');
    if (primary) primary.focus();
}

// Top cards: cheats (detected or confirmed), suspects and reoffenders
function renderSummary() {
    const players = currentReport.players;
    const cheats = players.filter(p => confirmationOf(p) || levelOf(p) === 'cheat');
    const suspects = players.filter(p => !confirmationOf(p) && levelOf(p) === 'sospechoso');
    const conduct = players.filter(p => levelOf(p) === 'conducta');
    const reo = players.filter(p => p.previousInfractions > 0);

    document.getElementById('sumPlayers').innerText = players.length;
    document.getElementById('sumPlayersSub').innerText = (players.length - cheats.length - suspects.length) + ' sin indicios' +
        (conduct.length ? ' · ' + conduct.length + ' por conducta' : '');

    const sf = document.getElementById('sumFlags');
    sf.innerText = cheats.length;
    sf.className = 'sum-value' + (cheats.length ? ' t-red' : '');
    const parts = [];
    if (cheats.length) parts.push(cheats.map(p => p.name).join(', '));
    if (suspects.length) parts.push(suspects.length + (suspects.length === 1 ? ' sospechoso' : ' sospechosos'));
    document.getElementById('sumFlagsSub').innerText = parts.length ? parts.join(' · ') : 'Nada sospechoso';

    const sr = document.getElementById('sumReo');
    sr.innerText = reo.length;
    sr.className = 'sum-value' + (reo.length ? ' t-yellow' : '');
    document.getElementById('sumReoSub').innerText = reo.length ? reo.map(p => p.name).join(', ') : 'Ninguno con antecedentes';
}

function loadDemoReport(report, meta) {
    report.players = report.players || [];
    currentReport = report;
    currentMeta = meta || {};
    timelineShown = 25;

    showView('audit');
    renderDemoMeta();

    const h = report.header;
    const minutes = Math.max(1, Math.round(h.totalTicks / (h.tickrate * 60)));
    const rounds = (h.roundStarts || []).length;

    document.getElementById('cacheNotice').style.display = currentMeta.fromCache ? 'flex' : 'none';
    document.getElementById('sumDuration').innerText = minutes + ' min';
    document.getElementById('sumDurationSub').innerText = (rounds ? rounds + ' rondas · ' : '') + h.tickrate + ' tick';
    renderSummary();

    renderPlayers();
    renderDetections();
    renderTimeline();
    renderAdmin();
    renderChat();
    renderHistory();
    document.getElementById('main').scrollTop = 0;
}
