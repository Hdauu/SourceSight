// ── Drag & drop ───────────────────────────────────────────────────
// WebView2 hands the real file paths to Go when the File objects travel with
// postMessageWithAdditionalObjects; older runtimes only give us the names.
const FILE_DROP_MESSAGE = '__sourcesight_files_dropped__';
const dropzone = document.getElementById('dropzone');

function uiLog(msg) {
    if (window.logClient) window.logClient(msg);
}

window.addEventListener('dragover', e => {
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy';
    dropzone.classList.add('dragover');
});
window.addEventListener('dragleave', e => { if (!e.relatedTarget) dropzone.classList.remove('dragover'); });
window.addEventListener('drop', e => {
    e.preventDefault();
    dropzone.classList.remove('dragover');
    const files = e.dataTransfer && e.dataTransfer.files;
    const wv = window.chrome && window.chrome.webview;
    uiLog('drop: archivos=' + (files ? files.length : 'null') +
        ' nombres=' + (files ? Array.from(files).map(f => f.name).join('|') : '') +
        ' tipos=' + (e.dataTransfer ? Array.from(e.dataTransfer.types || []).join('|') : '') +
        ' additionalObjects=' + !!(wv && wv.postMessageWithAdditionalObjects));
    if (!files || !files.length) return;
    if (wv && wv.postMessageWithAdditionalObjects) {
        try {
            wv.postMessageWithAdditionalObjects(FILE_DROP_MESSAGE, files);
        } catch (err) {
            uiLog('drop: postMessageWithAdditionalObjects fallo: ' + err);
            dropByName(files);
        }
    } else {
        dropByName(files);
    }
});

function dropByName(files) {
    if (!window.onDemoFilesDroppedByName) return;
    showLoading();
    window.onDemoFilesDroppedByName(Array.from(files).map(f => f.name));
}

window.addEventListener('keydown', e => {
    if (e.key === 'Escape') { closePlayer(); closeSettings(); }
});
