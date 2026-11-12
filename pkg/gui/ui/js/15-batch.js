// ── Batch: several demos analyzed in parallel ─────────────────────────
// Go sends batchStart once, batchUpdate every ~200 ms while it runs and batchDone
// at the end (pkg/gui/batch.go). Rows of finished demos open their report.

const BATCH_STATE = {
    queued: 'En cola',
    hashing: 'Verificando',
    parsing: 'Analizando',
    saving: 'Guardando',
    done: 'Lista',
    cached: 'Ya estaba',
    skipped: 'Repetida',
    error: 'Error',
    canceled: 'Cancelada'
};
const BATCH_ACTIVE = { hashing: true, parsing: true, saving: true };
const BATCH_FINISHED = { done: true, cached: true, skipped: true, error: true, canceled: true };

let batchRows = [];
let batchRunning = false;

function fmtSecs(s) {
    if (!s) return '';
    s = Math.round(s);
    return s < 60 ? s + ' s' : Math.floor(s / 60) + ' min ' + (s % 60) + ' s';
}

function batchStart(rows, info) {
    batchRunning = true;
    batchRows = rows;
    document.getElementById('batchCancel').style.display = 'inline-flex';
    document.getElementById('batchCancel').disabled = false;
    document.getElementById('batchClose').style.display = 'none';
    showView('batch');
    renderBatch(info);
}

function batchUpdate(rows, info) {
    batchRows = rows;
    renderBatch(info);
}

function batchDone(rows, info) {
    batchRunning = false;
    batchRows = rows;
    document.getElementById('batchCancel').style.display = 'none';
    document.getElementById('batchClose').style.display = 'inline-flex';
    renderBatch(info);
    refreshHistory();
}

function renderBatch(info) {
    const total = info.total || batchRows.length;
    if (info.done) {
        document.getElementById('batchTitle').innerText = info.canceled
            ? 'Análisis cancelado'
            : total + (total === 1 ? ' demo analizada' : ' demos analizadas');
        document.getElementById('batchSub').innerText = 'Tardó ' + fmtSecs(info.elapsed) + ' con ' + info.workers + ' en paralelo. Click en una demo para abrirla.';
    } else {
        document.getElementById('batchTitle').innerText = 'Analizando ' + total + ' demos';
        document.getElementById('batchSub').innerText = info.workers + ' en paralelo · ' + info.finished + ' de ' + total + ' listas' +
            (info.eta ? ' · faltan unos ' + fmtSecs(info.eta) : '');
    }

    // Overall progress: finished demos count whole, running ones by their percentage
    let progress = 0;
    batchRows.forEach(r => { progress += BATCH_FINISHED[r.state] ? 1 : r.pct / 100; });
    document.getElementById('batchBar').style.width = Math.round(100 * progress / Math.max(1, total)) + '%';

    // Result counts
    const count = pred => batchRows.filter(pred).length;
    const counts = [
        ['tag-red', count(r => r.level === 'cheat'), ' con cheat', ' con cheat'],
        ['tag-yellow', count(r => r.level === 'sospechoso'), ' sospechosa', ' sospechosas'],
        ['tag-plain', count(r => (r.state === 'done' || r.state === 'cached') && r.level !== 'cheat' && r.level !== 'sospechoso'), ' limpia', ' limpias'],
        ['tag-plain', count(r => r.state === 'skipped'), ' repetida', ' repetidas'],
        ['tag-red', count(r => r.state === 'error'), ' con error', ' con error']
    ].filter(c => c[1] > 0);
    document.getElementById('batchCounts').innerHTML = counts.map(c => '<span class="tag ' + c[0] + '">' + c[1] + (c[1] === 1 ? c[2] : c[3]) + '</span>').join('');

    // Running demos first, then queued, then finished (most recent work on top)
    const order = r => BATCH_ACTIVE[r.state] ? 0 : r.state === 'queued' ? 1 : 2;
    const rows = batchRows.slice().sort((a, b) => order(a) - order(b) || a.id - b.id);
    document.getElementById('batchBody').innerHTML = rows.map(r => {
        const finished = BATCH_FINISHED[r.state];
        const bar = r.state === 'error' ? ' fail' : finished ? ' done' : '';
        const result = r.state === 'error'
            ? '<span class="t-red" title="' + esc(r.error) + '">' + esc(r.error) + '</span>'
            : r.level ? (levelTag(r.level) + ' ' + (r.flags || []).map(flagTag).join(' ')) : '<span class="muted">–</span>';
        const clickable = r.historyId ? ' class="click" onclick="openBatchRow(\'' + r.historyId + '\')"' : '';
        return '<tr' + clickable + '>' +
            '<td class="player">' + esc(r.name) + '</td>' +
            '<td><span class="batch-state' + (BATCH_ACTIVE[r.state] ? ' active' : '') + '">' + BATCH_STATE[r.state] + '</span></td>' +
            '<td><div class="mini-progress' + bar + '"><div style="width:' + (finished ? 100 : r.pct) + '%"></div></div></td>' +
            '<td class="wrap">' + result + '</td>' +
            '<td class="r muted">' + fmtSecs(r.seconds) + '</td>' +
        '</tr>';
    }).join('');
}

function cancelBatch() {
    document.getElementById('batchCancel').disabled = true;
    if (window.cancelAnalysis) window.cancelAnalysis();
}

function openBatchRow(historyId) {
    openHistoryItem(historyId);
}

function closeBatch() {
    restoreIdle();
}
