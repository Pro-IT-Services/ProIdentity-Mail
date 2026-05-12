package admin

const adminIndexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>ProIdentity Mail Admin</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Public+Sans:wght@400;600;700&display=swap" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap" rel="stylesheet">
  <style nonce="__PROIDENTITY_CSP_NONCE__">
    :root {
      --background: #f7f9fb;
      --surface: #ffffff;
      --rail: #e1e4e8;
      --soft: #f1f4f7;
      --muted-soft: #e9edf2;
      --ink: #171a20;
      --muted: #596170;
      --outline: #cfd5dd;
      --primary: #4648d4;
      --primary-strong: #3032bd;
      --primary-soft: #e4e5ff;
      --success: #087443;
      --success-soft: #dcfae6;
      --warning: #a15c07;
      --warning-soft: #fff3d6;
      --danger: #b42318;
      --danger-soft: #fee4e2;
      --shadow: 0 8px 22px rgba(15, 23, 42, .06);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      background: var(--background);
      color: var(--ink);
      font: 13px/1.45 "Public Sans", system-ui, sans-serif;
      letter-spacing: 0;
    }
    button, input, select, textarea { font: inherit; }
    .material-symbols-outlined {
      font-variation-settings: "FILL" 0, "wght" 400, "GRAD" 0, "opsz" 24;
      font-size: 21px;
      line-height: 1;
    }
    .app { min-height: 100vh; padding-left: 236px; }
    aside {
      position: fixed;
      inset: 0 auto 0 0;
      z-index: 20;
      width: 236px;
      background: var(--rail);
      border-right: 1px solid rgba(118,117,134,.24);
      padding: 18px 12px;
      display: flex;
      flex-direction: column;
    }
    .brand { display: flex; align-items: center; gap: 12px; margin: 0 6px 30px; }
    .brand-mark {
      width: 36px;
      height: 36px;
      border-radius: 8px;
      background: var(--primary);
      color: white;
      display: grid;
      place-items: center;
      box-shadow: 0 8px 18px rgba(70,72,212,.22);
    }
    .brand h1 { margin: 0; color: var(--primary); font-size: 19px; line-height: 1; }
    .brand p { margin: 4px 0 0; color: #2d3340; font-size: 11px; font-weight: 700; letter-spacing: .08em; }
    nav { display: grid; gap: 4px; }
    .nav-item {
      width: 100%;
      min-height: 38px;
      border: 0;
      border-radius: 8px;
      background: transparent;
      color: #242a35;
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 0 12px;
      cursor: pointer;
      text-align: left;
    }
    .nav-item:hover { background: rgba(255,255,255,.48); }
    .nav-item.active {
      background: rgba(255,255,255,.66);
      color: var(--primary);
      font-weight: 700;
      box-shadow: inset -4px 0 0 var(--primary);
    }
    .sidebar-bottom { margin-top: auto; padding-top: 14px; border-top: 1px solid rgba(118,117,134,.24); display: grid; gap: 4px; }
    header {
      height: 58px;
      position: sticky;
      top: 0;
      z-index: 10;
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 18px;
      padding: 0 24px;
      border-bottom: 1px solid var(--outline);
      background: rgba(247,249,251,.94);
      backdrop-filter: blur(12px);
    }
    .title-area { min-width: 220px; }
    .title-area h2 { margin: 0; font-size: 18px; line-height: 1.1; }
    .title-area p { margin: 3px 0 0; color: var(--muted); font-size: 12px; }
    .search {
      width: min(430px, 42vw);
      min-height: 36px;
      border-radius: 8px;
      background: var(--soft);
      color: var(--muted);
      padding: 0 12px;
      display: flex;
      align-items: center;
      gap: 9px;
    }
    .search input { width: 100%; border: 0; outline: 0; background: transparent; color: var(--ink); }
    .top-actions { display: flex; align-items: center; gap: 10px; }
    .health-pill {
      min-height: 32px;
      border-radius: 999px;
      padding: 0 12px;
      display: inline-flex;
      align-items: center;
      gap: 7px;
      background: var(--muted-soft);
      color: #252b36;
      font-weight: 700;
      white-space: nowrap;
    }
    main { width: min(1420px, 100%); margin: 0 auto; padding: 24px 28px 36px; }
    .scope-bar {
      margin-bottom: 16px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: var(--surface);
      box-shadow: var(--shadow);
      padding: 12px 14px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
    }
    .scope-main { display: flex; align-items: end; gap: 12px; flex-wrap: wrap; min-width: 0; }
    .scope-picker { min-width: 230px; }
    .scope-picker select { width: 100%; }
    .scope-empty {
      min-height: 36px;
      border: 1px dashed var(--outline);
      border-radius: 8px;
      padding: 8px 10px;
      display: flex;
      align-items: center;
      color: var(--muted);
      font-weight: 700;
    }
    .scope-summary { display: flex; align-items: center; justify-content: end; gap: 8px; flex-wrap: wrap; }
    .context-field {
      min-height: 36px;
      border-radius: 8px;
      border: 1px solid var(--outline);
      background: var(--soft);
      color: var(--ink);
      padding: 8px 10px;
      display: flex;
      align-items: center;
      font-weight: 700;
    }
    .hero {
      display: flex;
      align-items: end;
      justify-content: space-between;
      gap: 20px;
      margin-bottom: 18px;
    }
    .hero h3 { margin: 0; font-size: 25px; line-height: 1.15; }
    .hero p { margin: 6px 0 0; color: var(--muted); font-size: 13px; }
    .actions { display: flex; gap: 9px; align-items: center; flex-wrap: wrap; }
    .button {
      min-height: 35px;
      border-radius: 8px;
      border: 1px solid var(--outline);
      background: var(--surface);
      color: #242936;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      padding: 0 13px;
      font-weight: 700;
      cursor: pointer;
      text-decoration: none;
      white-space: nowrap;
    }
    .button:hover { background: var(--soft); }
    .button.primary { border-color: var(--primary); background: var(--primary); color: white; box-shadow: 0 8px 18px rgba(70,72,212,.2); }
    .button.primary:hover { background: var(--primary-strong); }
    .button.danger { color: var(--danger); }
    .button:disabled { opacity: .55; cursor: default; }
    .spin { animation: spin 900ms linear infinite; }
    @keyframes spin { to { transform: rotate(360deg); } }
    .tabs {
      display: flex;
      gap: 6px;
      padding: 10px 12px 0;
      border-bottom: 1px solid var(--outline);
      background: #fbfcfd;
      overflow-x: auto;
    }
    .tab {
      min-height: 36px;
      border: 0;
      border-radius: 8px 8px 0 0;
      background: transparent;
      color: var(--muted);
      display: inline-flex;
      align-items: center;
      gap: 7px;
      padding: 0 12px;
      font-weight: 700;
      cursor: pointer;
      white-space: nowrap;
    }
    .tab.active {
      background: var(--surface);
      color: var(--primary);
      box-shadow: inset 0 -2px 0 var(--primary);
    }
    .grid { display: grid; gap: 16px; }
    .stats { grid-template-columns: repeat(4, minmax(0,1fr)); margin-bottom: 16px; }
    .two-col { grid-template-columns: minmax(0, 1.5fr) minmax(330px, .8fr); align-items: start; }
    .three-col { grid-template-columns: repeat(3, minmax(0, 1fr)); }
    .card {
      background: var(--surface);
      border: 1px solid var(--outline);
      border-radius: 8px;
      box-shadow: var(--shadow);
      overflow: hidden;
    }
    .card-body { padding: 16px; }
    .panel-head {
      min-height: 56px;
      padding: 13px 16px;
      border-bottom: 1px solid var(--outline);
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 14px;
    }
    .panel-head h4 { margin: 0; font-size: 16px; }
    .panel-head p { margin: 3px 0 0; color: var(--muted); font-size: 12px; }
    .stat { padding: 16px; min-height: 104px; display: flex; flex-direction: column; justify-content: space-between; }
    .stat-top { display: flex; justify-content: space-between; align-items: start; color: var(--muted); font-weight: 700; }
    .stat-icon { width: 36px; height: 36px; border-radius: 8px; display: grid; place-items: center; color: var(--primary); background: var(--primary-soft); }
    .stat-value { margin-top: 10px; font-size: 30px; line-height: 1; font-weight: 700; }
    .muted { color: var(--muted); }
    .small { font-size: 12px; }
    .hidden { display: none !important; }
    .form-grid { display: grid; grid-template-columns: repeat(2, minmax(0,1fr)); gap: 10px; }
    .form-grid.single { grid-template-columns: 1fr; }
    label { display: grid; gap: 5px; color: #384150; font-weight: 700; font-size: 12px; }
    input, select, textarea {
      min-height: 36px;
      border-radius: 8px;
      border: 1px solid var(--outline);
      background: white;
      color: var(--ink);
      padding: 0 10px;
      outline: 0;
    }
    textarea { min-height: 78px; padding: 9px 10px; resize: vertical; }
    input:focus, select:focus, textarea:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(70,72,212,.12); }
    .unit-row {
      display: grid;
      grid-template-columns: minmax(0, 1fr) 92px;
      gap: 8px;
    }
    .unit-row input, .unit-row select { width: 100%; }
    .full { grid-column: 1 / -1; }
    .table-wrap { overflow-x: auto; }
    table { width: 100%; border-collapse: collapse; }
    th, td { padding: 11px 14px; border-bottom: 1px solid #e7ebf0; text-align: left; vertical-align: middle; }
    th { color: #4c5564; font-size: 12px; background: #fafbfc; white-space: nowrap; }
    tbody tr:hover { background: #fbfcfe; }
    code {
      max-width: 580px;
      display: inline-block;
      border-radius: 6px;
      background: var(--soft);
      padding: 3px 6px;
      color: #202636;
      overflow-wrap: anywhere;
      white-space: normal;
    }
    .diff-box {
      max-height: 340px;
      overflow: auto;
      margin: 0;
      border: 1px solid #d7dce5;
      border-radius: 8px;
      background: #111827;
      color: #e5e7eb;
      padding: 12px;
      font: 12px/1.45 ui-monospace, SFMono-Regular, Consolas, "Liberation Mono", monospace;
      white-space: pre;
    }
    .drift-summary {
      display: grid;
      grid-template-columns: repeat(5, minmax(0,1fr));
      gap: 10px;
    }
    .drift-stat {
      border: 1px solid #e3e7ee;
      border-radius: 8px;
      background: #fbfcfd;
      padding: 10px;
      display: grid;
      gap: 3px;
    }
    .drift-stat strong { font-size: 20px; }
    .identity { display: flex; align-items: center; gap: 10px; }
    .initials {
      width: 30px;
      height: 30px;
      border-radius: 8px;
      display: grid;
      place-items: center;
      background: var(--primary-soft);
      color: var(--primary);
      font-weight: 700;
      flex: 0 0 auto;
    }
    .badge {
      min-height: 24px;
      border-radius: 999px;
      display: inline-flex;
      align-items: center;
      padding: 0 9px;
      font-weight: 700;
      font-size: 12px;
      background: var(--muted-soft);
      color: #384150;
      white-space: nowrap;
    }
    .badge.good { background: var(--success-soft); color: var(--success); }
    .badge.warn { background: var(--warning-soft); color: var(--warning); }
    .badge.bad { background: var(--danger-soft); color: var(--danger); }
    .checkbox-line { display: flex; align-items: center; gap: 8px; font-weight: 700; }
    .checkbox-line input { min-height: auto; }
    .step-list { display: grid; gap: 10px; }
    .step {
      border: 1px solid var(--outline);
      border-radius: 8px;
      padding: 13px;
      display: grid;
      gap: 12px;
      background: white;
    }
    .step-title { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
    .step-title strong { font-size: 14px; }
    .quick-actions {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 10px;
    }
    .quick-action {
      min-height: 84px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: #fbfcfd;
      padding: 12px;
      display: grid;
      gap: 7px;
      cursor: pointer;
      text-align: left;
      color: inherit;
    }
    .quick-action:hover { background: var(--soft); }
    .quick-action strong { display: flex; align-items: center; gap: 8px; }
    .record-grid { display: grid; gap: 8px; }
    .dns-record {
      border: 1px solid #e3e7ee;
      border-radius: 8px;
      padding: 10px;
      display: grid;
      grid-template-columns: 86px minmax(180px,.45fr) minmax(280px,1fr);
      gap: 10px;
      align-items: start;
      background: #fbfcfd;
    }
    .dns-record .purpose { grid-column: 2 / -1; color: var(--muted); font-size: 12px; }
    .setup-grid { display: grid; grid-template-columns: repeat(2, minmax(0,1fr)); gap: 10px; margin-top: 14px; }
    .setup-item {
      border: 1px solid #e3e7ee;
      border-radius: 8px;
      padding: 12px;
      background: #fbfcfd;
      display: grid;
      gap: 8px;
    }
    .setup-item h4 { margin: 0; font-size: 14px; }
    .setup-line { display: grid; grid-template-columns: 110px minmax(0,1fr); gap: 8px; align-items: start; }
    .provision-list { display: grid; gap: 8px; }
    .provision-action {
      border: 1px solid #e3e7ee;
      border-radius: 8px;
      padding: 10px;
      display: grid;
      gap: 7px;
      background: #fbfcfd;
    }
    .provision-action.conflict, .provision-action.blocked { border-color: #f4c36a; background: var(--warning-soft); }
    .provision-action.ok { border-color: #a8e8c4; background: var(--success-soft); }
    .provision-action.create { border-color: #cfd5dd; }
    .provision-meta { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
    .progress { height: 8px; border-radius: 999px; background: var(--muted-soft); overflow: hidden; }
    .progress span { display: block; height: 100%; border-radius: inherit; background: var(--primary); }
    .audit-tabs { display: flex; gap: 8px; flex-wrap: wrap; padding: 12px 14px; border-bottom: 1px solid var(--outline); }
    .audit-tab { border: 1px solid var(--outline); background: #fff; border-radius: 8px; padding: 9px 12px; font-weight: 800; display: inline-flex; align-items: center; gap: 8px; cursor: pointer; }
    .audit-tab.active { border-color: var(--primary); color: var(--primary); box-shadow: inset 0 -2px 0 var(--primary); }
    .audit-toolbar { display: grid; grid-template-columns: minmax(240px,1fr) 220px 180px; gap: 10px; padding: 12px 14px; border-bottom: 1px solid var(--outline); }
    .audit-list { display: grid; gap: 10px; padding: 14px; }
    .audit-card { border: 1px solid #e3e7ee; border-radius: 8px; padding: 12px; display: grid; gap: 8px; background: #fff; }
    .audit-card-head { display: grid; grid-template-columns: minmax(0,1fr) auto; gap: 12px; align-items: start; }
    .audit-card h4 { margin: 0; font-size: 15px; }
    .audit-summary { color: var(--muted); line-height: 1.45; }
    .audit-meta { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
    .audit-details { display: flex; gap: 6px; flex-wrap: wrap; }
    .audit-detail { background: #f3f6fa; border: 1px solid #e3e7ee; border-radius: 8px; padding: 5px 7px; font-size: 12px; }
    .audit-detail strong { margin-right: 4px; }
    .toast {
      position: fixed;
      right: 22px;
      bottom: 22px;
      z-index: 80;
      max-width: 420px;
      border-radius: 8px;
      padding: 12px 14px;
      background: #1f2937;
      color: white;
      box-shadow: 0 14px 34px rgba(15,23,42,.24);
    }
    .toast.error { background: var(--danger); }
    .login-cover {
      position: fixed;
      inset: 0;
      z-index: 100;
      display: grid;
      place-items: center;
      padding: 20px;
      background: var(--background);
    }
    .login-cover.push-mode { background: var(--background); color: var(--ink); }
    .login-card { width: min(430px, 100%); background: white; border: 1px solid var(--outline); border-radius: 8px; box-shadow: var(--shadow); padding: 22px; }
    .login-card h2 { margin: 0; font-size: 22px; }
    .login-card p { margin: 6px 0 18px; color: var(--muted); }
    .mfa-panel {
      margin-top: 14px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      padding: 13px;
      display: grid;
      gap: 11px;
      background: #fbfcfd;
    }
    .mfa-panel .actions { align-items: stretch; }
    .mfa-panel .button { justify-content: center; }
    .mfa-status { margin: 0; color: var(--muted); font-size: 12px; line-height: 1.45; }
    .login-status { margin-top: 12px; border: 1px solid var(--outline); border-radius: 8px; padding: 10px 12px; background: #fbfcfd; }
    .login-status.error { border-color: #fecaca; background: #fff7f7; color: #991b1b; }
    .mfa-code-control { position: relative; }
    .mfa-code-control input { width: 100%; padding-right: 42px; }
    .mfa-code-control .material-symbols-outlined {
      position: absolute;
      right: 10px;
      top: 50%;
      transform: translateY(-50%);
      color: var(--primary);
      pointer-events: none;
    }
    .auth-wait {
      min-height: 260px;
      display: grid;
      place-items: center;
      text-align: center;
    }
    .auth-wait-inner { max-width: 430px; display: grid; gap: 14px; justify-items: center; }
    .auth-wait-icon {
      width: 64px;
      height: 64px;
      border-radius: 18px;
      display: grid;
      place-items: center;
      background: var(--primary-soft);
      color: var(--primary);
    }
    .auth-wait h3 { margin: 0; font-size: 22px; }
    .auth-wait p { margin: 0; color: var(--muted); line-height: 1.45; }
    .auth-wait-status { display: inline-flex; align-items: center; gap: 8px; color: #4d6180; }
    .auth-wait-status::before { content: ""; width: 8px; height: 8px; border-radius: 999px; background: var(--primary); }
    .push-auth-view {
      width: min(430px, 100%);
      display: grid;
      gap: 18px;
    }
    .push-auth-brand { width: 100%; display: flex; align-items: center; gap: 12px; padding: 0 6px; }
    .push-auth-brand .brand-mark {
      width: 42px;
      height: 42px;
      background: var(--primary);
      color: white;
      border: 0;
      box-shadow: 0 10px 24px rgba(70,72,212,.2);
    }
    .push-auth-brand h2 { margin: 0; font-size: 21px; color: var(--primary); line-height: 1; }
    .push-auth-brand p { margin: 4px 0 0; color: #2d3340; font-size: 11px; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; }
    .push-auth-card {
      width: 100%;
      border: 1px solid var(--outline);
      border-radius: 8px;
      padding: 22px;
      background: white;
      box-shadow: var(--shadow);
      display: grid;
      gap: 20px;
      text-align: center;
    }
    .push-auth-head {
      display: flex;
      align-items: center;
      gap: 8px;
      color: var(--ink);
      font-weight: 800;
      border-bottom: 1px solid var(--outline);
      padding-bottom: 10px;
      text-align: left;
    }
    .push-auth-head .material-symbols-outlined { color: var(--primary); }
    .push-phone {
      width: 62px;
      height: 62px;
      border-radius: 16px;
      margin: 12px auto 0;
      display: grid;
      place-items: center;
      color: var(--primary);
      background: var(--primary-soft);
      border: 1px solid rgba(70,72,212,.16);
    }
    .push-auth-card h3 { margin: 0; color: var(--ink); font-size: 16px; }
    .push-auth-card p { margin: -10px 0 0; color: var(--muted); line-height: 1.45; }
    .push-auth-status { display: inline-flex; align-items: center; justify-content: center; gap: 8px; color: #4d6180; }
    .push-auth-status::before { content: ""; width: 8px; height: 8px; border-radius: 999px; background: var(--primary); }
    .push-auth-link {
      border: 0;
      background: transparent;
      color: var(--primary);
      text-decoration: underline;
      cursor: pointer;
      font: inherit;
      padding: 0;
    }
    .push-auth-code {
      display: grid;
      gap: 9px;
      text-align: left;
    }
    .push-auth-code label { color: var(--ink); }
    .push-auth-code input { background: white; color: var(--ink); border-color: var(--outline); }
    .push-auth-actions { display: flex; gap: 9px; justify-content: center; flex-wrap: wrap; }
    .push-auth-actions .button { background: white; color: var(--ink); border-color: var(--outline); }
    .push-auth-actions .button.primary { background: var(--primary); border-color: var(--primary); color: #fff; }
    .modal-backdrop {
      position: fixed;
      inset: 0;
      z-index: 90;
      display: grid;
      place-items: center;
      padding: 20px;
      background: rgba(17, 24, 39, .42);
    }
    .modal-card {
      width: min(960px, 100%);
      max-height: min(820px, calc(100vh - 40px));
      overflow: auto;
      background: white;
      border: 1px solid var(--outline);
      border-radius: 8px;
      box-shadow: 0 22px 70px rgba(15,23,42,.28);
    }
    .modal-head {
      min-height: 58px;
      padding: 13px 16px;
      border-bottom: 1px solid var(--outline);
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
    }
    .modal-head h3 { margin: 0; font-size: 18px; }
    .modal-body { padding: 16px; }
    .modal-actions {
      grid-column: 1 / -1;
      display: flex;
      justify-content: end;
      gap: 9px;
      border-top: 1px solid var(--outline);
      padding-top: 12px;
    }
    @media (max-width: 1000px) {
      .app { padding-left: 0; }
      aside { position: static; width: auto; }
      header { position: static; height: auto; padding: 14px; align-items: stretch; flex-direction: column; }
      .search { width: 100%; }
      main { padding: 18px 14px 28px; }
      .hero { align-items: stretch; flex-direction: column; }
      .scope-bar { align-items: stretch; flex-direction: column; }
      .scope-main { align-items: stretch; }
      .scope-picker { min-width: 100%; }
      .scope-summary { justify-content: start; }
      .stats, .two-col, .three-col { grid-template-columns: 1fr; }
      .quick-actions { grid-template-columns: 1fr; }
      .form-grid { grid-template-columns: 1fr; }
      .dns-record { grid-template-columns: 1fr; }
      .dns-record .purpose { grid-column: auto; }
      .setup-grid { grid-template-columns: 1fr; }
      .setup-line { grid-template-columns: 1fr; }
      .audit-toolbar { grid-template-columns: 1fr; }
      .audit-card-head { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <div class="login-cover" id="login-cover">
    <form class="login-card" id="login-form">
      <div class="brand" style="margin:0 0 18px 0">
        <div class="brand-mark"><span class="material-symbols-outlined">alternate_email</span></div>
        <div><h1>ProIdentity</h1><p>MAIL ADMIN</p></div>
      </div>
      <h2>Admin login</h2>
      <p>Use the server admin account to manage tenants, domains, users, shared mailboxes, security policy, and quarantine.</p>
      <div class="form-grid single">
        <label>Username<input name="username" autocomplete="username" required></label>
        <label>Password<input name="password" type="password" autocomplete="current-password" required></label>
        <button class="button primary" type="submit"><span class="material-symbols-outlined">login</span>Login</button>
      </div>
      <p class="login-status hidden" id="login-status"></p>
      <div class="mfa-panel hidden" id="mfa-panel">
        <div class="step-title"><strong id="mfa-title">Two-factor verification</strong><span class="badge" id="mfa-provider-badge">required</span></div>
        <p class="mfa-status" id="mfa-detail">Complete the second factor to finish signing in.</p>
        <label id="mfa-code-row">Authenticator code<div class="mfa-code-control"><input id="mfa-code" inputmode="numeric" autocomplete="one-time-code" placeholder="123456"><span class="material-symbols-outlined">shield</span></div></label>
        <div class="actions">
          <button class="button primary" type="button" id="mfa-submit"><span class="material-symbols-outlined">verified</span>Verify code</button>
          <button class="button" type="button" id="mfa-webauthn"><span class="material-symbols-outlined">passkey</span>Use hardware key</button>
          <button class="button" type="button" id="mfa-push"><span class="material-symbols-outlined">phonelink_lock</span>Check push</button>
          <button class="button" type="button" id="mfa-cancel"><span class="material-symbols-outlined">close</span>Cancel</button>
        </div>
      </div>
    </form>
    <section class="push-auth-view hidden" id="push-mfa-view" aria-live="polite">
      <div class="push-auth-brand">
        <div class="brand-mark"><span class="material-symbols-outlined">alternate_email</span></div>
        <div><h2>ProIdentity Mail</h2><p>Admin verification</p></div>
      </div>
      <div class="push-auth-card">
        <div class="push-auth-head"><span class="material-symbols-outlined">lock</span><span>Push Verification</span></div>
        <div class="push-phone"><span class="material-symbols-outlined">phone_iphone</span></div>
        <h3>Waiting for approval...</h3>
        <p>Check your phone for a push notification.</p>
        <div class="push-auth-status" id="push-mfa-status">Waiting for approval...</div>
        <button class="push-auth-link" type="button" id="push-mfa-manual"><span class="material-symbols-outlined" style="font-size:16px;vertical-align:-3px">key</span> Enter code manually</button>
        <div class="push-auth-code hidden" id="push-mfa-code-panel">
          <label>Authenticator code<div class="mfa-code-control"><input id="push-mfa-code" inputmode="numeric" autocomplete="one-time-code" placeholder="123456"><span class="material-symbols-outlined">shield</span></div></label>
          <div class="push-auth-actions">
            <button class="button primary" type="button" id="push-mfa-submit"><span class="material-symbols-outlined">verified</span>Verify code</button>
            <button class="button" type="button" id="push-mfa-check"><span class="material-symbols-outlined">refresh</span>Check push</button>
          </div>
        </div>
        <button class="push-auth-link" type="button" id="push-mfa-cancel">Cancel and try another method</button>
      </div>
    </section>
  </div>
  <div class="app">
    <aside>
      <div class="brand">
        <div class="brand-mark"><span class="material-symbols-outlined">alternate_email</span></div>
        <div><h1>ProIdentity</h1><p>MAIL ADMIN</p></div>
      </div>
      <nav id="nav"></nav>
      <div class="sidebar-bottom">
        <button class="nav-item" id="reload-nav"><span class="material-symbols-outlined">sync</span><span>Reload Data</span></button>
        <button class="nav-item" id="logout-nav"><span class="material-symbols-outlined">logout</span><span>Logout</span></button>
      </div>
    </aside>
    <header>
      <div class="title-area">
        <h2 id="page-title">Dashboard</h2>
        <p id="page-subtitle">Operational overview</p>
      </div>
      <div class="search">
        <span class="material-symbols-outlined">search</span>
        <input id="search" placeholder="Search current view...">
      </div>
      <div class="top-actions">
        <span class="health-pill" id="health-pill"><span class="material-symbols-outlined">monitor_heart</span><span>Checking</span></span>
        <button class="button" id="reload-top"><span class="material-symbols-outlined">refresh</span>Refresh</button>
      </div>
    </header>
    <main>
      <section class="scope-bar" id="scope-bar"></section>
      <section class="hero">
        <div>
          <h3 id="hero-title">Mail platform control center</h3>
          <p id="hero-text">Start with onboarding, then manage tenants, domains, users, DNS, security, quarantine, and audit activity.</p>
        </div>
        <div class="actions">
          <button class="button primary" id="start-onboarding"><span class="material-symbols-outlined">rocket_launch</span>Start setup</button>
          <button class="button" id="copy-discovery"><span class="material-symbols-outlined">content_copy</span>Copy discovery URL</button>
        </div>
      </section>
      <section id="view"></section>
    </main>
  </div>
  <div class="toast hidden" id="toast"></div>
  <div class="modal-backdrop hidden" id="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">
    <section class="modal-card">
      <div class="modal-head">
        <h3 id="modal-title">Edit tenant</h3>
        <button class="button" type="button" data-close-modal><span class="material-symbols-outlined">close</span>Close</button>
      </div>
      <div class="modal-body" id="modal-body"></div>
    </section>
  </div>
  <script nonce="__PROIDENTITY_CSP_NONCE__">
    const views = [
      ["dashboard", "Dashboard", "Operational overview", "space_dashboard"],
      ["onboarding", "Onboarding", "Tenant to first user setup", "fact_check"],
      ["tenants", "Tenants", "Organizations and customer boundaries", "apartment"],
      ["domains", "Domains", "Hosted domains, DNS, aliases, and catch-all", "dns"],
      ["users", "Users", "People, shared mailboxes, quota, and permissions", "group"],
      ["security", "Security", "Tenant spam, malware, and TLS policy", "shield_lock"],
      ["quarantine", "Quarantine", "Held spam and malware messages", "gpp_maybe"],
      ["audit", "Audit", "Admin and security activity", "receipt_long"],
      ["system", "System", "Service health and mail server behavior", "settings"]
    ];
    const state = {
      tenants: [], domains: [], users: [], tenantAdmins: [], aliases: [], catchAll: [], sharedPermissions: [], quarantine: [], audit: [], policies: [], rateLimits: [], mailSettings: null, adminMFA: null, totpEnrollment: null, pendingMFA: null, webAuthnRegistration: null, stepUpPromise: null, stepUpReject: null,
      view: "dashboard", domainTab: "domains", userTab: "people", systemTab: "mail", auditTab: "admin", auditAction: "", auditSeverity: "", auditSearch: "", selectedTenantId: "", selectedDomainId: "", dns: null, dnsCloudflare: null, domainTLS: null, configDrift: null, configDriftLoading: false, configApplyStatus: "", csrf: "", query: "", health: "checking", language: "en"
    };
    const i18nCatalog = __PROIDENTITY_I18N_CATALOG__;
    const supportedLanguages = [
      ["bg", "Bulgarian", "Български"],
      ["hr", "Croatian", "Hrvatski"],
      ["cs", "Czech", "Čeština"],
      ["da", "Danish", "Dansk"],
      ["nl", "Dutch", "Nederlands"],
      ["en", "English", "English"],
      ["fi", "Finnish", "Suomi"],
      ["fr", "French", "Français"],
      ["de", "German", "Deutsch"],
      ["el", "Greek", "Ελληνικά"],
      ["hu", "Hungarian", "Magyar"],
      ["ga", "Irish", "Gaeilge"],
      ["it", "Italian", "Italiano"],
      ["lv", "Latvian", "Latviešu"],
      ["lt", "Lithuanian", "Lietuvių"],
      ["pl", "Polish", "Polski"],
      ["pt", "Portuguese", "Português"],
      ["ro", "Romanian", "Română"],
      ["sk", "Slovak", "Slovenčina"],
      ["sl", "Slovenian", "Slovenščina"],
      ["es", "Spanish", "Español"],
      ["sv", "Swedish", "Svenska"]
    ];

    const $ = selector => document.querySelector(selector);
    const esc = value => String(value ?? "").replace(/[&<>"']/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#039;"}[char]));
    const dateText = value => value ? new Date(value).toLocaleString() : "-";
    const byID = (items, id) => items.find(item => String(item.id) === String(id));
    const tenantName = id => byID(state.tenants, id)?.name || ("Tenant " + (id || "-"));
    const domainName = id => byID(state.domains, id)?.name || ("Domain " + (id || "-"));
    const emailAddress = user => (user.local_part || "") + "@" + domainName(user.primary_domain_id);
    const emailFor = user => esc(emailAddress(user));
    const initials = value => String(value || "?").split(/[\s.-]+/).filter(Boolean).slice(0,2).map(part => part[0].toUpperCase()).join("") || "?";
    const selected = (a, b) => String(a) === String(b) ? "selected" : "";
    const checked = value => value ? "checked" : "";
    const languageOptions = current => supportedLanguages.map(item => "<option value=\"" + esc(item[0]) + "\" " + selected(current || "en", item[0]) + ">" + esc(item[2] + " / " + item[1]) + "</option>").join("");
    const translationSkipSelector = "code, pre, input, textarea, .material-symbols-outlined, .record-grid, .provision-list";
    function t(value) {
      const key = String(value ?? "");
      return (i18nCatalog[state.language] && i18nCatalog[state.language][key]) || (i18nCatalog.en && i18nCatalog.en[key]) || key;
    }
    function applyLanguage(code) {
      const normalized = supportedLanguages.some(item => item[0] === code) ? code : "en";
      state.language = normalized;
      document.documentElement.lang = normalized;
      translateUI(document.body);
    }
    function translateTextNode(node) {
      if (!node.nodeValue || !node.nodeValue.trim()) return;
      const parent = node.parentElement;
      if (!parent || parent.closest(translationSkipSelector)) return;
      const current = node.nodeValue.trim();
      const previous = node.__i18nOriginal || "";
      const original = (i18nCatalog.en && i18nCatalog.en[current] && current !== t(previous)) ? current : (previous || current);
      node.__i18nOriginal = original;
      const translated = t(original);
      node.nodeValue = node.nodeValue.replace(current, translated);
    }
    function translateAttributes(element) {
      ["placeholder", "title", "aria-label", "data-placeholder"].forEach(attr => {
        if (!element.hasAttribute(attr)) return;
        const dataName = "i18nOriginal" + attr.replace(/(^|-)([a-z])/g, (_, __, letter) => letter.toUpperCase());
        const current = element.getAttribute(attr);
        const previous = element.dataset[dataName] || "";
        const original = (i18nCatalog.en && i18nCatalog.en[current] && current !== t(previous)) ? current : (previous || current);
        element.dataset[dataName] = original;
        element.setAttribute(attr, t(original));
      });
    }
    function translateUI(root = document.body) {
      if (!root) return;
      const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {acceptNode: node => node.parentElement && !node.parentElement.closest(translationSkipSelector) ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_REJECT});
      const nodes = [];
      while (walker.nextNode()) nodes.push(walker.currentNode);
      nodes.forEach(translateTextNode);
      if (root.nodeType === Node.ELEMENT_NODE) translateAttributes(root);
      root.querySelectorAll?.("[placeholder], [title], [aria-label], [data-placeholder]").forEach(translateAttributes);
    }
    const visible = items => {
      const q = state.query.trim().toLowerCase();
      if (!q) return items;
      return items.filter(item => JSON.stringify(item).toLowerCase().includes(q));
    };
    const domainsForTenant = (tenantID = state.selectedTenantId) => state.domains.filter(item => !tenantID || String(item.tenant_id) === String(tenantID));
    const selectedTenantDomains = () => domainsForTenant();
    const selectedTenant = () => byID(state.tenants, state.selectedTenantId);
    const selectedDomain = () => byID(state.domains, state.selectedDomainId);
    const scopeTenantLabel = () => state.selectedTenantId ? tenantName(state.selectedTenantId) : "All tenants";
    const scopeDomainLabel = () => state.selectedDomainId ? domainName(state.selectedDomainId) : "All domains";
    const setupIncomplete = () => !state.tenants.length || !state.domains.length || !state.users.length;
    const activeViews = () => setupIncomplete() ? views : views.filter(item => item[0] !== "onboarding");
    const scopedUsers = () => state.users.filter(item =>
      (!state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId)) &&
      (!state.selectedDomainId || String(item.primary_domain_id) === String(state.selectedDomainId))
    );
    const matchesScopeTenant = item => !state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId);
    const matchesScopeDomain = domainID => !state.selectedDomainId || String(domainID) === String(state.selectedDomainId);
    const filteredUsers = () => visible(state.users.filter(item =>
      (!state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId)) &&
      (!state.selectedDomainId || String(item.primary_domain_id) === String(state.selectedDomainId))
    ));
    const usersInTenant = () => state.users.filter(item => !state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId));
    const sharedMailboxes = () => usersInTenant().filter(item => (item.mailbox_type || "user") === "shared");
    const normalUsers = () => usersInTenant().filter(item => (item.mailbox_type || "user") === "user");
    const filteredNormalUsers = () => visible(normalUsers().filter(item => !state.selectedDomainId || String(item.primary_domain_id) === String(state.selectedDomainId)));
    const filteredSharedMailboxes = () => visible(sharedMailboxes().filter(item => !state.selectedDomainId || String(item.primary_domain_id) === String(state.selectedDomainId)));
    const auditTabs = [
      ["admin", "Admin & access", ["admin", "auth"]],
      ["mail", "Mail security", ["mail_security"]],
      ["user", "User activity", ["user_activity"]],
      ["system", "System & DNS", ["system"]],
      ["all", "All", []]
    ];
    const auditTabCategories = key => (auditTabs.find(item => item[0] === key) || auditTabs[0])[2];
    const auditCategoryLabel = category => ({
      admin: "Admin",
      auth: "Access",
      security: "Security",
      mail_security: "Mail security",
      user_activity: "User activity",
      system: "System"
    }[category] || "Audit");
    function badge(value) {
      const text = String(value || "unknown");
      const low = text.toLowerCase();
      const cls = /active|ok|good|released|mark|checked|configured/.test(low) ? "good" : /held|pending|applying|working|quarantine|spam|warn|conflict|changes|blocked/.test(low) ? "warn" : /reject|delete|malware|failed|bad|error/.test(low) ? "bad" : "";
      return "<span class=\"badge " + cls + "\">" + esc(text) + "</span>";
    }
    function showStatus(message, isError) {
      const toast = $("#toast");
      toast.textContent = t(message);
      toast.className = "toast" + (isError ? " error" : "");
      clearTimeout(showStatus.timer);
      showStatus.timer = setTimeout(() => toast.classList.add("hidden"), 3800);
    }
    async function api(path, options) {
      const init = Object.assign({credentials: "same-origin", cache: "no-store", headers: {}}, options || {});
      const retryStepUp = init.retryStepUp !== false;
      delete init.retryStepUp;
      if (init.body && !init.headers["Content-Type"]) init.headers["Content-Type"] = "application/json";
      if (state.csrf && init.method && init.method !== "GET") init.headers["X-CSRF-Token"] = state.csrf;
      const response = await fetch(path, init);
      if (response.status === 204) return null;
      const data = await response.json().catch(() => ({}));
      if (response.status === 428 && retryStepUp) {
        await ensureAdminStepUp();
        return api(path, Object.assign({}, options || {}, {retryStepUp: false}));
      }
      if (!response.ok) throw new Error(data.error || response.statusText || "Request failed");
      return data;
    }
    async function ensureAdminStepUp() {
      if (state.stepUpPromise) return state.stepUpPromise;
      state.stepUpPromise = (async () => {
        const challenge = await api("/api/v1/session/step-up", {method: "POST", body: JSON.stringify({}), retryStepUp: false});
        await showAdminStepUpModal(challenge);
      })();
      try {
        await state.stepUpPromise;
      } finally {
        state.stepUpPromise = null;
      }
    }
    function showAdminStepUpModal(challenge) {
      return new Promise((resolve, reject) => {
        state.stepUpReject = reject;
        const isPush = challenge.provider === "proidentity";
        const isWebAuthn = challenge.provider === "webauthn";
        const title = isWebAuthn ? "Use your hardware key" : (isPush ? "Check your phone" : "Two-factor verification");
        const detail = isWebAuthn ? "Touch your passkey or security key to continue this admin action." : (isPush ? "Approve the request in ProIdentity Auth, or enter a hosted TOTP code." : "Enter your authenticator code to continue this admin action.");
        const codeRow = isWebAuthn ? "" : "<label>Authenticator code<div class=\"mfa-code-control\"><input id=\"step-up-code\" inputmode=\"numeric\" autocomplete=\"one-time-code\" placeholder=\"123456\"><span class=\"material-symbols-outlined\">shield</span></div></label>";
        const icon = isWebAuthn ? "passkey" : (isPush ? "phone_iphone" : "verified");
        const buttonText = isWebAuthn ? "Use hardware key" : (isPush ? "Check approval" : "Verify code");
        openModal("Confirm admin action",
          "<div class=\"step-list\"><div class=\"step\"><div class=\"step-title\"><div><h4>" + esc(title) + "</h4><p class=\"muted small\">" + esc(detail) + "</p></div>" + badge(challenge.provider || "mfa") + "</div>" +
          codeRow +
          "<p class=\"mfa-status\" id=\"step-up-status\">" + esc(isPush ? "Waiting for approval..." : "Fresh verification is required for dangerous changes.") + "</p>" +
          "<div class=\"modal-actions\"><button class=\"button\" type=\"button\" id=\"step-up-cancel\"><span class=\"material-symbols-outlined\">close</span>Cancel</button><button class=\"button primary\" type=\"button\" id=\"step-up-verify\"><span class=\"material-symbols-outlined\">" + esc(icon) + "</span>" + esc(buttonText) + "</button></div></div></div>");
        const status = $("#step-up-status");
        const finish = async () => {
          try {
            let result;
            if (isWebAuthn) {
              result = await runWebAuthnStepUp(challenge, status);
            } else {
              const code = $("#step-up-code")?.value || "";
              result = await api("/api/v1/session/step-up/verify", {method: "POST", body: JSON.stringify({mfa_token: challenge.mfa_token, code}), retryStepUp: false});
            }
            if (result && result.status === "pending") {
              status.textContent = t("Still waiting for approval...");
              return;
            }
            state.stepUpReject = null;
            closeModal();
            resolve();
          } catch (error) {
            if (String(error.message || "").includes("pending")) {
              status.textContent = t("Still waiting for approval...");
              return;
            }
            status.textContent = error.message;
            status.classList.add("error");
          }
        };
        $("#step-up-verify")?.addEventListener("click", finish);
        $("#step-up-cancel")?.addEventListener("click", () => {
          state.stepUpReject = null;
          closeModal();
          reject(new Error("Admin step-up was cancelled"));
        });
        $("#step-up-code")?.addEventListener("keydown", event => {
          if (event.key === "Enter") finish();
        });
        if (isWebAuthn) finish();
      });
    }
    function syncScope() {
      if (state.selectedTenantId && !state.tenants.some(item => String(item.id) === String(state.selectedTenantId))) {
        state.selectedTenantId = "";
      }
      const domains = domainsForTenant();
      if (state.selectedDomainId && !domains.some(item => String(item.id) === String(state.selectedDomainId))) {
        state.selectedDomainId = "";
      }
      if (state.dns && String(state.dns.domain_id) !== String(state.selectedDomainId)) {
        state.dns = null;
        state.dnsCloudflare = null;
      }
      if (state.domainTLS && String(state.domainTLS.domain_id) !== String(state.selectedDomainId)) state.domainTLS = null;
    }
    function setTenantScope(tenantID) {
      state.selectedTenantId = String(tenantID || "");
      if (!state.selectedTenantId || !domainsForTenant().some(item => String(item.id) === String(state.selectedDomainId))) state.selectedDomainId = "";
      state.dns = null;
      state.dnsCloudflare = null;
      state.domainTLS = null;
      state.query = "";
      $("#search").value = "";
      render();
    }
    function setDomainScope(domainID) {
      state.selectedDomainId = String(domainID || "");
      const domain = selectedDomain();
      if (domain && state.selectedTenantId && String(domain.tenant_id) !== String(state.selectedTenantId)) state.selectedTenantId = String(domain.tenant_id);
      state.dns = null;
      state.dnsCloudflare = null;
      state.domainTLS = null;
      state.query = "";
      $("#search").value = "";
      render();
    }
    async function bootstrapSession() {
      const response = await fetch("/api/v1/session", {credentials: "same-origin", cache: "no-store"});
      if (!response.ok) {
        $("#login-cover").classList.remove("hidden");
        return;
      }
      const data = await response.json();
      state.csrf = data.csrf_token || "";
      $("#login-cover").classList.add("hidden");
      await refresh();
    }
    async function refresh() {
      const [tenants, domains, users, tenantAdmins, aliases, catchAll, sharedPermissions, quarantine, audit, policies, rateLimits, mailSettings, adminMFA] = await Promise.all([
        api("/api/v1/tenants"), api("/api/v1/domains"), api("/api/v1/users"),
        api("/api/v1/tenant-admins"), api("/api/v1/aliases"), api("/api/v1/catch-all"), api("/api/v1/shared-permissions"),
        api("/api/v1/quarantine"), api("/api/v1/audit"), api("/api/v1/policies"), api("/api/v1/security/login-rate-limits"), api("/api/v1/mail-server-settings"), api("/api/v1/admin-mfa/settings")
      ]);
      state.tenants = tenants || [];
      state.domains = domains || [];
      state.users = users || [];
      state.tenantAdmins = tenantAdmins || [];
      state.aliases = aliases || [];
      state.catchAll = catchAll || [];
      state.sharedPermissions = sharedPermissions || [];
      state.quarantine = quarantine || [];
      state.audit = audit || [];
      state.policies = policies || [];
      state.rateLimits = rateLimits || [];
      state.mailSettings = mailSettings || null;
      state.adminMFA = adminMFA || null;
      applyLanguage(state.mailSettings?.default_language || "en");
      syncScope();
      if (setupIncomplete() && state.view === "dashboard") state.view = "onboarding";
      if (!setupIncomplete() && state.view === "onboarding") state.view = "dashboard";
      render();
      checkHealth();
    }
    async function checkHealth() {
      try {
        const data = await api("/healthz");
        state.health = data.status || "ok";
      } catch (error) {
        state.health = "failed";
      }
      $("#health-pill").innerHTML = "<span class=\"material-symbols-outlined\">monitor_heart</span><span>" + esc(state.health) + "</span>";
    }
    async function loadConfigDrift() {
      state.configDriftLoading = true;
      state.configApplyStatus = "";
      render();
      try {
        state.configDrift = await api("/api/v1/system/config-drift");
      } finally {
        state.configDriftLoading = false;
        render();
      }
    }
    async function requestConfigApply() {
      if (!confirm(t("Reload the whole live mail configuration from the database? This will apply Postfix, Dovecot, Rspamd, Nginx, and certificate helper files."))) return;
      state.configApplyStatus = "Reload request queued. The root apply service will sync live files and restart/reload affected services.";
      render();
      await api("/api/v1/system/config-apply", {method: "POST", body: JSON.stringify({})});
      showStatus("Configuration reload queued");
      setTimeout(() => loadConfigDrift().catch(error => showStatus(error.message, true)), 5000);
    }
    function setView(view) {
      state.view = view === "onboarding" && !setupIncomplete() ? "dashboard" : view;
      state.query = "";
      $("#search").value = "";
      render();
    }
    function renderNav() {
      $("#nav").innerHTML = activeViews().map(([id, label, , icon]) =>
        "<button class=\"nav-item " + (state.view === id ? "active" : "") + "\" data-view=\"" + id + "\"><span class=\"material-symbols-outlined\">" + icon + "</span><span>" + label + "</span></button>"
      ).join("");
    }
    function tabs(scope, items, current) {
      return "<div class=\"tabs\">" + items.map(item => "<button class=\"tab " + (item[0] === current ? "active" : "") + "\" data-tab-scope=\"" + esc(scope) + "\" data-tab=\"" + esc(item[0]) + "\"><span class=\"material-symbols-outlined\">" + esc(item[2]) + "</span>" + esc(item[1]) + "</button>").join("") + "</div>";
    }
    function tableCard(title, subtitle, headings, rows, emptyText, actionsHTML = "") {
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>" + esc(title) + "</h4><p>" + esc(subtitle) + "</p></div>" + actionsHTML + "</div>" + table(headings, rows, emptyText) + "</section>";
    }
    function tenantDomainActions(includeAllDomain = false) {
      const tenant = selectedTenant();
      const domain = selectedDomain();
      const badges = ["Tenant: " + (tenant ? tenant.name : "All tenants")];
      if (includeAllDomain || domain || !state.selectedDomainId) badges.push("Domain: " + (domain ? domain.name : "All domains"));
      return "<div class=\"actions\">" + badges.map(item => "<span class=\"badge\">" + esc(item) + "</span>").join("") + "</div>";
    }
    function renderScopeBar() {
      const bar = $("#scope-bar");
      const tenant = selectedTenant();
      const domain = selectedDomain();
      if (!state.tenants.length) {
        bar.innerHTML = "<div class=\"scope-main\"><div><strong>No tenant selected</strong><div class=\"muted small\">Create the first tenant to unlock domains, users, and routing.</div></div></div><div class=\"scope-summary\"><button class=\"button primary\" data-view=\"onboarding\"><span class=\"material-symbols-outlined\">rocket_launch</span>Start setup</button></div>";
        return;
      }
      const domains = domainsForTenant();
      const domainSelect = domains.length ? "<label class=\"scope-picker\">Domain<select id=\"global-domain\">" + domainOptions(state.selectedTenantId, state.selectedDomainId, true) + "</select></label>" : "<div class=\"scope-empty\">No domains in this scope yet</div>";
      const users = scopedUsers().length;
      const shared = scopedUsers().filter(item => (item.mailbox_type || "user") === "shared").length;
      const aliases = state.aliases.filter(item => matchesScopeTenant(item) && matchesScopeDomain(item.domain_id)).length;
      const catchAll = state.catchAll.filter(item => matchesScopeTenant(item) && matchesScopeDomain(item.domain_id)).length;
      bar.innerHTML = "<div class=\"scope-main\"><label class=\"scope-picker\">Tenant<select id=\"global-tenant\">" + tenantOptions(state.selectedTenantId, true) + "</select></label>" + domainSelect + "</div><div class=\"scope-summary\">" +
        "<span class=\"badge\">Users " + esc(users) + "</span><span class=\"badge\">Shared " + esc(shared) + "</span><span class=\"badge\">Aliases " + esc(aliases) + "</span><span class=\"badge " + (catchAll ? "good" : "") + "\">Catch-all " + esc(catchAll) + "</span>" +
        (domains.length ? "" : "<button class=\"button\" data-view=\"domains\"><span class=\"material-symbols-outlined\">add_link</span>Add domain</button>") + "</div>";
    }
    function render() {
      syncScope();
      if (!activeViews().some(item => item[0] === state.view)) state.view = setupIncomplete() ? "onboarding" : "dashboard";
      renderNav();
      renderScopeBar();
      const meta = views.find(item => item[0] === state.view) || views[0];
      $("#page-title").textContent = meta[1];
      $("#page-subtitle").textContent = meta[2];
      $("#hero-title").textContent = meta[1] === "Dashboard" ? "Mail platform control center" : meta[1];
      $("#hero-text").textContent = meta[2];
      $("#start-onboarding").classList.toggle("hidden", !setupIncomplete());
      const map = {dashboard: renderDashboard, onboarding: renderOnboarding, tenants: renderTenants, domains: renderDomains, users: renderUsers, security: renderSecurity, quarantine: renderQuarantine, audit: renderAudit, system: renderSystem};
      $("#view").innerHTML = (map[state.view] || renderDashboard)();
      translateUI(document.body);
    }
    function stat(label, value, icon, cls) {
      return "<article class=\"card stat\"><div class=\"stat-top\"><span>" + label + "</span><span class=\"stat-icon " + (cls || "") + "\"><span class=\"material-symbols-outlined\">" + icon + "</span></span></div><div><div class=\"stat-value\">" + esc(value) + "</div><div class=\"muted small\">Current live count</div></div></article>";
    }
    function renderDashboard() {
      const tenantID = state.selectedTenantId;
      const domainID = state.selectedDomainId;
      const held = state.quarantine.filter(item => (!tenantID || String(item.tenant_id) === String(tenantID)) && (!item.status || item.status === "held")).length;
      const domainCount = domainsForTenant(tenantID).length;
      const userCount = scopedUsers().length;
      const incomplete = setupIncomplete();
      const tasks = [
        ["Create tenant", state.tenants.length > 0],
        ["Add hosted domain", state.domains.length > 0],
        ["Review DNS records", !!state.dns],
        ["Create first user", state.users.length > 0],
        ["Confirm security policy", state.policies.length > 0]
      ];
      const scopeCard = incomplete ? "<section class=\"card\"><div class=\"panel-head\"><div><h4>Setup path</h4><p>The normal production order for the first customer or site.</p></div><button class=\"button primary\" data-view=\"onboarding\"><span class=\"material-symbols-outlined\">arrow_forward</span>Open onboarding</button></div><div class=\"card-body step-list\">" +
        tasks.map(task => "<div class=\"step-title\"><span>" + esc(task[0]) + "</span>" + badge(task[1] ? "ready" : "needed") + "</div>").join("") +
        "</div></section>" : "<section class=\"card\"><div class=\"panel-head\"><div><h4>Current scope</h4><p>Work inside the tenant and domain selected above.</p></div><button class=\"button\" data-view=\"domains\"><span class=\"material-symbols-outlined\">dns</span>Open domain</button></div><div class=\"card-body step-list\">" +
        "<div class=\"step-title\"><span>Tenant</span><strong>" + esc(scopeTenantLabel()) + "</strong></div><div class=\"step-title\"><span>Domain</span><strong>" + esc(scopeDomainLabel()) + "</strong></div><div class=\"step-title\"><span>Users in scope</span>" + badge(userCount) + "</div></div></section>";
      return (incomplete ? "<section class=\"card\" style=\"margin-bottom:16px\"><div class=\"panel-head\"><div><h4>Setup needs attention</h4><p>Create the first tenant, domain, and user before using advanced mail routing.</p></div><button class=\"button primary\" data-view=\"onboarding\"><span class=\"material-symbols-outlined\">rocket_launch</span>Continue setup</button></div></section>" : "") +
        "<div class=\"grid stats\">" +
        stat("Tenants", state.tenants.length, "apartment") + stat("Domains", domainCount, "dns") +
        stat("Users", userCount, "group") + stat("Held messages", held, "gpp_maybe") +
        "</div><section class=\"card\" style=\"margin-bottom:16px\"><div class=\"panel-head\"><div><h4>Common work</h4><p>Jump straight into the task without hunting through a long page.</p></div></div><div class=\"card-body quick-actions\">" +
        "<button class=\"quick-action\" data-create=\"user\"><strong><span class=\"material-symbols-outlined\">person_add</span>Create user</strong><span class=\"muted small\">Normal mailbox with quota and login.</span></button>" +
        "<button class=\"quick-action\" data-create=\"alias\"><strong><span class=\"material-symbols-outlined\">alternate_email</span>Add alias</strong><span class=\"muted small\">Alternate address under a tenant domain.</span></button>" +
        "<button class=\"quick-action\" data-create=\"catch-all\"><strong><span class=\"material-symbols-outlined\">all_inbox</span>Set catch-all</strong><span class=\"muted small\">Unknown recipients route to one mailbox.</span></button>" +
        "</div></section><div class=\"grid two-col\">" + scopeCard + "<section class=\"card\"><div class=\"panel-head\"><div><h4>Recent audit</h4><p>Latest admin/security actions.</p></div><button class=\"button\" data-view=\"audit\"><span class=\"material-symbols-outlined\">receipt_long</span>View audit</button></div><div class=\"table-wrap\"><table><tbody>" +
        visible(state.audit).slice(0,6).map(auditRow).join("") + emptyRows(state.audit, "No audit events yet.") +
        "</tbody></table></div></section></div>";
    }
    function renderOnboarding() {
      const selectedTenant = state.selectedTenantId || (state.tenants[0] && String(state.tenants[0].id)) || "";
      const domains = state.domains.filter(d => !selectedTenant || String(d.tenant_id) === String(selectedTenant));
      const selectedDomain = state.selectedDomainId || (domains[0] && String(domains[0].id)) || "";
      return "<div class=\"grid two-col\"><section class=\"card\"><div class=\"panel-head\"><div><h4>Guided setup</h4><p>Create in the order mail actually needs: tenant, domain, DNS, user.</p></div></div><div class=\"card-body step-list\">" +
        tenantStep() + domainStep(selectedTenant) + dnsStep(selectedDomain) + userStep(selectedTenant, selectedDomain) +
        "</div></section><section class=\"card\"><div class=\"panel-head\"><div><h4>Current scope</h4><p>Change tenant or domain once in the global bar above.</p></div></div><div class=\"card-body step-list\">" +
        "<div class=\"step\"><strong>" + esc(tenantName(selectedTenant)) + "</strong><span class=\"muted\">" + esc(domainName(selectedDomain)) + "</span><span class=\"muted small\">" + state.users.filter(u => String(u.tenant_id) === String(selectedTenant)).length + " users/shared mailboxes in this tenant</span></div>" +
        "</div></section></div>";
    }
    function hiddenInput(name, value) {
      return "<input type=\"hidden\" name=\"" + esc(name) + "\" value=\"" + esc(value || "") + "\">";
    }
    function contextField(label, value, name, id) {
      return hiddenInput(name, id) + "<label>" + esc(label) + "<span class=\"context-field\">" + esc(value || "Not selected") + "</span></label>";
    }
    function tenantField(tenantID) {
      return tenantID ? contextField("Tenant", tenantName(tenantID), "tenant_id", tenantID) : "<label>Tenant<select name=\"tenant_id\" required>" + tenantOptions(tenantID, false) + "</select></label>";
    }
    function domainField(tenantID, domainID, name = "primary_domain_id") {
      return domainID ? contextField("Domain", domainName(domainID), name, domainID) : "<label>Domain<select name=\"" + esc(name) + "\" required>" + domainOptions(tenantID, domainID, false) + "</select></label>";
    }
    function typeField(mailboxType) {
      return hiddenInput("mailbox_type", mailboxType) + "<label>Type<span class=\"context-field\">" + esc(mailboxType === "shared" ? "Shared mailbox" : "User mailbox") + "</span></label>";
    }
    function quotaParts(bytes) {
      const value = Number(bytes || 10737418240);
      if (value >= 1073741824 && value % 1073741824 === 0) return {value: value / 1073741824, unit: "gb"};
      return {value: Math.max(1, Math.round(value / 1048576)), unit: "mb"};
    }
    function quotaField(bytes) {
      const parts = quotaParts(bytes);
      return "<label>Mailbox storage quota<div class=\"unit-row\"><input name=\"quota_value\" type=\"number\" min=\"1\" step=\"1\" value=\"" + esc(parts.value) + "\"><select name=\"quota_unit\"><option value=\"gb\" " + selected(parts.unit, "gb") + ">GB</option><option value=\"mb\" " + selected(parts.unit, "mb") + ">MB</option></select></div></label>";
    }
    function statusOptions(current, values) {
      return values.map(value => "<option value=\"" + esc(value) + "\" " + selected(current, value) + ">" + esc(value) + "</option>").join("");
    }
    function resourceActions(type, id, label) {
      return "<div class=\"actions\"><button class=\"button\" data-edit=\"" + esc(type) + "\" data-id=\"" + esc(id) + "\"><span class=\"material-symbols-outlined\">edit</span>Edit</button><button class=\"button danger\" data-delete=\"" + esc(type) + "\" data-id=\"" + esc(id) + "\" data-label=\"" + esc(label) + "\"><span class=\"material-symbols-outlined\">delete</span>Remove</button></div>";
    }
    function createAction(type, label, icon, primary = false) {
      return "<button class=\"button " + (primary ? "primary" : "") + "\" data-create=\"" + esc(type) + "\"><span class=\"material-symbols-outlined\">" + esc(icon) + "</span>" + esc(label) + "</button>";
    }
    function tenantStep() {
      return "<div class=\"step\"><div class=\"step-title\"><strong>1. Tenant</strong>" + badge(state.tenants.length ? "ready" : "needed") + "</div><form class=\"form-grid\" data-form=\"tenant\">" +
        "<label>Name<input name=\"name\" placeholder=\"Example Company\" required></label><label>Slug<input name=\"slug\" placeholder=\"example-company\" required></label>" +
        "<button class=\"button primary full\" type=\"submit\"><span class=\"material-symbols-outlined\">add_business</span>Create tenant</button></form></div>";
    }
    function domainStep(tenantID) {
      return "<div class=\"step\"><div class=\"step-title\"><strong>2. Domain</strong>" + badge(state.domains.length ? "ready" : "needed") + "</div><form class=\"form-grid\" data-form=\"domain\">" +
        tenantField(tenantID) + "<label>Domain<input name=\"name\" placeholder=\"example.com\" required></label>" +
        "<button class=\"button primary full\" type=\"submit\"><span class=\"material-symbols-outlined\">add_link</span>Add domain</button></form></div>";
    }
    function dnsStep(domainID) {
      const records = state.dns && String(state.dns.domain_id) === String(domainID) ? state.dns.records : [];
      return "<div class=\"step\"><div class=\"step-title\"><strong>3. DNS records</strong>" + badge(records.length ? "loaded" : "load records") + "</div>" +
        "<div class=\"actions\"><span class=\"badge\">Domain: " + esc(domainName(domainID)) + "</span><button class=\"button\" data-load-dns=\"selected\" " + (domainID ? "" : "disabled") + "><span class=\"material-symbols-outlined\">dns</span>Load DNS</button></div>" +
        "<div class=\"record-grid\">" + renderDNSRecords(records) + "</div></div>";
    }
    function userStep(tenantID, domainID, mailboxType = "user") {
      return "<div class=\"step\"><div class=\"step-title\"><strong>4. User</strong>" + badge(state.users.length ? "ready" : "needed") + "</div><form class=\"form-grid\" data-form=\"user\">" +
        tenantField(tenantID) + domainField(tenantID, domainID) + typeField(mailboxType) + quotaField() +
        "<label>Local part<input name=\"local_part\" placeholder=\"marko\" required></label><label>Display name<input name=\"display_name\" placeholder=\"Marko Admin\"></label>" +
        (mailboxType === "shared" ? "" : "<label class=\"full\">Password<input name=\"password\" type=\"password\" autocomplete=\"new-password\" required></label>") +
        "<button class=\"button primary full\" type=\"submit\"><span class=\"material-symbols-outlined\">person_add</span>Create " + esc(mailboxType === "shared" ? "shared mailbox" : "user") + "</button></form></div>";
    }
    function renderTenants() {
      const actions = setupIncomplete() ? "<button class=\"button\" data-view=\"onboarding\"><span class=\"material-symbols-outlined\">add</span>Guided create</button>" : "";
      return "<div class=\"grid two-col\"><section class=\"card\"><div class=\"panel-head\"><div><h4>Tenants</h4><p>Each tenant is an isolated organization boundary.</p></div>" + actions + "</div>" +
        table(["Tenant", "Slug", "Status", "Created", "Actions"], visible(state.tenants).map(item => "<tr><td><div class=\"identity\"><span class=\"initials\">" + esc(initials(item.name)) + "</span><div><strong>" + esc(item.name) + "</strong><div class=\"muted small\">ID " + esc(item.id) + "</div></div></div></td><td><code>" + esc(item.slug) + "</code></td><td>" + badge(item.status) + "</td><td>" + esc(dateText(item.created_at)) + "</td><td><div class=\"actions\"><button class=\"button\" data-select-tenant=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">check_circle</span>Select</button>" + resourceActions("tenant", item.id, item.name) + "</div></td></tr>"), "No tenants match this view.") +
        "</section><section class=\"card\"><div class=\"panel-head\"><div><h4>Create tenant</h4><p>First step for every customer, site, or organization.</p></div></div><div class=\"card-body\">" + tenantStep() + "</div></section></div>";
    }
    function renderDomains() {
      const tenantID = state.selectedTenantId;
      const domainID = state.selectedDomainId;
      const domains = visible(domainsForTenant(tenantID));
      const aliasRows = state.aliases.filter(item => matchesScopeTenant(item) && matchesScopeDomain(item.domain_id));
      const catchRows = state.catchAll.filter(item => matchesScopeTenant(item) && matchesScopeDomain(item.domain_id));
      const tabBar = tabs("domain", [["domains", "Domains & DNS", "dns"], ["aliases", "Aliases", "alternate_email"], ["catchall", "Catch-all", "all_inbox"], ["certificates", "Certificates", "encrypted"]], state.domainTab);
      if (!state.tenants.length) {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Choose tenant first</h4><p>Create or select a tenant before adding domains.</p></div><button class=\"button primary\" data-view=\"onboarding\"><span class=\"material-symbols-outlined\">rocket_launch</span>Start setup</button></div></section>";
      }
      if (state.domainTab === "aliases") {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Domain aliases</h4><p>Aliases follow the current tenant/domain scope; choose one domain when creating a new alias.</p></div>" + tenantDomainActions(true) + "</div>" + tabBar +
          "<div class=\"card-body actions\">" + createAction("alias", "Add alias", "alternate_email", true) + "</div>" +
          table(["Alias", "Destination", "Tenant", "Created", "Actions"], visible(aliasRows).map(item => "<tr><td><strong>" + esc(item.source_local_part + "@" + domainName(item.domain_id)) + "</strong></td><td>" + esc(item.destination) + "</td><td>" + esc(tenantName(item.tenant_id)) + "</td><td>" + esc(dateText(item.created_at)) + "</td><td>" + resourceActions("alias", item.id, item.source_local_part + "@" + domainName(item.domain_id)) + "</td></tr>"), "No aliases in this scope.") + "</section>";
      }
      if (state.domainTab === "catchall") {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Catch-all mailbox</h4><p>Catch-all routes follow the current tenant/domain scope; choose one domain when changing routing.</p></div>" + tenantDomainActions(true) + "</div>" + tabBar +
          "<div class=\"card-body actions\">" + createAction("catch-all", "Set catch-all", "all_inbox", true) + "</div>" +
          table(["Domain", "Catch-all mailbox", "Tenant", "Status", "Actions"], visible(catchRows).map(item => "<tr><td>" + esc(domainName(item.domain_id)) + "</td><td>" + esc(item.destination) + "</td><td>" + esc(tenantName(item.tenant_id)) + "</td><td>" + badge(item.status) + "</td><td>" + resourceActions("catch-all", item.id, domainName(item.domain_id) + " catch-all") + "</td></tr>"), "No catch-all mailbox configured in this scope.") + "</section>";
      }
      if (state.domainTab === "certificates") {
        return renderDomainCertificates(tabBar);
      }
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Domains in " + esc(scopeTenantLabel()) + "</h4><p>Select one domain, or use All domains to review the whole scope.</p></div><div class=\"actions\">" + createAction("domain", "Add domain", "add_link", true) + "</div></div>" + tabBar +
        table(["Domain", "Tenant", "Status", "DKIM", "Actions"], domains.map(item => "<tr><td><strong>" + esc(item.name) + "</strong><div class=\"muted small\">ID " + esc(item.id) + "</div></td><td>" + esc(tenantName(item.tenant_id)) + "</td><td>" + badge(String(item.id) === String(domainID) ? "current" : item.status) + "</td><td><code>" + esc(item.dkim_selector || "mail") + "</code></td><td><div class=\"actions\"><button class=\"button\" data-select-domain=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">check_circle</span>Open</button><button class=\"button\" data-load-dns=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">dns</span>DNS</button><button class=\"button\" data-load-tls=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">encrypted</span>TLS</button><button class=\"button\" data-cloudflare-settings=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">cloud</span>Cloudflare</button>" + resourceActions("domain", item.id, item.name) + "</div></td></tr>"), "No domains in this scope yet.") +
        "</section>";
    }
    function renderDomainCertificates(tabBar) {
      const domainID = state.selectedDomainId;
      if (!domainID) {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Certificates</h4><p>Select one domain to manage DNS aliases, Let's Encrypt, custom cert paths, and mail SNI state.</p></div>" + tenantDomainActions(true) + "</div>" + tabBar + "<div class=\"card-body scope-empty\">Choose a domain from the global selector or open a domain row first.</div></section>";
      }
      const tls = state.domainTLS && String(state.domainTLS.domain_id) === String(domainID) ? state.domainTLS : null;
      if (!tls) {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Certificates for " + esc(domainName(domainID)) + "</h4><p>Load certificate settings and the current request progress for this domain.</p></div>" + tenantDomainActions(true) + "</div>" + tabBar + "<div class=\"card-body\"><button class=\"button primary\" data-load-tls=\"" + esc(domainID) + "\"><span class=\"material-symbols-outlined\">encrypted</span>Load certificate state</button></div></section>";
      }
      const settings = tls.settings || {};
      const hosts = settings.desired_hostnames || [];
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Certificates for " + esc(tls.domain || domainName(domainID)) + "</h4><p>Manage HTTPS certs, mail SNI certs, Cloudflare DNS challenge mode, and custom certificate paths.</p></div><div class=\"actions\">" +
        "<button class=\"button\" data-tls-settings=\"" + esc(domainID) + "\"><span class=\"material-symbols-outlined\">tune</span>TLS settings</button>" +
        "<button class=\"button primary\" data-tls-request=\"" + esc(domainID) + "\"><span class=\"material-symbols-outlined\">add_moderator</span>Request certificate</button>" +
        "<button class=\"button\" data-load-tls=\"" + esc(domainID) + "\"><span class=\"material-symbols-outlined\">refresh</span>Refresh</button></div></div>" + tabBar +
        "<div class=\"card-body step-list\">" +
        "<div class=\"step\"><div class=\"step-title\"><strong>Domain TLS policy</strong>" + badge(settings.tls_mode || "inherit") + "</div><div class=\"setup-line\"><strong>Challenge</strong><code>" + esc(settings.challenge_type || "dns-cloudflare") + "</code></div><div class=\"setup-line\"><strong>DNS aliases</strong><code>webmail: " + esc(settings.dns_webmail_alias_enabled ? "on" : "off") + " · madmin: " + esc(settings.dns_admin_alias_enabled ? "on" : "off") + "</code></div><div class=\"setup-line\"><strong>Applies to</strong><code>" + esc((settings.use_for_https ? "HTTPS " : "") + (settings.use_for_mail_sni ? "SMTP/IMAP/POP SNI" : "")) + "</code></div><div class=\"setup-line\"><strong>Requested names</strong><code>" + esc(hosts.length ? hosts.join(", ") : "No hostnames selected") + "</code></div></div>" +
        tlsCertificateTable(tls.certificates || []) + tlsJobList(tls.jobs || []) + "</div></section>";
    }
    function tlsCertificateTable(items) {
      return "<div><h4>Certificate inventory</h4>" + table(["Certificate", "Status", "Expires", "Used for", "Paths"], items.map(item => "<tr><td><strong>" + esc(item.common_name || "certificate " + item.id) + "</strong><div class=\"muted small\">" + esc((item.sans || []).join(", ")) + "</div></td><td>" + badge(item.status || "missing") + "</td><td>" + esc(item.not_after ? dateText(item.not_after) : "-") + "<div class=\"muted small\">" + esc(item.days_remaining || 0) + " days remaining</div></td><td>" + esc((item.used_for_https ? "HTTPS " : "") + (item.used_for_mail_sni ? "Mail SNI" : "") || "unused") + "</td><td><code>" + esc(item.cert_path || "-") + "</code></td></tr>"), "No certificates recorded for this domain yet.") + "</div>";
    }
    function tlsJobList(items) {
      return "<div><h4>Request progress</h4><div class=\"provision-list\">" + (items.length ? items.map(job => "<div class=\"provision-action " + esc(job.status || "") + "\"><div class=\"provision-meta\">" + badge(job.status || "queued") + "<strong>" + esc(job.job_type || "issue") + "</strong><code>" + esc(job.challenge_type || "dns-cloudflare") + "</code><code>" + esc((job.hostnames || []).join(", ")) + "</code></div><div class=\"progress\"><span style=\"width:" + esc(Math.max(0, Math.min(100, job.progress || 0))) + "%\"></span></div><div class=\"muted small\">" + esc(job.step || "queued") + " · " + esc(job.message || job.error || "Waiting") + "</div></div>").join("") : "<div class=\"scope-empty\">No certificate requests have been queued yet.</div>") + "</div></div>";
    }
    function renderUsers() {
      const tenantID = state.selectedTenantId;
      const domainID = state.selectedDomainId;
      const scopedIDs = new Set(scopedUsers().map(item => String(item.id)));
      const permissionRows = state.sharedPermissions.filter(item => (!tenantID || String(item.tenant_id) === String(tenantID)) && (!domainID || scopedIDs.has(String(item.shared_mailbox_id)) || scopedIDs.has(String(item.user_id))));
      const tenantAdminRows = (state.tenantAdmins || []).filter(item => (!tenantID || String(item.tenant_id) === String(tenantID)));
      const tabBar = tabs("user", [["people", "People", "person"], ["shared", "Shared mailboxes", "groups"], ["permissions", "Mailbox permissions", "key"], ["tenant-admins", "Tenant admins", "admin_panel_settings"]], state.userTab);
      const userRows = item => "<tr><td><div class=\"identity\"><span class=\"initials\">" + esc(initials(item.display_name || item.local_part)) + "</span><div><strong>" + emailFor(item) + "</strong><div class=\"muted small\">" + esc(item.display_name || "-") + "</div></div></div></td><td>" + esc(domainName(item.primary_domain_id)) + "</td><td>" + badge(item.status) + "</td><td>" + esc(formatBytes(item.quota_bytes)) + "</td><td>" + esc(dateText(item.created_at)) + "</td><td><div class=\"actions\">" + ((item.status || "").toLowerCase() === "locked" ? "<button class=\"button\" data-unlock-user=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">lock_open</span>Unlock</button>" : "") + "<button class=\"button\" data-reset-user-mfa=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">enhanced_encryption</span>Reset 2FA</button>" + resourceActions("user", item.id, emailAddress(item)) + "</div></td></tr>";
      if (state.userTab === "shared") {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Shared mailboxes for " + esc(scopeDomainLabel()) + "</h4><p>Group inboxes are created here; access is granted on the Permissions tab.</p></div><div class=\"actions\">" + createAction("shared-user", "Create shared mailbox", "group_add", true) + "</div></div>" + tabBar +
          table(["Shared mailbox", "Domain", "Status", "Storage quota", "Created", "Actions"], filteredSharedMailboxes().map(userRows), "No shared mailboxes in this scope.") + "</section>";
      }
      if (state.userTab === "permissions") {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Shared mailbox permissions</h4><p>Grant read, send as, send on behalf, and manage rights inside the selected scope.</p></div><div class=\"actions\">" + createAction("shared-permission", "Grant permission", "key", true) + "</div></div>" + tabBar +
          table(["Shared", "User", "Rights", "Actions"], visible(permissionRows).map(item => "<tr><td>" + esc(userLabel(item.shared_mailbox_id)) + "</td><td>" + esc(userLabel(item.user_id)) + "</td><td>" + esc(permissionText(item)) + "</td><td>" + resourceActions("shared-permission", item.id, userLabel(item.shared_mailbox_id) + " permission") + "</td></tr>"), "No shared permissions in this scope.") + "</section>";
      }
      if (state.userTab === "tenant-admins") {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Tenant admin permissions</h4><p>Allow mailbox users to manage the selected tenant in the admin panel with scoped rights.</p></div><div class=\"actions\">" + createAction("tenant-admin", "Add tenant admin", "admin_panel_settings", true) + "</div></div>" + tabBar +
          table(["Tenant", "User", "Role", "Status", "Actions"], visible(tenantAdminRows).map(item => "<tr><td>" + esc(tenantName(item.tenant_id)) + "</td><td>" + esc(userLabel(item.user_id)) + "</td><td>" + badge(item.role || "tenant_admin") + "</td><td>" + badge(item.status || "active") + "</td><td><button class=\"button danger\" data-delete=\"tenant-admin\" data-id=\"" + esc(item.id) + "\" data-label=\"" + esc(userLabel(item.user_id)) + "\"><span class=\"material-symbols-outlined\">delete</span>Remove</button></td></tr>"), "No tenant admins in this scope.") + "</section>";
      }
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>People for " + esc(scopeDomainLabel()) + "</h4><p>Normal user mailboxes with login, quota, webmail, IMAP, SMTP, CalDAV, and CardDAV access.</p></div><div class=\"actions\">" + createAction("user", "Create user", "person_add", true) + "</div></div>" + tabBar +
        table(["User", "Domain", "Status", "Storage quota", "Created", "Actions"], filteredNormalUsers().map(userRows), "No users in this scope.") + "</section>";
    }
    function renderSecurity() {
      const rows = state.policies.filter(item => !state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId));
      const lockedUsers = scopedUsers().filter(item => (item.status || "").toLowerCase() === "locked");
      const limits = visible(state.rateLimits || []);
      const policyCard = "<section class=\"card\"><div class=\"panel-head\"><div><h4>Security policy</h4><p>Spam action, malware action, and TLS requirements for the selected tenant.</p></div><button class=\"button\" data-refresh><span class=\"material-symbols-outlined\">refresh</span>Refresh</button></div>" +
        table(["Tenant", "Spam", "Malware", "Require TLS auth", "Actions"], visible(rows).map(item => "<tr><td><strong>" + esc(tenantName(item.tenant_id)) + "</strong><div class=\"muted small\">Tenant " + esc(item.tenant_id) + "</div></td><td><select data-policy-field=\"spam_action\" data-policy=\"" + esc(item.tenant_id) + "\"><option " + selected(item.spam_action, "mark") + ">mark</option><option " + selected(item.spam_action, "quarantine") + ">quarantine</option><option " + selected(item.spam_action, "reject") + ">reject</option></select></td><td><select data-policy-field=\"malware_action\" data-policy=\"" + esc(item.tenant_id) + "\"><option " + selected(item.malware_action, "quarantine") + ">quarantine</option><option " + selected(item.malware_action, "reject") + ">reject</option></select></td><td><input type=\"checkbox\" data-policy-field=\"require_tls_for_auth\" data-policy=\"" + esc(item.tenant_id) + "\" " + checked(item.require_tls_for_auth) + "></td><td><button class=\"button primary\" data-save-policy=\"" + esc(item.tenant_id) + "\"><span class=\"material-symbols-outlined\">save</span>Save</button></td></tr>"), "No policy found for this tenant.") + "</section>";
      const lockCard = "<section class=\"card\"><div class=\"panel-head\"><div><h4>Locked accounts</h4><p>Mailbox accounts automatically locked after repeated failed login attempts.</p></div>" + badge(lockedUsers.length) + "</div>" +
        table(["User", "Domain", "Status", "Actions"], visible(lockedUsers).map(item => "<tr><td><strong>" + emailFor(item) + "</strong><div class=\"muted small\">" + esc(item.display_name || "-") + "</div></td><td>" + esc(domainName(item.primary_domain_id)) + "</td><td>" + badge(item.status) + "</td><td><button class=\"button\" data-unlock-user=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">lock_open</span>Unlock</button></td></tr>"), "No locked accounts in this scope.") + "</section>";
      const limiterCard = "<section class=\"card\"><div class=\"panel-head\"><div><h4>Login protection</h4><p>Native IP, account, and IP plus account limiter state across admin, webmail, and mail protocols.</p></div><button class=\"button\" data-refresh><span class=\"material-symbols-outlined\">refresh</span>Refresh</button></div>" +
        table(["Service", "Subject", "Failures", "Lock", "Last seen", "Actions"], limits.map(item => "<tr><td>" + badge(item.service || "service") + "</td><td><strong>" + esc(item.subject || item.limiter_key || "-") + "</strong><div class=\"muted small\">" + esc(item.scope || "unknown") + "</div></td><td>" + esc(item.failure_count || 0) + "</td><td>" + (item.locked ? badge("blocked") : badge("tracking")) + "<div class=\"muted small\">" + esc(item.locked_until ? dateText(item.locked_until) : "No active block") + "</div></td><td>" + esc(dateText(item.last_failed_at || item.updated_at)) + "</td><td><button class=\"button\" data-clear-rate-limit=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">backspace</span>Clear</button></td></tr>"), "No login protection activity recorded yet.") + "</section>";
      return "<div class=\"grid\">" + policyCard + lockCard + limiterCard + "</div>";
    }
    function renderQuarantine() {
      const rows = state.quarantine.filter(item => !state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId));
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Quarantine</h4><p>Release false positives or delete messages that should not be delivered.</p></div><button class=\"button\" data-refresh><span class=\"material-symbols-outlined\">refresh</span>Refresh</button></div>" +
        table(["Verdict", "Recipient", "Sender", "Scanner", "Status", "Date", "Actions"], visible(rows).map(item => "<tr><td>" + badge(item.verdict) + "</td><td><strong>" + esc(item.recipient) + "</strong><div class=\"muted small\">Tenant " + esc(item.tenant_id) + "</div></td><td class=\"muted\">" + esc(item.sender || "-") + "</td><td>" + esc(item.scanner || "-") + "</td><td>" + badge(item.status || "held") + "</td><td>" + esc(dateText(item.created_at)) + "</td><td>" + quarantineActions(item) + "</td></tr>"), "No quarantine events for this tenant.") + "</section>";
    }
    function renderAudit() {
      const scoped = state.audit.filter(item => !state.selectedTenantId || !item.tenant_id || String(item.tenant_id) === String(state.selectedTenantId));
      const categories = auditTabCategories(state.auditTab);
      const baseRows = scoped.filter(item => !categories.length || categories.includes(item.category || "admin"));
      const actions = [...new Set(baseRows.map(item => item.action).filter(Boolean))].sort();
      const rows = auditFiltered(baseRows);
      const tabs = auditTabs.map(tab => {
        const count = scoped.filter(item => !tab[2].length || tab[2].includes(item.category || "admin")).length;
        return "<button class=\"audit-tab " + (state.auditTab === tab[0] ? "active" : "") + "\" data-audit-tab=\"" + esc(tab[0]) + "\">" + esc(tab[1]) + "<span class=\"badge\">" + esc(count) + "</span></button>";
      }).join("");
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Audit log</h4><p>Readable activity split by admin access, mail security, users, and system changes.</p></div><button class=\"button\" data-refresh><span class=\"material-symbols-outlined\">refresh</span>Refresh</button></div>" +
        "<div class=\"audit-tabs\">" + tabs + "</div><div class=\"audit-toolbar\">" +
        "<label>Find activity<input data-audit-search value=\"" + esc(state.auditSearch) + "\" placeholder=\"Search actor, target, detail...\"></label>" +
        "<label>Action<select data-audit-action><option value=\"\">All actions</option>" + actions.map(action => "<option value=\"" + esc(action) + "\" " + selected(state.auditAction, action) + ">" + esc(action) + "</option>").join("") + "</select></label>" +
        "<label>Severity<select data-audit-severity><option value=\"\">All severities</option><option value=\"info\" " + selected(state.auditSeverity, "info") + ">Info</option><option value=\"success\" " + selected(state.auditSeverity, "success") + ">Success</option><option value=\"warning\" " + selected(state.auditSeverity, "warning") + ">Warning</option><option value=\"danger\" " + selected(state.auditSeverity, "danger") + ">Danger</option></select></label>" +
        "</div><div class=\"audit-list\">" + (rows.length ? rows.map(auditCard).join("") : "<div class=\"scope-empty\">No audit events match these filters.</div>") + "</div></section>";
    }
    function renderSystem() {
      const tabBar = tabs("system", [["mail", "Mail behavior", "settings"], ["drift", "Config drift", "sync_problem"], ["mfa", "Admin MFA", "passkey"], ["proidentity", "ProIdentity Auth", "verified_user"]], state.systemTab);
      const health = "<section class=\"card\"><div class=\"panel-head\"><div><h4>Service health</h4><p>Live browser checks against the admin service.</p></div><button class=\"button\" data-check-health><span class=\"material-symbols-outlined\">monitor_heart</span>Check health</button></div><div class=\"card-body grid three-col\">" +
        stat("Admin API", state.health, "api") + stat("Session", state.csrf ? "active" : "login", "cookie") + stat("Data reload", "ready", "sync") + "</div></section>";
      if (state.systemTab === "drift") return "<div class=\"grid\">" + health + renderConfigDriftTab(tabBar) + "</div>";
      if (state.systemTab === "mfa") return "<div class=\"grid\">" + health + renderAdminMFATab(tabBar) + "</div>";
      if (state.systemTab === "proidentity") return "<div class=\"grid\">" + health + renderProIdentityAuthTab(tabBar) + "</div>";
      return "<div class=\"grid\">" + health + "<section class=\"card\"><div class=\"panel-head\"><div><h4>Mail server behavior</h4><p>Choose how DNS records and mail TLS identity are generated.</p></div>" + badge(state.mailSettings?.effective_hostname || "not loaded") + "</div>" + tabBar + "<div class=\"card-body\">" + mailServerSettingsForm() + "</div></section></div>";
    }
    function renderConfigDriftTab(tabBar) {
      const report = state.configDrift;
      const summary = report?.summary || {};
      const statusText = state.configDriftLoading ? "checking" : (report ? report.status : "not checked");
      const message = !report ? "Check drift to compare database-rendered configuration with the live server files." :
        report.status === "ok" ? "Live system matches database-rendered configuration." :
        report.status === "drift" ? "Live system differs from database-rendered configuration." :
        "Config drift check found errors.";
      const changed = report ? (report.items || []).filter(item => item.status !== "match") : [];
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Config drift</h4><p>Compare database desired state against live Postfix, Dovecot, Rspamd, and Nginx files.</p></div>" + badge(statusText) + "</div>" + tabBar +
        "<div class=\"card-body step-list\"><div class=\"step\"><div class=\"step-title\"><div><strong>" + esc(message) + "</strong><p class=\"muted small\">Use reload only after reviewing differences. The admin service queues the request; a root-owned systemd job performs the actual file sync and restarts.</p></div><div class=\"actions\"><button class=\"button\" data-check-config-drift><span class=\"material-symbols-outlined\">plagiarism</span>" + (state.configDriftLoading ? "Checking..." : "Check drift") + "</button><button class=\"button primary\" data-apply-config-drift " + (!report || report.status === "ok" || state.configDriftLoading ? "disabled" : "") + "><span class=\"material-symbols-outlined\">sync</span>Reload live config from DB</button></div></div>" + (state.configApplyStatus ? "<p class=\"muted small\">" + esc(state.configApplyStatus) + "</p>" : "") + "</div>" +
        (report ? "<div class=\"drift-summary\">" + driftStat("Files", summary.total || 0) + driftStat("Matching", summary.matching || 0) + driftStat("Changed", summary.drifted || 0) + driftStat("Missing", summary.missing_live || 0) + driftStat("Errors", summary.errors || 0) + "</div>" : "") +
        (report ? (changed.length ? changed.map(renderDriftItem).join("") : "<div class=\"scope-empty\">No differences found. Live configuration matches the database render.</div>") : "<div class=\"scope-empty\">No drift check has been run in this browser session.</div>") + "</div></section>";
    }
    function driftStat(label, value) {
      return "<div class=\"drift-stat\"><span class=\"muted small\">" + esc(label) + "</span><strong>" + esc(value) + "</strong></div>";
    }
    function renderDriftItem(item) {
      return "<div class=\"step\"><div class=\"step-title\"><div><strong>" + esc(item.label || item.id) + "</strong><p class=\"muted small\">" + esc(item.service || "system") + "</p></div>" + badge(item.status || "changed") + "</div>" +
        "<div class=\"setup-line\"><strong>Desired</strong><code>" + esc(item.desired_path || "-") + "</code></div><div class=\"setup-line\"><strong>Live</strong><code>" + esc(item.live_path || "-") + "</code></div>" +
        (item.error ? "<p class=\"muted small\">Error: " + esc(item.error) + "</p>" : "") +
        (item.diff ? "<pre class=\"diff-box\">" + esc(item.diff) + "</pre>" : "") + "</div>";
    }
    function renderAdminMFATab(tabBar) {
      const mfa = state.adminMFA || {};
      const enrollment = state.totpEnrollment;
      const provider = mfa.effective_provider || "disabled";
      const proIdentityActive = provider === "proidentity";
      const localText = proIdentityActive ? "Local TOTP is replaced by ProIdentity Auth while it is enabled." : (mfa.local_totp_enabled ? "Authenticator app is already set up. Use reconfigure only when rotating the secret." : "Use Google Authenticator, Microsoft Authenticator, 1Password, Bitwarden, or any app that supports standard TOTP.");
      const localButton = proIdentityActive ? "" : "<div class=\"actions\"><button class=\"button " + (mfa.local_totp_enabled ? "" : "primary") + "\" data-totp-enroll><span class=\"material-symbols-outlined\">qr_code_2</span>" + (mfa.local_totp_enabled ? "Reconfigure TOTP" : "Create setup QR") + "</button></div>";
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Admin MFA</h4><p>Configure admin second factors. ProIdentity Auth replaces local TOTP when enabled.</p></div>" + badge(provider) + "</div>" + tabBar +
        "<div class=\"card-body step-list\"><div class=\"step\"><div class=\"step-title\"><strong>Local TOTP authenticator</strong>" + badge(proIdentityActive ? "replaced" : (mfa.local_totp_enabled ? "enabled" : "not enabled")) + "</div><p class=\"muted small\">" + esc(localText) + "</p>" + localButton + "</div>" +
        (enrollment ? "<div class=\"step\"><div class=\"step-title\"><strong>Scan QR code</strong>" + badge("pending verification") + "</div><img alt=\"TOTP QR code\" style=\"width:220px;height:220px;border:1px solid var(--outline);border-radius:8px\" src=\"" + esc(enrollment.qr_data_url || "") + "\"><div class=\"setup-line\"><strong>Secret</strong><code>" + esc(enrollment.secret || "") + "</code></div><label>Verification code<input id=\"totp-verify-code\" inputmode=\"numeric\" autocomplete=\"one-time-code\" placeholder=\"123456\"></label><div class=\"actions\"><button class=\"button primary\" data-totp-verify><span class=\"material-symbols-outlined\">verified</span>Verify and enable</button></div></div>" : "") +
        "<div class=\"step\"><div class=\"step-title\"><strong>Native hardware keys</strong>" + badge((mfa.native_webauthn_credential_count || 0) > 0 ? (mfa.native_webauthn_credential_count + " registered") : "not registered") + "</div><p class=\"muted small\">Use browser WebAuthn with platform passkeys or physical security keys directly on this admin domain. This does not require ProIdentity Auth.</p><div class=\"actions\"><button class=\"button primary\" data-webauthn-register><span class=\"material-symbols-outlined\">passkey</span>Register hardware key</button></div></div></div></section>";
    }
    function renderProIdentityAuthTab(tabBar) {
      const mfa = state.adminMFA || {};
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>ProIdentity Auth</h4><p>Use your auth service for push approval and hosted TOTP verification.</p></div>" + badge(mfa.proidentity_enabled ? "enabled" : "disabled") + "</div>" + tabBar +
        "<div class=\"card-body\"><form class=\"form-grid\" data-proidentity-auth-form>" +
        "<label class=\"full\"><span class=\"checkbox-line\"><input name=\"enabled\" type=\"checkbox\" " + checked(mfa.proidentity_enabled) + "> Enable ProIdentity Auth for admin login</span></label>" +
        "<div class=\"step full\"><div class=\"step-title\"><strong>Service URL</strong>" + badge("fixed") + "</div><div class=\"setup-line\"><strong>Endpoint</strong><code>https://verify.proidentity.cloud</code></div><p class=\"muted small\">ProIdentity Auth always uses this service-provider endpoint. Only the API key, admin email, and approval timeout are configurable.</p></div>" +
        "<label>Admin user email<input name=\"user_email\" value=\"" + esc(mfa.proidentity_user_email || "") + "\" placeholder=\"admin@example.com\"></label>" +
        "<label>Service-provider API key<input name=\"api_key\" type=\"password\" value=\"\" placeholder=\"" + (mfa.proidentity_api_key_configured ? "Stored, leave blank to keep" : "Paste API key") + "\"></label>" +
        "<label>Approval timeout seconds<input name=\"timeout_seconds\" type=\"number\" min=\"30\" max=\"300\" value=\"" + esc(mfa.proidentity_timeout_seconds || 90) + "\"></label>" +
        "<div class=\"step full\"><div class=\"step-title\"><strong>Login behavior</strong>" + badge(mfa.effective_provider || "disabled") + "</div><p class=\"muted small\">When ProIdentity Auth is enabled and configured, local TOTP is not offered for admin login. The auth server provides push approval and hosted TOTP verification. Native hardware keys remain managed on the Admin MFA tab.</p></div>" +
        "<div class=\"step full\"><div class=\"step-title\"><strong>Hosted TOTP</strong>" + badge(mfa.proidentity_totp_enabled ? "enabled" : (mfa.proidentity_enabled ? "setup required" : "configure first")) + "</div><p class=\"muted small\">Registration verifies the TOTP code first, then requires ProIdentity push approval before the hosted TOTP method is saved.</p><div class=\"actions\"><button class=\"button\" type=\"button\" data-proidentity-totp-enroll><span class=\"material-symbols-outlined\">qr_code_2</span>" + (mfa.proidentity_totp_enabled ? "Reconfigure hosted TOTP" : "Create hosted TOTP QR") + "</button></div></div>" +
        "<div class=\"modal-actions\"><button class=\"button primary\" type=\"button\" data-save-proidentity-auth><span class=\"material-symbols-outlined\">save</span>Save ProIdentity Auth</button></div></form></div></section>";
    }
    function mailServerSettingsForm() {
      const settings = state.mailSettings || {};
      const headTenant = settings.head_tenant_id || state.selectedTenantId || "";
      const headDomain = settings.head_domain_id || state.selectedDomainId || "";
      return "<form class=\"form-grid\" data-mail-settings-form>" +
        "<label>Mail host mode<select name=\"hostname_mode\"><option value=\"shared\" " + selected(settings.hostname_mode || "shared", "shared") + ">Shared canonical hostname</option><option value=\"head-domain\" " + selected(settings.hostname_mode, "head-domain") + ">Head domain hostname</option><option value=\"per-domain\" " + selected(settings.hostname_mode, "per-domain") + ">Per-domain mail hostnames</option></select></label>" +
        "<label>Shared mail hostname<input name=\"mail_hostname\" value=\"" + esc(settings.mail_hostname || "") + "\" placeholder=\"mail.example.com\"></label>" +
        "<label>Head tenant<select name=\"head_tenant_id\">" + tenantOptions(headTenant, false) + "</select></label>" +
        "<label>Head domain<select name=\"head_domain_id\">" + domainOptions(headTenant, headDomain, false) + "</select></label>" +
        "<label>Default interface language<select name=\"default_language\">" + languageOptions(settings.default_language || "en") + "</select></label>" +
        "<label>Server public IPv4<input name=\"public_ipv4\" value=\"" + esc(settings.public_ipv4 || "") + "\" placeholder=\"176.107.29.40\"></label>" +
        "<label>Server public IPv6<input name=\"public_ipv6\" value=\"" + esc(settings.public_ipv6 || "") + "\" placeholder=\"Optional IPv6\"></label>" +
        "<label class=\"full\"><span class=\"checkbox-line\"><input name=\"sni_enabled\" type=\"checkbox\" " + checked(settings.sni_enabled) + "> Enable SNI certificate maps for SMTP, IMAP, and POP</span></label>" +
        "<div class=\"step full\"><div class=\"step-title\"><strong>Real client IP</strong>" + badge(settings.cloudflare_real_ip_enabled ? "Cloudflare" : "direct/proxy") + "</div><p class=\"muted small\">Enable this when public traffic reaches the internal Nginx through Cloudflare. Nginx will trust only Cloudflare source ranges and will pass the visitor IP to admin, webmail, rate limits, and ProIdentity push context.</p><label class=\"checkbox-line\"><input name=\"cloudflare_real_ip_enabled\" type=\"checkbox\" " + checked(settings.cloudflare_real_ip_enabled) + "> Behind Cloudflare proxy, use CF-Connecting-IP</label></div>" +
        "<div class=\"step full\"><div class=\"step-title\"><strong>Mailbox two-factor authentication</strong>" + badge(settings.force_mailbox_mfa ? "forced" : (settings.mailbox_mfa_enabled === false ? "disabled" : "available")) + "</div><p class=\"muted small\">When force is enabled, webmail asks users to set up 2FA after the password step. IMAP, SMTP, POP3, CalDAV, and CardDAV use app passwords.</p><label class=\"checkbox-line\"><input name=\"mailbox_mfa_enabled\" type=\"checkbox\" " + checked(settings.mailbox_mfa_enabled !== false) + "> Enable mailbox 2FA setup</label><label class=\"checkbox-line\"><input name=\"force_mailbox_mfa\" type=\"checkbox\" " + checked(settings.force_mailbox_mfa) + "> Force mailbox 2FA setup for users</label></div>" +
        "<div class=\"step full\"><div class=\"step-title\"><strong>DNS behavior</strong>" + badge(settings.effective_hostname || "pending") + "</div><p class=\"muted small\">Shared mode publishes MX to the canonical hostname and aliases mail.domain to it. Head-domain mode uses mail.head-domain as the canonical host. Per-domain mode publishes mail.domain A/AAAA records for every domain and should be paired with SNI certificates.</p></div>" +
        "<div class=\"modal-actions\"><button class=\"button primary\" type=\"button\" data-save-mail-settings><span class=\"material-symbols-outlined\">save</span>Save mail behavior</button></div></form>";
    }
    function table(headings, rows, emptyText) {
      return "<div class=\"table-wrap\"><table><thead><tr>" + headings.map(h => "<th>" + esc(h) + "</th>").join("") + "</tr></thead><tbody>" + rows.join("") + emptyRows(rows, emptyText) + "</tbody></table></div>";
    }
    function emptyRows(rows, text) {
      return rows.length ? "" : "<tr><td class=\"muted\" colspan=\"9\">" + esc(text) + "</td></tr>";
    }
    function tenantOptions(current, includeAll = false) {
      const rows = state.tenants.map(item => "<option value=\"" + esc(item.id) + "\" " + selected(item.id, current) + ">" + esc(item.name) + "</option>");
      if (rows.length && includeAll) rows.unshift("<option value=\"\" " + selected("", current) + ">All tenants</option>");
      return rows.length ? rows.join("") : "<option value=\"\">Create a tenant first</option>";
    }
    function domainOptions(tenantID, current, includeAll = false) {
      const domains = state.domains.filter(item => !tenantID || String(item.tenant_id) === String(tenantID));
      const rows = domains.map(item => "<option value=\"" + esc(item.id) + "\" " + selected(item.id, current) + ">" + esc(item.name) + (tenantID ? "" : " (" + esc(tenantName(item.tenant_id)) + ")") + "</option>");
      if (rows.length && includeAll) rows.unshift("<option value=\"\" " + selected("", current) + ">All domains</option>");
      return rows.length ? rows.join("") : "<option value=\"\">Create a domain first</option>";
    }
    function userEmailOptions(tenantID) {
      const rows = state.users.filter(item => !tenantID || String(item.tenant_id) === String(tenantID)).map(item => {
        const email = emailAddress(item);
        return "<option value=\"" + esc(email) + "\">" + esc(email) + " (" + esc(item.mailbox_type || "user") + ")</option>";
      });
      return rows.length ? rows.join("") : "<option value=\"\">Create a user first</option>";
    }
    function normalUserOptions(tenantID, domainID = "") {
      const rows = state.users.filter(item => (item.mailbox_type || "user") === "user" && (!tenantID || String(item.tenant_id) === String(tenantID)) && (!domainID || String(item.primary_domain_id) === String(domainID))).map(item => "<option value=\"" + esc(item.id) + "\">" + emailFor(item) + "</option>");
      return rows.length ? rows.join("") : "<option value=\"\">Create a user first</option>";
    }
    function sharedMailboxOptions(tenantID, domainID = "") {
      const rows = state.users.filter(item => (item.mailbox_type || "user") === "shared" && (!tenantID || String(item.tenant_id) === String(tenantID)) && (!domainID || String(item.primary_domain_id) === String(domainID))).map(item => "<option value=\"" + esc(item.id) + "\">" + emailFor(item) + "</option>");
      return rows.length ? rows.join("") : "<option value=\"\">Create a shared mailbox first</option>";
    }
    function userLabel(id) {
      const user = byID(state.users, id);
      return user ? emailAddress(user) : "User " + id;
    }
    function permissionText(item) {
      const rights = [];
      if (item.can_read) rights.push("read");
      if (item.can_send_as) rights.push("send as");
      if (item.can_send_on_behalf) rights.push("send on behalf");
      if (item.can_manage) rights.push("manage");
      return rights.join(", ") || "none";
    }
    function renderDNSRecords(records) {
      if (!records || !records.length) return "<div class=\"muted small\">Load records after selecting a domain. The backend generates MX, SPF, DKIM, DMARC, MTA-STS, and TLS reporting values where available.</div>";
      return records.map(record => "<div class=\"dns-record\"><strong>" + esc(record.type) + "</strong><code>" + esc(record.name) + "</code><code>" + esc(record.priority ? record.priority + " " : "") + esc(record.value) + "</code>" + (record.purpose ? "<div class=\"purpose\">" + esc(record.purpose) + (record.required ? " · required" : " · optional") + "</div>" : "") + "</div>").join("");
    }
    function renderClientSetup(items) {
      if (!items || !items.length) return "";
      return "<div class=\"setup-grid\">" + items.map(item => "<div class=\"setup-item\"><div class=\"step-title\"><h4>" + esc(item.client) + "</h4>" + badge(item.status || "supported") + "</div><div class=\"muted small\">" + esc(item.method || "") + "</div>" +
        (item.urls || []).map(url => "<div class=\"setup-line\"><strong>" + esc(url.label) + "</strong><code>" + esc(url.value) + "</code></div>").join("") +
        (item.dns || []).map(record => "<div class=\"setup-line\"><strong>" + esc(record.type) + "</strong><code>" + esc(record.name + " -> " + record.value) + "</code></div>").join("") +
        (item.notes || []).map(note => "<div class=\"muted small\">" + esc(note) + "</div>").join("") + "</div>").join("") + "</div>";
    }
    function renderDNSWarnings(warnings) {
      if (!warnings || !warnings.length) return "";
      return "<div class=\"step\">" + warnings.map(item => "<div class=\"provision-action blocked\"><div class=\"provision-meta\">" + badge("blocked") + "<strong>DNS not ready</strong></div><div class=\"muted small\">" + esc(item) + "</div></div>").join("") + "</div>";
    }
    function renderProvisionPlan(plan) {
      if (!plan) return "<div class=\"scope-empty\">Run a Cloudflare check to preview DNS changes.</div>";
      const actions = plan.actions || [];
      return "<div class=\"step\"><div class=\"step-title\"><strong>Provisioning plan</strong>" + badge(plan.status || "unknown") + "</div><p class=\"muted small\">" + esc(plan.summary || "") + "</p>" +
        "<div class=\"provision-list\">" + (actions.length ? actions.map(action => "<div class=\"provision-action " + esc(action.action || "") + "\"><div class=\"provision-meta\">" + badge(action.action || "review") + "<strong>" + esc(action.type) + "</strong><code>" + esc(action.name) + "</code><code>" + esc(action.priority ? action.priority + " " : "") + esc(action.value || "") + "</code></div><div class=\"muted small\">" + esc(action.reason || "") + "</div>" +
          ((action.existing || []).length ? "<div class=\"record-grid\">" + renderDNSRecords(action.existing) + "</div>" : "") + "</div>").join("") : "<div class=\"scope-empty\">No DNS changes needed.</div>") + "</div></div>";
    }
    function cloudflareBusyStep(title, message) {
      return "<div class=\"step\"><div class=\"step-title\"><div class=\"actions\"><span class=\"material-symbols-outlined spin\">progress_activity</span><strong>" + esc(title) + "</strong></div>" + badge("applying") + "</div><p class=\"muted small\">" + esc(message) + "</p></div>";
    }
    function setCloudflareBusy(busy, action) {
      const buttons = document.querySelectorAll("#modal-body [data-cloudflare-check], #modal-body [data-cloudflare-apply], #modal [data-close-modal]");
      buttons.forEach(button => button.disabled = !!busy);
      const applyButton = $("#modal-body [data-cloudflare-apply]");
      const checkButton = $("#modal-body [data-cloudflare-check]");
      if (applyButton) applyButton.innerHTML = busy && action === "apply" ? "<span class=\"material-symbols-outlined spin\">progress_activity</span>Applying DNS..." : "<span class=\"material-symbols-outlined\">cloud_sync</span>Apply DNS";
      if (checkButton) checkButton.innerHTML = busy && action === "check" ? "<span class=\"material-symbols-outlined spin\">progress_activity</span>Checking records..." : "<span class=\"material-symbols-outlined\">policy</span>Check records";
    }
    function cloudflareStatus(config) {
      config = config || {};
      return config.token_configured ? (config.status || "configured") : "not configured";
    }
    function cloudflareDNSAction(dns, config) {
      config = config || {};
      if (!config.token_configured) {
        return "<div class=\"step\"><div class=\"step-title\"><div><h4>Cloudflare auto setup</h4><p class=\"muted small\">Add the Cloudflare DNS token in this domain's settings before provisioning records.</p></div>" + badge("not configured") + "</div><div class=\"actions\"><button class=\"button\" data-cloudflare-settings=\"" + esc(dns.domain_id) + "\"><span class=\"material-symbols-outlined\">cloud</span>Domain Cloudflare settings</button></div></div>";
      }
      if (!dns.provisionable) {
        return "<div class=\"step\"><div class=\"step-title\"><div><h4>Cloudflare auto setup</h4><p class=\"muted small\">Cloudflare is configured, but DNS cannot be applied until the blocking warning is fixed.</p></div>" + badge("blocked") + "</div><button class=\"button primary\" disabled><span class=\"material-symbols-outlined\">cloud_off</span>Cloudflare auto setup</button></div>";
      }
      return "<div class=\"step\"><div class=\"step-title\"><div><h4>Cloudflare auto setup</h4><p class=\"muted small\">Cloudflare token is configured for this domain. Check first, then apply with backup if conflicts need replacement.</p></div>" + badge(cloudflareStatus(config)) + "</div><div class=\"actions\"><button class=\"button primary\" data-cloudflare-provision=\"" + esc(dns.domain_id) + "\"><span class=\"material-symbols-outlined\">cloud_sync</span>Cloudflare auto setup</button><button class=\"button\" data-cloudflare-settings=\"" + esc(dns.domain_id) + "\"><span class=\"material-symbols-outlined\">settings</span>Settings</button></div></div>";
    }
    function cloudflareSettingsForm(domainID, config) {
      config = config || {};
      return "<div class=\"step-list\"><div class=\"step\"><div class=\"step-title\"><div><h4>Domain Cloudflare settings</h4><p class=\"muted small\">Store the DNS edit token for this domain. The token is never returned to the browser.</p></div>" + badge(cloudflareStatus(config)) + "</div>" + (config.zone_name ? "<div class=\"setup-line\"><strong>Zone</strong><code>" + esc(config.zone_name) + "</code></div>" : "") + (config.last_error ? "<p class=\"muted small\">Last error: " + esc(config.last_error) + "</p>" : "") + "</div><form class=\"form-grid\" data-cloudflare-settings-form=\"" + esc(domainID) + "\">" +
        "<label>Cloudflare zone ID<input name=\"zone_id\" value=\"" + esc(config.zone_id || "") + "\" placeholder=\"Optional, auto-detect by domain\"></label>" +
        "<label>API token<input name=\"api_token\" type=\"password\" autocomplete=\"off\" placeholder=\"" + (config.token_configured ? "Configured - leave blank to keep" : "Cloudflare DNS edit token") + "\"></label>" +
        "<div class=\"modal-actions\"><button class=\"button\" type=\"button\" data-close-modal><span class=\"material-symbols-outlined\">close</span>Cancel</button><button class=\"button primary\" type=\"button\" data-cloudflare-save=\"" + esc(domainID) + "\"><span class=\"material-symbols-outlined\">key</span>Save Cloudflare settings</button></div></form></div>";
    }
    function cloudflareProvisionForm(domainID, config) {
      config = config || {};
      return "<div class=\"step-list\"><div class=\"step\"><div class=\"step-title\"><div><h4>Cloudflare DNS provisioning</h4><p class=\"muted small\">Preview desired MX, SPF, DKIM, DMARC, discovery, and service records against the existing zone.</p></div>" + badge(cloudflareStatus(config)) + "</div><div class=\"setup-line\"><strong>Zone</strong><code>" + esc(config.zone_name || config.zone_id || "auto-detect") + "</code></div><label class=\"checkbox-line\"><input id=\"cloudflare-replace\" type=\"checkbox\">Backup old records and replace conflicts</label><div class=\"modal-actions\"><button class=\"button\" type=\"button\" data-cloudflare-check=\"" + esc(domainID) + "\"><span class=\"material-symbols-outlined\">policy</span>Check records</button><button class=\"button primary\" type=\"button\" data-cloudflare-apply=\"" + esc(domainID) + "\"><span class=\"material-symbols-outlined\">cloud_sync</span>Apply DNS</button></div></div><div id=\"cloudflare-result\">" + renderProvisionPlan(null) + "</div></div>";
    }
    function tlsSettingsForm(domainID, tls) {
      const settings = (tls && tls.settings) || {};
      return "<form class=\"form-grid\" data-tls-settings-form=\"" + esc(domainID) + "\">" +
        "<label>TLS mode<select name=\"tls_mode\"><option value=\"inherit\" " + selected(settings.tls_mode || "inherit", "inherit") + ">Use system default</option><option value=\"letsencrypt-dns-cloudflare\" " + selected(settings.tls_mode, "letsencrypt-dns-cloudflare") + ">Let's Encrypt DNS challenge</option><option value=\"letsencrypt-http\" " + selected(settings.tls_mode, "letsencrypt-http") + ">Let's Encrypt HTTP challenge</option><option value=\"custom\" " + selected(settings.tls_mode, "custom") + ">Custom certificate paths</option><option value=\"disabled\" " + selected(settings.tls_mode, "disabled") + ">Disabled</option></select></label>" +
        "<label>Default challenge<select name=\"challenge_type\"><option value=\"dns-cloudflare\" " + selected(settings.challenge_type || "dns-cloudflare", "dns-cloudflare") + ">Cloudflare DNS</option><option value=\"http-01\" " + selected(settings.challenge_type, "http-01") + ">HTTP-01 webroot</option><option value=\"manual-dns\" " + selected(settings.challenge_type, "manual-dns") + ">Manual DNS</option><option value=\"custom-import\" " + selected(settings.challenge_type, "custom-import") + ">Custom import</option><option value=\"none\" " + selected(settings.challenge_type, "none") + ">None</option></select></label>" +
        "<label class=\"full\"><span class=\"checkbox-line\"><input name=\"dns_webmail_alias_enabled\" type=\"checkbox\" " + checked(settings.dns_webmail_alias_enabled !== false) + "> Publish optional webmail." + esc(tls?.domain || domainName(domainID)) + " DNS alias</span></label>" +
        "<label class=\"full\"><span class=\"checkbox-line\"><input name=\"dns_admin_alias_enabled\" type=\"checkbox\" " + checked(settings.dns_admin_alias_enabled !== false) + "> Publish optional madmin." + esc(tls?.domain || domainName(domainID)) + " DNS alias</span></label>" +
        "<label><span class=\"checkbox-line\"><input name=\"use_for_https\" type=\"checkbox\" " + checked(settings.use_for_https !== false) + "> Use for HTTPS</span></label><label><span class=\"checkbox-line\"><input name=\"use_for_mail_sni\" type=\"checkbox\" " + checked(settings.use_for_mail_sni !== false) + "> Use for Postfix/Dovecot SNI</span></label>" +
        "<label><span class=\"checkbox-line\"><input name=\"include_mail_hostname\" type=\"checkbox\" " + checked(settings.include_mail_hostname !== false) + "> Include mail.domain</span></label><label><span class=\"checkbox-line\"><input name=\"include_webmail_hostname\" type=\"checkbox\" " + checked(settings.include_webmail_hostname !== false) + "> Include webmail.domain</span></label>" +
        "<label><span class=\"checkbox-line\"><input name=\"include_admin_hostname\" type=\"checkbox\" " + checked(settings.include_admin_hostname !== false) + "> Include madmin.domain</span></label><label>Certificate name<input name=\"certificate_name\" value=\"" + esc(settings.certificate_name || "") + "\" placeholder=\"mail." + esc(tls?.domain || domainName(domainID)) + "\"></label>" +
        "<label class=\"full\">Custom fullchain path<input name=\"custom_cert_path\" value=\"" + esc(settings.custom_cert_path || "") + "\" placeholder=\"/etc/proidentity-mail/certs/domain/fullchain.pem\"></label>" +
        "<label>Custom key path<input name=\"custom_key_path\" value=\"" + esc(settings.custom_key_path || "") + "\" placeholder=\"/etc/proidentity-mail/certs/domain/privkey.pem\"></label><label>Custom chain path<input name=\"custom_chain_path\" value=\"" + esc(settings.custom_chain_path || "") + "\" placeholder=\"optional\"></label>" +
        "<div class=\"modal-actions\"><button class=\"button\" type=\"button\" data-close-modal><span class=\"material-symbols-outlined\">close</span>Cancel</button><button class=\"button primary\" type=\"button\" data-tls-save=\"" + esc(domainID) + "\"><span class=\"material-symbols-outlined\">save</span>Save TLS settings</button></div></form>";
    }
    function tlsRequestForm(domainID, tls) {
      const settings = (tls && tls.settings) || {};
      const hosts = settings.desired_hostnames || [];
      const hostChecks = hosts.length ? hosts.map(host => "<label class=\"full\"><span class=\"checkbox-line\"><input name=\"hostnames\" type=\"checkbox\" value=\"" + esc(host) + "\" checked> " + esc(host) + "</span></label>").join("") : "<div class=\"scope-empty full\">No hostnames selected in TLS settings.</div>";
      return "<form class=\"form-grid\" data-tls-request-form=\"" + esc(domainID) + "\">" +
        "<label>Request type<select name=\"job_type\"><option value=\"issue\">Issue new certificate</option><option value=\"renew\">Renew existing certificate</option><option value=\"deploy\">Deploy saved certificate</option><option value=\"check\">Check certificate state</option></select></label>" +
        "<label>Challenge<select name=\"challenge_type\"><option value=\"dns-cloudflare\" " + selected(settings.challenge_type || "dns-cloudflare", "dns-cloudflare") + ">Cloudflare DNS</option><option value=\"http-01\" " + selected(settings.challenge_type, "http-01") + ">HTTP-01 webroot</option><option value=\"manual-dns\" " + selected(settings.challenge_type, "manual-dns") + ">Manual DNS</option><option value=\"custom-import\" " + selected(settings.challenge_type, "custom-import") + ">Custom import</option></select></label>" +
        "<div class=\"step full\"><div class=\"step-title\"><strong>Certificate names</strong>" + badge(hosts.length ? hosts.length + " names" : "none") + "</div>" + hostChecks + "</div>" +
        "<div id=\"tls-request-result\" class=\"full\">" + tlsJobList(tls?.jobs || []) + "</div>" +
        "<div class=\"modal-actions\"><button class=\"button\" type=\"button\" data-close-modal><span class=\"material-symbols-outlined\">close</span>Close</button><button class=\"button primary\" type=\"button\" data-tls-queue=\"" + esc(domainID) + "\"><span class=\"material-symbols-outlined\">add_moderator</span>Queue certificate request</button></div></form>";
    }
    function quarantineActions(item) {
      if (item.status && item.status !== "held") return "<span class=\"muted small\">" + esc(item.resolution_note || "resolved") + "</span>";
      return "<div class=\"actions\"><button class=\"button\" data-quarantine-action=\"release\" data-quarantine-id=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">outbox</span>Release</button><button class=\"button danger\" data-quarantine-action=\"delete\" data-quarantine-id=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">delete</span>Delete</button></div>";
    }
    function auditRow(item) {
      return "<tr><td><strong>" + esc(item.title || item.action) + "</strong><div class=\"muted small\">" + esc(item.summary || dateText(item.created_at)) + "</div></td></tr>";
    }
    function auditFiltered(rows) {
      let out = rows;
      if (state.auditAction) out = out.filter(item => item.action === state.auditAction);
      if (state.auditSeverity) out = out.filter(item => (item.severity || "info") === state.auditSeverity);
      const q = (state.auditSearch || state.query || "").trim().toLowerCase();
      if (q) out = out.filter(item => JSON.stringify(item).toLowerCase().includes(q));
      return out;
    }
    function auditCard(item) {
      const details = (item.details || []).slice(0, 8).map(detail => "<span class=\"audit-detail\"><strong>" + esc(detail.label) + "</strong>" + esc(detail.value) + "</span>").join("");
      return "<article class=\"audit-card\"><div class=\"audit-card-head\"><div><h4>" + esc(item.title || item.action) + "</h4><div class=\"audit-summary\">" + esc(item.summary || "") + "</div></div><div class=\"audit-meta\">" + badge(item.severity || "info") + badge(auditCategoryLabel(item.category)) + "</div></div>" +
        "<div class=\"audit-meta\"><span><strong>Actor:</strong> " + esc(item.actor_label || item.actor_type || "-") + "</span><span><strong>Target:</strong> " + esc(item.target_label || item.target_id || "-") + "</span><span><strong>Tenant:</strong> " + esc(item.tenant_id ? tenantName(item.tenant_id) : "System wide") + "</span><span><strong>When:</strong> " + esc(dateText(item.created_at)) + "</span></div>" +
        (details ? "<div class=\"audit-details\">" + details + "</div>" : "") + "</article>";
    }
    function endpoint(path) {
      return "<div class=\"step-title\"><code>" + esc(path) + "</code><button class=\"button\" data-copy=\"" + esc(location.origin + path) + "\"><span class=\"material-symbols-outlined\">content_copy</span>Copy</button></div>";
    }
    function formatBytes(value) {
      const n = Number(value || 0);
      if (!n) return "-";
      if (n >= 1073741824) return (n / 1073741824).toFixed(1) + " GB";
      if (n >= 1048576) return (n / 1048576).toFixed(1) + " MB";
      return n + " B";
    }
    function formPayload(form) {
      const data = Object.fromEntries(new FormData(form).entries());
      ["tenant_id", "primary_domain_id", "domain_id", "shared_mailbox_id", "user_id"].forEach(key => { if (data[key]) data[key] = Number(data[key]); });
      if (data.quota_value) {
        const unit = String(data.quota_unit || "gb").toLowerCase();
        const multiplier = unit === "mb" ? 1048576 : 1073741824;
        data.quota_bytes = Math.round(Number(data.quota_value) * multiplier);
        delete data.quota_value;
        delete data.quota_unit;
      } else if (data.quota_gb) {
        data.quota_bytes = Number(data.quota_gb) * 1073741824;
        delete data.quota_gb;
      }
      ["can_read", "can_send_as", "can_send_on_behalf", "can_manage"].forEach(key => data[key] = data[key] === "on");
      if (data.password === "") delete data.password;
      return data;
    }
    function resourcePath(type, id) {
      const base = {
        "tenant": "/api/v1/tenants",
        "domain": "/api/v1/domains",
        "user": "/api/v1/users",
        "alias": "/api/v1/aliases",
        "catch-all": "/api/v1/catch-all",
        "shared-permission": "/api/v1/shared-permissions",
        "tenant-admin": "/api/v1/tenant-admins"
      }[type];
      if (!base) throw new Error("Unknown resource type");
      return base + "/" + id;
    }
    function resourceName(type) {
      return {"tenant":"tenant","domain":"domain","user":"user","alias":"alias","catch-all":"catch-all","shared-permission":"shared permission","tenant-admin":"tenant admin"}[type] || type;
    }
    function findResource(type, id) {
      const source = {
        "tenant": state.tenants,
        "domain": state.domains,
        "user": state.users,
        "alias": state.aliases,
        "catch-all": state.catchAll,
        "shared-permission": state.sharedPermissions,
        "tenant-admin": state.tenantAdmins
      }[type] || [];
      return byID(source, id);
    }
    function modalActions(label = "Save changes", icon = "save") {
      return "<div class=\"modal-actions\"><button class=\"button\" type=\"button\" data-close-modal><span class=\"material-symbols-outlined\">close</span>Cancel</button><button class=\"button primary\" type=\"submit\"><span class=\"material-symbols-outlined\">" + esc(icon) + "</span>" + esc(label) + "</button></div>";
    }
    function openModal(title, body) {
      $("#modal-title").textContent = t(title);
      $("#modal-body").innerHTML = body;
      $("#modal").classList.remove("hidden");
      translateUI($("#modal"));
      $("#modal-body input, #modal-body select, #modal-body textarea")?.focus();
    }
    function closeModal() {
      if (state.stepUpReject) {
        const reject = state.stepUpReject;
        state.stepUpReject = null;
        reject(new Error("Admin step-up was cancelled"));
      }
      $("#modal").classList.add("hidden");
      $("#modal-body").innerHTML = "";
    }
    function editForm(type, item) {
      if (type === "tenant") {
        return "<form class=\"form-grid\" data-edit-form=\"tenant\" data-id=\"" + esc(item.id) + "\"><label>Name<input name=\"name\" value=\"" + esc(item.name) + "\" required></label><label>Slug<input name=\"slug\" value=\"" + esc(item.slug) + "\" required></label><label>Status<select name=\"status\">" + statusOptions(item.status, ["active", "suspended"]) + "</select></label>" + modalActions() + "</form>";
      }
      if (type === "domain") {
        return "<form class=\"form-grid\" data-edit-form=\"domain\" data-id=\"" + esc(item.id) + "\"><label>Tenant<select name=\"tenant_id\" required>" + tenantOptions(item.tenant_id, false) + "</select></label><label>Domain<input name=\"name\" value=\"" + esc(item.name) + "\" required></label><label>Status<select name=\"status\">" + statusOptions(item.status, ["pending", "active", "disabled"]) + "</select></label><label>DKIM selector<input name=\"dkim_selector\" value=\"" + esc(item.dkim_selector || "mail") + "\" required></label>" + modalActions() + "</form>";
      }
      if (type === "user") {
        return "<form class=\"form-grid\" data-edit-form=\"user\" data-id=\"" + esc(item.id) + "\"><label>Tenant<select name=\"tenant_id\" required>" + tenantOptions(item.tenant_id, false) + "</select></label><label>Domain<select name=\"primary_domain_id\" required>" + domainOptions(item.tenant_id, item.primary_domain_id, false) + "</select></label><label>Type<select name=\"mailbox_type\"><option value=\"user\" " + selected(item.mailbox_type || "user", "user") + ">User mailbox</option><option value=\"shared\" " + selected(item.mailbox_type, "shared") + ">Shared mailbox</option></select></label><label>Status<select name=\"status\">" + statusOptions(item.status || "active", ["active", "locked", "disabled"]) + "</select></label>" + quotaField(item.quota_bytes) + "<label>Local part<input name=\"local_part\" value=\"" + esc(item.local_part) + "\" required></label><label>Display name<input name=\"display_name\" value=\"" + esc(item.display_name || "") + "\"></label><label class=\"full\">New password<input name=\"password\" type=\"password\" autocomplete=\"new-password\" placeholder=\"Leave blank to keep current password\"></label>" + modalActions() + "</form>";
      }
      if (type === "alias") {
        return "<form class=\"form-grid\" data-edit-form=\"alias\" data-id=\"" + esc(item.id) + "\"><label>Tenant<select name=\"tenant_id\" required>" + tenantOptions(item.tenant_id, false) + "</select></label><label>Domain<select name=\"domain_id\" required>" + domainOptions(item.tenant_id, item.domain_id, false) + "</select></label><label>Alias local part<input name=\"source_local_part\" value=\"" + esc(item.source_local_part) + "\" required></label><label>Destination<select name=\"destination\" required>" + userEmailOptions(item.tenant_id) + "</select></label>" + modalActions() + "</form>";
      }
      if (type === "catch-all") {
        return "<form class=\"form-grid\" data-edit-form=\"catch-all\" data-id=\"" + esc(item.id) + "\"><label>Tenant<select name=\"tenant_id\" required>" + tenantOptions(item.tenant_id, false) + "</select></label><label>Domain<select name=\"domain_id\" required>" + domainOptions(item.tenant_id, item.domain_id, false) + "</select></label><label>Catch-all mailbox<select name=\"destination\" required>" + userEmailOptions(item.tenant_id) + "</select></label><label>Status<select name=\"status\">" + statusOptions(item.status || "active", ["active", "disabled"]) + "</select></label>" + modalActions() + "</form>";
      }
      if (type === "shared-permission") {
        return "<form class=\"form-grid\" data-edit-form=\"shared-permission\" data-id=\"" + esc(item.id) + "\"><label>Tenant<select name=\"tenant_id\" required>" + tenantOptions(item.tenant_id, false) + "</select></label><label>Shared mailbox<select name=\"shared_mailbox_id\" required>" + sharedMailboxOptions(item.tenant_id) + "</select></label><label>User<select name=\"user_id\" required>" + normalUserOptions(item.tenant_id) + "</select></label><label><span><input name=\"can_read\" type=\"checkbox\" " + checked(item.can_read) + "> Read</span></label><label><span><input name=\"can_send_as\" type=\"checkbox\" " + checked(item.can_send_as) + "> Send as</span></label><label><span><input name=\"can_send_on_behalf\" type=\"checkbox\" " + checked(item.can_send_on_behalf) + "> Send on behalf</span></label><label><span><input name=\"can_manage\" type=\"checkbox\" " + checked(item.can_manage) + "> Manage</span></label>" + modalActions() + "</form>";
      }
      return "<div class=\"scope-empty\">This resource cannot be edited.</div>";
    }
    function createForm(type) {
      const domain = selectedDomain();
      const tenantID = state.selectedTenantId || (domain ? String(domain.tenant_id) : "");
      const domainID = state.selectedDomainId || "";
      if (type === "domain") {
        return "<form class=\"form-grid\" data-form=\"domain\">" + tenantField(tenantID) + "<label>Domain<input name=\"name\" placeholder=\"example.com\" required></label>" + modalActions("Add domain", "add_link") + "</form>";
      }
      if (type === "alias") {
        return "<form class=\"form-grid\" data-form=\"alias\">" + tenantField(tenantID) + domainField(tenantID, domainID, "domain_id") + "<label>Alias local part<input name=\"source_local_part\" placeholder=\"sales\" required></label><label>Destination<select name=\"destination\" required>" + userEmailOptions(tenantID) + "</select></label>" + modalActions("Create alias", "alternate_email") + "</form>";
      }
      if (type === "catch-all") {
        return "<form class=\"form-grid\" data-form=\"catch-all\">" + tenantField(tenantID) + domainField(tenantID, domainID, "domain_id") + "<label class=\"full\">Catch-all mailbox<select name=\"destination\" required>" + userEmailOptions(tenantID) + "</select></label>" + modalActions("Set catch-all mailbox", "all_inbox") + "</form>";
      }
      if (type === "user" || type === "shared-user") {
        const mailboxType = type === "shared-user" ? "shared" : "user";
        return "<form class=\"form-grid\" data-form=\"user\">" + tenantField(tenantID) + domainField(tenantID, domainID) + typeField(mailboxType) + quotaField() +
          "<label>Local part<input name=\"local_part\" placeholder=\"marko\" required></label><label>Display name<input name=\"display_name\" placeholder=\"Marko Admin\"></label>" +
          (mailboxType === "shared" ? "" : "<label class=\"full\">Password<input name=\"password\" type=\"password\" autocomplete=\"new-password\" required></label>") +
          modalActions(mailboxType === "shared" ? "Create shared mailbox" : "Create user", mailboxType === "shared" ? "group_add" : "person_add") + "</form>";
      }
      if (type === "shared-permission") {
        return "<form class=\"form-grid\" data-form=\"shared-permission\">" + tenantField(tenantID) + "<label>Shared mailbox<select name=\"shared_mailbox_id\" required>" + sharedMailboxOptions(tenantID, domainID) + "</select></label><label>User<select name=\"user_id\" required>" + normalUserOptions(tenantID, domainID) + "</select></label><label><span><input name=\"can_read\" type=\"checkbox\" checked> Read</span></label><label><span><input name=\"can_send_as\" type=\"checkbox\"> Send as</span></label><label><span><input name=\"can_send_on_behalf\" type=\"checkbox\"> Send on behalf</span></label><label><span><input name=\"can_manage\" type=\"checkbox\"> Manage</span></label>" + modalActions("Grant permission", "key") + "</form>";
      }
      if (type === "tenant-admin") {
        return "<form class=\"form-grid\" data-form=\"tenant-admin\">" + tenantField(tenantID) + "<label>User<select name=\"user_id\" required>" + normalUserOptions(tenantID, "") + "</select></label><label>Role<select name=\"role\"><option value=\"tenant_admin\">Tenant admin</option><option value=\"read_only\">Read only</option></select></label><label>Status<select name=\"status\"><option value=\"active\">active</option><option value=\"disabled\">disabled</option></select></label>" + modalActions("Add tenant admin", "admin_panel_settings") + "</form>";
      }
      return "<div class=\"scope-empty\">Select a tenant and domain first.</div>";
    }
    function openCreate(type) {
      const title = {
        "domain": "Add domain",
        "alias": "Create alias",
        "catch-all": "Set catch-all mailbox",
        "user": "Create user",
        "shared-user": "Create shared mailbox",
        "shared-permission": "Grant shared mailbox permission",
        "tenant-admin": "Add tenant admin"
      }[type] || "Create";
      openModal(title, createForm(type));
    }
    function openDNSModal(dns, cloudflareConfig) {
      openModal("DNS and auto setup for " + dns.domain, "<div class=\"step-list\"><div class=\"step\"><div class=\"step-title\"><div><h4>Publish DNS records</h4><p class=\"muted small\">These records use the configured mail server identity: <code>" + esc(dns.mail_host) + "</code>.</p></div>" + badge(dns.provisionable ? "ready" : "blocked") + "</div><div class=\"actions\"><button class=\"button\" data-tls-settings=\"" + esc(dns.domain_id) + "\"><span class=\"material-symbols-outlined\">encrypted</span>Optional webmail/madmin aliases and TLS</button></div></div>" + renderDNSWarnings(dns.warnings) + cloudflareDNSAction(dns, cloudflareConfig) + "<div class=\"record-grid\">" + renderDNSRecords(dns.records) + "</div><div><h4>Client setup support</h4><p class=\"muted small\">Thunderbird and Outlook can discover settings automatically when the CNAME records point here. Gmail app uses manual IMAP/SMTP setup.</p></div>" + renderClientSetup(dns.client_setup) + "</div>");
    }
    async function openCloudflareSettings(domainID) {
      const config = await api("/api/v1/domains/" + domainID + "/cloudflare");
      state.dnsCloudflare = config;
      openModal("Cloudflare settings for " + domainName(domainID), cloudflareSettingsForm(domainID, config));
    }
    async function openCloudflareProvisionModal(domainID) {
      const config = await api("/api/v1/domains/" + domainID + "/cloudflare");
      state.dnsCloudflare = config;
      if (!config.token_configured) {
        openModal("Cloudflare settings for " + domainName(domainID), cloudflareSettingsForm(domainID, config));
        showStatus("Add Cloudflare token before DNS provisioning", true);
        return;
      }
      openModal("Cloudflare DNS for " + domainName(domainID), cloudflareProvisionForm(domainID, config));
    }
    async function loadTLS(domainID, openRequest) {
      if (!domainID) throw new Error("Select a domain first");
      const tls = await api("/api/v1/domains/" + domainID + "/tls");
      state.domainTLS = tls;
      state.selectedDomainId = String(domainID);
      state.domainTab = "certificates";
      render();
      if (openRequest) openModal("Request certificate for " + (tls.domain || domainName(domainID)), tlsRequestForm(domainID, tls));
      showStatus("TLS state loaded for " + (tls.domain || domainName(domainID)));
      return tls;
    }
    async function openTLSSettings(domainID) {
      const tls = state.domainTLS && String(state.domainTLS.domain_id) === String(domainID) ? state.domainTLS : await loadTLS(domainID, false);
      openModal("TLS settings for " + (tls.domain || domainName(domainID)), tlsSettingsForm(domainID, tls));
    }
    async function openTLSRequest(domainID) {
      const tls = state.domainTLS && String(state.domainTLS.domain_id) === String(domainID) ? state.domainTLS : await loadTLS(domainID, false);
      openModal("Request certificate for " + (tls.domain || domainName(domainID)), tlsRequestForm(domainID, tls));
    }
    async function saveTLSSettings(domainID) {
      const form = $("#modal-body [data-tls-settings-form]");
      const data = Object.fromEntries(new FormData(form).entries());
      const payload = {
        dns_webmail_alias_enabled: data.dns_webmail_alias_enabled === "on",
        dns_admin_alias_enabled: data.dns_admin_alias_enabled === "on",
        tls_mode: data.tls_mode || "inherit",
        challenge_type: data.challenge_type || "dns-cloudflare",
        use_for_https: data.use_for_https === "on",
        use_for_mail_sni: data.use_for_mail_sni === "on",
        include_mail_hostname: data.include_mail_hostname === "on",
        include_webmail_hostname: data.include_webmail_hostname === "on",
        include_admin_hostname: data.include_admin_hostname === "on",
        certificate_name: data.certificate_name || "",
        custom_cert_path: data.custom_cert_path || "",
        custom_key_path: data.custom_key_path || "",
        custom_chain_path: data.custom_chain_path || ""
      };
      const settings = await api("/api/v1/domains/" + domainID + "/tls/settings", {method: "PUT", body: JSON.stringify(payload)});
      state.dns = null;
      state.dnsCloudflare = null;
      state.domainTLS = await api("/api/v1/domains/" + domainID + "/tls");
      $("#modal-body").innerHTML = tlsSettingsForm(domainID, state.domainTLS);
      render();
      showStatus("TLS settings saved");
      return settings;
    }
    async function queueTLSJob(domainID) {
      const form = $("#modal-body [data-tls-request-form]");
      const data = Object.fromEntries(new FormData(form).entries());
      const hostnames = [...form.querySelectorAll("input[name='hostnames']:checked")].map(input => input.value);
      const resultBox = $("#tls-request-result");
      if (resultBox) resultBox.innerHTML = cloudflareBusyStep("Queueing certificate request", "Saving the request so the root TLS worker can run certbot and deploy to Nginx, Postfix, and Dovecot.");
      const job = await api("/api/v1/domains/" + domainID + "/tls/jobs", {method: "POST", body: JSON.stringify({job_type: data.job_type || "issue", challenge_type: data.challenge_type || "dns-cloudflare", hostnames})});
      state.domainTLS = await api("/api/v1/domains/" + domainID + "/tls");
      if ($("#tls-request-result")) $("#tls-request-result").innerHTML = tlsJobList(state.domainTLS.jobs || []);
      render();
      showStatus("TLS request queued: job " + job.id);
    }
    async function saveCloudflareConfig(domainID) {
      const form = $("#modal-body [data-cloudflare-settings-form]");
      const data = Object.fromEntries(new FormData(form).entries());
      const config = await api("/api/v1/domains/" + domainID + "/cloudflare", {method: "PUT", body: JSON.stringify({zone_id: data.zone_id || "", api_token: data.api_token || ""})});
      state.dnsCloudflare = config;
      $("#modal-body").innerHTML = cloudflareSettingsForm(domainID, config);
      render();
      showStatus(config.token_configured ? "Cloudflare token saved" : "Cloudflare config saved");
    }
    async function checkCloudflare(domainID) {
      setCloudflareBusy(true, "check");
      const resultBox = $("#cloudflare-result");
      if (resultBox) resultBox.innerHTML = cloudflareBusyStep("Checking Cloudflare DNS", "Reading existing records and comparing them with the desired mail records.");
      showStatus("Checking Cloudflare DNS...");
      try {
        const plan = await api("/api/v1/domains/" + domainID + "/cloudflare/check", {method: "POST"});
        $("#cloudflare-result").innerHTML = renderProvisionPlan(plan);
        showStatus("Cloudflare DNS check completed");
      } catch (error) {
        if ($("#cloudflare-result")) $("#cloudflare-result").innerHTML = "<div class=\"step\"><div class=\"step-title\"><strong>Cloudflare check failed</strong>" + badge("error") + "</div><p class=\"muted small\">" + esc(error.message) + "</p></div>";
        throw error;
      } finally {
        setCloudflareBusy(false, "check");
      }
    }
    async function applyCloudflare(domainID) {
      const replace = !!$("#cloudflare-replace")?.checked;
      setCloudflareBusy(true, "apply");
      const resultBox = $("#cloudflare-result");
      if (resultBox) resultBox.innerHTML = cloudflareBusyStep("Applying Cloudflare DNS", "Creating, updating, and backing up records. Keep this window open until the result appears.");
      showStatus("Applying Cloudflare DNS...");
      try {
        const result = await api("/api/v1/domains/" + domainID + "/cloudflare/apply", {method: "POST", body: JSON.stringify({replace})});
        $("#cloudflare-result").innerHTML = "<div class=\"step\"><div class=\"step-title\"><strong>Applied</strong>" + badge("ok") + "</div><p class=\"muted small\">Changed " + esc(result.changed || 0) + " record(s). Backup ID " + esc(result.backup_id || "-") + ".</p></div>" + renderProvisionPlan(result.plan);
        showStatus("Cloudflare DNS applied");
      } catch (error) {
        if ($("#cloudflare-result")) $("#cloudflare-result").innerHTML = "<div class=\"step\"><div class=\"step-title\"><strong>Cloudflare apply failed</strong>" + badge("error") + "</div><p class=\"muted small\">" + esc(error.message) + "</p></div>";
        throw error;
      } finally {
        setCloudflareBusy(false, "apply");
      }
    }
    function openEdit(type, id) {
      const item = findResource(type, id);
      if (!item) {
        showStatus("Resource not found", true);
        return;
      }
      openModal("Edit " + resourceName(type), editForm(type, item));
      if (type === "alias") {
        const select = $("#modal-body select[name='destination']");
        if (select) select.value = item.destination;
      }
      if (type === "catch-all") {
        const select = $("#modal-body select[name='destination']");
        if (select) select.value = item.destination;
      }
      if (type === "shared-permission") {
        const shared = $("#modal-body select[name='shared_mailbox_id']");
        const user = $("#modal-body select[name='user_id']");
        if (shared) shared.value = String(item.shared_mailbox_id);
        if (user) user.value = String(item.user_id);
      }
    }
    async function submitEditForm(form) {
      const type = form.dataset.editForm;
      const id = form.dataset.id;
      const updated = await api(resourcePath(type, id), {method: "PUT", body: JSON.stringify(formPayload(form))});
      if (updated && type === "tenant") state.selectedTenantId = String(updated.id);
      if (updated && type === "domain") state.selectedDomainId = String(updated.id);
      closeModal();
      await refresh();
      showStatus("Saved " + resourceName(type));
    }
    async function deleteResource(type, id, label) {
      const detail = type === "tenant" || type === "domain" || type === "user" ? " This disables the record and preserves stored mail, groupware data, quarantine, and audit history." : " This removes the routing or permission row.";
      if (!confirm(t("Remove") + " " + t(resourceName(type)) + " " + label + "?" + detail)) return;
      await api(resourcePath(type, id), {method: "DELETE"});
      await refresh();
      showStatus("Removed " + resourceName(type));
    }
    async function submitForm(form) {
      const type = form.dataset.form;
      const inModal = !!form.closest("#modal");
      const data = formPayload(form);
      const path = type === "tenant" ? "/api/v1/tenants" : type === "domain" ? "/api/v1/domains" : type === "alias" ? "/api/v1/aliases" : type === "catch-all" ? "/api/v1/catch-all" : type === "shared-permission" ? "/api/v1/shared-permissions" : type === "tenant-admin" ? "/api/v1/tenant-admins" : "/api/v1/users";
      const created = await api(path, {method: "POST", body: JSON.stringify(data)});
      if (type === "tenant") state.selectedTenantId = String(created.id);
      if (type === "domain") state.selectedDomainId = String(created.id);
      if (type === "user") {
        state.selectedTenantId = String(created.tenant_id);
        state.userTab = created.mailbox_type === "shared" ? "shared" : "people";
      }
      if (type === "alias") {
        state.view = "domains";
        state.domainTab = "aliases";
      }
      if (type === "catch-all") {
        state.view = "domains";
        state.domainTab = "catchall";
      }
      if (type === "shared-permission") state.userTab = "permissions";
      if (type === "tenant-admin") state.userTab = "tenant-admins";
      form.reset();
      if (inModal) closeModal();
      await refresh();
      showStatus((type === "user" ? "User" : type === "catch-all" ? "Catch-all" : type === "shared-permission" ? "Shared permission" : type === "tenant-admin" ? "Tenant admin" : type[0].toUpperCase() + type.slice(1)) + " created");
    }
    async function loadDNS(id) {
      const selectedID = id === "selected" ? ($("#dns-domain")?.value || state.selectedDomainId) : id;
      if (!selectedID) throw new Error("Select a domain first");
      const [dns, cloudflareConfig] = await Promise.all([
        api("/api/v1/domains/" + selectedID + "/dns"),
        api("/api/v1/domains/" + selectedID + "/cloudflare").catch(() => null)
      ]);
      state.dns = dns;
      state.dnsCloudflare = cloudflareConfig;
      state.selectedDomainId = String(selectedID);
      render();
      openDNSModal(state.dns, state.dnsCloudflare);
      showStatus("DNS records loaded for " + state.dns.domain);
    }
    async function savePolicy(tenantID) {
      const fields = document.querySelectorAll("[data-policy=\"" + CSS.escape(String(tenantID)) + "\"]");
      const body = {};
      fields.forEach(field => body[field.dataset.policyField] = field.type === "checkbox" ? field.checked : field.value);
      await api("/api/v1/policies/" + tenantID, {method: "PUT", body: JSON.stringify(body)});
      await refresh();
      showStatus("Security policy saved");
    }
    async function saveMailSettings() {
      const form = $("[data-mail-settings-form]");
      const data = Object.fromEntries(new FormData(form).entries());
      const payload = {
        hostname_mode: data.hostname_mode || "shared",
        mail_hostname: data.mail_hostname || "",
        head_tenant_id: data.head_tenant_id ? Number(data.head_tenant_id) : null,
        head_domain_id: data.head_domain_id ? Number(data.head_domain_id) : null,
        public_ipv4: data.public_ipv4 || "",
        public_ipv6: data.public_ipv6 || "",
        sni_enabled: data.sni_enabled === "on",
        mailbox_mfa_enabled: data.mailbox_mfa_enabled === "on",
        force_mailbox_mfa: data.force_mailbox_mfa === "on",
        cloudflare_real_ip_enabled: data.cloudflare_real_ip_enabled === "on",
        default_language: data.default_language || "en"
      };
      state.mailSettings = await api("/api/v1/mail-server-settings", {method: "PUT", body: JSON.stringify(payload)});
      applyLanguage(state.mailSettings.default_language || payload.default_language || "en");
      state.dns = null;
      state.dnsCloudflare = null;
      state.domainTLS = null;
      await refresh();
      showStatus("Mail server behavior saved");
    }
    async function startTOTPEnrollment() {
      state.totpEnrollment = await api("/api/v1/admin-mfa/totp/enroll", {method: "POST", body: JSON.stringify({})});
      render();
      showStatus("TOTP setup QR created");
    }
    async function verifyTOTPEnrollment() {
      const code = $("#totp-verify-code")?.value || "";
      state.adminMFA = await api("/api/v1/admin-mfa/totp/verify", {method: "POST", body: JSON.stringify({code})});
      state.totpEnrollment = null;
      render();
      showStatus("Admin TOTP MFA enabled");
    }
    async function saveProIdentityAuth() {
      const form = $("[data-proidentity-auth-form]");
      const data = Object.fromEntries(new FormData(form).entries());
      state.adminMFA = await api("/api/v1/admin-mfa/proidentity", {method: "PUT", body: JSON.stringify({
        enabled: data.enabled === "on",
        api_key: data.api_key || "",
        user_email: data.user_email || "",
        timeout_seconds: Number(data.timeout_seconds || 90)
      })});
      state.totpEnrollment = null;
      render();
      showStatus("ProIdentity Auth settings saved");
    }
    async function startProIdentityTOTPEnrollment() {
      const enrollment = await api("/api/v1/admin-mfa/proidentity/totp/enroll", {method: "POST", body: JSON.stringify({})});
      const qrData = enrollment.qr_data_url || enrollment.data_url || enrollment.qr_code_data_url || enrollment.qr_png_data_url || (enrollment.base64_png ? "data:image/png;base64," + enrollment.base64_png : "");
      const otpauth = enrollment.otpauth_url || "";
      state.proIdentityTOTPEnrollment = enrollment;
      openModal("ProIdentity hosted TOTP", "<div class=\"step-list\" id=\"hosted-totp-setup\"><div class=\"step\"><div class=\"step-title\"><strong>1. Scan setup QR</strong>" + badge("hosted") + "</div>" + (qrData ? "<img alt=\"ProIdentity TOTP QR code\" style=\"width:220px;height:220px;border:1px solid var(--outline);border-radius:8px\" src=\"" + esc(qrData) + "\">" : "<div class=\"scope-empty\">The auth service did not return a QR image.</div>") + (otpauth ? "<div class=\"setup-line\"><strong>Setup URL</strong><code>" + esc(otpauth) + "</code></div>" : "") + "<p class=\"muted small\">Add this code to your authenticator app, then enter the generated code below.</p></div><div class=\"step\"><div class=\"step-title\"><strong>2. Verify code</strong>" + badge("required") + "</div><label>Authenticator code<div class=\"mfa-code-control\"><input id=\"hosted-totp-code\" inputmode=\"numeric\" autocomplete=\"one-time-code\" placeholder=\"123456\"><span class=\"material-symbols-outlined\">shield</span></div></label><p class=\"mfa-status\" id=\"hosted-totp-status\">After the code is verified, we will send a ProIdentity push approval to finish saving this method.</p><div class=\"actions\"><button class=\"button primary\" type=\"button\" data-proidentity-totp-verify><span class=\"material-symbols-outlined\">verified</span>Verify code</button><button class=\"button\" type=\"button\" data-close-modal><span class=\"material-symbols-outlined\">close</span>Cancel</button></div></div></div><div class=\"auth-wait hidden\" id=\"hosted-totp-phone\"><div class=\"auth-wait-inner\"><div class=\"auth-wait-icon\"><span class=\"material-symbols-outlined\">phone_iphone</span></div><h3>Check your phone</h3><p>We sent a confirmation request to the ProIdentity app on your registered device.</p><div class=\"auth-wait-status\" id=\"hosted-totp-phone-status\">Waiting for approval...</div><button class=\"button\" type=\"button\" data-proidentity-totp-confirm><span class=\"material-symbols-outlined\">refresh</span>Check approval</button><button class=\"button\" type=\"button\" data-close-modal>Cancel and try another method</button></div></div>");
      $("#hosted-totp-code")?.focus();
      showStatus("ProIdentity hosted TOTP enrollment created");
    }
    function setHostedTOTPStatus(message, isError = false) {
      const phonePane = $("#hosted-totp-phone");
      const el = phonePane && !phonePane.classList.contains("hidden") ? $("#hosted-totp-phone-status") : ($("#hosted-totp-status") || $("#hosted-totp-phone-status"));
      if (!el) return;
      el.textContent = t(message);
      el.style.color = isError ? "var(--danger)" : "";
    }
    async function verifyProIdentityHostedTOTP() {
      const code = ($("#hosted-totp-code")?.value || "").trim();
      if (!code) {
        setHostedTOTPStatus("Enter the 6-digit code from the authenticator app.", true);
        return;
      }
      setHostedTOTPStatus("Verifying code...");
      const challenge = await api("/api/v1/admin-mfa/proidentity/totp/verify", {method: "POST", body: JSON.stringify({code})});
      state.proIdentityTOTPChallenge = challenge;
      $("#hosted-totp-setup")?.classList.add("hidden");
      $("#hosted-totp-phone")?.classList.remove("hidden");
      await pollProIdentityHostedTOTP();
    }
    async function confirmProIdentityHostedTOTP() {
      const challenge = state.proIdentityTOTPChallenge;
      if (!challenge?.mfa_token) throw new Error("Hosted TOTP confirmation expired");
      const response = await fetch("/api/v1/admin-mfa/proidentity/totp/confirm", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({mfa_token: challenge.mfa_token})});
      const body = await response.json().catch(() => ({}));
      if (response.status === 202) return body;
      if (!response.ok) throw new Error(body.error || "Hosted TOTP confirmation failed");
      state.adminMFA = body;
      state.proIdentityTOTPChallenge = null;
      closeModal();
      render();
      showStatus("ProIdentity hosted TOTP enabled");
      return body;
    }
    async function pollProIdentityHostedTOTP() {
      $("#hosted-totp-phone-status").textContent = t("Waiting for approval...");
      for (let attempt = 0; attempt < 45; attempt++) {
        const body = await confirmProIdentityHostedTOTP();
        if (body?.proidentity_totp_enabled) return;
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
      throw new Error("ProIdentity push approval timed out");
    }
    function base64URLToBuffer(value) {
      const normalized = String(value || "").replace(/-/g, "+").replace(/_/g, "/");
      const padded = normalized + "=".repeat((4 - normalized.length % 4) % 4);
      const binary = atob(padded);
      const bytes = new Uint8Array(binary.length);
      for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
      return bytes.buffer;
    }
    function bufferToBase64URL(buffer) {
      const bytes = new Uint8Array(buffer || new ArrayBuffer(0));
      let binary = "";
      bytes.forEach(byte => binary += String.fromCharCode(byte));
      return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
    }
    function clonePublicKeyOptions(options) {
      return JSON.parse(JSON.stringify(options || {}));
    }
    function decodeCreationOptions(options) {
      const publicKey = clonePublicKeyOptions(options);
      publicKey.challenge = base64URLToBuffer(publicKey.challenge);
      publicKey.user.id = base64URLToBuffer(publicKey.user.id);
      if (publicKey.excludeCredentials) {
        publicKey.excludeCredentials = publicKey.excludeCredentials.map(item => Object.assign({}, item, {id: base64URLToBuffer(item.id)}));
      }
      return publicKey;
    }
    function decodeRequestOptions(options) {
      const publicKey = clonePublicKeyOptions(options);
      publicKey.challenge = base64URLToBuffer(publicKey.challenge);
      if (publicKey.allowCredentials) {
        publicKey.allowCredentials = publicKey.allowCredentials.map(item => Object.assign({}, item, {id: base64URLToBuffer(item.id)}));
      }
      return publicKey;
    }
    function encodeCredential(credential) {
      const response = {};
      for (const key of ["clientDataJSON", "attestationObject", "authenticatorData", "signature", "userHandle"]) {
        if (credential.response[key]) response[key] = bufferToBase64URL(credential.response[key]);
      }
      return {
        id: credential.id,
        rawId: bufferToBase64URL(credential.rawId),
        type: credential.type,
        response,
        clientExtensionResults: credential.getClientExtensionResults ? credential.getClientExtensionResults() : {}
      };
    }
    async function startWebAuthnRegistration() {
      if (!window.PublicKeyCredential || !navigator.credentials) throw new Error("This browser does not support hardware keys or passkeys");
      const name = "Admin hardware key " + new Date().toLocaleDateString();
      const challenge = await api("/api/v1/admin-mfa/webauthn/register/begin", {method: "POST", body: JSON.stringify({})});
      state.webAuthnRegistration = challenge;
      const credential = await navigator.credentials.create({publicKey: decodeCreationOptions(challenge.publicKey)});
      if (!credential) throw new Error("Hardware key registration was cancelled");
      state.adminMFA = await api("/api/v1/admin-mfa/webauthn/register/finish", {method: "POST", body: JSON.stringify({token: challenge.token, name, credential: encodeCredential(credential)})});
      state.webAuthnRegistration = null;
      render();
      showStatus("Hardware key registered");
    }
    async function resolveQuarantine(id, action) {
      const note = prompt(t("Resolution note"), t(action === "release" ? "false positive" : "malware/spam removed")) || "";
      await api("/api/v1/quarantine/" + id + "/" + action, {method: "POST", body: JSON.stringify({resolution_note: note})});
      await refresh();
      showStatus("Quarantine event " + (action === "delete" ? "deleted" : "released"));
    }
    async function unlockUser(id) {
      await api("/api/v1/users/" + id + "/unlock", {method: "POST"});
      await refresh();
      showStatus("User unlocked and login limiter entries cleared");
    }
    async function resetUserMFA(id) {
      await api("/api/v1/users/" + id + "/mfa/reset", {method: "POST"});
      await refresh();
      showStatus("User 2FA reset. If forced 2FA is enabled, setup will be required on next webmail login.");
    }
    async function clearRateLimit(id) {
      await api("/api/v1/security/login-rate-limits/" + id, {method: "DELETE"});
      await refresh();
      showStatus("Login protection entry cleared");
    }
    async function logout() {
      await api("/api/v1/session", {method: "DELETE"});
      state.csrf = "";
      $("#login-cover").classList.remove("hidden");
      showStatus("Logged out");
    }
    document.addEventListener("click", event => {
      if (event.target.id === "modal" || event.target.closest("[data-close-modal]")) {
        closeModal();
        return;
      }
      const tab = event.target.closest("[data-tab]");
      if (tab) {
        if (tab.dataset.tabScope === "domain") state.domainTab = tab.dataset.tab;
        if (tab.dataset.tabScope === "user") state.userTab = tab.dataset.tab;
        if (tab.dataset.tabScope === "system") state.systemTab = tab.dataset.tab;
        render();
        if (tab.dataset.tabScope === "system" && state.systemTab === "drift" && !state.configDrift && !state.configDriftLoading) {
          loadConfigDrift().catch(error => showStatus(error.message, true));
        }
        return;
      }
      const nextDomainTab = event.target.closest("[data-set-domain-tab]")?.dataset.setDomainTab;
      if (nextDomainTab) state.domainTab = nextDomainTab;
      const nextUserTab = event.target.closest("[data-set-user-tab]")?.dataset.setUserTab;
      if (nextUserTab) state.userTab = nextUserTab;
      const view = event.target.closest("[data-view]")?.dataset.view;
      if (view) setView(view);
      const refreshButton = event.target.closest("[data-refresh]");
      if (refreshButton) refresh().then(() => showStatus("Data refreshed")).catch(error => showStatus(error.message, true));
      const healthButton = event.target.closest("[data-check-health]");
      if (healthButton) checkHealth().then(() => showStatus("Health checked"));
      const driftButton = event.target.closest("[data-check-config-drift]");
      if (driftButton) {
        loadConfigDrift().catch(error => showStatus(error.message, true));
        return;
      }
      const applyDriftButton = event.target.closest("[data-apply-config-drift]");
      if (applyDriftButton) {
        requestConfigApply().catch(error => showStatus(error.message, true));
        return;
      }
      const dns = event.target.closest("[data-load-dns]")?.dataset.loadDns;
      if (dns) loadDNS(dns).catch(error => showStatus(error.message, true));
      const loadTLSID = event.target.closest("[data-load-tls]")?.dataset.loadTls;
      if (loadTLSID) {
        loadTLS(loadTLSID, false).catch(error => showStatus(error.message, true));
        return;
      }
      const tlsSettings = event.target.closest("[data-tls-settings]")?.dataset.tlsSettings;
      if (tlsSettings) {
        openTLSSettings(tlsSettings).catch(error => showStatus(error.message, true));
        return;
      }
      const tlsRequest = event.target.closest("[data-tls-request]")?.dataset.tlsRequest;
      if (tlsRequest) {
        openTLSRequest(tlsRequest).catch(error => showStatus(error.message, true));
        return;
      }
      const tlsSave = event.target.closest("[data-tls-save]")?.dataset.tlsSave;
      if (tlsSave) {
        saveTLSSettings(tlsSave).catch(error => showStatus(error.message, true));
        return;
      }
      const tlsQueue = event.target.closest("[data-tls-queue]")?.dataset.tlsQueue;
      if (tlsQueue) {
        queueTLSJob(tlsQueue).catch(error => showStatus(error.message, true));
        return;
      }
      const cloudflareSettings = event.target.closest("[data-cloudflare-settings]")?.dataset.cloudflareSettings;
      if (cloudflareSettings) {
        openCloudflareSettings(cloudflareSettings).catch(error => showStatus(error.message, true));
        return;
      }
      const cloudflareProvision = event.target.closest("[data-cloudflare-provision]")?.dataset.cloudflareProvision;
      if (cloudflareProvision) {
        openCloudflareProvisionModal(cloudflareProvision).catch(error => showStatus(error.message, true));
        return;
      }
      const cloudflareSave = event.target.closest("[data-cloudflare-save]")?.dataset.cloudflareSave;
      if (cloudflareSave) {
        saveCloudflareConfig(cloudflareSave).catch(error => showStatus(error.message, true));
        return;
      }
      const cloudflareCheck = event.target.closest("[data-cloudflare-check]")?.dataset.cloudflareCheck;
      if (cloudflareCheck) {
        checkCloudflare(cloudflareCheck).catch(error => showStatus(error.message, true));
        return;
      }
      const cloudflareApply = event.target.closest("[data-cloudflare-apply]")?.dataset.cloudflareApply;
      if (cloudflareApply) {
        applyCloudflare(cloudflareApply).catch(error => showStatus(error.message, true));
        return;
      }
      const save = event.target.closest("[data-save-policy]")?.dataset.savePolicy;
      if (save) savePolicy(save).catch(error => showStatus(error.message, true));
      const saveMail = event.target.closest("[data-save-mail-settings]");
      if (saveMail) {
        saveMailSettings().catch(error => showStatus(error.message, true));
        return;
      }
      const totpEnroll = event.target.closest("[data-totp-enroll]");
      if (totpEnroll) {
        startTOTPEnrollment().catch(error => showStatus(error.message, true));
        return;
      }
      const totpVerify = event.target.closest("[data-totp-verify]");
      if (totpVerify) {
        verifyTOTPEnrollment().catch(error => showStatus(error.message, true));
        return;
      }
      const saveProIdentity = event.target.closest("[data-save-proidentity-auth]");
      if (saveProIdentity) {
        saveProIdentityAuth().catch(error => showStatus(error.message, true));
        return;
      }
      const proIdentityTOTPEnroll = event.target.closest("[data-proidentity-totp-enroll]");
      if (proIdentityTOTPEnroll) {
        startProIdentityTOTPEnrollment().catch(error => showStatus(error.message, true));
        return;
      }
      const proIdentityTOTPVerify = event.target.closest("[data-proidentity-totp-verify]");
      if (proIdentityTOTPVerify) {
        verifyProIdentityHostedTOTP().catch(error => setHostedTOTPStatus(error.message, true));
        return;
      }
      const proIdentityTOTPConfirm = event.target.closest("[data-proidentity-totp-confirm]");
      if (proIdentityTOTPConfirm) {
        confirmProIdentityHostedTOTP().catch(error => setHostedTOTPStatus(error.message, true));
        return;
      }
      const webAuthnRegister = event.target.closest("[data-webauthn-register]");
      if (webAuthnRegister) {
        startWebAuthnRegistration().catch(error => showStatus(error.message, true));
        return;
      }
      const qButton = event.target.closest("[data-quarantine-action]");
      if (qButton) resolveQuarantine(qButton.dataset.quarantineId, qButton.dataset.quarantineAction).catch(error => showStatus(error.message, true));
      const unlockButton = event.target.closest("[data-unlock-user]");
      if (unlockButton) {
        if (confirm(t("Unlock this mailbox user and clear its failed-login limiter entries?"))) {
          unlockUser(unlockButton.dataset.unlockUser).catch(error => showStatus(error.message, true));
        }
        return;
      }
      const resetMFAButton = event.target.closest("[data-reset-user-mfa]");
      if (resetMFAButton) {
        if (confirm(t("Reset this user's mailbox 2FA? They will need to set it up again when force 2FA is enabled."))) {
          resetUserMFA(resetMFAButton.dataset.resetUserMfa).catch(error => showStatus(error.message, true));
        }
        return;
      }
      const clearRateLimitButton = event.target.closest("[data-clear-rate-limit]");
      if (clearRateLimitButton) {
        if (confirm(t("Clear this login protection entry?"))) {
          clearRateLimit(clearRateLimitButton.dataset.clearRateLimit).catch(error => showStatus(error.message, true));
        }
        return;
      }
      const auditTab = event.target.closest("[data-audit-tab]")?.dataset.auditTab;
      if (auditTab) {
        state.auditTab = auditTab;
        state.auditAction = "";
        state.auditSeverity = "";
        render();
        return;
      }
      const createButton = event.target.closest("[data-create]");
      if (createButton) {
        openCreate(createButton.dataset.create);
        return;
      }
      const editButton = event.target.closest("[data-edit]");
      if (editButton) {
        openEdit(editButton.getAttribute("data-edit"), editButton.dataset.id);
        return;
      }
      const deleteButton = event.target.closest("[data-delete]");
      if (deleteButton) {
        deleteResource(deleteButton.getAttribute("data-delete"), deleteButton.dataset.id, deleteButton.dataset.label || deleteButton.dataset.id).catch(error => showStatus(error.message, true));
        return;
      }
      const selectTenant = event.target.closest("[data-select-tenant]")?.dataset.selectTenant;
      if (selectTenant) {
        state.selectedTenantId = String(selectTenant);
        state.selectedDomainId = "";
        state.dns = null;
        state.dnsCloudflare = null;
        state.domainTLS = null;
        setView("domains");
      }
      const selectDomain = event.target.closest("[data-select-domain]")?.dataset.selectDomain;
      if (selectDomain) {
        state.selectedDomainId = String(selectDomain);
        state.dns = null;
        state.dnsCloudflare = null;
        state.domainTLS = null;
        state.userTab = "people";
        setView("users");
      }
      const copy = event.target.closest("[data-copy]")?.dataset.copy;
      if (copy) navigator.clipboard.writeText(copy).then(() => showStatus("Copied")).catch(error => showStatus("Copy failed: " + error.message, true));
    });
    document.addEventListener("submit", event => {
      const editForm = event.target.closest("[data-edit-form]");
      if (editForm) {
        event.preventDefault();
        submitEditForm(editForm).catch(error => showStatus(error.message, true));
        return;
      }
      const form = event.target.closest("[data-form]");
      if (!form) return;
      event.preventDefault();
      submitForm(form).catch(error => showStatus(error.message, true));
    });
    document.addEventListener("change", event => {
      if (event.target.id === "global-tenant") {
        setTenantScope(event.target.value);
      }
      if (event.target.id === "global-domain") {
        setDomainScope(event.target.value);
      }
      if (event.target.matches("[data-audit-action]")) {
        state.auditAction = event.target.value;
        render();
      }
      if (event.target.matches("[data-audit-severity]")) {
        state.auditSeverity = event.target.value;
        render();
      }
    });
    document.addEventListener("input", event => {
      if (event.target.matches("[data-audit-search]")) {
        state.auditSearch = event.target.value;
        render();
      }
    });
    document.addEventListener("keydown", event => {
      if (event.target.id === "hosted-totp-code" && event.key === "Enter") {
        event.preventDefault();
        verifyProIdentityHostedTOTP().catch(error => setHostedTOTPStatus(error.message, true));
      }
    });
    $("#search").addEventListener("input", event => { state.query = event.target.value; render(); });
    $("#reload-nav").addEventListener("click", () => refresh().then(() => showStatus("Data refreshed")).catch(error => showStatus(error.message, true)));
    $("#reload-top").addEventListener("click", () => refresh().then(() => showStatus("Data refreshed")).catch(error => showStatus(error.message, true)));
    $("#logout-nav").addEventListener("click", () => logout().catch(error => showStatus(error.message, true)));
    $("#start-onboarding").addEventListener("click", () => setView("onboarding"));
    $("#copy-discovery").addEventListener("click", () => navigator.clipboard.writeText(location.origin + "/.well-known/proidentity-mail/config.json?emailaddress=user@example.com").then(() => showStatus("Discovery URL copied")).catch(error => showStatus("Copy failed: " + error.message, true)));
    function setMFAStatus(message) {
      $("#mfa-detail").textContent = t(message);
    }
    function setLoginStatus(message, isError = false) {
      const el = $("#login-status");
      if (!el) return;
      el.textContent = message ? t(message) : "";
      el.className = "login-status" + (message ? "" : " hidden") + (isError ? " error" : "");
    }
    function setPushMFAStatus(message, isError = false) {
      const el = $("#push-mfa-status");
      if (!el) return;
      el.textContent = t(message);
      el.style.color = isError ? "#ff9494" : "";
    }
    function hideMFAPanel() {
      state.pendingMFA = null;
      state.mfaPolling = false;
      setLoginStatus("");
      $("#login-cover").classList.remove("push-mode");
      $("#login-form").classList.remove("hidden");
      $("#mfa-panel").classList.add("hidden");
      $("#push-mfa-view").classList.add("hidden");
      $("#push-mfa-code-panel").classList.add("hidden");
      $("#mfa-code").value = "";
      $("#push-mfa-code").value = "";
    }
    function showMFAPanel(challenge) {
      state.pendingMFA = challenge;
      if (challenge.provider === "proidentity") {
        showProIdentityPushView(challenge);
        return;
      }
      const providers = challenge.providers || [challenge.provider];
      const hasCode = challenge.provider === "totp" || providers.includes("totp");
      const hasPush = false;
      const hasWebAuthn = challenge.provider === "webauthn" || providers.includes("webauthn");
      $("#mfa-provider-badge").textContent = challenge.provider === "webauthn" ? t("hardware key") : "TOTP";
      $("#mfa-title").textContent = t("Two-factor verification");
      $("#mfa-code-row").classList.toggle("hidden", !hasCode);
      $("#mfa-submit").classList.toggle("hidden", !hasCode);
      $("#mfa-webauthn").classList.toggle("hidden", !hasWebAuthn);
      $("#mfa-push").classList.toggle("hidden", !hasPush);
      $("#mfa-panel").classList.remove("hidden");
      if (challenge.provider === "webauthn") {
        setMFAStatus("Use your registered passkey or security key to finish signing in.");
      } else {
        setMFAStatus("Enter the 6-digit code from your authenticator app.");
      }
      if (hasCode) $("#mfa-code").focus();
    }
    function showProIdentityPushView(challenge) {
      state.pendingMFA = challenge;
      $("#login-cover").classList.add("push-mode");
      $("#login-form").classList.add("hidden");
      $("#mfa-panel").classList.add("hidden");
      $("#push-mfa-view").classList.remove("hidden");
      $("#push-mfa-code-panel").classList.add("hidden");
      $("#push-mfa-code").value = "";
      setPushMFAStatus(t("Waiting for approval..."));
      pollProIdentityMFA(challenge).catch(error => setPushMFAStatus(error.message, true));
    }
    function showPushManualCode() {
      $("#push-mfa-code-panel").classList.remove("hidden");
      setPushMFAStatus(t("Enter the hosted TOTP code or approve the push request."));
      $("#push-mfa-code").focus();
    }
    async function finishAdminMFA(challenge, code = "") {
      const response = await fetch("/api/v1/session/mfa", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json"}, body: JSON.stringify({mfa_token: challenge.mfa_token, code})});
      const body = await response.json().catch(() => ({}));
      if (response.status === 202 && challenge.provider === "proidentity") return body;
      if (!response.ok) throw new Error(body.error || "MFA failed");
      state.csrf = body.csrf_token || "";
      hideMFAPanel();
      $("#login-cover").classList.add("hidden");
      $("#login-form").reset();
      await refresh();
      showStatus("Logged in");
      return body;
    }
    async function pollProIdentityMFA(challenge) {
      if (!challenge) throw new Error("MFA challenge expired");
      if (state.mfaPolling) return;
      state.mfaPolling = true;
      try {
        if (challenge.provider === "proidentity") setPushMFAStatus(t("Waiting for approval..."));
        else setMFAStatus(t("Waiting for approval..."));
        for (let attempt = 0; attempt < 45; attempt++) {
          if (!state.pendingMFA || state.pendingMFA.mfa_token !== challenge.mfa_token) return;
          const body = await finishAdminMFA(challenge);
          if (body && body.csrf_token) return;
          await new Promise(resolve => setTimeout(resolve, 2000));
        }
        throw new Error("ProIdentity Auth approval timed out");
      } finally {
        state.mfaPolling = false;
      }
    }
    async function verifyMFAFromPanel() {
      const challenge = state.pendingMFA;
      if (!challenge) throw new Error("MFA challenge expired");
      const code = $("#mfa-code").value.trim();
      if (!code) throw new Error("Authenticator code is required");
      setMFAStatus("Verifying code...");
      await finishAdminMFA(challenge, code);
    }
    async function verifyPushManualCode() {
      const challenge = state.pendingMFA;
      if (!challenge) throw new Error("MFA challenge expired");
      const code = $("#push-mfa-code").value.trim();
      if (!code) throw new Error("Authenticator code is required");
      setPushMFAStatus("Verifying code...");
      await finishAdminMFA(challenge, code);
    }
    async function runWebAuthnLogin() {
      const challenge = state.pendingMFA;
      if (!challenge || !challenge.publicKey) throw new Error("Hardware-key challenge expired");
      if (!window.PublicKeyCredential || !navigator.credentials) throw new Error("This browser does not support hardware keys or passkeys");
      setMFAStatus("Waiting for your passkey or security key...");
      const credential = await navigator.credentials.get({publicKey: decodeRequestOptions(challenge.publicKey)});
      if (!credential) throw new Error("Hardware-key verification was cancelled");
      const response = await fetch("/api/v1/session/mfa/webauthn", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json"}, body: JSON.stringify({mfa_token: challenge.mfa_token, credential: encodeCredential(credential)})});
      const body = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(body.error || "Hardware-key verification failed");
      state.csrf = body.csrf_token || "";
      hideMFAPanel();
      $("#login-cover").classList.add("hidden");
      $("#login-form").reset();
      await refresh();
      showStatus("Logged in");
    }
    async function runWebAuthnStepUp(challenge, statusEl) {
      if (!challenge || !challenge.publicKey) throw new Error("Hardware-key challenge expired");
      if (!window.PublicKeyCredential || !navigator.credentials) throw new Error("This browser does not support hardware keys or passkeys");
      if (statusEl) statusEl.textContent = t("Waiting for your passkey or security key...");
      const credential = await navigator.credentials.get({publicKey: decodeRequestOptions(challenge.publicKey)});
      if (!credential) throw new Error("Hardware-key verification was cancelled");
      return api("/api/v1/session/step-up/webauthn", {method: "POST", body: JSON.stringify({mfa_token: challenge.mfa_token, credential: encodeCredential(credential)}), retryStepUp: false});
    }
    $("#mfa-submit").addEventListener("click", () => verifyMFAFromPanel().catch(error => setMFAStatus(error.message)));
    $("#mfa-code").addEventListener("keydown", event => {
      if (event.key === "Enter") {
        event.preventDefault();
        verifyMFAFromPanel().catch(error => setMFAStatus(error.message));
      }
    });
    $("#mfa-webauthn").addEventListener("click", () => runWebAuthnLogin().catch(error => setMFAStatus(error.message)));
    $("#mfa-push").addEventListener("click", () => {
      state.mfaPolling = false;
      pollProIdentityMFA(state.pendingMFA).catch(error => setMFAStatus(error.message));
    });
    $("#mfa-cancel").addEventListener("click", hideMFAPanel);
    $("#push-mfa-manual").addEventListener("click", showPushManualCode);
    $("#push-mfa-submit").addEventListener("click", () => verifyPushManualCode().catch(error => setPushMFAStatus(error.message, true)));
    $("#push-mfa-check").addEventListener("click", () => {
      state.mfaPolling = false;
      pollProIdentityMFA(state.pendingMFA).catch(error => setPushMFAStatus(error.message, true));
    });
    $("#push-mfa-cancel").addEventListener("click", hideMFAPanel);
    $("#push-mfa-code").addEventListener("keydown", event => {
      if (event.key === "Enter") {
        event.preventDefault();
        verifyPushManualCode().catch(error => setPushMFAStatus(error.message, true));
      }
    });
    $("#login-form").addEventListener("submit", async event => {
      event.preventDefault();
      const form = event.currentTarget;
      const data = Object.fromEntries(new FormData(form).entries());
      try {
        setLoginStatus("Checking credentials...");
        const response = await fetch("/api/v1/session", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json"}, body: JSON.stringify(data)});
        const body = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(body.error || "Login failed");
        if (body.mfa_required) {
          setLoginStatus("");
          showMFAPanel(body);
          return;
        }
        state.csrf = body.csrf_token || "";
        hideMFAPanel();
        $("#login-cover").classList.add("hidden");
        form.reset();
        await refresh();
        showStatus("Logged in");
      } catch (error) {
        setLoginStatus(error.message, true);
        showStatus(error.message, true);
      }
    });
    render();
    bootstrapSession().catch(error => showStatus(error.message, true));
  </script>
</body>
</html>`
