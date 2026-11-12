// ── Resizable sidebar ─────────────────────────────────────────────
// Width is saved in the app config (the page has no web origin, so browser
// storage isn't reliable here).
const ASIDE_DEFAULT = 288, ASIDE_MIN = 220, ASIDE_MAX = 560;
const aside = document.getElementById('aside');
const resizer = document.getElementById('asideResizer');

function setAsideWidth(px, save) {
    const w = Math.round(Math.min(ASIDE_MAX, Math.max(ASIDE_MIN, px)));
    aside.style.width = w + 'px';
    if (save && window.saveSidebarWidth) window.saveSidebarWidth(w === ASIDE_DEFAULT ? 0 : w);
}

resizer.addEventListener('mousedown', e => {
    e.preventDefault();
    const startX = e.clientX, startW = aside.getBoundingClientRect().width;
    document.body.classList.add('resizing');
    const move = ev => setAsideWidth(startW + ev.clientX - startX, false);
    const up = () => {
        document.body.classList.remove('resizing');
        window.removeEventListener('mousemove', move);
        window.removeEventListener('mouseup', up);
        setAsideWidth(aside.getBoundingClientRect().width, true);
    };
    window.addEventListener('mousemove', move);
    window.addEventListener('mouseup', up);
});
resizer.addEventListener('dblclick', () => setAsideWidth(ASIDE_DEFAULT, true));
resizer.addEventListener('keydown', e => {
    if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return;
    e.preventDefault();
    setAsideWidth(aside.getBoundingClientRect().width + (e.key === 'ArrowRight' ? 16 : -16), true);
});
