package gui

const AppHTML = `<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SourceSight</title>
    <style>
        :root {
            --bg: #121314;
            --panel: #18191b;
            --panel-2: #1e1f22;
            --hover: #232528;
            --border: #2a2c30;
            --border-strong: #383b40;
            --text: #e7e5e1;
            --text-2: #b4b2ad;
            --muted: #85878c;
            --accent: #e07b2a;
            --accent-hover: #c96a1d;
            --accent-soft: rgba(224, 123, 42, 0.12);
            --red: #e5534b;
            --red-soft: rgba(229, 83, 75, 0.12);
            --yellow: #d4b23f;
            --yellow-soft: rgba(212, 178, 63, 0.12);
            --green: #57ab5a;
            --green-soft: rgba(87, 171, 90, 0.12);
            --chart: #d07024;
            --mono: "Cascadia Mono", Consolas, monospace;
            --r: 4px;
        }

        * { box-sizing: border-box; margin: 0; padding: 0; user-select: none; -webkit-user-drag: none; }

        body {
            background: var(--bg);
            color: var(--text);
            font: 13px/1.45 "Segoe UI", system-ui, sans-serif;
            -webkit-font-smoothing: antialiased;
            height: 100vh;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        button { font: inherit; color: inherit; }
        code, .mono { font-family: var(--mono); font-size: 11.5px; }
        .num { font-variant-numeric: tabular-nums; }
        .muted { color: var(--muted); }
        .t-red { color: var(--red) !important; }
        .t-yellow { color: var(--yellow) !important; }
        .t-green { color: var(--green) !important; }
        .strong { font-weight: 600; color: var(--text); }

        .ic { width: 15px; height: 15px; flex-shrink: 0; fill: none; stroke: currentColor; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
        .ic-sm { width: 13px; height: 13px; }

        ::-webkit-scrollbar { width: 10px; height: 10px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: #2c2e32; border: 3px solid var(--bg); border-radius: 6px; }
        ::-webkit-scrollbar-thumb:hover { background: #3a3d42; }

        /* ── Top bar ─────────────────────────────────────────── */
        header {
            height: 52px;
            flex-shrink: 0;
            display: flex;
            align-items: center;
            gap: 16px;
            padding: 0 16px 0 14px;
            background: var(--panel);
            border-bottom: 1px solid var(--border);
        }

        .brand { display: flex; align-items: center; gap: 9px; flex-shrink: 0; }
        .brand svg { width: 24px; height: 24px; }
        .brand-name { font-size: 14px; font-weight: 600; letter-spacing: -0.1px; }
        .brand-ver { font-size: 11px; color: var(--muted); }

        .demo-meta {
            flex: 1;
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 10px;
            padding-left: 16px;
            border-left: 1px solid var(--border);
            color: var(--muted);
            white-space: nowrap;
            overflow: hidden;
        }
        .demo-meta .map { color: var(--text); font-weight: 600; }
        .demo-meta span { overflow: hidden; text-overflow: ellipsis; }

        .actions { display: flex; align-items: center; gap: 6px; flex-shrink: 0; }

        .btn {
            height: 30px;
            padding: 0 11px;
            display: inline-flex;
            align-items: center;
            gap: 7px;
            background: var(--panel-2);
            border: 1px solid var(--border);
            border-radius: var(--r);
            color: var(--text-2);
            font-size: 12.5px;
            font-weight: 500;
            cursor: pointer;
            white-space: nowrap;
        }
        .btn:hover { background: var(--hover); border-color: var(--border-strong); color: var(--text); }
        .btn-primary { background: var(--accent); border-color: var(--accent); color: #1a1208; font-weight: 600; }
        .btn-primary:hover { background: var(--accent-hover); border-color: var(--accent-hover); color: #1a1208; }
        .btn-plain { background: transparent; border-color: transparent; }
        .btn-plain:hover { background: var(--hover); border-color: transparent; }
        .btn-icon { width: 30px; padding: 0; justify-content: center; }
        .btn-danger { color: #f0a29d; }
        .btn-danger:hover { color: #fff; border-color: var(--red); background: var(--red-soft); }
        .btn:focus-visible, .seg button:focus-visible, .hist-item:focus-visible, input:focus-visible, select:focus-visible {
            outline: 2px solid var(--accent);
            outline-offset: 1px;
        }

        /* ── Update banner ───────────────────────────────────── */
        .update-bar {
            display: none;
            align-items: center;
            gap: 12px;
            padding: 7px 16px;
            background: var(--accent-soft);
            border-bottom: 1px solid rgba(224, 123, 42, 0.3);
            color: var(--text);
            font-size: 12.5px;
        }
        .update-bar .grow { flex: 1; }

        /* ── Layout ──────────────────────────────────────────── */
        .layout { flex: 1; display: flex; min-height: 0; }

        aside {
            position: relative;
            width: 288px;
            flex-shrink: 0;
            display: flex;
            flex-direction: column;
            background: var(--panel);
            border-right: 1px solid var(--border);
            min-height: 0;
        }
        .aside-head {
            padding: 14px 14px 10px;
            display: flex;
            align-items: baseline;
            justify-content: space-between;
        }
        .aside-title { font-size: 12.5px; font-weight: 600; }
        .hist-list { flex: 1; overflow-y: auto; padding: 0 6px 10px; }
        .hist-empty { padding: 18px 10px; color: var(--muted); font-size: 12px; }

        .hist-item {
            position: relative;
            display: block;
            width: 100%;
            text-align: left;
            background: transparent;
            border: 0;
            border-left: 2px solid transparent;
            border-radius: 0 var(--r) var(--r) 0;
            padding: 7px 28px 7px 9px;
            cursor: pointer;
        }
        .hist-item:hover { background: var(--hover); }
        .hist-item.active { background: var(--panel-2); border-left-color: var(--accent); }
        .hist-item.flagged:not(.active) { border-left-color: rgba(229, 83, 75, 0.55); }
        .hist-item.suspect:not(.active) { border-left-color: rgba(212, 178, 63, 0.5); }
        .hist-tags { display: flex; gap: 3px; flex-wrap: wrap; margin-top: 3px; }
        .hist-tags .tag { font-size: 10.5px; padding: 0 5px; }

        .aside-tabs { display: flex; border-bottom: 1px solid var(--border); }
        .aside-tabs button {
            flex: 1;
            height: 40px;
            background: transparent;
            border: 0;
            color: var(--muted);
            font-size: 12.5px;
            font-weight: 600;
            cursor: pointer;
        }
        .aside-tabs button:hover { color: var(--text); }
        .aside-tabs button.on { color: var(--text); box-shadow: inset 0 -2px 0 var(--accent); }
        .aside-tabs .num { color: var(--muted); font-weight: 400; margin-left: 3px; }
        .aside-pane { flex: 1; display: flex; flex-direction: column; min-height: 0; }
        .aside-search { padding: 10px 10px 8px; display: flex; flex-direction: column; gap: 6px; }
        .seg-full { display: flex; }
        .seg-full button { flex: 1 1 auto; min-width: 0; padding: 0 6px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

        /* Drag handle on the sidebar edge */
        .aside-resizer {
            position: absolute;
            top: 0;
            right: -4px;
            width: 8px;
            height: 100%;
            cursor: col-resize;
            z-index: 20;
        }
        .aside-resizer::after {
            content: "";
            position: absolute;
            top: 0;
            bottom: 0;
            left: 3px;
            width: 2px;
            background: transparent;
            transition: background 0.12s;
        }
        .aside-resizer:hover::after, .aside-resizer:focus-visible::after, body.resizing .aside-resizer::after { background: var(--accent); }
        .aside-resizer:focus-visible { outline: none; }
        body.resizing { cursor: col-resize; user-select: none; }
        .aside-hint { font-size: 11.5px; color: var(--muted); }
        .aside-hint:empty { display: none; }

        /* Player page & overview */
        .page-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 14px; }
        .page-title { font-size: 20px; font-weight: 600; }
        .page-sub { color: var(--muted); font-size: 12.5px; margin-top: 2px; }
        .compare td.r { width: 140px; }
        .delta-up { color: var(--red); }
        .delta-down { color: var(--muted); }

        /* Staff confirmation */
        .tag-confirmed { color: #fff; background: #b8322b; }
        .confirm-box {
            display: flex;
            flex-direction: column;
            gap: 8px;
            padding: 10px 12px;
            border: 1px solid var(--border);
            border-radius: var(--r);
            font-size: 12.5px;
        }
        .confirm-box.confirmed {
            flex-direction: row;
            align-items: center;
            justify-content: space-between;
            border-color: rgba(229, 83, 75, 0.45);
            background: var(--red-soft);
            color: #f3c1bd;
        }
        .confirm-form { display: flex; gap: 6px; }

        /* Prompt */
        .prompt-text { color: var(--text-2); font-size: 13px; line-height: 1.5; }
        .prompt-actions { display: flex; gap: 6px; justify-content: flex-end; flex-wrap: wrap; }
        .hist-name { font-size: 12.5px; color: var(--text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
        .hist-sub { font-size: 11.5px; color: var(--muted); display: flex; gap: 6px; white-space: nowrap; overflow: hidden; }
        .hist-flag { color: var(--red); font-weight: 600; }
        .hist-remove {
            position: absolute;
            right: 4px;
            top: 50%;
            transform: translateY(-50%);
            width: 22px;
            height: 22px;
            display: none;
            align-items: center;
            justify-content: center;
            background: transparent;
            border: 0;
            border-radius: var(--r);
            color: var(--muted);
            cursor: pointer;
        }
        .hist-item:hover .hist-remove { display: flex; }
        .hist-remove:hover { background: var(--border); color: var(--text); }

        main { flex: 1; overflow-y: auto; min-width: 0; }
        .content { max-width: 1320px; margin: 0 auto; padding: 20px 24px 40px; display: flex; flex-direction: column; gap: 16px; }

        /* ── Empty / loading ─────────────────────────────────── */
        .dropzone {
            margin-top: 8vh;
            border: 1px dashed var(--border-strong);
            border-radius: 6px;
            padding: 64px 24px;
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 8px;
            text-align: center;
            cursor: pointer;
        }
        .dropzone:hover { border-color: var(--muted); }
        .dropzone.dragover { border-color: var(--accent); background: var(--accent-soft); }
        .dropzone .ic { width: 28px; height: 28px; color: var(--muted); margin-bottom: 8px; stroke-width: 1.5; }
        .dropzone-title { font-size: 16px; font-weight: 600; }
        .dropzone-sub { color: var(--muted); max-width: 420px; }

        .loading { display: none; flex-direction: column; align-items: center; gap: 12px; padding: 110px 20px; }
        .loading-text { color: var(--text-2); font-variant-numeric: tabular-nums; }
        .progress { width: 340px; max-width: 80%; height: 4px; background: var(--border); border-radius: 2px; overflow: hidden; }
        .progress-bar { width: 0; height: 100%; background: var(--accent); transition: width 0.15s linear; }

        /* ── Summary ─────────────────────────────────────────── */
        .notice {
            display: none;
            align-items: center;
            gap: 10px;
            padding: 8px 12px;
            border: 1px solid var(--border);
            border-radius: var(--r);
            background: var(--panel);
            color: var(--text-2);
            font-size: 12.5px;
        }
        .notice .grow { flex: 1; }

        .summary {
            display: grid;
            grid-template-columns: repeat(4, minmax(0, 1fr));
            background: var(--panel);
            border: 1px solid var(--border);
            border-radius: var(--r);
        }
        .sum-cell { padding: 12px 16px; min-width: 0; }
        .sum-cell + .sum-cell { border-left: 1px solid var(--border); }
        .sum-label { font-size: 12px; color: var(--muted); }
        .sum-value { font-size: 22px; font-weight: 600; font-variant-numeric: tabular-nums; margin-top: 1px; }
        .sum-sub { font-size: 12px; color: var(--muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

        /* ── Panels & tables ─────────────────────────────────── */
        .panel { background: var(--panel); border: 1px solid var(--border); border-radius: var(--r); overflow: hidden; }
        .panel-head {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
            padding: 10px 14px;
            border-bottom: 1px solid var(--border);
            min-height: 46px;
        }
        .panel-title { font-size: 13.5px; font-weight: 600; display: flex; align-items: baseline; gap: 8px; white-space: nowrap; }
        .panel-title .count { font-size: 12px; font-weight: 400; color: var(--muted); }
        .panel-tools { display: flex; align-items: center; gap: 8px; }

        .input {
            height: 30px;
            background: var(--bg);
            border: 1px solid var(--border);
            border-radius: var(--r);
            color: var(--text);
            padding: 0 10px;
            font: inherit;
            font-size: 12.5px;
            outline: none;
            width: 220px;
            user-select: text;
        }
        .input::placeholder { color: var(--muted); }
        .input:focus { border-color: var(--border-strong); }
        aside .input { width: 100%; }

        .seg { display: inline-flex; border: 1px solid var(--border); border-radius: var(--r); overflow: hidden; }
        .seg button {
            height: 28px;
            padding: 0 10px;
            background: transparent;
            border: 0;
            color: var(--muted);
            font-size: 12px;
            cursor: pointer;
        }
        .seg button + button { border-left: 1px solid var(--border); }
        .seg button:hover { color: var(--text); background: var(--hover); }
        .seg button.on { color: var(--text); background: var(--panel-2); box-shadow: inset 0 -2px 0 var(--accent); }

        .table-wrap { overflow-x: auto; }
        table { width: 100%; border-collapse: collapse; }
        th {
            text-align: left;
            font-size: 12px;
            font-weight: 500;
            color: var(--muted);
            padding: 8px 14px;
            border-bottom: 1px solid var(--border);
            white-space: nowrap;
            background: var(--panel);
        }
        th.r, td.r { text-align: right; }
        th.sortable { cursor: pointer; }
        th.sortable:hover { color: var(--text-2); }
        th.sorted { color: var(--text); }
        th.sorted::after {
            content: "";
            display: inline-block;
            margin-left: 5px;
            vertical-align: middle;
            border-left: 3.5px solid transparent;
            border-right: 3.5px solid transparent;
            border-top: 4.5px solid currentColor;
        }
        th.sorted.asc::after { transform: rotate(180deg); }

        td {
            padding: 8px 14px;
            border-bottom: 1px solid var(--border);
            color: var(--text-2);
            white-space: nowrap;
            font-variant-numeric: tabular-nums;
        }
        tbody tr:last-child td { border-bottom: 0; }
        tbody tr:hover td { background: var(--panel-2); }
        tr.click { cursor: pointer; }
        tr.mark-red td:first-child { box-shadow: inset 2px 0 0 var(--red); }
        tr.mark-yellow td:first-child { box-shadow: inset 2px 0 0 var(--yellow); }
        td.wrap { white-space: normal; }
        td.player { color: var(--text); font-weight: 500; }

        .empty td { text-align: center; color: var(--muted); padding: 22px 14px; white-space: normal; }

        .tag {
            display: inline-block;
            padding: 1px 6px;
            border-radius: 3px;
            font-size: 11.5px;
            font-weight: 600;
            white-space: nowrap;
        }
        .tag-red { color: var(--red); background: var(--red-soft); }
        .tag-yellow { color: var(--yellow); background: var(--yellow-soft); }
        .tag-green { color: var(--green); background: var(--green-soft); }
        .tag-plain { color: var(--text-2); background: var(--panel-2); font-weight: 500; }
        .tags { display: inline-flex; gap: 4px; flex-wrap: wrap; }

        .cmd {
            height: 24px;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 0 7px;
            background: transparent;
            border: 1px solid var(--border);
            border-radius: var(--r);
            color: var(--text-2);
            font-family: var(--mono);
            font-size: 11.5px;
            cursor: pointer;
        }
        .cmd .ic { color: var(--muted); }
        .cmd:hover { border-color: var(--border-strong); color: var(--text); }

        .play {
            height: 24px;
            display: inline-flex;
            align-items: center;
            gap: 5px;
            padding: 0 8px;
            background: transparent;
            border: 1px solid var(--border);
            border-radius: var(--r);
            color: var(--accent);
            font-size: 12px;
            font-weight: 500;
            cursor: pointer;
        }
        .play .ic { width: 10px; height: 10px; fill: currentColor; stroke: none; }
        .play:hover { border-color: var(--accent); background: var(--accent-soft); }

        .more { padding: 8px; text-align: center; border-top: 1px solid var(--border); }

        /* ── Modals ──────────────────────────────────────────── */
        .overlay {
            position: fixed;
            inset: 0;
            display: none;
            align-items: center;
            justify-content: center;
            padding: 24px;
            background: rgba(0, 0, 0, 0.6);
            z-index: 100;
        }
        .modal {
            width: 820px;
            max-width: 100%;
            max-height: 90vh;
            display: flex;
            flex-direction: column;
            background: var(--panel);
            border: 1px solid var(--border-strong);
            border-radius: 6px;
            box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
            overflow: hidden;
        }
        .modal-head {
            display: flex;
            align-items: flex-start;
            justify-content: space-between;
            gap: 14px;
            padding: 16px 18px 14px;
            border-bottom: 1px solid var(--border);
        }
        .modal-title { font-size: 17px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
        .modal-sub { color: var(--muted); font-size: 12px; margin-top: 2px; }
        .modal-actions { display: flex; gap: 6px; flex-wrap: wrap; justify-content: flex-end; }
        .modal-body { padding: 16px 18px 20px; overflow-y: auto; display: flex; flex-direction: column; gap: 20px; }

        .block-title { font-size: 12.5px; font-weight: 600; color: var(--text); margin-bottom: 8px; display: flex; align-items: baseline; gap: 8px; }
        .block-title .count { font-weight: 400; color: var(--muted); }

        .stats { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); border: 1px solid var(--border); border-radius: var(--r); }
        .stat { padding: 9px 12px; border-right: 1px solid var(--border); border-bottom: 1px solid var(--border); min-width: 0; }
        .stat:nth-child(5n) { border-right: 0; }
        .stat:nth-last-child(-n+5) { border-bottom: 0; }
        .stat-label { font-size: 11.5px; color: var(--muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
        .stat-value { font-size: 15px; font-weight: 600; font-variant-numeric: tabular-nums; }

        .callout {
            display: flex;
            gap: 10px;
            align-items: flex-start;
            padding: 9px 12px;
            border-radius: var(--r);
            border: 1px solid var(--border);
            font-size: 12.5px;
            color: var(--text-2);
        }
        .callout-red { border-color: rgba(229, 83, 75, 0.4); background: var(--red-soft); color: #f0b2ad; }

        .kv { display: grid; grid-template-columns: 150px 1fr; gap: 4px 12px; font-size: 12.5px; }
        .kv dt { color: var(--muted); }
        .kv dd { color: var(--text-2); }

        .list { display: flex; flex-direction: column; border: 1px solid var(--border); border-radius: var(--r); max-height: 240px; overflow-y: auto; }
        .list-row { display: flex; align-items: center; gap: 10px; padding: 7px 10px; border-bottom: 1px solid var(--border); font-size: 12.5px; }
        .list-row:last-child { border-bottom: 0; }
        .list-row .grow { flex: 1; min-width: 0; color: var(--text-2); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

        /* Chart */
        .chart { position: relative; border: 1px solid var(--border); border-radius: var(--r); padding: 10px 12px 6px; }
        .chart svg { display: block; width: 100%; height: auto; overflow: visible; }
        .chart .axis { font-size: 10.5px; fill: var(--muted); font-family: "Segoe UI", system-ui, sans-serif; }
        .chart .grid { stroke: var(--border); stroke-width: 1; }
        .chart .bar { fill: var(--chart); }
        .chart .hit { fill: transparent; cursor: default; }
        .chart .hit:hover + .bar, .chart .bar.hl { fill: #e8893a; }
        .tooltip {
            position: absolute;
            pointer-events: none;
            display: none;
            min-width: 150px;
            padding: 7px 9px;
            background: var(--bg);
            border: 1px solid var(--border-strong);
            border-radius: var(--r);
            font-size: 12px;
            color: var(--text-2);
            z-index: 5;
        }
        .tooltip b { color: var(--text); font-weight: 600; }
        .tooltip .row { display: flex; justify-content: space-between; gap: 14px; }

        .field { display: flex; flex-direction: column; gap: 6px; }
        .field label { font-size: 12.5px; font-weight: 600; }
        .field .help { font-size: 12px; color: var(--muted); }
        .field .input, .field select { width: 100%; }
        select.input { cursor: pointer; }
        .divider { height: 1px; background: var(--border); }
        a.link { color: var(--accent); cursor: pointer; text-decoration: none; }
        a.link:hover { text-decoration: underline; }

        .toast {
            position: fixed;
            right: 20px;
            bottom: 20px;
            max-width: 520px;
            padding: 9px 13px;
            background: var(--panel-2);
            border: 1px solid var(--border-strong);
            border-radius: var(--r);
            color: var(--text);
            font-size: 12.5px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
            opacity: 0;
            transform: translateY(8px);
            transition: opacity 0.18s, transform 0.18s;
            pointer-events: none;
            z-index: 200;
        }
        .toast.show { opacity: 1; transform: none; }

        @media (max-width: 1100px) {
            .summary { grid-template-columns: repeat(2, minmax(0, 1fr)); }
            .sum-cell:nth-child(3) { border-left: 0; }
            .sum-cell:nth-child(n+3) { border-top: 1px solid var(--border); }
            .stats { grid-template-columns: repeat(3, minmax(0, 1fr)); }
            .stat:nth-child(5n) { border-right: 1px solid var(--border); }
            .stat:nth-child(3n) { border-right: 0; }
            .btn-label { display: none; }
        }
        @media (max-width: 760px) {
            aside { display: none; }
            .demo-meta { display: none; }
        }
    </style>
</head>
<body>

    <header>
        <div class="brand">
            <svg viewBox="0 0 24 24" aria-hidden="true">
                <rect x="1" y="1" width="22" height="22" rx="5" fill="#1b1c1f" stroke="#2e3035"/>
                <g fill="none" stroke="#ecebe6" stroke-width="1.7" stroke-linecap="round">
                    <path d="M12 4.6v3.2M12 16.2v3.2M4.6 12h3.2M16.2 12h3.2"/>
                    <path d="M14.9 6.9a6 6 0 0 1 2.2 2.2M17.1 14.9a6 6 0 0 1-2.2 2.2M9.1 17.1a6 6 0 0 1-2.2-2.2M6.9 9.1a6 6 0 0 1 2.2-2.2"/>
                </g>
                <circle cx="12" cy="12" r="1.8" fill="#e07b2a"/>
            </svg>
            <span class="brand-name">SourceSight</span>
            <span class="brand-ver" id="appVersion"></span>
        </div>

        <div class="demo-meta" id="demoMeta"></div>

        <div class="actions">
            <button class="btn" id="btnReanalyze" style="display:none;" onclick="reanalyze()" title="Volver a analizar esta demo">
                <i data-icon="refresh"></i><span class="btn-label">Re-analizar</span>
            </button>
            <button class="btn" id="btnPNG" style="display:none;" onclick="window.exportPNGReport()" title="Guardar el reporte como imagen">
                <i data-icon="image"></i><span class="btn-label">PNG</span>
            </button>
            <button class="btn" id="btnSummary" style="display:none;" onclick="copyStaffSummary()" title="Copiar resumen en Markdown para Discord">
                <i data-icon="copy"></i><span class="btn-label">Resumen</span>
            </button>
            <button class="btn btn-primary" onclick="pickDemos()">
                <i data-icon="folder"></i><span class="btn-label">Abrir demos</span>
            </button>
            <button class="btn btn-plain btn-icon" onclick="openSettings()" title="Configuración">
                <i data-icon="settings"></i>
            </button>
        </div>
    </header>

    <div class="update-bar" id="updateBar">
        <span class="grow" id="updateText"></span>
        <a class="link" id="updateNotes" onclick="openUpdatePage()">Ver cambios</a>
        <button class="btn btn-primary" id="updateApply" onclick="startUpdate()">Actualizar</button>
        <button class="btn btn-plain btn-icon" onclick="dismissUpdate()" title="Ahora no"><i data-icon="x"></i></button>
    </div>

    <div class="layout">
        <aside id="aside">
            <div class="aside-resizer" id="asideResizer" role="separator" aria-orientation="vertical" aria-label="Ancho del panel" tabindex="0" title="Arrastrá para cambiar el ancho (doble click: volver al normal)"></div>
            <div class="aside-tabs" role="tablist">
                <button class="on" id="tabDemos" role="tab" onclick="setAsideTab('demos')">Demos <span class="num" id="histCount"></span></button>
                <button id="tabPlayers" role="tab" onclick="setAsideTab('players')">Jugadores</button>
            </div>

            <div class="aside-pane" id="paneDemos">
                <div class="aside-search">
                    <input class="input" id="histSearch" placeholder="Buscar demo, mapa o jugador" oninput="renderHistory()">
                    <div class="seg seg-full" id="histFilter">
                        <button class="on" onclick="setHistFilter('all', this)">Todas</button>
                        <button onclick="setHistFilter('cheat', this)">Cheat</button>
                        <button onclick="setHistFilter('suspect', this)">Sospechosas</button>
                        <button onclick="setHistFilter('clean', this)">Limpias</button>
                    </div>
                    <select class="input" id="histCategory" onchange="renderHistory()" aria-label="Categoría">
                        <option value="">Todas las categorías</option>
                    </select>
                </div>
                <div class="hist-list" id="histList"></div>
            </div>

            <div class="aside-pane" id="panePlayers" style="display:none;">
                <div class="aside-search">
                    <input class="input" id="playerSearch" placeholder="Nombre, SteamID o link de perfil" oninput="searchPlayers()">
                    <div class="aside-hint" id="playerHint"></div>
                </div>
                <div class="hist-list" id="playerList"></div>
            </div>
        </aside>

        <main id="main">
            <div class="content">

                <div class="dropzone" id="dropzone" onclick="pickDemos()">
                    <i data-icon="upload"></i>
                    <div class="dropzone-title">Soltá una demo acá</div>
                    <div class="dropzone-sub">También podés soltar varias o una carpeta entera, o hacer click para buscarlas. Todo se analiza en tu PC.</div>
                </div>

                <div class="loading" id="loadingState">
                    <div class="loading-text" id="loadingText">Abriendo demo...</div>
                    <div class="progress"><div class="progress-bar" id="loadingBar"></div></div>
                </div>

                <div id="overviewView" style="display:none; flex-direction:column; gap:16px;"></div>
                <div id="playerView" style="display:none; flex-direction:column; gap:16px;"></div>

                <div id="auditContent" style="display:none; flex-direction:column; gap:16px;">

                    <div class="notice" id="cacheNotice">
                        <i data-icon="history"></i>
                        <span class="grow" id="cacheNoticeText">Esta demo ya estaba analizada; se abrió el resultado guardado.</span>
                        <button class="btn" onclick="reanalyze()">Re-analizar</button>
                    </div>

                    <div class="summary">
                        <div class="sum-cell">
                            <div class="sum-label">Jugadores</div>
                            <div class="sum-value" id="sumPlayers">0</div>
                            <div class="sum-sub" id="sumPlayersSub"></div>
                        </div>
                        <div class="sum-cell">
                            <div class="sum-label">Cheats</div>
                            <div class="sum-value" id="sumFlags">0</div>
                            <div class="sum-sub" id="sumFlagsSub"></div>
                        </div>
                        <div class="sum-cell">
                            <div class="sum-label">Duración</div>
                            <div class="sum-value" id="sumDuration">0</div>
                            <div class="sum-sub" id="sumDurationSub"></div>
                        </div>
                        <div class="sum-cell">
                            <div class="sum-label">Reincidentes</div>
                            <div class="sum-value" id="sumReo">0</div>
                            <div class="sum-sub" id="sumReoSub"></div>
                        </div>
                    </div>

                    <section class="panel" id="playersCard">
                        <div class="panel-head">
                            <div class="panel-title">Jugadores <span class="count" id="playersCount"></span></div>
                            <div class="panel-tools">
                                <input class="input" id="searchPlayer" placeholder="Buscar jugador" oninput="renderPlayers()">
                                <div class="seg" id="playerFilter">
                                    <button class="on" onclick="setPlayerFilter('all', this)">Todos</button>
                                    <button onclick="setPlayerFilter('cheat', this)">Cheat</button>
                                    <button onclick="setPlayerFilter('suspect', this)">Sospechosos</button>
                                    <button onclick="setPlayerFilter('clean', this)">Limpios</button>
                                </div>
                            </div>
                        </div>
                        <div class="table-wrap">
                            <table>
                                <thead>
                                    <tr>
                                        <th class="sortable" data-key="name" onclick="sortPlayers('name')">Jugador</th>
                                        <th class="sortable r" data-key="totalJumps" onclick="sortPlayers('totalJumps')">Saltos</th>
                                        <th class="sortable r" data-key="perfRatio" onclick="sortPlayers('perfRatio')">Perfs</th>
                                        <th class="sortable r" data-key="oneTickRatio" onclick="sortPlayers('oneTickRatio')">1 tick</th>
                                        <th class="sortable r" data-key="maxStreak" onclick="sortPlayers('maxStreak')">Racha</th>
                                        <th class="sortable r" data-key="snapCount" onclick="sortPlayers('snapCount')">Snaps</th>
                                        <th class="sortable r" data-key="maxSpeed" onclick="sortPlayers('maxSpeed')">Vel. máx</th>
                                        <th class="r">Resultado</th>
                                    </tr>
                                </thead>
                                <tbody id="playersBody"></tbody>
                            </table>
                        </div>
                    </section>

                    <section class="panel" id="detectCard">
                        <div class="panel-head">
                            <div class="panel-title">Detecciones</div>
                            <span class="muted">Triggerbot, aim, prefire, macro, flash y humo, wallhack</span>
                        </div>
                        <div class="table-wrap">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Jugador</th>
                                        <th class="r">Reacción</th>
                                        <th class="r">Snaps / silent</th>
                                        <th class="r">Macro</th>
                                        <th class="r">Cegado / humo</th>
                                        <th class="r">Wallhack</th>
                                        <th class="r" title="Peleas donde hizo daño a menos de 100 ms de que el rival quedara a la vista">Prefire</th>
                                        <th class="r">Resultado</th>
                                    </tr>
                                </thead>
                                <tbody id="detectBody"></tbody>
                            </table>
                        </div>
                    </section>

                    <section class="panel" id="timelineCard">
                        <div class="panel-head">
                            <div class="panel-title">Incidentes <span class="count" id="timelineCount"></span></div>
                            <span class="muted">El comando ya incluye unos segundos antes de la jugada</span>
                        </div>
                        <div class="table-wrap">
                            <table>
                                <thead>
                                    <tr>
                                        <th class="r">Ronda</th>
                                        <th>Tipo</th>
                                        <th>Jugador</th>
                                        <th>Comando</th>
                                        <th></th>
                                        <th>Detalle</th>
                                    </tr>
                                </thead>
                                <tbody id="timelineBody"></tbody>
                            </table>
                        </div>
                        <div class="more" id="timelineMore" style="display:none;">
                            <button class="btn btn-plain" onclick="showMoreTimeline()">Mostrar 25 más</button>
                        </div>
                    </section>

                    <section class="panel" id="adminCard">
                        <div class="panel-head">
                            <div class="panel-title">Comandos de admins <span class="count" id="adminCount"></span></div>
                        </div>
                        <div class="table-wrap">
                            <table>
                                <thead>
                                    <tr>
                                        <th class="r">Ronda</th>
                                        <th>Admin</th>
                                        <th>Comando</th>
                                        <th>Objetivo</th>
                                        <th>Contexto</th>
                                        <th></th>
                                    </tr>
                                </thead>
                                <tbody id="adminBody"></tbody>
                            </table>
                        </div>
                    </section>

                    <section class="panel" id="chatCard">
                        <div class="panel-head">
                            <div class="panel-title">Chat <span class="count" id="chatCount"></span></div>
                            <div class="seg" id="chatFilter">
                                <button class="on" onclick="setChatFilter('all', this)">Todo</button>
                                <button onclick="setChatFilter('toxic', this)">Tóxico</button>
                                <button onclick="setChatFilter('ghosting', this)">Ghosting</button>
                                <button onclick="setChatFilter('spam', this)">Spam</button>
                            </div>
                        </div>
                        <div class="table-wrap">
                            <table>
                                <thead>
                                    <tr>
                                        <th class="r">Ronda</th>
                                        <th>Jugador</th>
                                        <th>Canal</th>
                                        <th>Mensaje</th>
                                        <th>Tipo</th>
                                        <th>Tick</th>
                                        <th></th>
                                    </tr>
                                </thead>
                                <tbody id="chatBody"></tbody>
                            </table>
                        </div>
                    </section>
                </div>
            </div>
        </main>
    </div>

    <!-- Player detail -->
    <div class="overlay" id="playerModal" onclick="closePlayer()">
        <div class="modal" onclick="event.stopPropagation()">
            <div class="modal-head">
                <div style="min-width:0;">
                    <div class="modal-title" id="pmName"></div>
                    <div class="modal-sub mono" id="pmSub"></div>
                </div>
                <div class="modal-actions" id="pmActions"></div>
            </div>
            <div class="modal-body" id="pmBody"></div>
        </div>
    </div>

    <!-- Settings -->
    <div class="overlay" id="settingsModal" onclick="closeSettings()">
        <div class="modal" style="width:600px;" onclick="event.stopPropagation()">
            <div class="modal-head">
                <div class="modal-title">Configuración</div>
                <button class="btn btn-plain btn-icon" onclick="closeSettings()" title="Cerrar"><i data-icon="x"></i></button>
            </div>
            <div class="modal-body">
                <div class="field">
                    <label for="cfgCSGOPath">Carpeta de CS:GO</label>
                    <div style="display:flex; gap:6px;">
                        <input class="input mono" id="cfgCSGOPath" placeholder="C:\Program Files (x86)\Steam\steamapps\common\csgo legacy\csgo">
                        <button class="btn" onclick="browseCSGO()">Buscar</button>
                    </div>
                    <div class="help">Se usa para abrir cada incidente directamente en el juego.</div>
                </div>

                <div class="field">
                    <label for="cfgPreroll">Segundos antes de la jugada</label>
                    <select class="input" id="cfgPreroll">
                        <option value="1.5">1,5 s</option>
                        <option value="2.5" selected>2,5 s</option>
                        <option value="4.0">4 s</option>
                        <option value="6.0">6 s</option>
                    </select>
                </div>

                <div style="display:flex; justify-content:flex-end; gap:6px;">
                    <button class="btn" onclick="closeSettings()">Cancelar</button>
                    <button class="btn btn-primary" onclick="saveSettings()">Guardar</button>
                </div>

                <div class="divider"></div>

                <div>
                    <div class="block-title">Umbrales de detección</div>
                    <dl class="kv">
                        <dt>Autobunny</dt><dd>75% de perfs o más, o 5 perfs seguidos</dd>
                        <dt>Triggerbot</dt><dd>reacción media menor a 80 ms con desvío menor a 15 ms</dd>
                        <dt>Aimlock / spinbot</dt><dd>5 o más giros de más de 45° en un tick dentro de 1 segundo</dd>
                        <dt>Wallhack</dt><dd>seguir con la mira a un enemigo oculto durante 1 s y 20° de recorrido, 3 veces</dd>
                        <dt>Silent aim</dt><dd>3 impactos con más de 35° de diferencia entre la mira y el daño</dd>
                    </dl>
                </div>

                <div class="divider"></div>

                <div>
                    <div class="block-title">Acerca de</div>
                    <dl class="kv">
                        <dt>Versión</dt><dd id="aboutVersion"></dd>
                        <dt>Código fuente</dt><dd><a class="link" id="aboutRepo" onclick="openRepo()"></a></dd>
                        <dt>Exportaciones</dt><dd><a class="link" onclick="window.openExportFolder()">Documentos\SourceSight</a></dd>
                    </dl>
                    <div style="margin-top:10px;">
                        <button class="btn" onclick="checkUpdates(true)"><i data-icon="refresh"></i>Buscar actualizaciones</button>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Question from Go (duplicate demos) -->
    <div class="overlay" id="promptModal">
        <div class="modal" style="width:520px;" role="dialog" aria-labelledby="promptTitle">
            <div class="modal-head">
                <div class="modal-title" id="promptTitle"></div>
            </div>
            <div class="modal-body">
                <div class="prompt-text" id="promptText"></div>
                <div class="prompt-actions" id="promptActions"></div>
            </div>
        </div>
    </div>

    <div class="toast" id="toast"></div>

    <script>
        let currentReport = null;
        let currentMeta = null;
        let historyItems = [];
        let playerFilter = 'all';
        let chatFilter = 'all';
        let sortKey = 'perfRatio';
        let sortAsc = false;
        let timelineShown = 25;
        let updateInfo = null;
        let appInfo = { version: '', repoUrl: '' };

        // ── Icons ─────────────────────────────────────────────────────────
        const ICONS = {
            folder: '<path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/>',
            image: '<rect width="18" height="18" x="3" y="3" rx="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21"/>',
            copy: '<rect width="14" height="14" x="8" y="8" rx="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>',
            settings: '<path d="M21 4h-7M10 4H3M21 12h-9M8 12H3M21 20h-5M12 20H3M14 2v4M8 10v4M16 18v4"/>',
            upload: '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><path d="M17 8l-5-5-5 5"/><path d="M12 3v12"/>',
            play: '<path d="M6 3l14 9-14 9z"/>',
            x: '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
            refresh: '<path d="M3 12a9 9 0 0 1 15.5-6.3L21 8"/><path d="M21 3v5h-5"/><path d="M21 12a9 9 0 0 1-15.5 6.3L3 16"/><path d="M8 16H3v5"/>',
            history: '<path d="M3 12a9 9 0 1 0 3-6.7L3 8"/><path d="M3 3v5h5"/><path d="M12 7v5l4 2"/>',
            external: '<path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>',
            file: '<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/><path d="M10 13h4M10 17h4"/>',
            ban: '<circle cx="12" cy="12" r="10"/><path d="m4.9 4.9 14.2 14.2"/>'
        };

        function icon(name, cls) {
            return '<svg class="ic' + (cls ? ' ' + cls : '') + '" viewBox="0 0 24 24" aria-hidden="true">' + (ICONS[name] || '') + '</svg>';
        }

        function hydrateIcons(root) {
            (root || document).querySelectorAll('i[data-icon]').forEach(el => {
                el.outerHTML = icon(el.getAttribute('data-icon'), el.className);
            });
        }
        hydrateIcons();

        // ── Helpers ───────────────────────────────────────────────────────
        function esc(s) {
            if (s === null || s === undefined) return '';
            return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
        }

        // Server names usually carry decorative emoji; drop pictographs and dangling separators
        function stripEmoji(s) {
            return String(s || '')
                .replace(/[\p{Extended_Pictographic}\u{FE0F}\u{200D}]/gu, '')
                .replace(/\s*([|»•·])\s*(?=[|»•·]|$)/g, '')
                .replace(/\s{2,}/g, ' ')
                .trim();
        }


        function roundOf(tick) {
            const starts = (currentReport && currentReport.header.roundStarts) || [];
            let lo = 0, hi = starts.length;
            while (lo < hi) {
                const mid = (lo + hi) >> 1;
                if (starts[mid] > tick) hi = mid; else lo = mid + 1;
            }
            return lo;
        }

        function cmdButton(cmd, label) {
            return '<button class="cmd" onclick="event.stopPropagation(); copyText(\'' + esc(cmd) + '\', \'Copiado: ' + esc(cmd) + '\')" title="Copiar">' + esc(label || cmd) + icon('copy', 'ic-sm') + '</button>';
        }

        function playButton(tick) {
            return '<button class="play" onclick="event.stopPropagation(); playInGame(' + tick + ')" title="Abrir en CS:GO">' + icon('play') + 'Ver</button>';
        }

        const RED_TYPES = /TRIGGER|SILENT|WALL/;
        function typeTag(type) {
            const cls = RED_TYPES.test(type) ? 'tag-red' : 'tag-yellow';
            return '<span class="tag ' + cls + '">' + esc(type) + '</span>';
        }

        // ── Levels & staff confirmations ──────────────────────────────────
        // Levels come from Go (core/level.go); "confirmed" is a staff decision.
        let FLAG_LEVELS = {};
        let confirmedMap = {};
        const LEVEL_LABEL = { cheat: 'Cheat detectado', sospechoso: 'Sospechoso', conducta: 'Conducta', limpio: 'Limpio' };
        const LEVEL_TAG = { cheat: 'tag-red', sospechoso: 'tag-yellow', conducta: 'tag-plain' };

        function levelOf(p) { return p.level || 'limpio'; }
        function confirmationOf(p) { return confirmedMap[String(p.steamId)] || null; }
        // flagged: cheat or suspicious (worth a look); conduct-only players are not flagged
        function flagged(p) { const l = levelOf(p); return l === 'cheat' || l === 'sospechoso' || !!confirmationOf(p); }

        function flagTag(f) {
            return '<span class="tag ' + (LEVEL_TAG[FLAG_LEVELS[f] || 'sospechoso'] || 'tag-plain') + '">' + esc(f) + '</span>';
        }

        function flagTags(p) {
            return '<span class="tags">' + (p.flags || []).map(flagTag).join('') + '</span>';
        }

        function levelTag(level) {
            if (!level || level === 'limpio') return '<span class="muted">Limpio</span>';
            return '<span class="tag ' + LEVEL_TAG[level] + '">' + LEVEL_LABEL[level] + '</span>';
        }

        function confirmedTag() { return '<span class="tag tag-confirmed">Confirmado</span>'; }

        // Compact cell for tables: confirmation or level, then the detections
        function resultCell(p) {
            const conf = confirmationOf(p);
            const head = conf ? confirmedTag() : (levelOf(p) === 'limpio' ? '' : levelTag(levelOf(p)));
            if (!head && !(p.flags || []).length) return '<span class="muted">Limpio</span>';
            return '<span class="tags">' + head + (p.flags || []).map(flagTag).join('') + '</span>';
        }

        function rowMark(p) {
            if (confirmationOf(p) || levelOf(p) === 'cheat') return ' mark-red';
            if (levelOf(p) === 'sospechoso') return ' mark-yellow';
            return '';
        }

        function refreshConfirmed(then) {
            if (!window.listConfirmed) { if (then) then(); return; }
            window.listConfirmed().then(m => {
                confirmedMap = m || {};
                if (then) then();
            });
        }

        // Re-render whatever is on screen after a confirmation changes
        function rerenderAll() {
            if (currentReport) { renderPlayers(); renderDetections(); renderSummary(); }
            renderPlayerList();
            if (document.getElementById('playerView').style.display === 'flex' && activePlayer) openPlayerPage(activePlayer);
        }

        // Block shown in the player modal and on the player page
        function confirmBlock(steamId, name, level) {
            const conf = confirmedMap[steamId];
            if (conf) {
                return '<div class="confirm-box confirmed">' +
                    '<div><strong>Confirmado como cheater</strong> por el staff el ' + fmtDate(conf.date) +
                    (conf.note ? '<div class="muted" style="margin-top:2px;">' + esc(conf.note) + '</div>' : '') + '</div>' +
                    '<button class="btn" onclick="setConfirmed(\'' + steamId + '\', false)">Quitar confirmación</button></div>';
            }
            const hint = level === 'cheat' || level === 'sospechoso'
                ? 'La app solo marca indicios. Si viste la demo y es cheat, confirmalo: queda guardado para este SteamID.'
                : 'Si por otra vía sabés que usa cheats, podés confirmarlo igual.';
            return '<div class="confirm-box" id="confirmBox-' + steamId + '">' +
                '<div class="muted">' + hint + '</div>' +
                '<div class="confirm-form">' +
                    '<input class="input" id="confirmNote-' + steamId + '" placeholder="Nota (opcional): qué viste, en qué ronda" style="flex:1;">' +
                    '<button class="btn btn-danger" onclick="setConfirmed(\'' + steamId + '\', true, \'' + esc(String(name).replace(/['\\]/g, '')) + '\')">Confirmar cheater</button>' +
                '</div></div>';
        }

        function setConfirmed(steamId, on, name) {
            if (!window.setConfirmed) return;
            const noteEl = document.getElementById('confirmNote-' + steamId);
            const note = on && noteEl ? noteEl.value.trim() : '';
            window.setConfirmed(steamId, name || '', note, on).then(() => {
                refreshConfirmed(() => {
                    rerenderAll();
                    const modalOpen = document.getElementById('playerModal').style.display === 'flex';
                    if (modalOpen && currentReport) {
                        const p = currentReport.players.find(x => String(x.steamId) === steamId);
                        if (p) openPlayer(p);
                    }
                    showToast(on ? 'Marcado como cheater confirmado' : 'Confirmación quitada');
                });
            });
        }

        function emptyRow(cols, text) {
            return '<tr class="empty"><td colspan="' + cols + '">' + text + '</td></tr>';
        }

        // ── Called from Go ────────────────────────────────────────────────
        function setProgress(pct, label) {
            document.getElementById('loadingBar').style.width = pct + '%';
            document.getElementById('loadingText').innerText = label || ('Analizando... ' + pct + '%');
        }

        // Only one main view is visible at a time
        const VIEWS = { drop: 'dropzone', loading: 'loadingState', audit: 'auditContent', overview: 'overviewView', player: 'playerView' };
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

        // ── History ───────────────────────────────────────────────────────
        function refreshHistory() {
            if (!window.listHistory) return;
            window.listHistory().then(items => {
                historyItems = items || [];
                renderHistory();
            });
        }

        let histFilter = 'all';
        let asideTab = 'demos';

        function setHistFilter(f, el) {
            histFilter = f;
            document.querySelectorAll('#histFilter button').forEach(b => b.classList.toggle('on', b === el));
            renderHistory();
        }

        function renderHistory() {
            const list = document.getElementById('histList');
            const q = document.getElementById('histSearch').value.toLowerCase();
            const catSel = document.getElementById('histCategory');
            document.getElementById('histCount').innerText = historyItems.length || '';

            // Category options come from what is actually in the history
            const cats = [...new Set(historyItems.flatMap(e => e.flagTypes || []))].sort();
            const current = catSel.value;
            catSel.innerHTML = '<option value="">Todas las categorías</option>' +
                cats.map(c => '<option value="' + esc(c) + '"' + (c === current ? ' selected' : '') + '>' + esc(c) + '</option>').join('');
            const cat = cats.includes(current) ? current : '';

            const items = historyItems.filter(e => {
                const lvl = e.level || 'limpio';
                if (histFilter === 'cheat' && lvl !== 'cheat') return false;
                if (histFilter === 'suspect' && lvl !== 'sospechoso') return false;
                if (histFilter === 'clean' && (lvl === 'cheat' || lvl === 'sospechoso')) return false;
                if (cat && !(e.flagTypes || []).includes(cat)) return false;
                return !q || (e.demoName + ' ' + e.mapName + ' ' + (e.flagged || []).join(' ')).toLowerCase().includes(q);
            });

            if (historyItems.length === 0) {
                list.innerHTML = '<div class="hist-empty">Las demos que analices van a aparecer acá.</div>';
                return;
            }
            if (items.length === 0) {
                list.innerHTML = '<div class="hist-empty">Ninguna demo coincide con el filtro.</div>';
                return;
            }

            const activeId = currentMeta && currentMeta.historyId;
            list.innerHTML = items.map(e => {
                const lvl = e.level || 'limpio';
                const date = (e.analyzedAt || '').slice(0, 10).split('-').reverse().join('/');
                const tags = (e.flagTypes || []).map(flagTag).join('');
                const levelText = { cheat: '<span class="hist-flag">cheat</span>', sospechoso: '<span class="t-yellow">sospechosa</span>', conducta: '<span>conducta</span>', limpio: '<span>limpia</span>' }[lvl];
                return '<div class="hist-item' + (e.id === activeId ? ' active' : '') + (lvl === 'cheat' ? ' flagged' : (lvl === 'sospechoso' ? ' suspect' : '')) + '" tabindex="0" onclick="openHistoryItem(\'' + e.id + '\')" onkeydown="if(event.key===\'Enter\')openHistoryItem(\'' + e.id + '\')" title="' + esc(e.demoPath) + '">' +
                    '<div class="hist-name">' + esc(e.demoName) + '</div>' +
                    '<div class="hist-sub"><span>' + esc(e.mapName) + '</span><span>' + date + '</span>' + levelText + '</div>' +
                    (tags ? '<div class="hist-tags">' + tags + '</div>' : '') +
                    '<button class="hist-remove" onclick="event.stopPropagation(); removeHistoryItem(\'' + e.id + '\')" title="Quitar del historial (no borra la demo)">' + icon('x', 'ic-sm') + '</button>' +
                    '</div>';
            }).join('');
        }

        function openHistoryItem(id) {
            if (window.openHistory) window.openHistory(id);
        }

        function removeHistoryItem(id) {
            if (window.removeHistory) window.removeHistory(id).then(() => {
                refreshHistory();
                if (asideTab === 'players') { searchPlayers(); showOverview(); }
            });
        }

        // ── Players across all demos ──────────────────────────────────────
        let playerRows = [];
        let activePlayer = null;
        let searchTimer = null;

        function setAsideTab(tab) {
            asideTab = tab;
            document.getElementById('tabDemos').classList.toggle('on', tab === 'demos');
            document.getElementById('tabPlayers').classList.toggle('on', tab === 'players');
            document.getElementById('paneDemos').style.display = tab === 'demos' ? 'flex' : 'none';
            document.getElementById('panePlayers').style.display = tab === 'players' ? 'flex' : 'none';
            if (tab === 'players') {
                searchPlayers();
                if (activePlayer) openPlayerPage(activePlayer); else showOverview();
            } else {
                restoreIdle();
            }
        }

        function searchPlayers() {
            clearTimeout(searchTimer);
            searchTimer = setTimeout(() => {
                if (!window.listPlayers) return;
                const q = document.getElementById('playerSearch').value;
                window.listPlayers(q).then(res => {
                    playerRows = res.rows || [];
                    document.getElementById('playerHint').innerText = res.hint ||
                        (res.steamId && !playerRows.length ? 'Ese SteamID no aparece en ninguna demo analizada.' : '');
                    renderPlayerList();
                    if (res.steamId && playerRows.length === 1) openPlayerPage(playerRows[0].steamId);
                });
            }, 150);
        }

        function renderPlayerList() {
            const list = document.getElementById('playerList');
            if (!playerRows.length) {
                list.innerHTML = '<div class="hist-empty">' + (historyItems.length ? 'Ningún jugador coincide.' : 'Analizá demos para ver jugadores acá.') + '</div>';
                return;
            }
            list.innerHTML = playerRows.slice(0, 300).map(r => {
                const conf = confirmedMap[r.steamId];
                const tags = (conf ? confirmedTag() : '') + (r.flagTypes || []).map(flagTag).join('');
                const mark = conf || r.level === 'cheat' ? ' flagged' : (r.level === 'sospechoso' ? ' suspect' : '');
                return '<div class="hist-item' + (r.steamId === activePlayer ? ' active' : '') + mark + '" tabindex="0" onclick="openPlayerPage(\'' + r.steamId + '\')" onkeydown="if(event.key===\'Enter\')openPlayerPage(\'' + r.steamId + '\')">' +
                    '<div class="hist-name">' + esc(r.name) + '</div>' +
                    '<div class="hist-sub"><span>' + r.demos + (r.demos === 1 ? ' demo' : ' demos') + '</span>' +
                    (r.flaggedDemos ? (r.level === 'cheat' ? '<span class="hist-flag">cheat en ' : '<span class="t-yellow">sospechoso en ') + r.flaggedDemos + '</span>' : (r.level === 'conducta' ? '<span>conducta</span>' : '')) + '</div>' +
                    (tags ? '<div class="hist-tags">' + tags + '</div>' : '') +
                    '</div>';
            }).join('');
        }

        function fmtDate(ts) { return (ts || '').slice(0, 10).split('-').reverse().join('/'); }
        function fmtNum(v, suffix) { return v === null || v === undefined ? '–' : v + (suffix || ''); }

        // One row of "this player vs everyone": the player's value turns red when it is
        // well above the average of all players (twice as high, or +15 points of perfs).
        function compareRow(label, mine, avg, suffix, higherIsWorse, perfLike) {
            let cls = '';
            if (mine !== null && mine !== undefined && avg !== null && avg !== undefined && higherIsWorse) {
                const worse = perfLike ? mine - avg >= 15 : (avg > 0 ? mine >= avg * 2 && mine - avg >= 1 : mine >= 1);
                if (worse) cls = 't-red';
            }
            return '<tr><td>' + label + '</td><td class="r ' + cls + '">' + fmtNum(mine, suffix) + '</td><td class="r muted">' + fmtNum(avg, suffix) + '</td></tr>';
        }

        function showOverview() {
            activePlayer = null;
            renderPlayerList();
            if (!window.getOverview) return;
            window.getOverview().then(ov => {
                const el = document.getElementById('overviewView');
                const a = ov.averages || {};
                const flags = Object.keys(ov.flagCounts || {}).sort((x, y) => ov.flagCounts[y] - ov.flagCounts[x]);
                el.innerHTML =
                    '<div class="page-head"><div><div class="page-title">Todas las demos</div>' +
                    '<div class="page-sub">Promedios de cada jugador en cada demo del historial. Elegí un jugador a la izquierda o buscalo por SteamID.</div></div></div>' +
                    '<div class="summary">' +
                        '<div class="sum-cell"><div class="sum-label">Demos</div><div class="sum-value">' + ov.demos + '</div><div class="sum-sub">' + ov.cheatDemos + ' con cheat · ' + ov.suspectDemos + ' sospechosas</div></div>' +
                        '<div class="sum-cell"><div class="sum-label">Jugadores distintos</div><div class="sum-value">' + ov.players + '</div><div class="sum-sub">' +
                            ov.cheatPlayers + ' cheat · ' + ov.suspectPlayers + ' sospechosos · ' + Object.keys(confirmedMap).length + ' confirmados</div></div>' +
                        '<div class="sum-cell"><div class="sum-label">Perfs promedio</div><div class="sum-value">' + fmtNum(a.perfRatio, '%') + '</div><div class="sum-sub">jugadores con 5 saltos o más</div></div>' +
                        '<div class="sum-cell"><div class="sum-label">Reacción promedio</div><div class="sum-value">' + fmtNum(a.reactionMs, ' ms') + '</div><div class="sum-sub">' + fmtNum(a.snapsPerDemo) + ' snaps por demo</div></div>' +
                    '</div>' +
                    '<section class="panel"><div class="panel-head"><div class="panel-title">Detecciones por categoría</div></div>' +
                    '<div class="table-wrap"><table><thead><tr><th>Categoría</th><th class="r">Veces</th></tr></thead><tbody>' +
                    (flags.length ? flags.map(f => '<tr><td>' + flagTag(f) + '</td><td class="r">' + ov.flagCounts[f] + '</td></tr>').join('')
                                  : emptyRow(2, 'Todavía no hay detecciones en el historial.')) +
                    '</tbody></table></div></section>';
                showView('overview');
            });
        }

        function openPlayerPage(steamId) {
            if (!window.getPlayerDetail) return;
            closePlayer();
            if (asideTab !== 'players') {
                asideTab = 'players';
                document.getElementById('tabDemos').classList.remove('on');
                document.getElementById('tabPlayers').classList.add('on');
                document.getElementById('paneDemos').style.display = 'none';
                document.getElementById('panePlayers').style.display = 'flex';
                searchPlayers();
            }
            activePlayer = steamId;
            renderPlayerList();
            window.getPlayerDetail(steamId).then(d => {
                const el = document.getElementById('playerView');
                if (!d) {
                    el.innerHTML = '<div class="muted">Este jugador no aparece en el historial.</div>';
                    showView('player');
                    return;
                }
                const a = d.averages || {}, g = (d.global && d.global.averages) || {};
                const flagKeys = Object.keys(d.flagCounts || {});
                el.innerHTML =
                    '<div class="page-head"><div style="min-width:0;"><div class="page-title">' + esc(d.name) + '</div>' +
                    '<div class="page-sub mono">SteamID64 ' + esc(d.steamId) + '</div>' +
                    (d.names.length > 1 ? '<div class="page-sub">También jugó como: ' + d.names.filter(n => n !== d.name).map(esc).join(', ') + '</div>' : '') +
                    '</div><div class="modal-actions">' +
                        '<button class="btn" onclick="copyText(\'' + d.steamId + '\', \'SteamID copiado\')">' + icon('copy', 'ic-sm') + 'SteamID</button>' +
                        '<button class="btn" onclick="openSteam(\'' + d.steamId + '\')">' + icon('external', 'ic-sm') + 'Steam</button>' +
                    '</div></div>' +
                    '<div style="display:flex; flex-direction:column; gap:8px;">' +
                        '<div>' + (confirmedMap[d.steamId] ? confirmedTag() : levelTag(d.level)) +
                        ' <span class="muted" style="font-size:12px;">peor resultado en todas sus demos</span></div>' +
                        confirmBlock(d.steamId, d.name, d.level) +
                    '</div>' +
                    '<div class="summary">' +
                        '<div class="sum-cell"><div class="sum-label">Demos</div><div class="sum-value">' + d.demos + '</div><div class="sum-sub">del ' + fmtDate(d.firstSeen) + ' al ' + fmtDate(d.lastSeen) + '</div></div>' +
                        '<div class="sum-cell"><div class="sum-label">Marcado en</div><div class="sum-value' + (d.flaggedDemos ? ' t-red' : '') + '">' + d.flaggedDemos + '</div><div class="sum-sub">' + (flagKeys.length ? flagKeys.map(f => f + ' ×' + d.flagCounts[f]).join(', ') : 'ninguna demo') + '</div></div>' +
                        '<div class="sum-cell"><div class="sum-label">Incidentes</div><div class="sum-value">' + d.totals.incidents + '</div><div class="sum-sub">' + fmtNum(a.incidentsPerDemo) + ' por demo</div></div>' +
                        '<div class="sum-cell"><div class="sum-label">Kills cegado / humo</div><div class="sum-value">' + d.totals.blind + ' / ' + d.totals.smoke + '</div><div class="sum-sub">macro: ' + d.totals.macro + '</div></div>' +
                    '</div>' +
                    '<section class="panel"><div class="panel-head"><div class="panel-title">Promedio del jugador vs. todos</div><span class="muted">En rojo, lo que está muy por encima del promedio general</span></div>' +
                    '<div class="table-wrap"><table class="compare"><thead><tr><th></th><th class="r">Este jugador</th><th class="r">Promedio general</th></tr></thead><tbody>' +
                        compareRow('Perfs de bhop', a.perfRatio, g.perfRatio, '%', true, true) +
                        compareRow('Reacción (triggerbot)', a.reactionMs, g.reactionMs, ' ms', false) +
                        compareRow('Prefire: daño apenas aparece el rival', a.prefirePct, g.prefirePct, '%', true, true) +
                        compareRow('Snaps por demo', a.snapsPerDemo, g.snapsPerDemo, '', true) +
                        compareRow('Silent aim por demo', a.silentPerDemo, g.silentPerDemo, '', true) +
                        compareRow('Wallhack por demo', a.wallPerDemo, g.wallPerDemo, '', true) +
                        compareRow('Incidentes por demo', a.incidentsPerDemo, g.incidentsPerDemo, '', true) +
                    '</tbody></table></div></section>' +
                    '<section class="panel"><div class="panel-head"><div class="panel-title">Demos <span class="count">' + d.demos + '</span></div><span class="muted">Click para abrir la demo</span></div>' +
                    '<div class="table-wrap"><table><thead><tr><th>Fecha</th><th>Demo</th><th>Mapa</th><th class="r">Perfs</th><th class="r">Reacción</th><th class="r">Snaps</th><th class="r">Wallhack</th><th class="r">Resultado</th></tr></thead><tbody>' +
                    d.appearances.map(ap =>
                        '<tr class="click' + (ap.flags && ap.flags.length ? ' mark-red' : '') + '" onclick="openDemoFromPlayer(\'' + ap.entryId + '\')">' +
                            '<td class="num">' + fmtDate(ap.analyzedAt) + '</td>' +
                            '<td class="player">' + esc(ap.demoName) + '</td>' +
                            '<td class="muted">' + esc(ap.mapName) + '</td>' +
                            '<td class="r">' + (ap.totalJumps >= 5 ? ap.perfRatio.toFixed(1) + '%' : '<span class="muted">–</span>') + '</td>' +
                            '<td class="r">' + fmtNum(ap.reactionMs, ' ms') + '</td>' +
                            '<td class="r">' + ap.snaps + '</td>' +
                            '<td class="r">' + ap.wallhack + '</td>' +
                            '<td class="r">' + (ap.flags && ap.flags.length ? '<span class="tags">' + ap.flags.map(flagTag).join('') + '</span>' : '<span class="muted">Limpio</span>') + '</td>' +
                        '</tr>'
                    ).join('') +
                    '</tbody></table></div></section>';
                showView('player');
            });
        }

        function openDemoFromPlayer(entryId) {
            asideTab = 'demos';
            document.getElementById('tabDemos').classList.add('on');
            document.getElementById('tabPlayers').classList.remove('on');
            document.getElementById('paneDemos').style.display = 'flex';
            document.getElementById('panePlayers').style.display = 'none';
            openHistoryItem(entryId);
        }

        // ── Players ───────────────────────────────────────────────────────
        function setPlayerFilter(f, el) {
            playerFilter = f;
            document.querySelectorAll('#playerFilter button').forEach(b => b.classList.toggle('on', b === el));
            renderPlayers();
        }

        function sortPlayers(key) {
            if (sortKey === key) sortAsc = !sortAsc;
            else { sortKey = key; sortAsc = key === 'name'; }
            renderPlayers();
        }

        function renderPlayers() {
            if (!currentReport) return;
            document.querySelectorAll('#playersCard th.sortable').forEach(th => {
                const on = th.getAttribute('data-key') === sortKey;
                th.classList.toggle('sorted', on);
                th.classList.toggle('asc', on && sortAsc);
            });

            const q = document.getElementById('searchPlayer').value.toLowerCase();
            let players = currentReport.players.slice();
            if (q) players = players.filter(p => p.name.toLowerCase().includes(q));
            if (playerFilter === 'cheat') players = players.filter(p => confirmationOf(p) || levelOf(p) === 'cheat');
            if (playerFilter === 'suspect') players = players.filter(p => !confirmationOf(p) && levelOf(p) === 'sospechoso');
            if (playerFilter === 'clean') players = players.filter(p => !flagged(p));

            players.sort((a, b) => {
                const va = a[sortKey], vb = b[sortKey];
                if (typeof va === 'string') return sortAsc ? va.localeCompare(vb) : vb.localeCompare(va);
                return sortAsc ? va - vb : vb - va;
            });

            document.getElementById('playersCount').innerText = players.length;
            const body = document.getElementById('playersBody');
            if (players.length === 0) {
                body.innerHTML = emptyRow(8, 'Ningún jugador coincide.');
                return;
            }
            body.innerHTML = '';
            players.forEach(p => {
                const tr = document.createElement('tr');
                tr.className = 'click' + rowMark(p);
                tr.onclick = () => openPlayer(p);
                const perfCls = p.perfRatio >= 75 ? 't-red' : (p.perfRatio >= 55 ? 't-yellow' : '');
                tr.innerHTML =
                    '<td class="player">' + esc(p.name) + '</td>' +
                    '<td class="r">' + p.totalJumps + '</td>' +
                    '<td class="r ' + perfCls + '">' + p.perfRatio.toFixed(1) + '% <span class="muted">' + p.perfectHops + '</span></td>' +
                    '<td class="r">' + p.oneTickRatio.toFixed(1) + '%</td>' +
                    '<td class="r ' + (p.maxStreak >= 5 ? 't-red' : '') + '">' + p.maxStreak + '</td>' +
                    '<td class="r ' + (p.flags && p.flags.includes('AIMLOCK') ? 't-red' : '') + '">' + p.snapCount + '</td>' +
                    '<td class="r">' + p.maxSpeed + ' <span class="muted">u/s</span></td>' +
                    '<td class="r">' + resultCell(p) + '</td>';
                body.appendChild(tr);
            });
        }

        function renderDetections() {
            // Everyone worth a look except bhop-only cases, which the players table already shows
            const rows = currentReport.players.filter(p => flagged(p) &&
                (confirmationOf(p) || p.flags.some(f => f !== 'AUTOBUNNY' && f !== 'SOSPECHOSO-BHOP')));
            const body = document.getElementById('detectBody');
            if (rows.length === 0) {
                body.innerHTML = emptyRow(8, 'Nadie superó los umbrales de triggerbot, aimlock, silent aim, macro, wallhack, prefire, kills cegado ni humo.');
                return;
            }
            body.innerHTML = '';
            rows.forEach(p => {
                const tr = document.createElement('tr');
                tr.className = 'click' + rowMark(p);
                tr.onclick = () => openPlayer(p);
                const hasTrig = p.triggerAvgMs !== null && p.triggerAvgMs !== undefined;
                tr.innerHTML =
                    '<td class="player">' + esc(p.name) + '</td>' +
                    '<td class="r ' + (hasTrig && p.triggerAvgMs < 80 ? 't-red' : '') + '">' + (hasTrig ? p.triggerAvgMs + ' ms <span class="muted">±' + (p.triggerSigma || 0).toFixed(0) + '</span>' : '<span class="muted">–</span>') + '</td>' +
                    '<td class="r">' + p.snapCount + ' / ' + p.silentCount + '</td>' +
                    '<td class="r">' + (p.macroHits || '<span class="muted">–</span>') + '</td>' +
                    '<td class="r">' + p.flashKills + ' / ' + p.smokeKills + '</td>' +
                    '<td class="r">' + (p.wallTrackCount || '<span class="muted">–</span>') + '</td>' +
                    '<td class="r' + ((p.flags || []).includes('PREFIRE') ? ' t-red' : '') + '">' + (p.peekEngagements ? p.prefireRatio + '%' : '<span class="muted">–</span>') + '</td>' +
                    '<td class="r">' + resultCell(p) + '</td>';
                body.appendChild(tr);
            });
        }

        function renderTimeline() {
            const all = currentReport.timeline || [];
            const card = document.getElementById('timelineCard');
            card.style.display = all.length ? 'block' : 'none';
            if (!all.length) return;

            const shown = Math.min(timelineShown, all.length);
            document.getElementById('timelineCount').innerText = all.length;
            document.getElementById('timelineBody').innerHTML = all.slice(0, shown).map(inc =>
                '<tr>' +
                    '<td class="r muted">' + roundOf(inc.tick) + '</td>' +
                    '<td>' + typeTag(inc.type) + '</td>' +
                    '<td class="player">' + esc(inc.player) + '</td>' +
                    '<td>' + cmdButton(inc.gotoCmd) + '</td>' +
                    '<td>' + playButton(inc.tick) + '</td>' +
                    '<td class="wrap muted" style="min-width:260px;">' + esc(inc.note) + '</td>' +
                '</tr>'
            ).join('');
            document.getElementById('timelineMore').style.display = shown < all.length ? 'block' : 'none';
        }

        function showMoreTimeline() {
            timelineShown += 25;
            renderTimeline();
        }

        function renderAdmin() {
            const log = currentReport.adminLog || [];
            document.getElementById('adminCount').innerText = log.length || '';
            const body = document.getElementById('adminBody');
            if (!log.length) {
                body.innerHTML = emptyRow(6, 'No hubo comandos de castigo de admins en esta demo.');
                return;
            }
            body.innerHTML = log.map(ev =>
                '<tr class="' + (ev.isRevenge ? 'mark-red' : '') + '">' +
                    '<td class="r muted">' + ev.round + '</td>' +
                    '<td class="player">' + esc(ev.adminName) + '</td>' +
                    '<td><code>' + esc(ev.command) + '</code></td>' +
                    '<td>' + esc(ev.targetName) + '</td>' +
                    '<td class="wrap">' + (ev.isRevenge ? '<span class="tag tag-red">Venganza</span> ' : '') + '<span class="muted">' + esc(ev.note || '') + '</span></td>' +
                    '<td class="r">' + (ev.tick ? playButton(ev.tick) : '') + '</td>' +
                '</tr>'
            ).join('');
        }

        function setChatFilter(f, el) {
            chatFilter = f;
            document.querySelectorAll('#chatFilter button').forEach(b => b.classList.toggle('on', b === el));
            renderChat();
        }

        function renderChat() {
            const all = currentReport.chatLog || [];
            let msgs = all;
            if (chatFilter === 'toxic') msgs = all.filter(m => m.isToxicity);
            if (chatFilter === 'ghosting') msgs = all.filter(m => m.isGhosting);
            if (chatFilter === 'spam') msgs = all.filter(m => m.category === 'SPAM BIND');

            document.getElementById('chatCount').innerText = all.length ? (msgs.length === all.length ? all.length : msgs.length + ' de ' + all.length) : '';
            const body = document.getElementById('chatBody');
            if (!all.length) { body.innerHTML = emptyRow(7, 'La demo no tiene mensajes de chat.'); return; }
            if (!msgs.length) { body.innerHTML = emptyRow(7, 'Ningún mensaje de este tipo.'); return; }

            body.innerHTML = msgs.map(m => {
                let tag = '<span class="muted">–</span>';
                if (m.isToxicity) tag = '<span class="tag tag-red">Tóxico</span>';
                else if (m.isGhosting) tag = '<span class="tag tag-yellow">Ghosting</span>';
                else if (m.category === 'SPAM BIND') tag = '<span class="tag tag-yellow">Spam</span>';
                return '<tr class="' + (m.isToxicity ? 'mark-red' : '') + '">' +
                    '<td class="r muted">' + m.round + '</td>' +
                    '<td class="player">' + esc(m.senderName) + '</td>' +
                    '<td class="muted">' + (m.isTeamOnly ? 'Equipo' : 'Global') + '</td>' +
                    '<td class="wrap" style="min-width:240px; color:var(--text); user-select:text;">' + esc(m.text) + '</td>' +
                    '<td>' + tag + '</td>' +
                    '<td>' + cmdButton(m.gotoCmd, String(m.tick)) + '</td>' +
                    '<td class="r">' + playButton(m.tick) + '</td>' +
                '</tr>';
            }).join('');
        }

        // ── Player detail ─────────────────────────────────────────────────
        function stat(label, value, cls) {
            return '<div class="stat"><div class="stat-label">' + label + '</div><div class="stat-value ' + (cls || '') + '">' + value + '</div></div>';
        }

        function openPlayer(p) {
            const isCheat = !!confirmationOf(p) || levelOf(p) === 'cheat';
            const hasSteam = !!p.steamId && p.steamId !== "0";
            const sid = String(p.steamId);

            document.getElementById('pmName').innerText = p.name;
            document.getElementById('pmSub').innerText = hasSteam ? 'SteamID64 ' + sid : 'Sin SteamID (bot o demo sin datos)';

            let actions = '';
            if (hasSteam) actions += '<button class="btn" onclick="openSteam(\'' + sid + '\')">' + icon('external', 'ic-sm') + 'Steam</button>';
            if (hasSteam) actions += '<button class="btn" onclick="exportEvidence(\'' + sid + '\')">' + icon('file', 'ic-sm') + 'Exportar evidencia</button>';
            if (hasSteam && isCheat) actions += '<button class="btn btn-danger" onclick="copyBan(\'' + sid + '\')">' + icon('ban', 'ic-sm') + 'Copiar ban</button>';
            actions += '<button class="btn btn-plain btn-icon" onclick="closePlayer()" title="Cerrar">' + icon('x') + '</button>';
            document.getElementById('pmActions').innerHTML = actions;

            const hasTrig = p.triggerAvgMs !== null && p.triggerAvgMs !== undefined;
            let html = '';

            if (p.previousInfractions > 0) {
                html += '<div class="callout callout-red">Tiene ' + p.previousInfractions + (p.previousInfractions === 1 ? ' detección anterior' : ' detecciones anteriores') + ' en demos que analizaste antes.</div>';
            }

            html += '<div><div class="block-title">Resultado</div>' + resultCell(p) +
                (levelOf(p) !== 'limpio' && !confirmationOf(p)
                    ? '<div class="muted" style="margin-top:6px; font-size:12px;">' + ({
                        cheat: 'Superó un umbral fuerte de detección. Revisá la demo antes de sancionar.',
                        sospechoso: 'Hay indicios, pero no alcanzan para decir que es cheat. Conviene mirar la demo.',
                        conducta: 'No es cheat: es algo de comportamiento para el staff (AFK, flashes al equipo).'
                    })[levelOf(p)] + '</div>'
                    : '') +
                '</div>';
            if (hasSteam) html += '<div>' + confirmBlock(sid, p.name, levelOf(p)) + '</div>';

            html += '<div><div class="block-title">Métricas</div><div class="stats">' +
                stat('Saltos', p.totalJumps) +
                stat('Perfs', p.perfRatio.toFixed(1) + '%', p.perfRatio >= 75 ? 't-red' : (p.perfRatio >= 55 ? 't-yellow' : '')) +
                stat('Presión de 1 tick', p.oneTickRatio.toFixed(1) + '%') +
                stat('Racha de perfs', p.maxStreak, p.maxStreak >= 5 ? 't-red' : '') +
                stat('Vel. máxima', p.maxSpeed + ' u/s') +
                stat('Reacción', hasTrig ? p.triggerAvgMs + ' ms' : '–', hasTrig && p.triggerAvgMs < 80 ? 't-red' : '') +
                stat('Snaps', p.snapCount, p.flags && p.flags.includes('AIMLOCK') ? 't-red' : '') +
                stat('Silent aim', p.silentCount, p.silentCount >= 3 ? 't-red' : '') +
                stat('Wallhack', p.wallTrackCount || 0, p.wallTrackCount >= 3 ? 't-red' : '') +
                stat('Macro', p.macroHits || 0, p.macroHits > 0 ? 't-yellow' : '') +
                stat('Kills cegado', p.flashKills || 0, p.flashKills > 0 ? 't-yellow' : '') +
                stat('Kills por humo', p.smokeKills || 0, p.smokeKills > 0 ? 't-yellow' : '') +
                stat('Teamflash', (p.teamflashSecs || 0).toFixed(1) + ' s', p.teamflashSecs >= 35 ? 't-yellow' : '') +
                stat('AFK', p.isAfk ? 'Sí' : 'No', p.isAfk ? 't-yellow' : '') +
                stat('Prefire', p.peekEngagements ? p.prefireRatio + '% de ' + p.peekEngagements : '–', (p.flags || []).includes('PREFIRE') ? 't-red' : (p.prefireRatio >= 30 && p.peekEngagements >= 10 ? 't-yellow' : '')) +
                '</div></div>';

            const incs = p.incidents || [];
            if (incs.length && (currentReport.header.roundStarts || []).length) {
                html += '<div><div class="block-title">Incidentes por ronda</div><div class="chart" id="roundChart"></div></div>';
            }

            html += '<div><div class="block-title">Incidentes <span class="count">' + incs.length + '</span></div>';
            if (incs.length) {
                html += '<div class="list">' + incs.slice().sort((a, b) => a.tick - b.tick).map(inc =>
                    '<div class="list-row">' +
                        '<span class="muted num" style="width:46px; flex-shrink:0;">R' + roundOf(inc.tick) + '</span>' +
                        typeTag(inc.type) +
                        '<span class="grow" title="' + esc(inc.note) + '">' + esc(inc.note) + '</span>' +
                        cmdButton(inc.gotoCmd) + playButton(inc.tick) +
                    '</div>'
                ).join('') + '</div>';
            } else {
                html += '<div class="muted">Sin incidentes.</div>';
            }
            html += '</div>';

            html += '<div id="pmProfile"></div>';

            if (p.chatMessages && p.chatMessages.length) {
                html += '<div><div class="block-title">Chat <span class="count">' + p.chatMessages.length + '</span></div><div class="list" style="max-height:180px;">' +
                    p.chatMessages.map(m =>
                        '<div class="list-row">' +
                            '<span class="muted num" style="width:46px; flex-shrink:0;">R' + m.round + '</span>' +
                            (m.isToxicity ? '<span class="tag tag-red">Tóxico</span>' : '') +
                            '<span class="grow" style="color:var(--text); user-select:text;">' + esc(m.text) + '</span>' +
                            playButton(m.tick) +
                        '</div>'
                    ).join('') + '</div></div>';
            }

            document.getElementById('pmBody').innerHTML = html;
            document.getElementById('playerModal').style.display = 'flex';
            document.getElementById('pmBody').scrollTop = 0;

            if (document.getElementById('roundChart')) drawRoundChart(document.getElementById('roundChart'), incs);
            if (hasSteam && window.getPlayerDetail) {
                window.getPlayerDetail(sid).then(d => renderProfile(d));
            }
        }

        function renderProfile(d) {
            const el = document.getElementById('pmProfile');
            if (!el || !d) return;
            const a = d.averages || {};
            const flagKeys = Object.keys(d.flagCounts || {});
            el.innerHTML = '<div class="block-title">En todas tus demos <span class="count">' + d.demos + '</span></div>' +
                '<dl class="kv">' +
                    '<dt>Marcado en</dt><dd>' + (d.flaggedDemos ? '<span class="t-red">' + d.flaggedDemos + ' de ' + d.demos + '</span> · ' + flagKeys.map(f => esc(f) + ' ×' + d.flagCounts[f]).join(', ') : 'ninguna') + '</dd>' +
                    '<dt>Perfs promedio</dt><dd>' + fmtNum(a.perfRatio, '%') + '</dd>' +
                    '<dt>Reacción promedio</dt><dd>' + fmtNum(a.reactionMs, ' ms') + '</dd>' +
                    '<dt>Snaps por demo</dt><dd>' + fmtNum(a.snapsPerDemo) + '</dd>' +
                    (d.names.length > 1 ? '<dt>Otros nombres</dt><dd>' + d.names.filter(n => n !== d.name).map(esc).join(', ') + '</dd>' : '') +
                '</dl>' +
                '<div style="margin-top:10px;"><button class="btn" onclick="openPlayerPage(\'' + d.steamId + '\')">Ver ficha completa</button></div>';
        }

        // Single-series bar chart: incidents per round, with a per-bar hover breakdown.
        function drawRoundChart(host, incs) {
            const rounds = currentReport.header.roundStarts.length;
            const buckets = Array.from({ length: rounds + 1 }, () => ({ total: 0, types: {} }));
            incs.forEach(inc => {
                const r = Math.min(roundOf(inc.tick), rounds);
                buckets[r].total++;
                buckets[r].types[inc.type] = (buckets[r].types[inc.type] || 0) + 1;
            });
            const data = buckets.slice(1);
            const max = Math.max(1, ...data.map(d => d.total));

            const W = Math.max(320, host.clientWidth - 24), H = 130, padL = 26, padB = 18, padT = 6;
            const plotW = W - padL, plotH = H - padB - padT;
            const band = plotW / rounds;
            const gap = Math.min(2, band * 0.3);
            const barW = Math.max(1, band - gap);
            const y = v => padT + plotH - (v / max) * plotH;

            let svg = '<svg viewBox="0 0 ' + W + ' ' + H + '" role="img" aria-label="Incidentes por ronda: máximo ' + max + ' en una ronda">';
            svg += '<line class="grid" x1="' + padL + '" x2="' + W + '" y1="' + y(max) + '" y2="' + y(max) + '"/>';
            svg += '<line class="grid" x1="' + padL + '" x2="' + W + '" y1="' + (padT + plotH) + '" y2="' + (padT + plotH) + '"/>';
            svg += '<text class="axis" x="' + (padL - 6) + '" y="' + (y(max) + 4) + '" text-anchor="end">' + max + '</text>';
            svg += '<text class="axis" x="' + (padL - 6) + '" y="' + (padT + plotH + 4) + '" text-anchor="end">0</text>';

            const step = rounds > 24 ? 5 : (rounds > 12 ? 2 : 1);
            data.forEach((d, i) => {
                const x = padL + i * band + gap / 2;
                svg += '<rect class="hit" data-i="' + i + '" x="' + (padL + i * band) + '" y="' + padT + '" width="' + band + '" height="' + plotH + '"/>';
                if (d.total > 0) {
                    const top = y(d.total), h = padT + plotH - top, r = Math.min(4, barW / 2, h);
                    svg += '<path class="bar" data-i="' + i + '" d="M' + x + ',' + (padT + plotH) + 'V' + (top + r) + 'Q' + x + ',' + top + ' ' + (x + r) + ',' + top +
                        'H' + (x + barW - r) + 'Q' + (x + barW) + ',' + top + ' ' + (x + barW) + ',' + (top + r) + 'V' + (padT + plotH) + 'Z"/>';
                }
                if ((i + 1) % step === 0 || i === 0) {
                    svg += '<text class="axis" x="' + (padL + i * band + band / 2) + '" y="' + (H - 4) + '" text-anchor="middle">' + (i + 1) + '</text>';
                }
            });
            svg += '</svg><div class="tooltip" id="chartTip"></div>';
            host.innerHTML = svg;

            const tip = host.querySelector('#chartTip');
            host.querySelectorAll('.hit').forEach(hit => {
                hit.addEventListener('mouseenter', () => {
                    const i = +hit.getAttribute('data-i');
                    const d = data[i];
                    host.querySelectorAll('.bar').forEach(b => b.classList.toggle('hl', b.getAttribute('data-i') === String(i)));
                    tip.innerHTML = '<div class="row"><b>Ronda ' + (i + 1) + '</b><b class="num">' + d.total + '</b></div>' +
                        Object.keys(d.types).sort((a, b) => d.types[b] - d.types[a]).map(t => '<div class="row"><span>' + esc(t) + '</span><span class="num">' + d.types[t] + '</span></div>').join('');
                    tip.style.display = 'block';
                    const hostBox = host.getBoundingClientRect(), box = hit.getBoundingClientRect();
                    let left = box.left - hostBox.left + box.width / 2 - tip.offsetWidth / 2;
                    left = Math.max(4, Math.min(left, hostBox.width - tip.offsetWidth - 4));
                    tip.style.left = left + 'px';
                    tip.style.top = '-' + (tip.offsetHeight - 4) + 'px';
                });
                hit.addEventListener('mouseleave', () => {
                    tip.style.display = 'none';
                    host.querySelectorAll('.bar.hl').forEach(b => b.classList.remove('hl'));
                });
            });
        }

        function closePlayer() { document.getElementById('playerModal').style.display = 'none'; }

        function findPlayer(sid) { return currentReport.players.find(p => String(p.steamId) === sid); }

        function openSteam(sid) { openLink('https://steamcommunity.com/profiles/' + sid); }

        function exportEvidence(sid) { if (window.exportPlayerEvidence) window.exportPlayerEvidence(sid); }

        function copyBan(sid) {
            const p = findPlayer(sid);
            if (!p) return;
            const reason = (p.flags || []).length ? p.flags.join(' + ') : 'confirmado por staff';
            const cmd = 'sm_ban #' + sid + ' 0 "SourceSight: ' + reason + '"';
            copyText(cmd, 'Comando de ban copiado');
        }

        // ── Actions ───────────────────────────────────────────────────────
        function pickDemos() {
            if (window.selectDemoFile) window.selectDemoFile();
        }

        function reanalyze() {
            if (window.reanalyzeCurrent) window.reanalyzeCurrent();
        }

        function playInGame(tick) {
            if (window.launchInGame) window.launchInGame(tick);
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
                    md += '- R' + roundOf(inc.tick) + ' ' + inc.player + ' ' + inc.type + ': ' + q + inc.gotoCmd + q + '\n';
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

        // ── Settings & updates ────────────────────────────────────────────
        function openSettings() {
            if (window.getConfig) {
                window.getConfig().then(s => {
                    try {
                        const cfg = JSON.parse(s);
                        if (cfg.csgoPath) document.getElementById('cfgCSGOPath').value = cfg.csgoPath;
                        if (cfg.prerollSeconds) document.getElementById('cfgPreroll').value = cfg.prerollSeconds.toFixed(1);
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
            if (!window.saveConfig) { closeSettings(); return; }
            window.saveConfig(p, preroll)
                .then(() => { showToast('Configuración guardada'); closeSettings(); })
                .catch(err => showToast('No se pudo guardar: ' + err));
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

        // ── Startup ───────────────────────────────────────────────────────
        let readyFired = false;
        function notifyReady() {
            if (readyFired) return;
            readyFired = true;
            if (window.getAppInfo) {
                window.getAppInfo().then(info => {
                    appInfo = info;
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
                    } catch (e) {}
                });
            }
            const levels = window.getFlagLevels ? window.getFlagLevels().then(m => { FLAG_LEVELS = m || {}; }) : Promise.resolve();
            const confirmed = new Promise(resolve => refreshConfirmed(resolve));
            Promise.all([levels, confirmed]).finally(() => {
                refreshHistory();
                if (window.onAppReady) window.onAppReady();
            });
            setTimeout(() => checkUpdates(false), 1500);
        }
        window.addEventListener('DOMContentLoaded', notifyReady);
        setTimeout(notifyReady, 150);
    </script>
</body>
</html>`
