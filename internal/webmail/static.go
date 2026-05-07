package webmail

const webmailIndexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>ProIdentity Webmail</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Public+Sans:wght@400;600;700&display=swap" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap" rel="stylesheet">
  <style>
    :root {
      --background: #f7f9fb;
      --surface: #ffffff;
      --surface-soft: #f2f4f6;
      --surface-muted: #eceef0;
      --ink: #191c1e;
      --muted: #464554;
      --outline: #c7c4d7;
      --outline-strong: #767586;
      --primary: #4648d4;
      --primary-soft: #e1e0ff;
      --secondary-soft: #dae2fd;
      --danger: #ba1a1a;
      --success: #079455;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      height: 100vh;
      overflow: hidden;
      background: var(--background);
      color: var(--ink);
      font: 13px/1.45 "Public Sans", system-ui, sans-serif;
      letter-spacing: 0;
    }
    body.auth-locked header,
    body.auth-locked .app {
      visibility: hidden;
      pointer-events: none;
    }
    .material-symbols-outlined {
      font-variation-settings: "FILL" 0, "wght" 400, "GRAD" 0, "opsz" 24;
      font-size: 24px;
      line-height: 1;
    }
    header {
      height: 54px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 20px;
      border-bottom: 1px solid var(--outline);
      background: var(--background);
    }
    .brand { display: flex; align-items: center; gap: 12px; color: var(--primary); }
    .brand h1 { margin: 0; font-size: 22px; line-height: 1; font-weight: 700; }
    .top-actions { display: flex; align-items: center; gap: 18px; }
    .search {
      width: min(390px, 34vw);
      min-height: 38px;
      display: flex;
      align-items: center;
      gap: 14px;
      background: #e3e5e7;
      border-radius: 999px;
      padding: 0 16px;
      color: var(--outline-strong);
      font-weight: 600;
      letter-spacing: .08em;
    }
    .search input { width: 100%; border: 0; outline: 0; background: transparent; color: var(--ink); font: inherit; }
    .avatar {
      width: 34px;
      height: 34px;
      border: 0;
      border-radius: 50%;
      background: #5d60f0;
      color: white;
      display: grid;
      place-items: center;
      font-weight: 700;
      cursor: pointer;
    }
    .account-menu {
      position: fixed;
      top: 48px;
      right: 14px;
      z-index: 70;
      width: min(320px, calc(100vw - 28px));
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: white;
      box-shadow: 0 18px 44px rgba(15,23,42,.2);
      padding: 14px;
      display: grid;
      gap: 12px;
    }
    .account-menu.hidden { display: none; }
    .account-menu strong { overflow-wrap: anywhere; }
    .account-actions { display: flex; gap: 8px; justify-content: flex-end; }
    .security-note {
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: #f8f8fc;
      color: var(--muted);
      padding: 12px;
      display: flex;
      gap: 10px;
      align-items: flex-start;
      font-size: 12px;
    }
    .app {
      height: calc(100vh - 54px);
      display: grid;
      grid-template-columns: 240px 340px minmax(400px, 1fr) 54px;
      overflow: hidden;
    }
    body.workspace-wide .app { grid-template-columns: 240px minmax(0, 1fr) 54px; }
    body.workspace-wide .list-pane { display: none; }
    aside {
      background: var(--surface-soft);
      border-right: 1px solid var(--outline);
      display: flex;
      flex-direction: column;
      padding: 18px 12px;
      min-width: 0;
    }
    .compose {
      min-height: 48px;
      border: 0;
      border-radius: 12px;
      background: #5b5ef1;
      color: white;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 10px;
      font: inherit;
      font-weight: 700;
      font-size: 17px;
      cursor: pointer;
      box-shadow: 0 10px 22px rgba(70,72,212,.22);
      margin-bottom: 22px;
    }
    .folder-list { display: grid; gap: 8px; }
    .folder {
      min-height: 40px;
      border: 0;
      border-radius: 999px;
      background: transparent;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      padding: 0 14px;
      color: #272a3a;
      font: inherit;
      font-size: 15px;
      cursor: pointer;
    }
    .folder span:first-child { display: flex; align-items: center; gap: 12px; }
    .folder.active { background: var(--secondary-soft); color: #3f465c; font-weight: 700; }
    .count {
      min-width: 32px;
      border-radius: 999px;
      padding: 2px 10px;
      font-size: 12px;
      font-weight: 700;
    }
    .count.unread { background: var(--primary); color: white; }
    .count.total { background: transparent; color: var(--muted); padding-right: 0; }
    .count.hidden { display: none; }
    .labels {
      margin-top: auto;
      border-top: 1px solid var(--outline);
      padding: 18px 8px 0;
      display: grid;
      gap: 10px;
      color: var(--muted);
    }
    .labels h3 { margin: 0 0 4px; color: var(--outline-strong); font-size: 12px; letter-spacing: .18em; }
    .label { display: flex; align-items: center; gap: 12px; font-size: 15px; }
    .dot { width: 12px; height: 12px; border-radius: 50%; background: var(--primary); }
    .dot.danger { background: #d11c1c; }
    .dot.dark { background: #56627a; }
    .list-pane {
      background: var(--surface);
      border-right: 1px solid var(--outline);
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
    .pane-head {
      min-height: 52px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 16px;
      border-bottom: 1px solid var(--outline);
      background: var(--surface-soft);
    }
    .pane-head h2 { margin: 0; font-size: 19px; }
    .message-list { overflow: auto; flex: 1; }
    .message {
      border: 0;
      border-bottom: 1px solid var(--outline);
      background: var(--surface);
      width: 100%;
      text-align: left;
      padding: 14px 16px;
      cursor: pointer;
      border-left: 4px solid transparent;
      display: grid;
      gap: 5px;
      color: var(--ink);
    }
    .message:hover { background: #fafaff; }
    .message.active { background: rgba(218,226,253,.35); border-left-color: var(--primary); }
    .message.unread .from, .message.unread .subject { font-weight: 800; }
    .message.unread .subject:before {
      content: "";
      display: inline-block;
      width: 7px;
      height: 7px;
      border-radius: 999px;
      background: var(--primary);
      margin-right: 7px;
      vertical-align: 1px;
    }
    .message-top { display: flex; justify-content: space-between; gap: 12px; }
    .from { font-weight: 700; font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .time { color: #6b6880; font-size: 12px; font-weight: 700; letter-spacing: .04em; white-space: nowrap; }
    .subject { font-weight: 700; font-size: 15px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .preview { color: var(--muted); font-size: 13px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
    .tag {
      justify-self: start;
      border-radius: 4px;
      padding: 2px 7px;
      background: rgba(70,72,212,.1);
      border: 1px solid rgba(70,72,212,.24);
      color: var(--primary);
      font-size: 10px;
      font-weight: 700;
      letter-spacing: .06em;
    }
    .reader {
      background: var(--surface);
      overflow: auto;
      display: flex;
      flex-direction: column;
      min-width: 0;
    }
    .toolbar {
      min-height: 50px;
      border-bottom: 1px solid var(--outline);
      display: flex;
      align-items: center;
      gap: 18px;
      padding: 0 22px;
      color: #292b3c;
      background: var(--surface);
    }
    .toolbar.hidden { display: none; }
    .tool-button {
      width: 34px;
      height: 34px;
      border: 0;
      border-radius: 8px;
      background: transparent;
      color: inherit;
      display: grid;
      place-items: center;
      cursor: pointer;
    }
    .tool-button:hover { background: var(--surface-soft); }
    .message-meta { margin-top: 18px; display: grid; gap: 10px; }
    .reader-content { padding: 6px 28px 36px; max-width: 760px; }
    .reader h2 { margin: 0 0 20px; font-size: 28px; line-height: 1.15; }
    .sender-row { display: flex; align-items: center; justify-content: space-between; gap: 20px; margin-bottom: 26px; }
    .sender { display: flex; align-items: center; gap: 12px; }
    .sender-icon {
      width: 40px;
      height: 40px;
      border-radius: 50%;
      background: var(--surface-muted);
      color: var(--primary);
      display: grid;
      place-items: center;
    }
    .sender strong { font-size: 18px; }
    .body { font-size: 16px; line-height: 1.55; }
    .recommend {
      margin: 18px 0;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: var(--surface-muted);
      padding: 18px;
    }
    .recommend h3 { margin: 0 0 12px; color: var(--primary); font-size: 15px; letter-spacing: .08em; }
    .mini-grid { display: grid; gap: 10px; margin-top: 16px; }
    .mini-row {
      border: 1px solid var(--outline);
      border-radius: 8px;
      padding: 12px;
      background: #fbfcff;
      display: flex;
      justify-content: space-between;
      gap: 12px;
      align-items: center;
    }
    .workspace-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 14px; }
    .workspace-head h2 { margin-bottom: 6px; }
    .muted { color: var(--muted); }
    .compact-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
    .secondary-button, .danger-button {
      min-height: 34px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: white;
      color: var(--ink);
      padding: 0 10px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 6px;
      font: inherit;
      font-weight: 700;
      cursor: pointer;
    }
    .secondary-button .material-symbols-outlined, .danger-button .material-symbols-outlined { font-size: 20px; }
    .danger-button { color: var(--danger); border-color: rgba(186,26,26,.32); }
    .editor-toolbar {
      display: flex;
      align-items: center;
      gap: 6px;
      border: 1px solid var(--outline);
      border-bottom: 0;
      border-radius: 8px 8px 0 0;
      background: var(--surface-soft);
      padding: 6px;
    }
    .editor {
      width: 100%;
      min-height: 220px;
      max-height: 46vh;
      overflow: auto;
      border: 1px solid var(--outline);
      border-radius: 0 0 8px 8px;
      padding: 12px;
      font: 15px/1.55 "Public Sans", system-ui, sans-serif;
      outline: 0;
      white-space: pre-wrap;
    }
    .editor:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(70,72,212,.12); }
    .folder-tools { display: grid; gap: 8px; margin-top: 12px; }
    .folder-tools .secondary-button { width: 100%; }
    .pill-row { display: flex; flex-wrap: wrap; gap: 8px; }
    select {
      min-height: 38px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      padding: 8px 12px;
      font: inherit;
      background: white;
    }
    .connect-box {
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: #fbfcff;
      padding: 14px;
      margin: 12px 0 18px;
      display: grid;
      gap: 8px;
    }
    .connect-row { display: grid; grid-template-columns: 90px minmax(0, 1fr); gap: 10px; align-items: center; }
    .connect-row code { overflow-wrap: anywhere; font-size: 12px; }
    .connect-box.hidden { display: none; }
    .workspace-tools { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
    .contact-list, .agenda-list { display: grid; gap: 10px; }
    .contact-card, .event-card {
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: white;
      min-height: 64px;
      padding: 12px 14px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
    }
    .contact-main { display: flex; align-items: center; gap: 12px; min-width: 0; }
    .contact-initials {
      width: 38px;
      height: 38px;
      border-radius: 50%;
      background: var(--secondary-soft);
      color: #30384d;
      display: grid;
      place-items: center;
      font-weight: 800;
      flex: none;
    }
    .calendar-layout {
      display: grid;
      grid-template-columns: minmax(280px, 420px) minmax(320px, 1fr);
      gap: 18px;
      align-items: start;
    }
    .month-card {
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: white;
      padding: 14px;
    }
    .month-grid {
      display: grid;
      grid-template-columns: repeat(7, minmax(0, 1fr));
      gap: 6px;
      margin-top: 12px;
    }
    .month-cell {
      min-height: 42px;
      border: 1px solid var(--outline);
      border-radius: 6px;
      padding: 6px;
      color: var(--muted);
      background: #fbfbff;
      font-size: 12px;
      font-weight: 700;
    }
    .month-cell.has-event { border-color: var(--primary); color: var(--primary); background: var(--primary-soft); }
    .day-name {
      color: var(--outline-strong);
      font-size: 11px;
      font-weight: 800;
      text-align: center;
      text-transform: uppercase;
    }
    .rail {
      border-left: 1px solid var(--outline);
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 24px;
      padding-top: 26px;
      background: var(--surface-soft);
    }
    .rail .bottom { margin-top: auto; display: grid; gap: 24px; padding-bottom: 22px; }
    .login {
      position: fixed;
      inset: 54px 0 0;
      background: rgba(247,249,251,.92);
      backdrop-filter: blur(12px);
      z-index: 30;
      display: grid;
      place-items: center;
      padding: 24px;
    }
    .login.hidden { display: none; }
    .modal {
      position: fixed;
      inset: auto 28px 28px auto;
      width: min(520px, calc(100vw - 56px));
      z-index: 35;
      border: 1px solid var(--outline);
      border-radius: 12px;
      background: white;
      box-shadow: 0 20px 48px rgba(15,23,42,.18);
      display: grid;
      gap: 12px;
      padding: 18px;
    }
    .modal.hidden { display: none; }
    .modal-head { display: flex; justify-content: space-between; align-items: center; }
    .modal-head h2 { margin: 0; font-size: 18px; }
    .modal .modal-actions { display: flex; justify-content: flex-end; gap: 8px; }
    .compose-backdrop {
      position: fixed;
      inset: 0;
      z-index: 34;
      background: rgba(25, 28, 30, .18);
      backdrop-filter: blur(8px);
    }
    .compose-backdrop.hidden { display: none; }
    #compose-modal {
      inset: 72px 72px 34px 300px;
      width: auto;
      min-width: min(920px, calc(100vw - 40px));
      max-width: 1120px;
      margin: 0 auto;
      padding: 0;
      gap: 0;
      border-radius: 12px;
      overflow: hidden;
      grid-template-rows: auto minmax(0, 1fr) auto;
    }
    #compose-modal.expanded { inset: 24px; max-width: none; }
    .compose-head {
      min-height: 74px;
      padding: 0 24px;
      border-bottom: 1px solid var(--outline);
      background: var(--surface-soft);
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    .compose-title { display: flex; align-items: center; gap: 14px; }
    .compose-title h2 { margin: 0; font-size: 20px; }
    .compose-body { display: flex; min-height: 0; flex-direction: column; background: white; }
    .compose-fields { padding: 18px 28px 8px; display: grid; gap: 0; }
    .compose-row {
      min-height: 48px;
      border-bottom: 1px solid rgba(199,196,215,.58);
      display: flex;
      align-items: center;
      gap: 14px;
    }
    .compose-row-label {
      width: 54px;
      color: var(--on-surface-variant, var(--muted));
      font-weight: 700;
      font-size: 14px;
      letter-spacing: .02em;
    }
    .compose-row input {
      border: 0;
      outline: 0;
      min-height: 38px;
      background: transparent;
      box-shadow: none;
      flex: 1;
      padding: 0;
      font-size: 15px;
    }
    .compose-row input:focus { outline: 0; box-shadow: none; border-color: transparent; }
    .recipient-wrap { flex: 1; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; min-width: 0; }
    .recipient-chip {
      min-height: 28px;
      border-radius: 999px;
      background: var(--secondary-soft);
      color: #30384d;
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 0 9px 0 12px;
      font-size: 13px;
      font-weight: 700;
    }
    .recipient-chip button {
      border: 0;
      background: transparent;
      color: inherit;
      width: 18px;
      height: 18px;
      padding: 0;
      display: grid;
      place-items: center;
      cursor: pointer;
    }
    .recipient-chip .material-symbols-outlined { font-size: 16px; }
    .compose-field-toggle {
      border: 0;
      background: transparent;
      color: var(--primary);
      font-weight: 700;
      cursor: pointer;
      padding: 4px 2px;
    }
    .compose-field-toggle:hover { text-decoration: underline; }
    .compose-row.optional.hidden { display: none; }
    .compose-toolbar {
      min-height: 58px;
      padding: 8px 28px;
      background: var(--surface-soft);
      border-bottom: 1px solid rgba(199,196,215,.52);
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .compose-toolbar-group {
      display: flex;
      align-items: center;
      gap: 4px;
      padding-right: 10px;
      border-right: 1px solid var(--outline);
    }
    .compose-toolbar-group:last-child { border-right: 0; padding-right: 0; }
    .compose-editor-shell { flex: 1; min-height: 0; overflow: auto; padding: 24px 30px; }
    #compose-editor {
      min-height: 100%;
      max-height: none;
      border: 0;
      border-radius: 0;
      padding: 0;
      box-shadow: none;
      color: var(--ink);
      font-size: 17px;
      line-height: 1.65;
      white-space: pre-wrap;
    }
    #compose-editor:focus { border-color: transparent; box-shadow: none; }
    #compose-editor:empty:before {
      content: attr(data-placeholder);
      color: #8b8da0;
      pointer-events: none;
    }
    .compose-attachments {
      padding: 14px 28px;
      border-top: 1px solid var(--outline);
      background: var(--surface-soft);
      display: flex;
      gap: 12px;
      flex-wrap: wrap;
      align-items: center;
    }
    .attachment-chip {
      min-height: 54px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: white;
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 9px 12px;
      color: var(--muted);
      font-weight: 700;
    }
    .attachment-chip.disabled {
      border-style: dashed;
      opacity: .78;
    }
    .compose-footer {
      min-height: 78px;
      padding: 14px 28px;
      border-top: 1px solid var(--outline);
      background: var(--surface-soft);
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 18px;
    }
    .compose-footer-left, .compose-footer-right { display: flex; align-items: center; gap: 12px; }
    .send-button {
      min-height: 46px;
      border: 0;
      border-radius: 8px;
      background: var(--primary);
      color: white;
      padding: 0 28px;
      display: inline-flex;
      align-items: center;
      gap: 14px;
      font-weight: 700;
      cursor: pointer;
      box-shadow: 0 10px 22px rgba(70,72,212,.24);
    }
    .send-button:hover { background: #3032bd; }
    .security-compose-badge {
      position: fixed;
      right: 26px;
      bottom: 96px;
      z-index: 36;
      border: 1px solid var(--primary);
      border-radius: 999px;
      background: rgba(255,255,255,.92);
      color: var(--primary);
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 10px 18px;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: .14em;
      text-transform: uppercase;
      box-shadow: 0 12px 24px rgba(15,23,42,.12);
      pointer-events: none;
    }
    .security-compose-badge.hidden { display: none; }
    .status-dot { width: 8px; height: 8px; border-radius: 999px; background: var(--primary); }
    textarea {
      width: 100%;
      min-height: 150px;
      resize: vertical;
      border: 1px solid var(--outline);
      border-radius: 8px;
      padding: 10px 12px;
      font: inherit;
    }
    .login-card {
      width: min(460px, 100%);
      border: 1px solid var(--outline);
      border-radius: 16px;
      background: white;
      box-shadow: 0 20px 48px rgba(15,23,42,.12);
      padding: 24px;
      display: grid;
      gap: 16px;
    }
    .login-card h2 { margin: 0; font-size: 24px; }
    label { display: grid; gap: 7px; color: var(--muted); font-size: 12px; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; }
    input {
      min-height: 38px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      padding: 8px 12px;
      font: inherit;
    }
    input:focus, select:focus { outline: 2px solid rgba(70,72,212,.18); border-color: var(--primary); }
    .primary-button {
      min-height: 40px;
      border: 0;
      border-radius: 8px;
      background: var(--primary);
      color: white;
      font: inherit;
      font-weight: 700;
      cursor: pointer;
    }
    .primary-button:disabled { opacity: .65; cursor: wait; }
    .error { color: var(--danger); min-height: 20px; }
    .error.info { color: var(--muted); }
    .toast {
      position: fixed;
      right: 22px;
      bottom: 22px;
      z-index: 80;
      max-width: min(360px, calc(100vw - 44px));
      border: 1px solid rgba(70,72,212,.24);
      border-radius: 8px;
      background: #ffffff;
      color: var(--ink);
      box-shadow: 0 18px 38px rgba(15,23,42,.18);
      padding: 12px 14px;
      font-weight: 700;
    }
    .toast.error-toast {
      border-color: rgba(186,26,26,.32);
      color: var(--danger);
    }
    .toast.hidden { display: none; }
    @media (max-width: 1100px) {
      body { overflow: auto; height: auto; }
      .app { height: auto; grid-template-columns: 1fr; }
      aside, .list-pane, .reader, .rail { min-height: auto; }
      .rail { display: none; }
      .search { width: 44vw; }
      .calendar-layout { grid-template-columns: 1fr; }
      #compose-modal { inset: 12px; min-width: 0; }
      .compose-head, .compose-fields, .compose-toolbar, .compose-editor-shell, .compose-attachments, .compose-footer { padding-left: 16px; padding-right: 16px; }
      .compose-footer { align-items: stretch; flex-direction: column; }
      .compose-footer-left, .compose-footer-right { width: 100%; justify-content: space-between; }
      .send-button { flex: 1; justify-content: center; }
      .security-compose-badge { display: none; }
    }
  </style>
</head>
<body class="auth-locked">
  <header>
    <div class="brand"><span class="material-symbols-outlined">shield</span><h1>ProIdentity Mail</h1></div>
    <div class="top-actions">
      <div class="search"><span class="material-symbols-outlined">search</span><input id="search" placeholder="Search emails..."></div>
      <button class="tool-button" type="button" id="refresh" title="Refresh"><span class="material-symbols-outlined">refresh</span></button>
      <button class="avatar" type="button" id="avatar" title="Account">--</button>
    </div>
  </header>
  <div class="account-menu hidden" id="account-menu">
    <div>
      <div class="muted">Signed in as</div>
      <strong id="account-email">--</strong>
    </div>
    <div class="security-note"><span class="material-symbols-outlined">info</span><span>Mailbox security badges are shown only when message authentication data exists for the selected message.</span></div>
    <div class="account-actions">
      <button class="secondary-button" type="button" id="account-close">Close</button>
      <button class="danger-button" type="button" id="account-logout"><span class="material-symbols-outlined">logout</span>Logout</button>
    </div>
  </div>

  <div class="app">
    <aside>
      <button class="compose" type="button"><span class="material-symbols-outlined">edit</span>Compose</button>
      <nav class="folder-list" id="folder-list"></nav>
      <div class="folder-tools">
        <button class="secondary-button" type="button" id="add-folder"><span class="material-symbols-outlined">create_new_folder</span>New folder</button>
        <button class="secondary-button" type="button" id="open-filters"><span class="material-symbols-outlined">filter_alt</span>Filters</button>
      </div>
      <div class="labels">
        <h3>Mail Tools</h3>
        <div class="label"><span class="dot danger"></span>Spam training</div>
        <div class="label"><span class="dot"></span>Custom folders</div>
        <div class="label"><span class="dot dark"></span>Server filters</div>
      </div>
    </aside>

    <section class="list-pane">
      <div class="pane-head"><h2>Inbox</h2><span class="material-symbols-outlined">filter_list</span></div>
      <div class="message-list" id="messages"></div>
    </section>

    <section class="reader">
      <div class="toolbar">
        <button class="tool-button" type="button" id="archive-message" title="Archive"><span class="material-symbols-outlined">archive</span></button>
        <button class="tool-button" type="button" id="mark-spam" title="Mark as spam"><span class="material-symbols-outlined">report</span></button>
        <button class="tool-button" type="button" id="mark-ham" title="Mark as not spam"><span class="material-symbols-outlined">verified</span></button>
        <button class="tool-button" type="button" id="trash-message" title="Delete"><span class="material-symbols-outlined">delete</span></button>
        <span style="height:24px;width:1px;background:var(--outline)"></span>
        <button class="tool-button" type="button" id="reply-message" title="Reply"><span class="material-symbols-outlined">reply</span></button>
        <button class="tool-button" type="button" id="reply-all-message" title="Reply all"><span class="material-symbols-outlined">reply_all</span></button>
        <button class="tool-button" type="button" id="forward-message" title="Forward"><span class="material-symbols-outlined">forward</span></button>
      </div>
      <article class="reader-content" id="reader">
        <h2>Select a message</h2>
        <div class="body">Load your mailbox, then choose a message from the inbox list to inspect the sender, subject, and preview.</div>
      </article>
    </section>

    <aside class="rail">
      <button class="tool-button" type="button" id="open-calendar" title="Calendar"><span class="material-symbols-outlined">calendar_month</span></button>
      <button class="tool-button" type="button" id="open-filters-rail" title="Filters"><span class="material-symbols-outlined">filter_alt</span></button>
      <button class="tool-button" type="button" id="open-contacts" title="Contacts"><span class="material-symbols-outlined">contacts</span></button>
      <div class="bottom">
        <button class="tool-button" type="button" id="open-folders-rail" title="Folders"><span class="material-symbols-outlined">folder_managed</span></button>
        <button class="tool-button" type="button" id="logout" title="Logout"><span class="material-symbols-outlined">logout</span></button>
      </div>
    </aside>
  </div>

  <div class="login" id="login-panel">
    <form class="login-card" id="login">
      <div class="brand"><span class="material-symbols-outlined">shield</span><h1>ProIdentity Mail</h1></div>
      <h2>Secure mailbox login</h2>
      <label>Email<input name="email" autocomplete="username" required></label>
      <label>Password<input name="password" type="password" autocomplete="current-password" required></label>
      <button class="primary-button" type="submit">Load Mailbox</button>
      <div class="error" id="error"></div>
    </form>
  </div>

  <div class="compose-backdrop hidden" id="compose-backdrop"></div>
  <form class="modal hidden" id="compose-modal">
    <div class="compose-head">
      <div class="compose-title"><span class="material-symbols-outlined" style="color:var(--primary)">edit_note</span><h2>New Message</h2></div>
      <div class="compact-actions">
        <button class="tool-button" type="button" id="expand-compose" title="Expand"><span class="material-symbols-outlined">open_in_full</span></button>
        <button class="tool-button" type="button" id="close-compose" title="Close"><span class="material-symbols-outlined">close</span></button>
      </div>
    </div>
    <div class="compose-body">
      <div class="compose-fields">
        <div class="compose-row">
          <span class="compose-row-label">To</span>
          <div class="recipient-wrap" id="to-chip-row">
            <input name="to" autocomplete="email" placeholder="Add recipients...">
          </div>
          <button class="compose-field-toggle" type="button" data-show-compose-field="cc-row">Cc</button>
          <button class="compose-field-toggle" type="button" data-show-compose-field="bcc-row">Bcc</button>
        </div>
        <div class="compose-row optional hidden" id="cc-row">
          <span class="compose-row-label">Cc</span>
          <input name="cc" autocomplete="email" placeholder="Add carbon copy recipients...">
        </div>
        <div class="compose-row optional hidden" id="bcc-row">
          <span class="compose-row-label">Bcc</span>
          <input name="bcc" autocomplete="email" placeholder="Add blind copy recipients...">
        </div>
        <div class="compose-row">
          <span class="compose-row-label">Subject</span>
          <input name="subject" placeholder="Enter message subject" required>
        </div>
      </div>
      <div class="compose-toolbar">
        <div class="compose-toolbar-group">
          <button class="tool-button" type="button" data-editor-command="bold" title="Bold"><span class="material-symbols-outlined">format_bold</span></button>
          <button class="tool-button" type="button" data-editor-command="italic" title="Italic"><span class="material-symbols-outlined">format_italic</span></button>
          <button class="tool-button" type="button" data-editor-command="underline" title="Underline"><span class="material-symbols-outlined">format_underlined</span></button>
        </div>
        <div class="compose-toolbar-group">
          <button class="tool-button" type="button" data-editor-command="insertUnorderedList" title="Bullet list"><span class="material-symbols-outlined">format_list_bulleted</span></button>
          <button class="tool-button" type="button" data-editor-command="insertOrderedList" title="Numbered list"><span class="material-symbols-outlined">format_list_numbered</span></button>
        </div>
        <div class="compose-toolbar-group">
          <button class="tool-button" type="button" data-editor-link title="Link"><span class="material-symbols-outlined">link</span></button>
          <button class="tool-button" type="button" data-editor-image title="Image URL"><span class="material-symbols-outlined">image</span></button>
        </div>
        <div class="compose-toolbar-group">
          <button class="tool-button" type="button" data-editor-clear title="Clear formatting"><span class="material-symbols-outlined">format_clear</span></button>
        </div>
        <div style="flex:1"></div>
      </div>
      <div class="compose-editor-shell">
        <div class="editor" id="compose-editor" contenteditable="true" spellcheck="true" data-placeholder="Write your message..."></div>
        <input type="hidden" name="body">
      </div>
      <div class="compose-attachments">
        <button class="attachment-chip disabled" type="button" disabled title="Attachment sending is not implemented yet"><span class="material-symbols-outlined">add_circle</span><span>Add file<br><span class="muted">coming soon</span></span></button>
      </div>
    </div>
    <div class="compose-footer">
      <div class="compose-footer-left">
        <button class="send-button" type="submit"><span>Send Message</span><span class="material-symbols-outlined">send</span></button>
        <button class="secondary-button" type="button" id="save-draft"><span class="material-symbols-outlined">save</span>Save Draft</button>
      </div>
      <div class="compose-footer-right">
        <div class="error" id="compose-error"></div>
        <button class="danger-button" type="button" id="discard-compose" title="Discard"><span class="material-symbols-outlined">delete</span></button>
      </div>
    </div>
  </form>
  <div class="security-compose-badge hidden" id="compose-security-badge"><span class="status-dot"></span><span>Encrypted Session</span><span class="material-symbols-outlined">verified_user</span></div>
  <div class="toast hidden" id="toast"></div>

  <form class="modal hidden" id="folder-modal">
    <div class="modal-head"><h2>New folder</h2><button class="tool-button" type="button" id="close-folder" title="Close"><span class="material-symbols-outlined">close</span></button></div>
    <label>Folder name<input name="name" placeholder="Projects" required></label>
    <div class="modal-actions">
      <button class="secondary-button" type="button" id="cancel-folder">Cancel</button>
      <button class="primary-button" type="submit">Create Folder</button>
    </div>
    <div class="error" id="folder-error"></div>
  </form>

  <form class="modal hidden" id="filter-modal">
    <div class="modal-head"><h2 id="filter-title">Mail filter</h2><button class="tool-button" type="button" id="close-filter" title="Close"><span class="material-symbols-outlined">close</span></button></div>
    <input type="hidden" name="id">
    <label>Name<input name="name" placeholder="Move invoices" required></label>
    <label>Field<select name="field"><option value="from">From</option><option value="to">To</option><option value="subject" selected>Subject</option><option value="body">Body</option></select></label>
    <label>Match<select name="operator"><option value="contains">Contains</option><option value="equals">Equals</option><option value="starts_with">Starts with</option><option value="ends_with">Ends with</option></select></label>
    <label>Value<input name="value" placeholder="invoice" required></label>
    <label>Action<select name="action"><option value="move">Move to folder</option><option value="mark_spam">Mark spam</option><option value="delete">Delete</option><option value="keep">Keep in inbox</option></select></label>
    <label>Destination folder<select name="folder" id="filter-folder"></select></label>
    <label><span><input name="enabled" type="checkbox" checked> Enabled</span></label>
    <div class="modal-actions">
      <button class="secondary-button" type="button" id="cancel-filter">Cancel</button>
      <button class="primary-button" type="submit">Save Filter</button>
    </div>
    <div class="error" id="filter-error"></div>
  </form>

  <form class="modal hidden" id="contact-modal">
    <div class="modal-head"><h2 id="contact-title">Contact</h2><button class="tool-button" type="button" id="close-contact" title="Close"><span class="material-symbols-outlined">close</span></button></div>
    <input type="hidden" name="id">
    <label>Name<input name="name" autocomplete="name" required></label>
    <label>Email<input name="email" type="email" autocomplete="email" required></label>
    <div class="modal-actions">
      <button class="secondary-button" type="button" id="cancel-contact">Cancel</button>
      <button class="primary-button" type="submit">Save Contact</button>
    </div>
    <div class="error" id="contact-error"></div>
  </form>

  <form class="modal hidden" id="event-modal">
    <div class="modal-head"><h2 id="event-title">Calendar Event</h2><button class="tool-button" type="button" id="close-event" title="Close"><span class="material-symbols-outlined">close</span></button></div>
    <input type="hidden" name="id">
    <label>Title<input name="title" required></label>
    <label>Starts<input name="starts_at" type="datetime-local" required></label>
    <label>Ends<input name="ends_at" type="datetime-local" required></label>
    <div class="modal-actions">
      <button class="secondary-button" type="button" id="cancel-event">Cancel</button>
      <button class="primary-button" type="submit">Save Event</button>
    </div>
    <div class="error" id="event-error"></div>
  </form>

  <script>
    const state = { csrf: "", email: "", messages: [], selected: null, folder: "inbox", folders: [], filters: [], contacts: [], events: [], view: "mail" };
    let toastTimer = null;
    const esc = value => String(value ?? "").replace(/[&<>"']/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"}[char]));
    const initials = email => String(email || "--").split("@")[0].split(/[._-]+/).filter(Boolean).slice(0, 2).map(part => part[0]).join("").toUpperCase() || "--";
    const messageTime = item => item.date ? new Date(item.date).toLocaleString([], {month: "short", day: "numeric", hour: "2-digit", minute: "2-digit"}) : "";
    const dateKey = value => {
      const date = value ? new Date(value) : new Date();
      if (Number.isNaN(date.getTime())) return "";
      const month = String(date.getMonth() + 1).padStart(2, "0");
      const day = String(date.getDate()).padStart(2, "0");
      return date.getFullYear() + "-" + month + "-" + day;
    };
    const dateTimeLocal = value => {
      const date = value ? new Date(value) : new Date(Date.now() + 3600000);
      const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
      return local.toISOString().slice(0, 16);
    };
    const shortFrom = value => String(value || "Unknown").replace(/<.*>/, "").replace(/"/g, "").trim() || "Unknown";
    const serviceBase = () => location.origin.replace(/^http:/, "https:");
    const emailOnly = value => {
      const match = String(value || "").match(/<([^>]+)>/);
      return (match ? match[1] : String(value || "")).replace(/"/g, "").trim();
    };
    const prefixedSubject = (prefix, subject) => {
      const text = String(subject || "");
      return text.toLowerCase().startsWith(prefix.toLowerCase()) ? text : prefix + text;
    };
    const setAuthenticated = value => {
      document.body.classList.toggle("auth-locked", !value);
      document.querySelector("#account-menu").classList.add("hidden");
    };
    const folderCountHTML = folder => {
      const unread = Number(folder.unread || 0);
      const total = Number(folder.total || 0);
      if (unread > 0) return "<span class=\"count unread\">" + esc(unread) + "</span>";
      if (total > 0) return "<span class=\"count total\">" + esc(total) + "</span>";
      return "<span class=\"count hidden\"></span>";
    };
    const renderCurrentView = () => {
      if (state.view === "mail") render();
      else if (state.view === "contacts") renderContactsView();
      else if (state.view === "calendar") renderCalendarView();
      else if (state.view === "filters") renderFiltersView();
    };
    const updateViewChrome = (placeholder = "Search emails...") => {
      document.querySelector(".toolbar").classList.toggle("hidden", state.view !== "mail");
      document.body.classList.toggle("workspace-wide", state.view !== "mail");
      document.querySelector("#search").placeholder = placeholder;
    };
    const api = async (path, options = {}) => {
      const response = await fetch(path, {credentials: "same-origin", cache: "no-store", ...options, headers: {"Content-Type": "application/json", ...(state.csrf ? {"X-CSRF-Token": state.csrf} : {}), ...(options.headers || {})}});
      if (!response.ok) {
        let message = "Request failed";
        try { message = (await response.json()).error || message; } catch {}
        throw new Error(message);
      }
      if (response.status === 204) return null;
      return response.json();
    };
    function showToast(message, isError = false) {
      const box = document.querySelector("#toast");
      clearTimeout(toastTimer);
      box.textContent = message;
      box.className = "toast" + (isError ? " error-toast" : "");
      toastTimer = setTimeout(() => box.classList.add("hidden"), 3600);
    }
    async function loadMessages() {
      state.view = "mail";
      await loadFolders();
      const response = await fetch("/api/v1/messages?limit=100&folder=" + encodeURIComponent(state.folder), { credentials: "same-origin", cache: "no-store" });
      if (!response.ok) {
        document.querySelector("#login-panel").classList.remove("hidden");
        throw new Error("Mailbox authentication failed");
      }
      state.messages = await response.json();
      state.selected = state.messages[0] || null;
      render();
    }
    async function loadFolders() {
      try {
        state.folders = await api("/api/v1/folders");
      } catch {
        state.folders = [
          {id: "inbox", name: "Inbox", system: true, total: 0},
          {id: "archive", name: "Archive", system: true, total: 0},
          {id: "spam", name: "Spam", system: true, total: 0},
          {id: "trash", name: "Trash", system: true, total: 0}
        ];
      }
      renderFolders();
    }
    async function bootstrapSession() {
      const response = await fetch("/api/v1/session", {credentials: "same-origin", cache: "no-store"});
      if (!response.ok) {
        setAuthenticated(false);
        document.querySelector("#login-panel").classList.remove("hidden");
        return;
      }
      const body = await response.json();
      state.csrf = body.csrf_token || "";
      state.email = body.email || "";
      setAuthenticated(true);
      document.querySelector("#login-panel").classList.add("hidden");
      await loadMessages();
    }
    async function moveSelected(folder) {
      if (!state.selected) {
        showToast("Select a message first", true);
        return;
      }
      const response = await fetch("/api/v1/messages/" + encodeURIComponent(state.selected.id) + "/move", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({folder})});
      if (!response.ok) throw new Error("Move failed");
      await loadMessages();
      showToast(folder === "trash" ? "Message moved to trash" : "Message moved to " + folder);
    }
    function folderIcon(folder) {
      const id = String(folder.id || "").toLowerCase();
      if (id === "inbox") return "inbox";
      if (id === "sent") return "send";
      if (id === "archive") return "archive";
      if (id === "trash") return "delete";
      if (id === "spam") return "report";
      return "folder";
    }
    function renderFolders() {
      const folders = state.folders.length ? state.folders : [{id: "inbox", name: "Inbox", system: true, total: state.messages.length}];
      document.querySelector("#folder-list").innerHTML = folders.map(folder =>
        "<button class=\"folder " + (String(folder.id) === String(state.folder) ? "active" : "") + "\" data-folder=\"" + esc(folder.id) + "\"><span><span class=\"material-symbols-outlined\">" + folderIcon(folder) + "</span>" + esc(folder.name) + "</span>" + folderCountHTML(folder) + "</button>"
      ).join("");
      document.querySelectorAll("[data-folder]").forEach(item => item.addEventListener("click", async () => {
        state.folder = item.dataset.folder;
        state.selected = null;
        await loadMessages();
      }));
    }
    function folderOptions(current) {
      return state.folders.filter(folder => folder.id !== "trash").map(folder => "<option value=\"" + esc(folder.id) + "\" " + (String(folder.id) === String(current) ? "selected" : "") + ">" + esc(folder.name) + "</option>").join("");
    }
    async function saveFolder(event) {
      event.preventDefault();
      const form = event.currentTarget;
      document.querySelector("#folder-error").textContent = "";
      try {
        const created = await api("/api/v1/folders", {method: "POST", body: JSON.stringify({name: form.elements.name.value.trim()})});
        state.folder = created.id;
        form.reset();
        form.classList.add("hidden");
        await loadMessages();
      } catch (error) {
        document.querySelector("#folder-error").textContent = error.message;
      }
    }
    async function deleteFolder(name) {
      if (!confirm("Delete folder " + name + "? Messages in it will be removed from this folder.")) return;
      await api("/api/v1/folders/" + encodeURIComponent(name), {method: "DELETE"});
      state.folder = "inbox";
      await loadMessages();
    }
    async function loadFiltersView() {
      state.view = "filters";
      await loadFolders();
      state.filters = await api("/api/v1/filters");
      renderFiltersView();
    }
    function renderFiltersView() {
      updateViewChrome("Search filters...");
      document.querySelector("#reader").innerHTML =
        "<div class=\"workspace-head\"><div><h2>Filters</h2><div class=\"muted\">Rules saved for this mailbox. Delivery-time execution is the next mail pipeline step.</div></div><button class=\"primary-button\" id=\"add-filter\" type=\"button\">Add Filter</button></div>" +
        "<div class=\"mini-grid\">" + (state.filters.length ? state.filters.map(item => "<div class=\"mini-row\"><div><strong>" + esc(item.name) + "</strong><div class=\"muted\">" + esc(item.field) + " " + esc(item.operator) + " \"" + esc(item.value) + "\" -> " + esc(item.action) + (item.folder ? " " + esc(item.folder) : "") + "</div></div><div class=\"compact-actions\">" + (item.enabled ? "<span class=\"tag\">ENABLED</span>" : "<span class=\"tag\">OFF</span>") + "<button class=\"secondary-button\" data-edit-filter=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">edit</span>Edit</button><button class=\"danger-button\" data-delete-filter=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">delete</span>Delete</button></div></div>").join("") : "<div class=\"mini-row\"><div><strong>No filters yet</strong><div class=\"muted\">Create rules for sender, recipient, subject, or body matching.</div></div></div>") + "</div>";
      document.querySelector("#add-filter").addEventListener("click", () => openFilterModal());
    }
    function openFilterModal(filter = {}) {
      const form = document.querySelector("#filter-modal");
      form.reset();
      form.elements.id.value = filter.id || "";
      form.elements.name.value = filter.name || "";
      form.elements.field.value = filter.field || "subject";
      form.elements.operator.value = filter.operator || "contains";
      form.elements.value.value = filter.value || "";
      form.elements.action.value = filter.action || "move";
      document.querySelector("#filter-folder").innerHTML = folderOptions(filter.folder || state.folder || "inbox");
      form.elements.folder.value = filter.folder || state.folder || "inbox";
      form.elements.enabled.checked = filter.enabled !== false;
      document.querySelector("#filter-title").textContent = filter.id ? "Edit Filter" : "Add Filter";
      document.querySelector("#filter-error").textContent = "";
      form.classList.remove("hidden");
    }
    async function saveFilter(event) {
      event.preventDefault();
      const form = event.currentTarget;
      const id = form.elements.id.value;
      const payload = {
        name: form.elements.name.value.trim(),
        field: form.elements.field.value,
        operator: form.elements.operator.value,
        value: form.elements.value.value.trim(),
        action: form.elements.action.value,
        folder: form.elements.folder.value,
        enabled: form.elements.enabled.checked
      };
      try {
        if (id) await api("/api/v1/filters/" + encodeURIComponent(id), {method: "PUT", body: JSON.stringify(payload)});
        else await api("/api/v1/filters", {method: "POST", body: JSON.stringify(payload)});
        form.classList.add("hidden");
        await loadFiltersView();
      } catch (error) {
        document.querySelector("#filter-error").textContent = error.message;
      }
    }
    async function deleteFilter(id) {
      await api("/api/v1/filters/" + encodeURIComponent(id), {method: "DELETE"});
      await loadFiltersView();
    }
    async function loadContactsView() {
      state.view = "contacts";
      state.contacts = await api("/api/v1/contacts");
      renderContactsView();
    }
    function filteredContacts() {
      const q = document.querySelector("#search").value.trim().toLowerCase();
      if (!q) return state.contacts;
      return state.contacts.filter(item => [item.name, item.email].some(value => String(value || "").toLowerCase().includes(q)));
    }
    function renderContactsView() {
      updateViewChrome("Search contacts...");
      const carddav = serviceBase() + "/dav/addressbooks/" + encodeURIComponent(state.email) + "/default/";
      const rows = filteredContacts();
      document.querySelector("#reader").innerHTML =
        "<div class=\"workspace-head\"><div><h2>Contacts</h2><div class=\"muted\">People available to webmail and CardDAV clients.</div></div><div class=\"workspace-tools\"><button class=\"secondary-button\" id=\"toggle-contact-sync\" type=\"button\"><span class=\"material-symbols-outlined\">settings_ethernet</span>Sync info</button><button class=\"primary-button\" id=\"add-contact\" type=\"button\">Add Contact</button></div></div>" +
        "<div class=\"connect-box hidden\" id=\"contact-sync-info\"><strong>Phone contact source</strong><div class=\"connect-row\"><span class=\"muted\">Server</span><code>" + esc(carddav) + "</code></div><div class=\"connect-row\"><span class=\"muted\">Username</span><code>" + esc(state.email) + "</code></div></div>" +
        "<div class=\"contact-list\">" + (rows.length ? rows.map(item => "<div class=\"contact-card\"><div class=\"contact-main\"><div class=\"contact-initials\">" + esc(initials(item.name || item.email)) + "</div><div><strong>" + esc(item.name) + "</strong><div class=\"muted\">" + esc(item.email) + "</div></div></div><div class=\"compact-actions\"><button class=\"secondary-button\" data-edit-contact=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">edit</span>Edit</button><button class=\"danger-button\" data-delete-contact=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">delete</span>Delete</button></div></div>").join("") : "<div class=\"mini-row\"><div><strong>No contacts found</strong><div class=\"muted\">Add a contact or adjust search.</div></div></div>") + "</div>";
      document.querySelector("#add-contact").addEventListener("click", () => openContactModal());
      document.querySelector("#toggle-contact-sync").addEventListener("click", () => document.querySelector("#contact-sync-info").classList.toggle("hidden"));
    }
    function openContactModal(contact = {}) {
      const form = document.querySelector("#contact-modal");
      form.reset();
      form.elements.id.value = contact.id || "";
      form.elements.name.value = contact.name || "";
      form.elements.email.value = contact.email || "";
      document.querySelector("#contact-title").textContent = contact.id ? "Edit Contact" : "Add Contact";
      document.querySelector("#contact-error").textContent = "";
      form.classList.remove("hidden");
    }
    async function saveContact(event) {
      event.preventDefault();
      const form = event.currentTarget;
      const id = form.elements.id.value;
      const payload = {name: form.elements.name.value.trim(), email: form.elements.email.value.trim()};
      try {
        if (id) await api("/api/v1/contacts/" + encodeURIComponent(id), {method: "PUT", body: JSON.stringify(payload)});
        else await api("/api/v1/contacts", {method: "POST", body: JSON.stringify(payload)});
        form.classList.add("hidden");
        await loadContactsView();
      } catch (error) {
        document.querySelector("#contact-error").textContent = error.message;
      }
    }
    async function deleteContact(id) {
      await api("/api/v1/contacts/" + encodeURIComponent(id), {method: "DELETE"});
      await loadContactsView();
    }
    async function loadCalendarView() {
      state.view = "calendar";
      state.events = await api("/api/v1/calendar");
      renderCalendarView();
    }
    function filteredEvents() {
      const q = document.querySelector("#search").value.trim().toLowerCase();
      if (!q) return state.events;
      return state.events.filter(item => [item.title, item.starts_at, item.ends_at].some(value => String(value || "").toLowerCase().includes(q)));
    }
    function renderMonthGrid(events) {
      const today = new Date();
      const first = new Date(today.getFullYear(), today.getMonth(), 1);
      const last = new Date(today.getFullYear(), today.getMonth() + 1, 0);
      const eventDays = new Set(events.map(item => dateKey(item.starts_at)));
      const cells = ["Sun","Mon","Tue","Wed","Thu","Fri","Sat"].map(day => "<div class=\"day-name\">" + day + "</div>");
      for (let i = 0; i < first.getDay(); i++) cells.push("<div></div>");
      for (let day = 1; day <= last.getDate(); day++) {
        const key = dateKey(new Date(today.getFullYear(), today.getMonth(), day));
        cells.push("<div class=\"month-cell " + (eventDays.has(key) ? "has-event" : "") + "\">" + day + "</div>");
      }
      return "<div class=\"month-card\"><strong>" + esc(today.toLocaleString([], {month: "long", year: "numeric"})) + "</strong><div class=\"month-grid\">" + cells.join("") + "</div></div>";
    }
    function renderCalendarView() {
      updateViewChrome("Search calendar...");
      const caldav = serviceBase() + "/dav/calendars/" + encodeURIComponent(state.email) + "/default/";
      const rows = filteredEvents().sort((a, b) => new Date(a.starts_at) - new Date(b.starts_at));
      document.querySelector("#reader").innerHTML =
        "<div class=\"workspace-head\"><div><h2>Calendar</h2><div class=\"muted\">Month view and agenda for CalDAV events.</div></div><div class=\"workspace-tools\"><button class=\"secondary-button\" id=\"toggle-calendar-sync\" type=\"button\"><span class=\"material-symbols-outlined\">settings_ethernet</span>Sync info</button><button class=\"primary-button\" id=\"add-event\" type=\"button\">Add Event</button></div></div>" +
        "<div class=\"connect-box hidden\" id=\"calendar-sync-info\"><strong>Phone calendar source</strong><div class=\"connect-row\"><span class=\"muted\">Server</span><code>" + esc(caldav) + "</code></div><div class=\"connect-row\"><span class=\"muted\">Username</span><code>" + esc(state.email) + "</code></div></div>" +
        "<div class=\"calendar-layout\">" + renderMonthGrid(rows) + "<div class=\"agenda-list\">" + (rows.length ? rows.map(item => "<div class=\"event-card\"><div><strong>" + esc(item.title) + "</strong><div class=\"muted\">" + esc(new Date(item.starts_at).toLocaleString()) + " - " + esc(new Date(item.ends_at).toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'})) + "</div></div><div class=\"compact-actions\"><button class=\"secondary-button\" data-edit-event=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">edit</span>Edit</button><button class=\"danger-button\" data-delete-event=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">delete</span>Delete</button></div></div>").join("") : "<div class=\"mini-row\"><div><strong>No events found</strong><div class=\"muted\">Create an event or adjust search.</div></div></div>") + "</div></div>";
      document.querySelector("#add-event").addEventListener("click", () => openEventModal());
      document.querySelector("#toggle-calendar-sync").addEventListener("click", () => document.querySelector("#calendar-sync-info").classList.toggle("hidden"));
    }
    function openEventModal(item = {}) {
      const form = document.querySelector("#event-modal");
      form.reset();
      form.elements.id.value = item.id || "";
      form.elements.title.value = item.title || "";
      form.elements.starts_at.value = dateTimeLocal(item.starts_at);
      form.elements.ends_at.value = dateTimeLocal(item.ends_at || new Date(Date.now() + 7200000));
      document.querySelector("#event-title").textContent = item.id ? "Edit Event" : "Add Event";
      document.querySelector("#event-error").textContent = "";
      form.classList.remove("hidden");
    }
    async function saveEvent(event) {
      event.preventDefault();
      const form = event.currentTarget;
      const id = form.elements.id.value;
      const payload = {title: form.elements.title.value.trim(), starts_at: new Date(form.elements.starts_at.value).toISOString(), ends_at: new Date(form.elements.ends_at.value).toISOString()};
      try {
        if (id) await api("/api/v1/calendar/" + encodeURIComponent(id), {method: "PUT", body: JSON.stringify(payload)});
        else await api("/api/v1/calendar", {method: "POST", body: JSON.stringify(payload)});
        form.classList.add("hidden");
        await loadCalendarView();
      } catch (error) {
        document.querySelector("#event-error").textContent = error.message;
      }
    }
    async function deleteEvent(id) {
      await api("/api/v1/calendar/" + encodeURIComponent(id), {method: "DELETE"});
      await loadCalendarView();
    }
    async function reportSelected(verdict) {
      if (!state.selected) {
        showToast("Select a message first", true);
        return;
      }
      const response = await fetch("/api/v1/messages/" + encodeURIComponent(state.selected.id) + "/report", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({verdict})});
      if (!response.ok) throw new Error("Message report failed");
      await loadMessages();
      showToast(verdict === "spam" ? "Marked as spam" : "Marked as not spam");
    }
    async function selectedDetail() {
      if (!state.selected) throw new Error("Select a message first");
      const response = await fetch("/api/v1/messages/" + encodeURIComponent(state.selected.id), {credentials: "same-origin", cache: "no-store"});
      return response.ok ? response.json() : state.selected;
    }
    async function openResponse(mode) {
      const item = await selectedDetail();
      const form = document.querySelector("#compose-modal");
      const sender = emailOnly(item.from || state.selected.from);
      const originalBody = String(item.body || state.selected.preview || "");
      form.reset();
      clearRecipientChips();
      if (mode === "forward") {
        form.elements.to.value = "";
        form.elements.subject.value = prefixedSubject("Fwd: ", item.subject || state.selected.subject || "");
        document.querySelector("#compose-editor").innerText = "\n\nForwarded message\nFrom: " + (item.from || state.selected.from || "") + "\nTo: " + (item.to || state.selected.to || state.email || "") + "\n\n" + originalBody;
      } else {
        addRecipientChip(sender);
        form.elements.subject.value = prefixedSubject("Re: ", item.subject || state.selected.subject || "");
        document.querySelector("#compose-editor").innerText = "\n\nOn " + messageTime(item) + ", " + (item.from || sender) + " wrote:\n" + originalBody.split("\n").map(line => "> " + line).join("\n");
      }
      document.querySelector("#compose-error").textContent = "";
      openCompose();
    }
    function openCompose() {
      document.querySelector("#compose-backdrop").classList.remove("hidden");
      document.querySelector("#compose-modal").classList.remove("hidden");
      document.querySelector("#compose-security-badge").classList.remove("hidden");
      setTimeout(() => document.querySelector("#compose-editor").focus(), 0);
    }
    function closeCompose() {
      document.querySelector("#compose-backdrop").classList.add("hidden");
      document.querySelector("#compose-modal").classList.add("hidden");
      document.querySelector("#compose-security-badge").classList.add("hidden");
    }
    function clearRecipientChips() {
      document.querySelectorAll(".recipient-chip").forEach(item => item.remove());
    }
    function addRecipientChip(email) {
      const value = String(email || "").trim();
      if (!value) return;
      const input = document.querySelector("#to-chip-row input[name='to']");
      const chip = document.createElement("span");
      chip.className = "recipient-chip";
      chip.dataset.recipient = value;
      chip.innerHTML = "<span>" + esc(value) + "</span><button type=\"button\" title=\"Remove\"><span class=\"material-symbols-outlined\">cancel</span></button>";
      chip.querySelector("button").addEventListener("click", () => chip.remove());
      input.before(chip);
      input.value = "";
    }
    function normalizeRecipients() {
      const input = document.querySelector("#to-chip-row input[name='to']");
      String(input.value || "").split(",").map(item => item.trim()).filter(Boolean).forEach(addRecipientChip);
      return [...document.querySelectorAll(".recipient-chip")].map(item => item.dataset.recipient);
    }
    function loadDraft() {
      try {
        const draft = JSON.parse(localStorage.getItem("proidentity-compose-draft") || "{}");
        const form = document.querySelector("#compose-modal");
        if (draft.to) String(draft.to).split(",").map(item => item.trim()).filter(Boolean).forEach(addRecipientChip);
        form.elements.cc.value = draft.cc || "";
        form.elements.bcc.value = draft.bcc || "";
        form.elements.subject.value = draft.subject || "";
        document.querySelector("#compose-editor").innerText = draft.body || "";
      } catch {}
    }
    function saveDraft() {
      const form = document.querySelector("#compose-modal");
      normalizeRecipients();
      const draft = {
        to: [...document.querySelectorAll(".recipient-chip")].map(item => item.dataset.recipient).join(", "),
        cc: form.elements.cc.value,
        bcc: form.elements.bcc.value,
        subject: form.elements.subject.value,
        body: document.querySelector("#compose-editor").innerText
      };
      localStorage.setItem("proidentity-compose-draft", JSON.stringify(draft));
      document.querySelector("#compose-error").className = "error info";
      document.querySelector("#compose-error").textContent = "Draft saved in this browser";
    }
    function filteredMessages() {
      const q = document.querySelector("#search").value.trim().toLowerCase();
      if (!q) return state.messages;
      return state.messages.filter(item => [item.from, item.to, item.subject, item.preview].some(value => String(value || "").toLowerCase().includes(q)));
    }
    function render() {
      if (state.view !== "mail") return;
      updateViewChrome("Search emails...");
      document.querySelector("#avatar").textContent = initials(state.email);
      document.querySelector("#account-email").textContent = state.email || "--";
      document.querySelector(".pane-head h2").textContent = state.folder.charAt(0).toUpperCase() + state.folder.slice(1);
      renderFolders();
      const list = filteredMessages();
      document.querySelector("#messages").innerHTML = list.map((item, index) => {
        const active = state.selected && state.selected.id === item.id ? " active" : "";
        const unread = item.unread ? " unread" : "";
        const tag = /spam|security|dkim|spf|tls/i.test(item.subject || item.preview || "") ? "<span class=\"tag\">SECURITY</span>" : "<span class=\"tag\">MAIL</span>";
        return "<button class=\"message" + active + unread + "\" data-id=\"" + esc(item.id) + "\"><div class=\"message-top\"><span class=\"from\">" + esc(shortFrom(item.from)) + "</span><span class=\"time\">" + esc(messageTime(item)) + "</span></div><div class=\"subject\">" + esc(item.subject || "(no subject)") + "</div><div class=\"preview\">" + esc(item.preview || "") + "</div>" + tag + "</button>";
      }).join("");
      renderReader();
    }
    function renderReader() {
      const item = state.selected;
      if (!item) {
        document.querySelector("#reader").innerHTML = "<h2>No messages yet</h2><div class=\"body\">New mail delivered by Postfix and Dovecot will appear here after refresh.</div>";
        return;
      }
          fetch("/api/v1/messages/" + encodeURIComponent(item.id), { credentials: "same-origin", cache: "no-store" })
        .then(response => response.ok ? response.json() : item)
        .then(detail => {
          document.querySelector("#reader").innerHTML =
        "<h2>" + esc(item.subject || "(no subject)") + "</h2>" +
        "<div class=\"sender-row\"><div class=\"sender\"><div class=\"sender-icon\"><span class=\"material-symbols-outlined\">business</span></div><div><strong>" + esc(shortFrom(item.from)) + "</strong><div class=\"muted\">" + esc(item.from || "") + "</div><div class=\"muted\">To: " + esc(item.to || state.email) + "</div></div></div><div class=\"muted\">" + esc(messageTime(item)) + "</div></div>" +
        "<div class=\"body\"><p>" + esc(detail.body || item.preview || "Message body is empty.").replace(/\n/g, "<br>") + "</p><div class=\"message-meta\"><div class=\"security-note\"><span class=\"material-symbols-outlined\">info</span><span>Sender authentication results are not stored for this message yet. No SPF, DKIM, TLS, or identity verification claim is being made.</span></div><div class=\"recommend\"><h3>MESSAGE SUMMARY</h3><ul><li>Mailbox: " + esc(item.mailbox) + "</li><li>Size: " + esc(item.size_bytes) + " bytes</li><li>Message ID: " + esc(item.id) + "</li></ul></div></div></div>";
        });
    }
    document.querySelector("#login").addEventListener("submit", async event => {
      event.preventDefault();
      const form = event.currentTarget;
      const button = form.querySelector("button[type='submit']");
      const data = new FormData(form);
      state.email = String(data.get("email") || "");
      const errorBox = document.querySelector("#error");
      errorBox.className = "error info";
      errorBox.textContent = "Signing in...";
      button.disabled = true;
      try {
        const response = await fetch("/api/v1/session", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json"}, body: JSON.stringify({email: state.email, password: String(data.get("password") || "")})});
        const body = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(body.error || "Mailbox authentication failed");
        state.csrf = body.csrf_token;
        setAuthenticated(true);
        document.querySelector("#login-panel").classList.add("hidden");
        await loadMessages();
        errorBox.textContent = "";
      } catch (error) {
        setAuthenticated(false);
        document.querySelector("#login-panel").classList.remove("hidden");
        errorBox.className = "error";
        errorBox.textContent = error.message || "Mailbox login failed";
      } finally {
        button.disabled = false;
      }
    });
    document.querySelector(".compose").addEventListener("click", () => {
      const form = document.querySelector("#compose-modal");
      form.reset();
      clearRecipientChips();
      document.querySelector("#compose-editor").innerHTML = "";
      document.querySelector("#compose-error").textContent = "";
      loadDraft();
      openCompose();
    });
    document.querySelector("#close-compose").addEventListener("click", closeCompose);
    document.querySelector("#compose-backdrop").addEventListener("click", closeCompose);
    document.querySelector("#discard-compose").addEventListener("click", () => {
      document.querySelector("#compose-modal").reset();
      clearRecipientChips();
      document.querySelector("#compose-editor").innerHTML = "";
      localStorage.removeItem("proidentity-compose-draft");
      closeCompose();
    });
    document.querySelector("#save-draft").addEventListener("click", saveDraft);
    document.querySelector("#expand-compose").addEventListener("click", () => document.querySelector("#compose-modal").classList.toggle("expanded"));
    document.querySelectorAll("[data-show-compose-field]").forEach(button => button.addEventListener("click", () => document.querySelector("#" + button.dataset.showComposeField).classList.remove("hidden")));
    document.querySelector("#to-chip-row input[name='to']").addEventListener("keydown", event => {
      if (event.key === "Enter" || event.key === ",") {
        event.preventDefault();
        addRecipientChip(event.currentTarget.value.replace(",", ""));
      }
    });
    document.querySelector("#close-contact").addEventListener("click", () => document.querySelector("#contact-modal").classList.add("hidden"));
    document.querySelector("#cancel-contact").addEventListener("click", () => document.querySelector("#contact-modal").classList.add("hidden"));
    document.querySelector("#contact-modal").addEventListener("submit", saveContact);
    document.querySelector("#add-folder").addEventListener("click", () => document.querySelector("#folder-modal").classList.remove("hidden"));
    document.querySelector("#close-folder").addEventListener("click", () => document.querySelector("#folder-modal").classList.add("hidden"));
    document.querySelector("#cancel-folder").addEventListener("click", () => document.querySelector("#folder-modal").classList.add("hidden"));
    document.querySelector("#folder-modal").addEventListener("submit", saveFolder);
    document.querySelector("#open-filters").addEventListener("click", () => loadFiltersView().catch(error => showToast(error.message, true)));
    document.querySelector("#open-filters-rail").addEventListener("click", () => loadFiltersView().catch(error => showToast(error.message, true)));
    document.querySelector("#open-folders-rail").addEventListener("click", async () => {
      try {
        state.view = "folders";
        updateViewChrome("Search folders...");
        await loadFolders();
        document.querySelector("#reader").innerHTML = "<div class=\"workspace-head\"><div><h2>Folders</h2><div class=\"muted\">Create custom folders and open them from the left rail.</div></div><button class=\"primary-button\" id=\"add-folder-inline\" type=\"button\">New Folder</button></div><div class=\"mini-grid\">" + state.folders.map(folder => "<div class=\"mini-row\"><div><strong>" + esc(folder.name) + "</strong><div class=\"muted\">" + esc(folder.total || 0) + " messages</div></div><div class=\"compact-actions\"><button class=\"secondary-button\" data-open-folder=\"" + esc(folder.id) + "\"><span class=\"material-symbols-outlined\">folder_open</span>Open</button>" + (folder.system ? "" : "<button class=\"danger-button\" data-delete-folder=\"" + esc(folder.id) + "\"><span class=\"material-symbols-outlined\">delete</span>Delete</button>") + "</div></div>").join("") + "</div>";
        document.querySelector("#add-folder-inline").addEventListener("click", () => document.querySelector("#folder-modal").classList.remove("hidden"));
      } catch (error) {
        showToast(error.message, true);
      }
    });
    document.querySelector("#close-filter").addEventListener("click", () => document.querySelector("#filter-modal").classList.add("hidden"));
    document.querySelector("#cancel-filter").addEventListener("click", () => document.querySelector("#filter-modal").classList.add("hidden"));
    document.querySelector("#filter-modal").addEventListener("submit", saveFilter);
    document.querySelector("#close-event").addEventListener("click", () => document.querySelector("#event-modal").classList.add("hidden"));
    document.querySelector("#cancel-event").addEventListener("click", () => document.querySelector("#event-modal").classList.add("hidden"));
    document.querySelector("#event-modal").addEventListener("submit", saveEvent);
    document.querySelector("#compose-modal").addEventListener("submit", async event => {
      event.preventDefault();
      const form = event.currentTarget;
      document.querySelector("#compose-error").textContent = "";
      document.querySelector("#compose-error").className = "error";
      form.elements.body.value = document.querySelector("#compose-editor").innerText.trim();
      const data = new FormData(event.currentTarget);
      const recipients = [normalizeRecipients().join(","), data.get("cc"), data.get("bcc")].flatMap(value => String(value || "").split(",").map(item => item.trim()).filter(Boolean));
      const payload = {to: recipients, subject: String(data.get("subject") || ""), body: String(data.get("body") || "")};
      if (!payload.to.length) {
        document.querySelector("#compose-error").textContent = "Add at least one recipient";
        return;
      }
      const response = await fetch("/api/v1/send", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify(payload)});
      if (!response.ok) {
        document.querySelector("#compose-error").textContent = "Send failed";
        return;
      }
      form.reset();
      clearRecipientChips();
      localStorage.removeItem("proidentity-compose-draft");
      document.querySelector("#compose-editor").innerHTML = "";
      closeCompose();
      await loadMessages();
    });
    document.querySelector("#refresh").addEventListener("click", () => loadMessages().then(() => showToast("Mailbox refreshed")).catch(error => showToast(error.message, true)));
    document.querySelector("#mark-spam").addEventListener("click", () => reportSelected("spam").catch(error => showToast(error.message, true)));
    document.querySelector("#mark-ham").addEventListener("click", () => reportSelected("ham").catch(error => showToast(error.message, true)));
    document.querySelector("#archive-message").addEventListener("click", () => moveSelected("archive").catch(error => showToast(error.message, true)));
    document.querySelector("#trash-message").addEventListener("click", () => moveSelected("trash").catch(error => showToast(error.message, true)));
    document.querySelector("#reply-message").addEventListener("click", () => openResponse("reply").catch(error => showToast(error.message, true)));
    document.querySelector("#reply-all-message").addEventListener("click", () => openResponse("reply").catch(error => showToast(error.message, true)));
    document.querySelector("#forward-message").addEventListener("click", () => openResponse("forward").catch(error => showToast(error.message, true)));
    document.querySelector("#open-contacts").addEventListener("click", () => loadContactsView().catch(error => showToast(error.message, true)));
    document.querySelector("#open-calendar").addEventListener("click", () => loadCalendarView().catch(error => showToast(error.message, true)));
    document.querySelector("#logout").addEventListener("click", async () => {
      await api("/api/v1/session", {method: "DELETE"});
      state.csrf = "";
      state.email = "";
      setAuthenticated(false);
      document.querySelector("#login-panel").classList.remove("hidden");
    });
    document.querySelector("#avatar").addEventListener("click", () => document.querySelector("#account-menu").classList.toggle("hidden"));
    document.querySelector("#account-close").addEventListener("click", () => document.querySelector("#account-menu").classList.add("hidden"));
    document.querySelector("#account-logout").addEventListener("click", () => document.querySelector("#logout").click());
    document.querySelectorAll("[data-editor-command]").forEach(button => button.addEventListener("click", () => {
      document.querySelector("#compose-editor").focus();
      document.execCommand(button.dataset.editorCommand, false, null);
    }));
    document.querySelector("[data-editor-link]").addEventListener("click", () => {
      const url = prompt("Link URL");
      if (!url) return;
      document.querySelector("#compose-editor").focus();
      document.execCommand("createLink", false, url);
    });
    document.querySelector("[data-editor-image]").addEventListener("click", () => {
      const url = prompt("Image URL");
      if (!url) return;
      document.querySelector("#compose-editor").focus();
      document.execCommand("insertImage", false, url);
    });
    document.querySelector("[data-editor-clear]").addEventListener("click", () => {
      document.querySelector("#compose-editor").focus();
      document.execCommand("removeFormat", false, null);
    });
    document.querySelector("#search").addEventListener("input", renderCurrentView);
    document.addEventListener("click", event => {
      const editContact = event.target.closest("[data-edit-contact]");
      if (editContact) {
        openContactModal(state.contacts.find(item => item.id === editContact.dataset.editContact) || {});
        return;
      }
      const deleteContactButton = event.target.closest("[data-delete-contact]");
      if (deleteContactButton) {
        deleteContact(deleteContactButton.dataset.deleteContact).catch(error => showToast(error.message, true));
        return;
      }
      const editEvent = event.target.closest("[data-edit-event]");
      if (editEvent) {
        openEventModal(state.events.find(item => item.id === editEvent.dataset.editEvent) || {});
        return;
      }
      const deleteEventButton = event.target.closest("[data-delete-event]");
      if (deleteEventButton) {
        deleteEvent(deleteEventButton.dataset.deleteEvent).catch(error => showToast(error.message, true));
        return;
      }
      const editFilter = event.target.closest("[data-edit-filter]");
      if (editFilter) {
        openFilterModal(state.filters.find(item => item.id === editFilter.dataset.editFilter) || {});
        return;
      }
      const deleteFilterButton = event.target.closest("[data-delete-filter]");
      if (deleteFilterButton) {
        deleteFilter(deleteFilterButton.dataset.deleteFilter).catch(error => showToast(error.message, true));
        return;
      }
      const openFolder = event.target.closest("[data-open-folder]");
      if (openFolder) {
        state.folder = openFolder.dataset.openFolder;
        loadMessages().catch(error => showToast(error.message, true));
        return;
      }
      const deleteFolderButton = event.target.closest("[data-delete-folder]");
      if (deleteFolderButton) {
        deleteFolder(deleteFolderButton.dataset.deleteFolder).catch(error => showToast(error.message, true));
        return;
      }
      const button = event.target.closest("[data-id]");
      if (!button) return;
      state.selected = state.messages.find(item => item.id === button.dataset.id) || null;
      render();
    });
    bootstrapSession().catch(error => {
      setAuthenticated(false);
      document.querySelector("#error").textContent = error.message;
    });
  </script>
</body>
</html>
`
