// Page errors go to the app log (%APPDATA%\SourceSight\sourcesight.log), so a part of
// the page that doesn't show can be diagnosed. The same error is logged once.
(function () {
    const logged = new Set();
    const report = msg => {
        if (logged.has(msg) || logged.size > 50) return;
        logged.add(msg);
        if (window.logClient) window.logClient('error de la página: ' + msg);
    };
    window.addEventListener('error', e => report((e.message || 'error') + ' (' + (e.filename || '').split('/').pop() + ':' + e.lineno + ':' + e.colno + ')'));
    window.addEventListener('unhandledrejection', e => {
        const r = e.reason;
        report('promesa: ' + (r && r.message ? r.message : String(r)));
    });
})();

let currentReport = null;
let currentMeta = null;
let historyItems = [];
let playerFilter = 'all';
let chatFilter = 'all';
let sortKey = 'level';
let sortAsc = false;
let timelineShown = 25;
let updateInfo = null;
let appInfo = { version: '', repoUrl: '' };
