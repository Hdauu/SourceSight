// ── Startup ───────────────────────────────────────────────────────
let readyFired = false;
function notifyReady() {
    if (readyFired) return;
    readyFired = true;
    if (window.getAppInfo) {
        window.getAppInfo().then(info => {
            appInfo = info;
            renderHistory();
            const v = info.version === 'dev' ? 'dev' : info.version;
            document.getElementById('appVersion').innerText = v;
            document.getElementById('aboutVersion').innerText = v;
            document.getElementById('aboutRepo').innerText = info.repoUrl.replace('https://', '');
        });
    }
    // Flag levels and staff confirmations must be known before anything renders
    if (window.getConfig) {
        window.getConfig().then(s => {
            try {
                const w = JSON.parse(s).sidebarWidth;
                if (w) setAsideWidth(w, false);
                updateArchiveDot(JSON.parse(s).archive || {});
            } catch (e) {}
        });
    }
    const levels = window.getFlagLevels ? window.getFlagLevels().then(m => { FLAG_LEVELS = m || {}; }) : Promise.resolve();
    const confirmed = new Promise(resolve => refreshConfirmed(resolve));
    Promise.all([levels, confirmed]).finally(() => {
        refreshHistory();
        if (!currentReport) goSection('home');
        if (window.onAppReady) window.onAppReady();
    });
    setTimeout(() => checkUpdates(false), 1500);
}
window.addEventListener('DOMContentLoaded', notifyReady);
setTimeout(notifyReady, 150);
