// ── Report tabs ───────────────────────────────────────────────────────
// The report is split in tabs (Jugadores, Incidentes, Mapa, SHIFT, Chat y admins):
// each section of #auditContent says which tab it belongs to with data-tab. The last
// tab used is remembered on this PC.

let reportTab = 'players';
try { reportTab = localStorage.getItem('reportTab') || 'players'; } catch (e) { /* storage unavailable */ }

function setReportTab(tab) {
    reportTab = tab;
    try { localStorage.setItem('reportTab', tab); } catch (e) { /* storage unavailable */ }
    document.querySelectorAll('#reportTabs button').forEach(b => b.classList.toggle('on', b.dataset.tab === tab));
    document.querySelectorAll('#auditContent [data-tab]').forEach(el => {
        if (el.tagName !== 'BUTTON') el.style.display = el.dataset.tab === tab ? '' : 'none';
    });
    if (tab === 'map' && typeof drawMap === 'function') drawMap();
}

// Counts next to each tab, refreshed with every report
function updateTabCounts() {
    const r = currentReport || {};
    const set = (id, n) => { document.getElementById(id).innerText = n ? String(n) : ''; };
    set('tabCountPlayers', (r.players || []).length);
    set('tabCountIncidents', (r.timeline || []).length);
    set('tabCountShift', (r.shiftLog || []).length);
    set('tabCountChat', (r.chatLog || []).length + (r.adminLog || []).length);
    // A tab that would be empty isn't a good place to land: fall back to players
    const empty = { incidents: !(r.timeline || []).length };
    setReportTab(empty[reportTab] ? 'players' : reportTab);
}
