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
  <style nonce="__PROIDENTITY_CSP_NONCE__">
    :root {
      --background: #f4f5f7;
      --surface: #ffffff;
      --surface-soft: #f7f7f8;
      --surface-muted: #eceff3;
      --ink: #191c1e;
      --muted: #464554;
      --outline: #d6d8df;
      --outline-strong: #767586;
      --primary: #4648d4;
      --primary-soft: #e8ebff;
      --secondary-soft: #d7e9ff;
      --outlook-blue: #0f6cbd;
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
      height: 94px;
      display: flex;
      flex-direction: column;
      align-items: stretch;
      justify-content: flex-start;
      padding: 0;
      border-bottom: 1px solid var(--outline);
      background: #fbfbfc;
    }
    .titlebar {
      height: 46px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 20px;
      padding: 0 18px;
      border-bottom: 1px solid var(--outline);
      background: #fbfbfc;
    }
    .brand { display: flex; align-items: center; gap: 10px; color: var(--primary); min-width: 188px; }
    .brand h1 { margin: 0; font-size: 18px; line-height: 1; font-weight: 800; }
    .brand .material-symbols-outlined { font-size: 25px; }
    .top-actions { display: flex; align-items: center; gap: 18px; }
    .search {
      width: min(420px, 32vw);
      min-height: 34px;
      display: flex;
      align-items: center;
      gap: 10px;
      background: #eceef1;
      border: 1px solid transparent;
      border-radius: 7px;
      padding: 0 12px;
      color: var(--outline-strong);
      font-weight: 500;
    }
    .search:focus-within { background: white; border-color: rgba(70,72,212,.36); }
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
      height: calc(100vh - 94px);
      display: grid;
      grid-template-columns: 276px 380px minmax(420px, 1fr) 54px;
      overflow: hidden;
    }
    body.workspace-wide .app { grid-template-columns: 276px minmax(0, 1fr) 54px; }
    body.workspace-wide .list-pane { display: none; }
    aside {
      background: #f3f4f6;
      border-right: 1px solid var(--outline);
      display: flex;
      flex-direction: column;
      padding: 10px 8px;
      min-width: 0;
      min-height: 0;
      overflow-x: hidden;
      overflow-y: auto;
      overscroll-behavior: contain;
    }
    .compose {
      min-height: 38px;
      border: 0;
      border-radius: 6px;
      background: var(--primary);
      color: white;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      font: inherit;
      font-weight: 700;
      font-size: 14px;
      cursor: pointer;
      box-shadow: 0 8px 18px rgba(70,72,212,.18);
      margin: 2px 4px 12px;
    }
    .mailbox-list { display: grid; gap: 4px; margin-bottom: 8px; }
    .mailbox-card {
      border: 0;
      border-radius: 7px;
      background: transparent;
      display: grid;
      grid-template-columns: 32px minmax(0, 1fr) auto;
      align-items: center;
      gap: 9px;
      min-height: 50px;
      padding: 6px 8px;
      color: var(--ink);
      font: inherit;
      text-align: left;
      cursor: pointer;
    }
    .mailbox-card.active { background: #e9effc; }
    .mailbox-avatar, .message-avatar, .reader-avatar {
      border-radius: 50%;
      background: linear-gradient(135deg, var(--primary), var(--outlook-blue));
      color: white;
      display: grid;
      place-items: center;
      font-weight: 800;
    }
    .mailbox-avatar { width: 32px; height: 32px; font-size: 12px; }
    .mailbox-details { min-width: 0; display: grid; gap: 1px; }
    .mailbox-name { display: block; min-width: 0; font-weight: 800; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .mailbox-address { display: block; min-width: 0; color: var(--muted); font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .mailbox-kind {
      justify-self: end;
      max-width: 62px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      border-radius: 999px;
      background: rgba(70,72,212,.08);
      color: var(--muted);
      padding: 2px 6px;
      font-size: 9px;
      font-weight: 800;
      text-transform: uppercase;
      letter-spacing: .04em;
    }
    .nav-section-title {
      margin: 12px 10px 6px;
      color: #343741;
      font-weight: 800;
      display: flex;
      align-items: center;
      gap: 6px;
      border: 0;
      background: transparent;
      padding: 0;
      font: inherit;
      cursor: pointer;
    }
    .nav-section-title .material-symbols-outlined { transition: transform .16s ease; }
    .nav-section-title.collapsed .material-symbols-outlined { transform: rotate(-90deg); }
    .nav-section-body.collapsed { display: none; }
    .folder-list { display: grid; gap: 2px; }
    .folder {
      min-height: 36px;
      border: 0;
      border-radius: 6px;
      background: transparent;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      padding: 0 10px 0 12px;
      color: #272a3a;
      font: inherit;
      font-size: 14px;
      cursor: pointer;
    }
    .folder span:first-child { display: flex; align-items: center; gap: 10px; min-width: 0; }
    .folder .material-symbols-outlined { font-size: 20px; }
    .folder.active { background: #cfe5ff; color: #182233; font-weight: 800; }
    .folder.drop-allowed {
      outline: 2px solid var(--primary);
      background: var(--primary-soft);
    }
    .folder.drop-denied {
      outline: 2px solid rgba(186,26,26,.35);
      background: rgba(186,26,26,.08);
    }
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
      padding: 12px 8px 0;
      display: grid;
      gap: 8px;
      color: var(--muted);
    }
    .labels h3 { margin: 0 0 4px; color: var(--outline-strong); font-size: 12px; letter-spacing: .18em; }
    .label { display: flex; align-items: center; gap: 12px; font-size: 15px; }
    .dot { width: 12px; height: 12px; border-radius: 50%; background: var(--primary); }
    .dot.danger { background: #d11c1c; }
    .dot.dark { background: #56627a; }
    .profile-card {
      margin: auto 0 0;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: white;
      min-height: 58px;
      padding: 8px;
      display: grid;
      grid-template-columns: 36px minmax(0, 1fr);
      gap: 9px;
      align-items: center;
      color: var(--ink);
      text-align: left;
      font: inherit;
      cursor: pointer;
    }
    .profile-card:hover { border-color: rgba(70,72,212,.42); background: #fbfcff; }
    .profile-avatar {
      width: 36px;
      height: 36px;
      border-radius: 50%;
      background: var(--primary);
      color: white;
      display: grid;
      place-items: center;
      font-weight: 800;
    }
    .profile-lines { min-width: 0; display: grid; gap: 1px; }
    .profile-name { font-weight: 800; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .profile-email { color: var(--muted); font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .list-pane {
      background: var(--surface);
      border-right: 1px solid var(--outline);
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
    .pane-head {
      min-height: 58px;
      display: grid;
      grid-template-columns: minmax(0, 1fr) auto;
      align-items: start;
      justify-content: space-between;
      gap: 12px;
      padding: 10px 14px;
      border-bottom: 1px solid var(--outline);
      background: #fbfbfc;
    }
    .pane-head h2 { margin: 0; font-size: 18px; line-height: 1.2; }
    .pane-subtitle { color: var(--muted); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .message-tabs {
      min-height: 45px;
      display: grid;
      grid-template-columns: 1fr 1fr auto;
      align-items: end;
      border-bottom: 1px solid var(--outline);
      background: white;
    }
    .message-tab {
      height: 45px;
      border: 0;
      border-bottom: 3px solid transparent;
      background: transparent;
      color: #2d3038;
      font: inherit;
      font-size: 15px;
      cursor: pointer;
    }
    .message-tab.active { border-bottom-color: var(--primary); font-weight: 800; }
    .message-list-tools { display: flex; align-items: center; justify-content: center; padding: 0 10px 8px; color: var(--muted); }
    .message-list { overflow: auto; flex: 1; }
    .tool-button.active { background: var(--primary-soft); color: var(--primary); }
    .message-group {
      position: sticky;
      top: 0;
      z-index: 2;
      min-height: 32px;
      padding: 7px 14px 5px;
      display: flex;
      align-items: center;
      gap: 8px;
      border-bottom: 1px solid #eceef3;
      background: rgba(255,255,255,.95);
      color: #363946;
      font-weight: 700;
    }
    .message-group .material-symbols-outlined { font-size: 19px; }
    .message {
      border: 0;
      border-bottom: 1px solid var(--outline);
      background: var(--surface);
      width: 100%;
      text-align: left;
      padding: 11px 12px;
      cursor: pointer;
      border-left: 3px solid transparent;
      display: grid;
      grid-template-columns: 24px 38px minmax(0, 1fr);
      gap: 10px;
      align-items: start;
      color: var(--ink);
    }
    .message:hover { background: #f8fbff; }
    .message.active,
    .message.selected { background: #eef6ff; border-left-color: var(--outlook-blue); }
    .message[draggable="true"] { cursor: grab; }
    .message.dragging { opacity: .55; }
    .message.unread .from, .message.unread .subject { font-weight: 800; }
    .message.unread .from:before {
      content: "";
      display: block;
      position: absolute;
      left: -33px;
      top: 8px;
      width: 7px;
      height: 7px;
      border-radius: 999px;
      background: var(--primary);
    }
    .message-caret { color: var(--muted); padding-top: 8px; }
    .message-caret .material-symbols-outlined { font-size: 19px; }
    .message-avatar { width: 38px; height: 38px; font-size: 13px; }
    .message-list.compact .message { grid-template-columns: 20px 30px minmax(0, 1fr); gap: 8px; padding: 7px 10px; }
    .message-list.compact .message-avatar { width: 30px; height: 30px; font-size: 11px; }
    .message-list.compact .message-caret { padding-top: 5px; }
    .message-list.compact .subject { font-size: 14px; }
    .message-list.compact .preview { -webkit-line-clamp: 1; }
    .message-body { min-width: 0; display: grid; gap: 3px; }
    .message-top { display: flex; justify-content: space-between; gap: 12px; }
    .from { position: relative; font-weight: 700; font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
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
      background: #ffffff;
      overflow: auto;
      display: flex;
      flex-direction: column;
      min-width: 0;
    }
    .toolbar {
      min-height: 48px;
      border-bottom: 1px solid var(--outline);
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 0 18px;
      color: #292b3c;
      background: #ffffff;
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
    .tool-button.action-hidden { display: none; }
    .message-context-menu {
      position: fixed;
      z-index: 90;
      width: 220px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: white;
      box-shadow: 0 18px 42px rgba(15,23,42,.18);
      padding: 6px;
      display: grid;
      gap: 2px;
    }
    .message-context-menu.hidden { display: none; }
    .context-menu-item {
      width: 100%;
      min-height: 34px;
      border: 0;
      border-radius: 6px;
      background: transparent;
      color: var(--ink);
      padding: 0 9px;
      display: flex;
      align-items: center;
      gap: 8px;
      font: inherit;
      font-weight: 700;
      text-align: left;
      cursor: pointer;
    }
    .context-menu-item:hover { background: var(--surface-soft); }
    .context-menu-item.danger { color: var(--danger); }
    .context-menu-item .material-symbols-outlined { font-size: 19px; }
    .message-meta { margin-top: auto; padding: 18px 0 32px; display: grid; gap: 8px; }
    .reader-content { padding: 26px 40px 42px; max-width: 920px; width: 100%; }
    .reader-content.mail-reader { min-height: 100%; display: flex; flex-direction: column; }
    .reader h2 { margin: 0 0 22px; font-size: 25px; line-height: 1.2; font-weight: 800; }
    .sender-row { display: flex; align-items: center; justify-content: space-between; gap: 20px; margin-bottom: 24px; }
    .sender { display: flex; align-items: center; gap: 12px; }
    .sender-icon {
      width: 42px;
      height: 42px;
      border-radius: 50%;
      background: linear-gradient(135deg, var(--primary), var(--outlook-blue));
      color: white;
      display: grid;
      place-items: center;
      font-weight: 800;
    }
    .sender strong { font-size: 18px; }
    .body { font-size: 16px; line-height: 1.55; }
    .message-display-body { flex: 0 0 auto; min-height: 120px; }
    .plain-body { white-space: pre-wrap; margin: 0; }
    .message-auth-panel {
      margin: 0 0 14px;
      min-height: 40px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: #f8f8fc;
      color: var(--muted);
      padding: 9px 12px;
      display: flex;
      align-items: center;
      gap: 10px;
      font-size: 12px;
    }
    .message-trust-banner {
      border-color: rgba(70,72,212,.28);
      background: #f8f8ff;
      color: #2d3354;
      justify-content: space-between;
      flex-wrap: wrap;
    }
    .message-auth-items { display: flex; align-items: center; gap: 14px; flex-wrap: wrap; }
    .message-auth-item { display: inline-flex; align-items: center; gap: 6px; font-weight: 800; letter-spacing: .04em; text-transform: uppercase; }
    .message-auth-item .material-symbols-outlined { color: var(--primary); font-size: 19px; }
    .message-trust-label { margin-left: auto; font-size: 11px; font-weight: 900; letter-spacing: .06em; text-transform: uppercase; }
    .message-local-delivery {
      border-color: rgba(15,108,189,.22);
      background: #f5fbff;
      color: #25415f;
    }
    .message-local-delivery strong { color: #102f4f; }
    .message-auth-copy { display: grid; gap: 2px; }
    .message-auth-badge {
      margin-left: auto;
      border-radius: 999px;
      background: rgba(15,108,189,.1);
      color: #25415f;
      padding: 3px 9px;
      font-size: 10px;
      font-weight: 900;
      letter-spacing: .08em;
      text-transform: uppercase;
      white-space: nowrap;
    }
    .external-source-banner {
      margin: 0 0 14px;
      border: 1px solid #f0bf4f;
      border-radius: 8px;
      background: #fff7df;
      color: #5d3d00;
      padding: 12px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 14px;
      font-size: 13px;
    }
    .external-source-banner.allowed {
      border-color: rgba(7,148,85,.28);
      background: #eefaf4;
      color: #075e45;
    }
    .external-source-banner strong { color: inherit; }
    .external-source-actions {
      display: flex;
      flex-wrap: wrap;
      justify-content: flex-end;
      gap: 8px;
    }
    .external-source-actions .secondary-button:disabled {
      opacity: .55;
      cursor: not-allowed;
      background: #f3f4f6;
    }
    .mail-html-frame {
      width: 100%;
      min-height: 260px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: white;
      display: block;
    }
    .recommend {
      margin: 18px 0;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: var(--surface-muted);
      padding: 18px;
    }
    .recommend h3 { margin: 0 0 12px; color: var(--primary); font-size: 15px; letter-spacing: .08em; }
    .message-summary {
      width: min(620px, 100%);
      margin: 0;
      padding: 10px 12px;
      font-size: 12px;
    }
    .message-summary h3 { margin: 0 0 6px; font-size: 11px; letter-spacing: .12em; }
    .message-summary ul { margin: 0; padding-left: 18px; display: grid; gap: 2px; }
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
    .sidebar-separator {
      height: 1px;
      background: var(--outline);
      margin: 14px 8px;
      flex: 0 0 auto;
    }
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
    #profile-modal {
      inset: 50% auto auto 50%;
      transform: translate(-50%, -50%);
      width: min(740px, calc(100vw - 36px));
      max-height: min(720px, calc(100vh - 44px));
      overflow: auto;
      padding: 0;
      gap: 0;
    }
    #profile-modal .modal-head {
      min-height: 60px;
      padding: 0 18px 0 22px;
      border-bottom: 1px solid var(--outline);
      background: var(--surface-soft);
    }
    .profile-settings-body { padding: 18px 22px 0; display: grid; gap: 14px; }
    .profile-settings-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(260px, .85fr); gap: 14px; align-items: start; }
    .profile-account-summary {
      border: 1px solid var(--outline);
      border-radius: 10px;
      background: #fbfcff;
      padding: 14px;
      display: grid;
      gap: 8px;
    }
    .profile-account-summary strong { font-size: 16px; overflow-wrap: anywhere; }
    .profile-account-summary code {
      display: block;
      width: 100%;
      overflow-wrap: anywhere;
      border-radius: 7px;
      background: var(--surface-muted);
      padding: 8px 10px;
      font-size: 12px;
    }
    .profile-settings-body .connect-box { margin: 0; }
    #profile-modal textarea { min-height: 180px; resize: vertical; }
    #profile-modal .modal-actions {
      position: sticky;
      bottom: 0;
      padding: 14px 18px;
      border-top: 1px solid var(--outline);
      background: rgba(255,255,255,.94);
      backdrop-filter: blur(8px);
    }
    #delete-confirm-modal {
      inset: 50% auto auto 50%;
      transform: translate(-50%, -50%);
      width: min(460px, calc(100vw - 36px));
    }
    #delete-confirm-modal p { margin: 0; color: var(--muted); }
    #message-delete-subject {
      display: block;
      margin-top: 6px;
      overflow-wrap: anywhere;
      color: var(--ink);
    }
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
    .compose-body { display: flex; min-height: 0; overflow: hidden; flex-direction: column; background: white; }
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
    .compose-row select {
      border: 0;
      outline: 0;
      min-height: 38px;
      background: transparent;
      box-shadow: none;
      flex: 1;
      padding: 0;
      font-size: 15px;
    }
    .compose-row select:focus { outline: 0; box-shadow: none; border-color: transparent; }
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
    .compose-toolbar-select {
      min-height: 34px;
      border: 0;
      border-radius: 7px;
      background: white;
      padding: 0 8px;
      font: inherit;
      font-weight: 700;
    }
    .compose-toolbar-group:last-child { border-right: 0; padding-right: 0; }
    .compose-editor-shell { flex: 1 1 auto; min-height: 0; overflow: auto; padding: 24px 30px; }
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
      flex: 0 0 auto;
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
    .attachment-input { display: none; }
    .attachment-chip.disabled {
      border-style: dashed;
      opacity: .78;
    }
    .attachment-list {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }
    .attachment-chip button {
      border: 0;
      background: transparent;
      color: inherit;
      cursor: pointer;
      display: grid;
      place-items: center;
    }
    .compose-footer {
      flex: 0 0 auto;
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
    .mfa-panel {
      border: 1px solid var(--outline);
      border-radius: 8px;
      padding: 14px;
      display: grid;
      gap: 12px;
      background: #fbfcff;
    }
    .mfa-panel.hidden { display: none; }
    .mfa-panel h3 { margin: 0; font-size: 16px; }
    .mfa-panel img { width: 180px; height: 180px; border: 1px solid var(--outline); border-radius: 8px; background: #fff; }
    .mfa-panel .actions { display: flex; gap: 8px; flex-wrap: wrap; }
    .mailbox-push-card { display: grid; justify-items: center; gap: 10px; text-align: center; padding: 8px 4px; }
    .mailbox-push-card.hidden, #mailbox-push-check.hidden, #mailbox-mfa-code-row.hidden, #mailbox-mfa-submit.hidden { display: none; }
    .mailbox-push-icon { width: 56px; height: 56px; border-radius: 16px; display: grid; place-items: center; background: #eef5ff; color: var(--primary); }
    .mailbox-push-status { display: inline-flex; align-items: center; justify-content: center; gap: 8px; color: var(--muted); }
    .mailbox-push-status::before { content: ""; width: 8px; height: 8px; border-radius: 999px; background: var(--primary); }
    .link-button { border: 0; background: transparent; color: #315a96; text-decoration: underline; cursor: pointer; font: inherit; }
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
    .mobile-switcher,
    .mobile-compose-button { display: none; }
    @media (max-width: 1180px) {
      .app {
        grid-template-columns: 244px 340px minmax(0, 1fr);
      }
      body.workspace-wide .app { grid-template-columns: 244px minmax(0, 1fr); }
      .rail { display: none; }
      .reader-content { padding: 22px 28px 36px; max-width: none; }
      .search { width: min(360px, 30vw); }
      .mailbox-card { grid-template-columns: 30px minmax(0, 1fr); }
      .mailbox-kind { grid-column: 2; justify-self: start; max-width: 100%; }
      .compose { margin-left: 2px; margin-right: 2px; }
    }
    @media (max-width: 980px) {
      header { height: 92px; }
      .app {
        height: calc(100vh - 92px);
        height: calc(100dvh - 92px);
        grid-template-columns: 220px 318px minmax(0, 1fr);
      }
      body.workspace-wide .app { grid-template-columns: 220px minmax(0, 1fr); }
      .titlebar { padding: 0 12px; gap: 12px; }
      .brand { min-width: 150px; }
      .brand h1 { font-size: 16px; }
      .search { width: min(300px, 32vw); }
      .toolbar {
        min-width: 0;
        overflow-x: auto;
        overscroll-behavior-x: contain;
        scrollbar-width: thin;
        padding: 0 10px;
      }
      .tool-button { flex: 0 0 auto; }
      aside { padding: 8px 6px; }
      .reader-content { padding: 20px 22px 34px; }
      .reader h2 { font-size: 22px; }
      .calendar-layout { grid-template-columns: 1fr; }
      #compose-modal { inset: 12px; min-width: 0; }
      #profile-modal { inset: 12px; transform: none; width: auto; max-height: calc(100vh - 24px); }
      .profile-settings-grid { grid-template-columns: 1fr; }
      .compose-head, .compose-fields, .compose-toolbar, .compose-editor-shell, .compose-attachments, .compose-footer { padding-left: 16px; padding-right: 16px; }
      .compose-footer { align-items: stretch; flex-direction: column; }
      .compose-footer-left, .compose-footer-right { width: 100%; justify-content: space-between; }
      .send-button { flex: 1; justify-content: center; }
      .security-compose-badge { display: none; }
    }
    @media (max-width: 760px) {
      :root { --mobile-nav-height: 66px; }
      body {
        height: 100vh;
        height: 100dvh;
        overflow: hidden;
      }
      header {
        height: 94px;
        position: sticky;
        top: 0;
        z-index: 20;
      }
      .titlebar { height: 46px; padding: 0 10px; }
      .brand { min-width: 0; gap: 8px; }
      .brand h1 {
        max-width: 150px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .top-actions { gap: 8px; min-width: 0; }
      .search {
        width: min(240px, 42vw);
        min-height: 36px;
        padding: 0 10px;
      }
      .avatar { width: 36px; height: 36px; flex: 0 0 auto; }
      .toolbar {
        min-height: 48px;
        padding: 0 8px;
        gap: 6px;
        overflow-x: auto;
        scrollbar-width: none;
        -webkit-overflow-scrolling: touch;
      }
      .toolbar::-webkit-scrollbar { display: none; }
      .tool-button { width: 44px; height: 44px; border-radius: 10px; }
      .app {
        height: calc(100vh - 94px - var(--mobile-nav-height));
        height: calc(100dvh - 94px - var(--mobile-nav-height));
        display: block;
        overflow: hidden;
      }
      .app > aside:not(.rail),
      .list-pane,
      .reader {
        width: 100%;
        height: 100%;
        min-height: 0;
      }
      .rail { display: none; }
      body.mobile-pane-sidebar .app > aside:not(.rail),
      body.mobile-pane-list .list-pane,
      body.mobile-pane-reader .reader {
        display: flex;
      }
      body.mobile-pane-sidebar .list-pane,
      body.mobile-pane-sidebar .reader,
      body.mobile-pane-list .app > aside:not(.rail),
      body.mobile-pane-list .reader,
      body.mobile-pane-reader .app > aside:not(.rail),
      body.mobile-pane-reader .list-pane {
        display: none;
      }
      body.workspace-wide .app > aside:not(.rail),
      body.workspace-wide .list-pane {
        display: none;
      }
      body.workspace-wide .reader {
        display: flex;
      }
      aside {
        border-right: 0;
        padding: 12px 10px calc(18px + env(safe-area-inset-bottom));
      }
      .compose { min-height: 46px; font-size: 15px; }
      .folder { min-height: 44px; }
      .profile-card { margin-top: 18px; min-height: 64px; }
      .pane-head {
        min-height: 58px;
        padding: 9px 12px;
      }
      .message-tabs { min-height: 44px; }
      .message-tab { height: 44px; font-size: 14px; }
      .message-list { -webkit-overflow-scrolling: touch; }
      .message {
        min-height: 72px;
        grid-template-columns: 20px 36px minmax(0, 1fr);
        gap: 8px;
        padding: 10px 10px;
      }
      .message-avatar { width: 36px; height: 36px; }
      .message-caret { padding-top: 7px; }
      .message-list.compact .message { min-height: 58px; }
      .reader { overflow: auto; -webkit-overflow-scrolling: touch; }
      .reader-content {
        padding: 16px 14px 28px;
        max-width: none;
      }
      .reader h2 { font-size: 20px; margin-bottom: 16px; }
      .sender-row {
        align-items: flex-start;
        flex-direction: column;
        gap: 10px;
      }
      .external-source-banner {
        align-items: stretch;
        flex-direction: column;
      }
      .mobile-switcher {
        position: fixed;
        left: 0;
        right: 0;
        bottom: 0;
        z-index: 45;
        height: var(--mobile-nav-height);
        padding: 6px 8px calc(6px + env(safe-area-inset-bottom));
        border-top: 1px solid var(--outline);
        background: rgba(255,255,255,.96);
        backdrop-filter: blur(12px);
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 8px;
      }
      .mobile-switcher button {
        min-width: 0;
        min-height: 48px;
        border: 0;
        border-radius: 12px;
        background: transparent;
        color: var(--muted);
        display: grid;
        place-items: center;
        gap: 1px;
        font: inherit;
        font-size: 11px;
        font-weight: 800;
        cursor: pointer;
      }
      .mobile-switcher button.active {
        background: var(--primary-soft);
        color: var(--primary);
      }
      .mobile-switcher .material-symbols-outlined { font-size: 23px; }
      .mobile-compose-button {
        position: fixed;
        right: 16px;
        bottom: calc(var(--mobile-nav-height) + 14px);
        z-index: 44;
        width: 56px;
        height: 56px;
        border: 0;
        border-radius: 18px;
        background: var(--primary);
        color: white;
        box-shadow: 0 16px 32px rgba(70,72,212,.28);
        align-items: center;
        justify-content: center;
        cursor: pointer;
      }
      body.mobile-pane-sidebar .mobile-compose-button,
      body.workspace-wide .mobile-compose-button,
      body.auth-locked .mobile-compose-button,
      body.auth-locked .mobile-switcher {
        display: none;
      }
      body:not(.mobile-pane-sidebar):not(.workspace-wide):not(.auth-locked) .mobile-compose-button {
        display: flex;
      }
      .modal {
        width: calc(100vw - 24px);
        max-height: calc(100dvh - 24px);
        overflow: auto;
      }
      #compose-modal,
      #compose-modal.expanded {
        inset: 0;
        width: 100vw;
        height: 100vh;
        height: 100dvh;
        min-width: 0;
        max-width: none;
        border-radius: 0;
      }
      .compose-head {
        min-height: 56px;
        padding: 0 12px;
      }
      .compose-title h2 { font-size: 18px; }
      .compose-fields { padding-top: 8px; }
      .compose-row {
        min-height: 46px;
        gap: 8px;
      }
      .compose-row-label { width: 44px; font-size: 13px; }
      .compose-toolbar {
        min-height: 52px;
        overflow-x: auto;
        flex-wrap: nowrap;
        -webkit-overflow-scrolling: touch;
      }
      .compose-toolbar-group { flex: 0 0 auto; }
      .compose-editor-shell {
        padding-top: 16px;
        padding-bottom: 16px;
      }
      #compose-editor { font-size: 16px; }
      .compose-attachments {
        max-height: 112px;
        overflow: auto;
      }
      .compose-footer {
        min-height: 72px;
        padding-bottom: calc(12px + env(safe-area-inset-bottom));
      }
      .attachment-chip { min-height: 46px; }
    }
    @media (max-width: 420px) {
      .brand h1 { max-width: 118px; font-size: 15px; }
      .search { width: min(162px, 40vw); }
      .search input::placeholder { color: transparent; }
      .message {
        grid-template-columns: 18px 34px minmax(0, 1fr);
        padding-left: 8px;
        padding-right: 8px;
      }
      .message-avatar { width: 34px; height: 34px; }
      .time { font-size: 11px; }
      .compose-footer-left,
      .compose-footer-right {
        flex-wrap: wrap;
      }
    }
  </style>
</head>
<body class="auth-locked mobile-pane-list">
  <header>
    <div class="titlebar">
      <div class="brand"><span class="material-symbols-outlined">shield</span><h1>ProIdentity Mail</h1></div>
      <div class="top-actions">
        <div class="search"><span class="material-symbols-outlined">search</span><input id="search" placeholder="Search mail"></div>
        <button class="avatar" type="button" id="avatar" title="Account">--</button>
      </div>
    </div>
    <div class="toolbar">
      <button class="tool-button" type="button" id="refresh" title="Refresh"><span class="material-symbols-outlined">refresh</span></button>
      <button class="tool-button" type="button" id="archive-message" title="Archive"><span class="material-symbols-outlined">archive</span></button>
      <button class="tool-button" type="button" id="mark-spam" title="Mark as spam"><span class="material-symbols-outlined">report</span></button>
      <button class="tool-button" type="button" id="mark-ham" title="Mark as not spam"><span class="material-symbols-outlined">verified</span></button>
      <button class="tool-button" type="button" id="trash-message" title="Delete"><span class="material-symbols-outlined">delete</span></button>
      <span style="height:24px;width:1px;background:var(--outline);margin:0 8px"></span>
      <button class="tool-button" type="button" id="reply-message" title="Reply"><span class="material-symbols-outlined">reply</span></button>
      <button class="tool-button" type="button" id="reply-all-message" title="Reply all"><span class="material-symbols-outlined">reply_all</span></button>
      <button class="tool-button" type="button" id="forward-message" title="Forward"><span class="material-symbols-outlined">forward</span></button>
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
      <button class="compose" type="button"><span class="material-symbols-outlined">edit</span>New mail</button>
      <button class="nav-section-title" type="button" data-toggle-section="mailboxes"><span class="material-symbols-outlined">expand_more</span>Mailboxes</button>
      <div class="mailbox-list nav-section-body" id="mailbox-list" data-section="mailboxes"></div>
      <div class="sidebar-separator" aria-hidden="true"></div>
      <button class="nav-section-title" type="button" data-toggle-section="folders"><span class="material-symbols-outlined">expand_more</span>Folders</button>
      <nav class="folder-list nav-section-body" id="folder-list" data-section="folders"></nav>
      <div class="folder-tools">
        <button class="secondary-button" type="button" id="add-folder"><span class="material-symbols-outlined">create_new_folder</span>New folder</button>
      <button class="secondary-button" type="button" id="open-filters"><span class="material-symbols-outlined">filter_alt</span>Filters</button>
      </div>
      <div class="sidebar-separator" aria-hidden="true"></div>
      <button class="profile-card" type="button" id="profile-card" title="Profile and signature settings">
        <span class="profile-avatar" id="profile-avatar">--</span>
        <span class="profile-lines"><span class="profile-name" id="profile-name">User profile</span><span class="profile-email" id="profile-email">--</span></span>
      </button>
    </aside>

    <section class="list-pane">
      <div class="pane-head"><div><h2>Inbox</h2><div class="pane-subtitle" id="pane-mailbox">Personal mailbox</div></div><button class="tool-button" type="button" id="open-filters-pane" title="Mail filters"><span class="material-symbols-outlined">filter_list</span></button></div>
      <div class="message-tabs"><button class="message-tab active" type="button" data-message-tab="focused" title="Focused shows normal person-to-person inbox mail" aria-label="Focused inbox messages">Focused</button><button class="message-tab" type="button" data-message-tab="other" title="Other shows newsletters, notifications, receipts, and automated mail" aria-label="Other automated and low priority messages">Other</button><div class="message-list-tools"><button class="tool-button" type="button" id="toggle-message-density" title="Toggle compact message list" aria-pressed="false"><span class="material-symbols-outlined">view_agenda</span></button></div></div>
      <div class="message-list" id="messages" tabindex="0"></div>
    </section>

    <section class="reader">
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
  <button class="mobile-compose-button" type="button" id="mobile-compose" title="New mail"><span class="material-symbols-outlined">edit</span></button>
  <nav class="mobile-switcher" aria-label="Mobile mail panes">
    <button type="button" data-mobile-pane="sidebar"><span class="material-symbols-outlined">menu</span><span>Folders</span></button>
    <button class="active" type="button" data-mobile-pane="list"><span class="material-symbols-outlined">mail</span><span>Mail</span></button>
    <button type="button" data-mobile-pane="reader"><span class="material-symbols-outlined">draft</span><span>Read</span></button>
  </nav>
  <div class="message-context-menu hidden" id="message-context-menu" role="menu"></div>

  <form class="modal hidden delete-confirm-modal" id="delete-confirm-modal">
    <div class="modal-head"><h2 id="delete-confirm-title">Remove message</h2><button class="tool-button" type="button" id="close-delete-confirm" title="Close"><span class="material-symbols-outlined">close</span></button></div>
    <p id="message-delete-copy">Do you really want to remove this message?</p>
    <strong id="message-delete-subject"></strong>
    <div class="modal-actions">
      <button class="secondary-button" type="button" id="cancel-delete-confirm">Cancel</button>
      <button class="danger-button" type="submit" id="confirm-delete-action"><span class="material-symbols-outlined">delete</span>Remove</button>
    </div>
    <div class="error" id="delete-confirm-error"></div>
  </form>

  <div class="login" id="login-panel">
    <form class="login-card" id="login">
      <div class="brand"><span class="material-symbols-outlined">shield</span><h1>ProIdentity Mail</h1></div>
      <h2>Secure mailbox login</h2>
      <label>Email<input name="email" autocomplete="username" required></label>
      <label>Password<input name="password" type="password" autocomplete="current-password" required></label>
      <button class="primary-button" type="submit">Load Mailbox</button>
      <div class="mfa-panel hidden" id="mailbox-mfa-panel">
        <h3 id="mailbox-mfa-title">Two-factor verification</h3>
        <div class="muted" id="mailbox-mfa-copy">Enter the code from your authenticator app.</div>
        <div class="mailbox-push-card hidden" id="mailbox-push-card">
          <div class="mailbox-push-icon"><span class="material-symbols-outlined">phone_iphone</span></div>
          <strong>Waiting for approval...</strong>
          <div class="mailbox-push-status" id="mailbox-push-status">Waiting for approval...</div>
          <button class="link-button" type="button" id="mailbox-push-manual">Enter code manually</button>
        </div>
        <div id="mailbox-mfa-qr"></div>
        <label id="mailbox-mfa-code-row">Authenticator code<input id="mailbox-mfa-code" inputmode="numeric" autocomplete="one-time-code" placeholder="123456"></label>
        <div class="actions">
          <button class="primary-button" type="button" id="mailbox-mfa-submit">Verify code</button>
          <button class="secondary-button hidden" type="button" id="mailbox-push-check">Check push</button>
          <button class="secondary-button" type="button" id="mailbox-mfa-cancel">Cancel</button>
        </div>
      </div>
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
          <span class="compose-row-label">From</span>
          <select name="from" id="compose-from" aria-label="Sender mailbox"></select>
        </div>
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
          <select class="compose-toolbar-select" data-editor-command="formatBlock" title="Paragraph style">
            <option value="div">Normal</option>
            <option value="h2">Heading</option>
            <option value="blockquote">Quote</option>
            <option value="pre">Code</option>
          </select>
          <button class="tool-button" type="button" data-editor-command="bold" title="Bold"><span class="material-symbols-outlined">format_bold</span></button>
          <button class="tool-button" type="button" data-editor-command="italic" title="Italic"><span class="material-symbols-outlined">format_italic</span></button>
          <button class="tool-button" type="button" data-editor-command="underline" title="Underline"><span class="material-symbols-outlined">format_underlined</span></button>
          <button class="tool-button" type="button" data-editor-command="strikeThrough" title="Strikethrough"><span class="material-symbols-outlined">strikethrough_s</span></button>
        </div>
        <div class="compose-toolbar-group">
          <button class="tool-button" type="button" data-editor-command="insertUnorderedList" title="Bullet list"><span class="material-symbols-outlined">format_list_bulleted</span></button>
          <button class="tool-button" type="button" data-editor-command="insertOrderedList" title="Numbered list"><span class="material-symbols-outlined">format_list_numbered</span></button>
          <button class="tool-button" type="button" data-editor-command="outdent" title="Decrease indent"><span class="material-symbols-outlined">format_indent_decrease</span></button>
          <button class="tool-button" type="button" data-editor-command="indent" title="Increase indent"><span class="material-symbols-outlined">format_indent_increase</span></button>
        </div>
        <div class="compose-toolbar-group">
          <button class="tool-button" type="button" data-editor-command="justifyLeft" title="Align left"><span class="material-symbols-outlined">format_align_left</span></button>
          <button class="tool-button" type="button" data-editor-command="justifyCenter" title="Align center"><span class="material-symbols-outlined">format_align_center</span></button>
          <button class="tool-button" type="button" data-editor-command="justifyRight" title="Align right"><span class="material-symbols-outlined">format_align_right</span></button>
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
        <input type="hidden" name="body_html">
      </div>
      <div class="compose-attachments">
        <label class="attachment-chip" for="attachment-input"><span class="material-symbols-outlined">add_circle</span><span>Add file<br><span class="muted">up to 25 MB total</span></span></label>
        <input class="attachment-input" id="attachment-input" name="attachments" type="file" multiple>
        <div class="attachment-list" id="attachment-list"></div>
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

  <form class="modal hidden" id="profile-modal">
    <div class="modal-head"><h2>Account settings</h2><button class="tool-button" type="button" id="close-profile" title="Close"><span class="material-symbols-outlined">close</span></button></div>
    <div class="profile-settings-body">
      <section class="profile-account-summary">
        <span class="muted small">Signed in as</span>
        <strong id="profile-modal-name">--</strong>
        <code id="profile-modal-email">--</code>
        <span class="muted small">Name and mailbox address are managed by an administrator so routing, aliases, and storage stay aligned.</span>
      </section>
      <div class="profile-settings-grid">
        <section>
          <label>Language<select name="language" id="profile-language"></select></label>
          <label>Signature<textarea name="signature_html" rows="7" placeholder="Regards,&#10;Your name"></textarea></label>
          <label><span><input name="signature_auto_add" type="checkbox"> Add signature automatically to new mail</span></label>
        </section>
        <section class="connect-box">
          <strong>Change password</strong>
          <label>Current password<input id="password-current" name="current_password" type="password" autocomplete="current-password"></label>
          <label>New password<input id="password-new" name="new_password" type="password" autocomplete="new-password"></label>
          <label>Confirm new password<input id="password-confirm" name="confirm_password" type="password" autocomplete="new-password"></label>
        </section>
        <section class="connect-box">
          <strong>Mailbox 2FA</strong>
          <div class="muted" id="mailbox-security-summary">Loading security state...</div>
          <div id="profile-mfa-enroll"></div>
        </section>
        <section class="connect-box">
          <strong>App passwords</strong>
          <div class="muted" id="app-password-summary">Use app passwords for IMAP, SMTP, POP3, CalDAV, and CardDAV clients.</div>
          <label>App password name<input id="app-password-name" type="text" placeholder="Thunderbird laptop"></label>
          <button class="secondary-button" type="button" id="create-app-password">Create app password</button>
          <div id="app-password-secret"></div>
          <div id="app-password-list"></div>
        </section>
      </div>
    </div>
    <div class="modal-actions">
      <button class="secondary-button" type="button" id="cancel-profile">Cancel</button>
      <button class="primary-button" type="submit">Save settings</button>
    </div>
    <div class="error" id="profile-error"></div>
  </form>

  <script nonce="__PROIDENTITY_CSP_NONCE__">
    const cspNonce = document.currentScript ? document.currentScript.nonce : "";
    const state = { csrf: "", email: "", language: "en", messages: [], selected: null, selectedIds: new Set(), selectionAnchor: null, folder: "inbox", folders: [], filters: [], contacts: [], events: [], view: "mail", dragging: null, allowExternalSources: new Set(), contentTrust: [], mailboxes: [], activeMailbox: "", messageTab: "focused", compactList: false, collapsedSections: new Set(), profile: {}, mailboxSecurity: null, appPasswords: [], pendingMailboxMFA: null, mailboxMFAPolling: false, attachments: [], currentDraftId: "", deleteSelectionMode: "", mobilePane: "list" };
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
    let toastTimer = null;
    const mobileLayoutQuery = window.matchMedia("(max-width: 760px)");
    function isMobileLayout() {
      return mobileLayoutQuery.matches;
    }
    function setMobilePane(pane) {
      const next = ["sidebar", "list", "reader"].includes(pane) ? pane : "list";
      state.mobilePane = next;
      document.body.classList.remove("mobile-pane-sidebar", "mobile-pane-list", "mobile-pane-reader");
      document.body.classList.add("mobile-pane-" + next);
      document.querySelectorAll("[data-mobile-pane]").forEach(button => {
        button.classList.toggle("active", button.dataset.mobilePane === next);
        button.setAttribute("aria-pressed", button.dataset.mobilePane === next ? "true" : "false");
      });
    }
    function syncMobilePane() {
      if (state.view !== "mail") {
        setMobilePane("reader");
        return;
      }
      setMobilePane(state.mobilePane || "list");
    }
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
    const mailboxLabel = value => String(value || "").split("@")[0] || "Mailbox";
    const languageOptions = current => supportedLanguages.map(item => "<option value=\"" + esc(item[0]) + "\" " + (String(current || "en") === item[0] ? "selected" : "") + ">" + esc(item[2] + " / " + item[1]) + "</option>").join("");
    const translationSkipSelector = "code, pre, input, textarea, .material-symbols-outlined, .message-display-body, .mail-html-frame, #messages .message-body";
    function t(value) {
      const key = String(value ?? "");
      return (i18nCatalog[state.language] && i18nCatalog[state.language][key]) || (i18nCatalog.en && i18nCatalog.en[key]) || key;
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
    const applyLanguage = code => {
      const normalized = supportedLanguages.some(item => item[0] === code) ? code : "en";
      document.documentElement.lang = normalized;
      state.language = normalized;
      translateUI(document.body);
    };
    const avatarStyle = value => {
      const palette = [["#4648d4", "#0f6cbd"], ["#0b7a75", "#2563eb"], ["#8b5cf6", "#db2777"], ["#b45309", "#dc2626"], ["#047857", "#65a30d"]];
      const text = String(value || "");
      let hash = 0;
      for (let i = 0; i < text.length; i++) hash = (hash * 31 + text.charCodeAt(i)) >>> 0;
      const pair = palette[hash % palette.length];
      return "background:linear-gradient(135deg," + pair[0] + "," + pair[1] + ")";
    };
    const serviceBase = () => location.origin.replace(/^http:/, "https:");
    const emailOnly = value => {
      const match = String(value || "").match(/<([^>]+)>/);
      return (match ? match[1] : String(value || "")).replace(/"/g, "").trim();
    };
    const emailDomain = value => {
      const address = emailOnly(value).toLowerCase();
      const at = address.lastIndexOf("@");
      return at > -1 ? address.slice(at + 1) : "";
    };
    const publicTrustDomains = new Set([
      "aol.com", "azet.sk", "centrum.sk", "fastmail.com", "gmail.com", "gmx.com", "gmx.net", "google.com", "googlemail.com",
      "hotmail.com", "icloud.com", "laposte.net", "libero.it", "live.com", "mail.com", "mail.ru", "me.com", "msn.com",
      "outlook.com", "outlook.xyz", "pm.me", "post.cz", "proton.me", "protonmail.com", "seznam.cz", "t-online.de",
      "tutanota.com", "web.de", "yahoo.com", "yandex.com", "yandex.ru", "ymail.com", "zoho.com"
    ]);
    const normalizeContentTrustDomain = value => {
      let domain = String(value || "").trim().toLowerCase().replace(/^@+/, "").replace(/\.+$/, "");
      const at = domain.lastIndexOf("@");
      if (at >= 0) domain = domain.slice(at + 1);
      return domain;
    };
    function isBlockedPublicTrustDomain(value) {
      const domain = normalizeContentTrustDomain(value);
      if (!domain) return false;
      for (const provider of publicTrustDomains) {
        if (domain === provider || domain.endsWith("." + provider)) return true;
      }
      return false;
    }
    function contentTrustHas(scope, value) {
      const normalized = scope === "domain" ? normalizeContentTrustDomain(value) : emailOnly(value).toLowerCase();
      return state.contentTrust.some(entry => String(entry.scope || "").toLowerCase() === scope && String(entry.value || "").toLowerCase() === normalized);
    }
    function messageContentTrust(detail) {
      const sender = emailOnly(detail && detail.from).toLowerCase();
      const domain = emailDomain(sender);
      const trustedSender = sender && contentTrustHas("sender", sender);
      const trustedDomain = domain && contentTrustHas("domain", domain);
      return { sender, domain, trustedSender, trustedDomain, blockedPublicDomain: isBlockedPublicTrustDomain(domain), allowed: trustedSender || trustedDomain };
    }
    const splitAddresses = value => String(value || "").split(/[;,]/).map(item => item.trim()).filter(Boolean);
    function localServerDomains() {
      ensureMailboxes();
      const domains = new Set([emailDomain(state.email), emailDomain(currentMailbox().address || currentMailbox().id)]);
      state.mailboxes.forEach(mailbox => domains.add(emailDomain(mailbox.address || mailbox.id)));
      domains.delete("");
      return domains;
    }
    function isLocalServerMessage(detail) {
      const fromDomain = emailDomain(detail && detail.from);
      if (!fromDomain) return false;
      const domains = localServerDomains();
      if (domains.has(fromDomain)) return true;
      const recipientDomains = splitAddresses((detail && detail.to) || state.email).map(emailDomain).filter(Boolean);
      return recipientDomains.includes(fromDomain);
    }
    const prefixedSubject = (prefix, subject) => {
      const text = String(subject || "");
      return text.toLowerCase().startsWith(prefix.toLowerCase()) ? text : prefix + text;
    };
    const restoreExternalSources = markup => String(markup || "").replace(/\sdata-external-(src|srcset|poster|background)=/gi, " $1=");
    function messageAuthStatus(detail) {
      const auth = detail && detail.auth;
      if (!auth || (!auth.spf && !auth.dkim && !auth.dmarc && !auth.tls)) {
        if (isLocalServerMessage(detail)) {
          return "<div class=\"message-auth-panel message-local-delivery\"><span class=\"material-symbols-outlined\">home_mail</span><span class=\"message-auth-copy\"><strong>Local server delivery</strong><span>External sender checks are not required for local mail.</span></span><span class=\"message-auth-badge\">LOCAL</span></div>";
        }
        return "<div class=\"message-auth-panel\"><span class=\"material-symbols-outlined\">info</span><span>Authentication details are not available for this message. No delivery problem was detected.</span></div>";
      }
      const items = [];
      if (auth.spf) items.push(["shield", "SPF", auth.spf]);
      if (auth.dkim) items.push(["verified_user", "DKIM", auth.dkim]);
      if (auth.dmarc) items.push(["policy", "DMARC", auth.dmarc]);
      if (auth.tls) items.push(["lock", "TLS", auth.tls]);
      const itemHTML = items.map(item => "<span class=\"message-auth-item\"><span class=\"material-symbols-outlined\">" + esc(item[0]) + "</span>" + esc(item[1]) + ": " + esc(String(item[2]).toUpperCase()) + "</span>").join("");
      const trust = auth.trusted ? "<span class=\"message-trust-label\">TRUSTED SENDER IDENTITY VERIFIED</span>" : "<span class=\"message-trust-label\">SENDER IDENTITY NOT VERIFIED</span>";
      return "<div class=\"message-auth-panel message-trust-banner\"><span class=\"message-auth-items\">" + itemHTML + "</span>" + trust + "</div>";
    }
    const mailFrameDocument = (detail, allowExternal) => {
      const markup = allowExternal ? restoreExternalSources(detail.html || "") : String(detail.html || "");
      return "<!doctype html><html><head><meta charset=\"utf-8\"><base target=\"_blank\"><style nonce=\"" + cspNonce + "\">html,body{margin:0;padding:0;background:#fff;color:#191c1e;font:15px/1.55 Arial,Helvetica,sans-serif;overflow-wrap:anywhere;}body{padding:16px;}img{max-width:100%;height:auto;}table{max-width:100%;border-collapse:collapse;}a{color:#4648d4;}</style></head><body>" + markup + "</body></html>";
    };
    const resizeMailFrame = frame => {
      try {
        const doc = frame.contentDocument || frame.contentWindow.document;
        const height = Math.min(Math.max(doc.documentElement.scrollHeight, doc.body.scrollHeight, 260) + 12, 12000);
        frame.style.height = height + "px";
      } catch {}
    };
    const setAuthenticated = value => {
      document.body.classList.toggle("auth-locked", !value);
      document.querySelector("#account-menu").classList.add("hidden");
    };
    const currentMailbox = () => state.mailboxes.find(item => item.id === state.activeMailbox) || state.mailboxes[0] || {id: state.email, name: mailboxLabel(state.email), address: state.email, kind: "Personal"};
    const mailboxAddress = mailbox => String((mailbox && (mailbox.address || mailbox.id)) || "").trim().toLowerCase();
    const isPersonalMailbox = mailbox => mailboxAddress(mailbox) === String(state.email || "").trim().toLowerCase() || String((mailbox && mailbox.kind) || "").toLowerCase() === "personal";
    const canSendAsMailbox = mailbox => isPersonalMailbox(mailbox) || (mailbox && (mailbox.can_send_as === true || mailbox.canSendAs === true));
    const sendableMailboxes = () => {
      ensureMailboxes();
      const seen = new Set();
      const allowed = state.mailboxes.filter(mailbox => {
        const address = mailboxAddress(mailbox);
        if (!address || seen.has(address) || !canSendAsMailbox(mailbox)) return false;
        seen.add(address);
        return true;
      });
      return allowed.length ? allowed : [{id: state.email, name: mailboxLabel(state.email), address: state.email, kind: "personal"}];
    };
    const preferredSenderAddress = requested => {
      const desired = String(requested || "").trim().toLowerCase();
      const options = sendableMailboxes();
      if (desired) {
        const match = options.find(mailbox => mailboxAddress(mailbox) === desired || String(mailbox.id || "").toLowerCase() === desired);
        if (match) return mailboxAddress(match);
      }
      const active = currentMailbox();
      if (canSendAsMailbox(active)) return mailboxAddress(active);
      return mailboxAddress(options[0]);
    };
    const mailboxParam = () => encodeURIComponent((currentMailbox().address || state.activeMailbox || state.email || "").toLowerCase());
    const withMailbox = path => path + (path.includes("?") ? "&" : "?") + "mailbox=" + mailboxParam();
    const ensureMailboxes = () => {
      if (!state.email) {
        state.mailboxes = [];
        state.activeMailbox = "";
        return;
      }
      if (!state.mailboxes.length) {
        state.mailboxes = [{id: state.email, name: mailboxLabel(state.email), address: state.email, kind: "Personal"}];
      }
      if (!state.activeMailbox) state.activeMailbox = state.mailboxes[0].id;
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
      translateAttributes(document.querySelector("#search"));
      syncMobilePane();
    };
    const folderKey = value => String(value || "").trim().replace(/^\./, "").toLowerCase();
    const systemFolderKeys = new Set(["inbox", "sent", "drafts", "draft", "archive", "spam", "trash"]);
    function canMarkSpamInFolder(folder = state.folder) {
      const id = folderKey(folder);
      return id === "inbox" || (id !== "" && !systemFolderKeys.has(id));
    }
    function canMarkNotSpamInFolder(folder = state.folder) {
      return folderKey(folder) === "spam";
    }
    function updateToolbarActions() {
      document.querySelector("#mark-spam").classList.toggle("action-hidden", !canMarkSpamInFolder());
      document.querySelector("#mark-ham").classList.toggle("action-hidden", !canMarkNotSpamInFolder());
    }
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
    async function loadProfile() {
      try {
        state.profile = await api("/api/v1/profile");
      } catch {
        state.profile = {email: state.email, display_name: mailboxLabel(state.email), first_name: "", last_name: "", signature_html: "", signature_auto_add: false, language: "en"};
      }
      applyLanguage(state.profile.language || "en");
      renderProfileCard();
    }
    async function loadContentTrust() {
      try {
        const entries = await api("/api/v1/content-trust");
        state.contentTrust = Array.isArray(entries) ? entries : [];
      } catch {
        state.contentTrust = [];
      }
    }
    async function loadSecuritySettings() {
      const [security, passwords] = await Promise.all([
        api("/api/v1/mfa").catch(() => null),
        api("/api/v1/app-passwords").catch(() => [])
      ]);
      state.mailboxSecurity = security;
      state.appPasswords = Array.isArray(passwords) ? passwords : [];
      renderProfileSecurity();
    }
    function renderProfileSecurity() {
      const security = state.mailboxSecurity || {};
      const summary = document.querySelector("#mailbox-security-summary");
      const appPasswordSummary = document.querySelector("#app-password-summary");
      const enroll = document.querySelector("#profile-mfa-enroll");
      const list = document.querySelector("#app-password-list");
      if (summary) summary.textContent = t(security.totp_enabled ? "Mailbox 2FA is enabled for webmail sign-in." : (security.force_mfa ? "Mailbox 2FA setup is required before normal webmail use." : "Mailbox 2FA is available for this account."));
      if (appPasswordSummary) appPasswordSummary.textContent = t(security.totp_enabled ? "Use app passwords for IMAP, SMTP, POP3, CalDAV, and CardDAV clients." : "App passwords are for mail and sync apps. Enable mailbox 2FA before creating long-lived device passwords.");
      if (enroll) enroll.innerHTML = security.totp_enabled ? "<span class=\"tag\">" + esc(t("2FA enabled")) + "</span>" : "<div class=\"actions\"><button class=\"secondary-button\" type=\"button\" id=\"profile-start-mfa\"><span class=\"material-symbols-outlined\">qr_code_2</span>" + esc(t("Set up 2FA")) + "</button></div><div id=\"profile-mfa-setup\"></div>";
      if (list) {
        list.innerHTML = (state.appPasswords || []).length
          ? (state.appPasswords || []).map(item => "<div class=\"mini-row\"><div><strong>" + esc(item.name || "App password") + "</strong><div class=\"muted\">" + esc((item.protocols || []).join(", ")) + (item.last_used_at ? " · last used " + esc(messageTime({date:item.last_used_at})) : "") + "</div></div><button class=\"secondary-button\" type=\"button\" data-revoke-app-password=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">delete</span>Revoke</button></div>").join("")
          : "<div class=\"muted\">" + esc(t("No app passwords yet.")) + "</div>";
      }
      const start = document.querySelector("#profile-start-mfa");
      if (start) start.addEventListener("click", () => startProfileMFAEnrollment().catch(error => document.querySelector("#profile-error").textContent = error.message));
      document.querySelectorAll("[data-revoke-app-password]").forEach(button => button.addEventListener("click", () => revokeAppPassword(button.dataset.revokeAppPassword).catch(error => document.querySelector("#profile-error").textContent = error.message)));
      translateUI(document.querySelector("#profile-modal"));
    }
    async function startProfileMFAEnrollment() {
      const enrollment = await api("/api/v1/mfa/totp/enroll", {method: "POST", body: JSON.stringify({})});
      const target = document.querySelector("#profile-mfa-setup");
      target.innerHTML = "<div class=\"mfa-panel\"><h3>" + esc(t("Scan setup QR")) + "</h3>" + (enrollment.qr_data_url ? "<img alt=\"Mailbox 2FA QR code\" src=\"" + esc(enrollment.qr_data_url) + "\">" : "") + "<label>" + esc(t("Authenticator code")) + "<input id=\"profile-mfa-code\" inputmode=\"numeric\" autocomplete=\"one-time-code\" placeholder=\"123456\"></label><button class=\"primary-button\" type=\"button\" id=\"profile-mfa-verify\">" + esc(t("Verify and enable")) + "</button></div>";
      document.querySelector("#profile-mfa-verify").addEventListener("click", () => verifyProfileMFAEnrollment().catch(error => document.querySelector("#profile-error").textContent = error.message));
    }
    async function verifyProfileMFAEnrollment() {
      const code = document.querySelector("#profile-mfa-code")?.value.trim() || "";
      if (!code) throw new Error("Enter the authenticator code.");
      state.mailboxSecurity = await api("/api/v1/mfa/totp/verify", {method: "POST", body: JSON.stringify({code})});
      renderProfileSecurity();
      showToast("Mailbox 2FA enabled");
    }
    async function createAppPassword() {
      const name = document.querySelector("#app-password-name").value.trim();
      const created = await api("/api/v1/app-passwords", {method: "POST", body: JSON.stringify({name, protocols:["imap","smtp","pop3","dav"]})});
      document.querySelector("#app-password-secret").innerHTML = "<div class=\"external-source-banner allowed\"><div><strong>" + esc(t("Copy this app password now.")) + "</strong><br><code>" + esc(created.secret || "") + "</code></div></div>";
      document.querySelector("#app-password-name").value = "";
      await loadSecuritySettings();
    }
    async function revokeAppPassword(id) {
      if (!confirm(t("Revoke this app password? Devices using it will stop syncing."))) return;
      await api("/api/v1/app-passwords/" + encodeURIComponent(id), {method: "DELETE"});
      await loadSecuritySettings();
      showToast("App password revoked");
    }
    async function addContentTrust(scope, value) {
      scope = String(scope || "").toLowerCase();
      value = scope === "domain" ? normalizeContentTrustDomain(value) : emailOnly(value).toLowerCase();
      if (!value) throw new Error("Nothing to trust for this message");
      if (scope === "domain" && isBlockedPublicTrustDomain(value)) {
        throw new Error("Domain trust is disabled for public mail providers. Trust this sender instead.");
      }
      const created = await api("/api/v1/content-trust", {method: "POST", body: JSON.stringify({scope, value})});
      state.contentTrust = state.contentTrust.filter(entry => !(String(entry.scope || "").toLowerCase() === scope && String(entry.value || "").toLowerCase() === value));
      state.contentTrust.push(created);
      renderReader();
      showToast(scope === "domain" ? "Domain trusted for your account" : "Sender trusted for your account");
    }
    function renderProfileCard() {
      const profile = state.profile || {};
      const displayName = profile.display_name || [profile.first_name, profile.last_name].filter(Boolean).join(" ") || mailboxLabel(state.email);
      const email = profile.email || state.email || "--";
      document.querySelector("#profile-avatar").textContent = initials(displayName || email);
      document.querySelector("#profile-avatar").style.cssText = avatarStyle(email);
      document.querySelector("#profile-name").textContent = displayName || t("User profile");
      document.querySelector("#profile-email").textContent = email;
    }
    function openProfileModal() {
      const form = document.querySelector("#profile-modal");
      const profile = state.profile || {};
      form.reset();
      document.querySelector("#profile-modal-name").textContent = profile.display_name || [profile.first_name, profile.last_name].filter(Boolean).join(" ") || mailboxLabel(state.email);
      document.querySelector("#profile-modal-email").textContent = profile.email || state.email || "";
      form.elements.language.innerHTML = languageOptions(profile.language || state.language || "en");
      form.elements.language.value = profile.language || state.language || "en";
      form.elements.signature_html.value = profile.signature_html || "";
      form.elements.signature_auto_add.checked = profile.signature_auto_add === true;
      document.querySelector("#profile-error").textContent = "";
      renderProfileSecurity();
      form.classList.remove("hidden");
      translateUI(form);
      loadSecuritySettings().catch(error => document.querySelector("#profile-error").textContent = error.message || "Security settings failed to load");
    }
    async function saveProfile(event) {
      event.preventDefault();
      const form = event.currentTarget;
      const error = document.querySelector("#profile-error");
      error.textContent = "";
      const payload = {
        signature_html: form.elements.signature_html.value.trim(),
        signature_auto_add: form.elements.signature_auto_add.checked,
        language: form.elements.language.value || "en"
      };
      try {
        state.profile = await api("/api/v1/profile", {method: "PUT", body: JSON.stringify(payload)});
        applyLanguage(state.profile.language || payload.language || "en");
        const currentPassword = form.elements.current_password.value;
        const newPassword = form.elements.new_password.value;
        const confirmPassword = form.elements.confirm_password.value;
        if (currentPassword || newPassword || confirmPassword) {
          if (newPassword !== confirmPassword) throw new Error("New password confirmation does not match");
          await api("/api/v1/password", {method: "POST", body: JSON.stringify({current_password: currentPassword, new_password: newPassword})});
        }
        renderProfileCard();
        form.classList.add("hidden");
        showToast("Settings saved");
      } catch (err) {
        error.textContent = err.message || "Profile save failed";
      }
    }
    function signatureMarkup() {
      const signature = String((state.profile && state.profile.signature_html) || "").trim();
      if (!signature) return "";
      const looksHTML = /<\/?[a-z][\s\S]*>/i.test(signature);
      const body = looksHTML ? linkifySignatureHTML(signature) : linkifySignatureText(signature);
      return "<div data-signature=\"true\"><br>" + body + "</div>";
    }
    function normalizeSignaturePhone(value, labeled) {
      const raw = String(value || "").trim();
      const startsInternational = raw.startsWith("+");
      const digits = raw.replace(/\D/g, "");
      if (digits.length < 7 || digits.length > 15) return "";
      if (!labeled && !startsInternational && !raw.startsWith("0")) return "";
      return (startsInternational ? "+" : "") + digits;
    }
    function signatureLinkMatches(text) {
      const matches = [];
      const emailPattern = /\b(?:mail|email|e-mail)?\s*:?\s*([A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,})/gi;
      const phonePattern = /(?:\b((?:phone|mobile|cell|tel|telephone)\s*:?\s*)|(^|[^\w@]))(\+?\d[\d\s().-]{5,}\d)/gi;
      let match;
      while ((match = emailPattern.exec(text)) !== null) {
        const address = match[1];
        const offset = match.index + match[0].lastIndexOf(address);
        matches.push({start: offset, end: offset + address.length, href: "mailto:" + address, text: address});
      }
      while ((match = phonePattern.exec(text)) !== null) {
        const label = match[1] || "";
        const phone = match[3];
        const href = normalizeSignaturePhone(phone, label !== "");
        if (!href) continue;
        const offset = match.index + match[0].lastIndexOf(phone);
        matches.push({start: offset, end: offset + phone.length, href: "tel:" + href, text: phone});
      }
      matches.sort((left, right) => left.start - right.start || right.end - left.end);
      const filtered = [];
      let cursor = 0;
      for (const item of matches) {
        if (item.start < cursor) continue;
        filtered.push(item);
        cursor = item.end;
      }
      return filtered;
    }
    function linkifySignatureText(text) {
      const value = String(text || "");
      let output = "";
      let cursor = 0;
      for (const match of signatureLinkMatches(value)) {
        output += esc(value.slice(cursor, match.start));
        output += "<a href=\"" + esc(match.href) + "\">" + esc(match.text) + "</a>";
        cursor = match.end;
      }
      output += esc(value.slice(cursor));
      return output.replace(/\n/g, "<br>");
    }
    function linkifySignatureHTML(html) {
      const template = document.createElement("template");
      template.innerHTML = String(html || "");
      const nodes = [];
      const walker = document.createTreeWalker(template.content, NodeFilter.SHOW_TEXT, {
        acceptNode(node) {
          if (!node.nodeValue || !node.nodeValue.trim()) return NodeFilter.FILTER_REJECT;
          const parent = node.parentElement;
          if (parent && parent.closest("a")) return NodeFilter.FILTER_REJECT;
          return NodeFilter.FILTER_ACCEPT;
        }
      });
      while (walker.nextNode()) nodes.push(walker.currentNode);
      nodes.forEach(node => {
        const linked = linkifySignatureText(node.nodeValue);
        if (linked === esc(node.nodeValue).replace(/\n/g, "<br>")) return;
        const span = document.createElement("span");
        span.innerHTML = linked;
        node.replaceWith(...span.childNodes);
      });
      return template.innerHTML;
    }
    function applyAutoSignature() {
      if (!state.profile || state.profile.signature_auto_add !== true) return;
      const editor = document.querySelector("#compose-editor");
      if (!editor || editor.querySelector("[data-signature='true']")) return;
      editor.insertAdjacentHTML("beforeend", signatureMarkup());
    }
    function clearAttachments() {
      state.attachments = [];
      const input = document.querySelector("#attachment-input");
      if (input) input.value = "";
      renderAttachments();
    }
    function renderAttachments() {
      const box = document.querySelector("#attachment-list");
      if (!box) return;
      box.innerHTML = state.attachments.map((file, index) => "<span class=\"attachment-chip\"><span class=\"material-symbols-outlined\">attach_file</span><span>" + esc(file.name) + "<br><span class=\"muted\">" + esc(Math.ceil(file.size / 1024)) + " KB</span></span><button type=\"button\" data-remove-attachment=\"" + index + "\" title=\"Remove attachment\"><span class=\"material-symbols-outlined\">close</span></button></span>").join("");
      box.querySelectorAll("[data-remove-attachment]").forEach(button => button.addEventListener("click", () => {
        state.attachments.splice(Number(button.dataset.removeAttachment), 1);
        renderAttachments();
      }));
    }
    async function responseError(response, fallback) {
      let message = fallback;
      try { message = (await response.json()).error || message; } catch {}
      return new Error(message);
    }
    function showToast(message, isError = false) {
      const box = document.querySelector("#toast");
      clearTimeout(toastTimer);
      box.textContent = t(message);
      box.className = "toast" + (isError ? " error-toast" : "");
      toastTimer = setTimeout(() => box.classList.add("hidden"), 3600);
    }
    async function loadMessages() {
      state.view = "mail";
      await loadMailboxes();
      await loadFolders();
      const response = await fetch(withMailbox("/api/v1/messages?limit=100&folder=" + encodeURIComponent(state.folder)), { credentials: "same-origin", cache: "no-store" });
      if (!response.ok) {
        document.querySelector("#login-panel").classList.remove("hidden");
        throw new Error("Mailbox authentication failed");
      }
      state.messages = await response.json();
      const visibleMessages = filteredMessages();
      state.selected = visibleMessages[0] || state.messages[0] || null;
      state.selectedIds = new Set(state.selected ? [state.selected.id] : []);
      state.selectionAnchor = state.selected ? state.selected.id : null;
      render();
    }
    async function loadMailboxes() {
      try {
        const mailboxes = await api("/api/v1/mailboxes");
        state.mailboxes = Array.isArray(mailboxes) && mailboxes.length ? mailboxes : [];
      } catch {
        state.mailboxes = [];
      }
      ensureMailboxes();
      if (!state.mailboxes.some(item => String(item.id) === String(state.activeMailbox))) {
        state.activeMailbox = state.mailboxes[0] ? state.mailboxes[0].id : "";
      }
      renderMailboxes();
      const compose = document.querySelector("#compose-modal");
      if (compose && !compose.classList.contains("hidden")) populateFromSelector();
    }
    async function loadFolders() {
      try {
        state.folders = await api(withMailbox("/api/v1/folders"));
      } catch {
        state.folders = [
          {id: "inbox", name: "Inbox", system: true, total: 0},
          {id: "drafts", name: "Drafts", system: true, total: 0},
          {id: "sent", name: "Sent", system: true, total: 0},
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
      ensureMailboxes();
      setAuthenticated(true);
      document.querySelector("#login-panel").classList.add("hidden");
      await loadProfile();
      await loadContentTrust();
      await loadMessages();
    }
    async function finishMailboxLogin(body) {
      state.csrf = body.csrf_token || "";
      state.email = body.email || state.email || "";
      state.pendingMailboxMFA = null;
      hideMailboxMFAPanel();
      setAuthenticated(true);
      document.querySelector("#login-panel").classList.add("hidden");
      await loadProfile();
      await loadContentTrust();
      await loadMessages();
    }
    function hideMailboxMFAPanel() {
      state.mailboxMFAPolling = false;
      document.querySelector("#mailbox-mfa-panel").classList.add("hidden");
      document.querySelector("#mailbox-push-card").classList.add("hidden");
      document.querySelector("#mailbox-push-check").classList.add("hidden");
      document.querySelector("#mailbox-mfa-code-row").classList.remove("hidden");
      document.querySelector("#mailbox-mfa-submit").classList.remove("hidden");
      document.querySelector("#mailbox-mfa-code").value = "";
      document.querySelector("#mailbox-mfa-qr").innerHTML = "";
    }
    function setMailboxPushStatus(message, isError = false) {
      const el = document.querySelector("#mailbox-push-status");
      if (!el) return;
      el.textContent = t(message);
      el.style.color = isError ? "var(--danger)" : "";
    }
    async function showMailboxMFAPanel(body) {
      const panel = document.querySelector("#mailbox-mfa-panel");
      const qrBox = document.querySelector("#mailbox-mfa-qr");
      const title = document.querySelector("#mailbox-mfa-title");
      const copy = document.querySelector("#mailbox-mfa-copy");
      state.pendingMailboxMFA = {token: body.mfa_token || "", purpose: body.mfa_setup_required ? "setup" : "login", provider: body.provider || "totp"};
      qrBox.innerHTML = "";
      document.querySelector("#mailbox-push-card").classList.add("hidden");
      document.querySelector("#mailbox-push-check").classList.add("hidden");
      document.querySelector("#mailbox-mfa-code-row").classList.remove("hidden");
      document.querySelector("#mailbox-mfa-submit").classList.remove("hidden");
      if (state.pendingMailboxMFA.provider === "proidentity") {
        title.textContent = t("Check your phone");
        copy.textContent = t("We sent a sign-in request to the ProIdentity app on your registered device.");
        document.querySelector("#mailbox-push-card").classList.remove("hidden");
        document.querySelector("#mailbox-push-check").classList.remove("hidden");
        document.querySelector("#mailbox-mfa-code-row").classList.add("hidden");
        document.querySelector("#mailbox-mfa-submit").classList.add("hidden");
        setMailboxPushStatus(t("Waiting for approval..."));
        panel.classList.remove("hidden");
        pollMailboxProIdentityMFA(state.pendingMailboxMFA).catch(error => setMailboxPushStatus(error.message, true));
        return;
      } else if (state.pendingMailboxMFA.purpose === "setup") {
        title.textContent = t("Set up mailbox 2FA");
        copy.textContent = t("Scan this QR code with an authenticator app, then enter the generated code.");
        const enroll = await fetch("/api/v1/mfa/totp/enroll", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json"}, body: JSON.stringify({mfa_token: state.pendingMailboxMFA.token})});
        const enrollment = await enroll.json().catch(() => ({}));
        if (!enroll.ok) throw new Error(enrollment.error || "Could not create 2FA setup");
        if (enrollment.qr_data_url) qrBox.innerHTML = "<img alt=\"Mailbox 2FA QR code\" src=\"" + esc(enrollment.qr_data_url) + "\">";
      } else {
        title.textContent = t("Two-factor verification");
        copy.textContent = t("Enter the code from your authenticator app to open this mailbox.");
      }
      panel.classList.remove("hidden");
      document.querySelector("#mailbox-mfa-code").focus();
    }
    async function finishMailboxMFA(code = "") {
      const pending = state.pendingMailboxMFA;
      if (!pending || !pending.token) throw new Error("Two-factor challenge expired. Sign in again.");
      const endpoint = pending.purpose === "setup" ? "/api/v1/mfa/totp/verify" : "/api/v1/session/mfa";
      const response = await fetch(endpoint, {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json"}, body: JSON.stringify({mfa_token: pending.token, code})});
      const body = await response.json().catch(() => ({}));
      if (response.status === 202 && pending.provider === "proidentity") return body;
      if (!response.ok) throw new Error(body.error || "Two-factor verification failed");
      await finishMailboxLogin(body);
      return body;
    }
    async function verifyMailboxMFA() {
      const pending = state.pendingMailboxMFA;
      if (!pending || !pending.token) throw new Error("Two-factor challenge expired. Sign in again.");
      const code = document.querySelector("#mailbox-mfa-code").value.trim();
      if (!code) throw new Error("Enter the authenticator code.");
      await finishMailboxMFA(code);
    }
    async function pollMailboxProIdentityMFA(pending) {
      if (!pending || !pending.token) throw new Error("Two-factor challenge expired. Sign in again.");
      if (state.mailboxMFAPolling) return;
      state.mailboxMFAPolling = true;
      try {
        setMailboxPushStatus(t("Waiting for approval..."));
        for (let attempt = 0; attempt < 45; attempt++) {
          if (!state.pendingMailboxMFA || state.pendingMailboxMFA.token !== pending.token) return;
          const body = await finishMailboxMFA("");
          if (body && body.csrf_token) return;
          await new Promise(resolve => setTimeout(resolve, 2000));
        }
        throw new Error("ProIdentity Auth approval timed out");
      } finally {
        state.mailboxMFAPolling = false;
      }
    }
    function showMailboxPushManualCode() {
      document.querySelector("#mailbox-mfa-code-row").classList.remove("hidden");
      document.querySelector("#mailbox-mfa-submit").classList.remove("hidden");
      setMailboxPushStatus(t("Enter the hosted TOTP code or approve the push request."));
      document.querySelector("#mailbox-mfa-code").focus();
    }
    function startCompose() {
      resetComposeState();
      populateFromSelector();
      loadDraft();
      applyAutoSignature();
      openCompose();
    }
    const selectedIDs = () => filteredMessages().filter(item => state.selectedIds.has(item.id)).map(item => item.id);
    const selectedCount = () => selectedIDs().length || (state.selected ? 1 : 0);
    function activeSelectionIDs() {
      const ids = selectedIDs();
      if (ids.length) return ids;
      return state.selected ? [state.selected.id] : [];
    }
    function selectMessageByID(id, additive = false) {
      const item = state.messages.find(row => row.id === id) || null;
      if (!item) return;
      if (!additive) state.selectedIds.clear();
      state.selectedIds.add(item.id);
      state.selected = item;
      state.selectionAnchor = item.id;
    }
    function selectMessageRange(toID) {
      const list = filteredMessages();
      if (!list.length) return;
      const anchorID = state.selectionAnchor || (state.selected && state.selected.id) || list[0].id;
      const start = Math.max(0, list.findIndex(item => item.id === anchorID));
      const end = Math.max(0, list.findIndex(item => item.id === toID));
      const left = Math.min(start, end);
      const right = Math.max(start, end);
      state.selectedIds.clear();
      for (let i = left; i <= right; i++) state.selectedIds.add(list[i].id);
      state.selected = list[end] || list[left] || null;
    }
    function handleMessageSelection(id, event = {}) {
      if (event.shiftKey) {
        selectMessageRange(id);
      } else if (event.ctrlKey || event.metaKey) {
        if (state.selectedIds.has(id)) state.selectedIds.delete(id);
        else state.selectedIds.add(id);
        state.selected = state.messages.find(item => item.id === id) || state.selected;
        state.selectionAnchor = id;
      } else {
        selectMessageByID(id);
      }
      render();
      if (isMobileLayout() && !(event.shiftKey || event.ctrlKey || event.metaKey)) setMobilePane("reader");
      focusMessage(id);
    }
    function focusMessage(id) {
      setTimeout(() => {
        const target = document.querySelector(".message[data-id=\"" + CSS.escape(id) + "\"]");
        if (target) target.focus({preventScroll: true});
      }, 0);
    }
    function extendKeyboardSelection(direction) {
      const list = filteredMessages();
      if (!list.length) return;
      const currentID = state.selected ? state.selected.id : list[0].id;
      const current = Math.max(0, list.findIndex(item => item.id === currentID));
      const next = Math.max(0, Math.min(list.length - 1, current + direction));
      if (state.selectionAnchor == null) state.selectionAnchor = currentID;
      selectMessageRange(list[next].id);
      render();
      focusMessage(list[next].id);
    }
    async function moveSelected(folder) {
      const ids = activeSelectionIDs();
      if (!ids.length) {
        showToast("Select messages first", true);
        return;
      }
      const response = ids.length === 1
        ? await fetch(withMailbox("/api/v1/messages/" + encodeURIComponent(ids[0]) + "/move"), {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({folder})})
        : await fetch(withMailbox("/api/v1/messages/batch/move"), {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({ids, folder})});
      if (!response.ok) throw await responseError(response, "Move failed");
      await loadMessages();
      showToast((ids.length === 1 ? "Message" : ids.length + " messages") + (folder === "trash" ? " moved to trash" : " moved to " + folder));
    }
    function restoreTargetForMessage(message) {
      const origin = folderKey(message && message.trash_origin);
      if (origin === "sent") return "sent";
      if (origin && !systemFolderKeys.has(origin)) return origin;
      return "inbox";
    }
    async function restoreSelectedFromTrash() {
      const ids = activeSelectionIDs();
      if (!ids.length) {
        showToast("Select messages first", true);
        return;
      }
      for (const id of ids) {
        const message = state.messages.find(item => item.id === id) || {};
        const folder = restoreTargetForMessage(message);
        const response = await fetch(withMailbox("/api/v1/messages/" + encodeURIComponent(id) + "/move"), {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({folder})});
        if (!response.ok) throw await responseError(response, "Restore failed");
      }
      await loadMessages();
      showToast(ids.length === 1 ? "Message restored" : ids.length + " messages restored");
    }
    async function permanentlyDeleteSelected() {
      const ids = activeSelectionIDs();
      if (!ids.length) {
        showToast("Select messages first", true);
        return;
      }
      const response = ids.length === 1
        ? await fetch(withMailbox("/api/v1/messages/" + encodeURIComponent(ids[0]) + "/delete"), {method: "DELETE", credentials: "same-origin", cache: "no-store", headers: {"X-CSRF-Token": state.csrf}})
        : await fetch(withMailbox("/api/v1/messages/batch/delete"), {method: "DELETE", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({ids})});
      if (!response.ok) throw await responseError(response, "Delete failed");
      state.selected = null;
      state.selectedIds.clear();
      state.selectionAnchor = null;
      await loadMessages();
      showToast(ids.length === 1 ? "Message deleted forever" : ids.length + " messages deleted forever");
    }
    function deleteSelectionDescription() {
      const ids = activeSelectionIDs();
      if (ids.length === 1) {
        const message = state.messages.find(item => item.id === ids[0]) || state.selected || {};
        return message.subject || message.preview || t("(no subject)");
      }
      return ids.length + " " + t("selected messages");
    }
    function showDeleteConfirmation(mode = state.folder === "trash" ? "permanent" : "trash") {
      const ids = activeSelectionIDs();
      if (!ids.length) {
        showToast("Select messages first", true);
        return;
      }
      state.deleteSelectionMode = mode;
      const deleteTitle = mode === "permanent" ? t("Delete forever") : t("Move to Trash");
      const deleteCopy = ids.length === 1
        ? (mode === "permanent" ? t("Do you really want to permanently remove this message?") : t("Do you really want to remove this message?"))
        : (mode === "permanent" ? t("Do you really want to permanently remove these messages?") : t("Do you really want to remove these messages?"));
      document.querySelector("#delete-confirm-title").textContent = deleteTitle;
      document.querySelector("#message-delete-copy").textContent = deleteCopy;
      document.querySelector("#message-delete-subject").textContent = deleteSelectionDescription();
      document.querySelector("#confirm-delete-action").innerHTML = "<span class=\"material-symbols-outlined\">" + (mode === "permanent" ? "delete_forever" : "delete") + "</span>" + esc(deleteTitle);
      document.querySelector("#delete-confirm-error").textContent = "";
      document.querySelector("#delete-confirm-modal").classList.remove("hidden");
      translateUI(document.querySelector("#delete-confirm-modal"));
    }
    function hideDeleteConfirmation() {
      document.querySelector("#delete-confirm-modal").classList.add("hidden");
      state.deleteSelectionMode = "";
    }
    async function confirmDeleteSelection(event) {
      event.preventDefault();
      try {
        if (state.deleteSelectionMode === "permanent") await permanentlyDeleteSelected();
        else await moveSelected("trash");
        hideDeleteConfirmation();
      } catch (error) {
        document.querySelector("#delete-confirm-error").textContent = error.message || "Delete failed";
      }
    }
    async function deleteSelectedForever() {
      showDeleteConfirmation("permanent");
    }
    function canDropOneMessage(message, source, target, targetFolder) {
      const origin = String(message.trash_origin || "").toLowerCase();
      if (source === "spam") return target === "trash";
      if (source === "sent") return target === "trash";
      if (source === "inbox") return target === "trash" || (target && !targetFolder.system);
      if (source === "trash") {
        if (origin === "sent") return target === "sent";
        return target === "inbox" || (target && !targetFolder.system);
      }
      return false;
    }
    function canDropMessage(targetFolder) {
      if (!state.dragging) return false;
      const source = String(state.dragging.sourceFolder || "").toLowerCase();
      const target = String(targetFolder.id || "").toLowerCase();
      const ids = Array.isArray(state.dragging.ids) && state.dragging.ids.length ? state.dragging.ids : [state.dragging.id];
      return ids.every(id => canDropOneMessage(state.messages.find(row => row.id === id) || {}, source, target, targetFolder));
    }
    async function moveDraggedMessage(targetFolder) {
      if (!state.dragging) return;
      if (!canDropMessage(targetFolder)) {
        showToast("This move is not allowed", true);
        return;
      }
      const ids = Array.isArray(state.dragging.ids) && state.dragging.ids.length ? state.dragging.ids : [state.dragging.id];
      const response = ids.length === 1
        ? await fetch(withMailbox("/api/v1/messages/" + encodeURIComponent(ids[0]) + "/move"), {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({folder: targetFolder.id})})
        : await fetch(withMailbox("/api/v1/messages/batch/move"), {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({ids, folder: targetFolder.id})});
      if (!response.ok) throw await responseError(response, "Move failed");
      await loadMessages();
      showToast((ids.length === 1 ? "Message" : ids.length + " messages") + " moved to " + targetFolder.name);
    }
    function folderIcon(folder) {
      const id = String(folder.id || "").toLowerCase();
      if (id === "inbox") return "inbox";
      if (id === "drafts") return "draft";
      if (id === "sent") return "send";
      if (id === "archive") return "archive";
      if (id === "trash") return "delete";
      if (id === "spam") return "report";
      return "folder";
    }
    function renderMailboxes() {
      ensureMailboxes();
      document.querySelector("#mailbox-list").innerHTML = state.mailboxes.map(mailbox =>
        "<button class=\"mailbox-card " + (String(mailbox.id) === String(state.activeMailbox) ? "active" : "") + "\" type=\"button\" data-mailbox=\"" + esc(mailbox.id) + "\"><span class=\"mailbox-avatar\" style=\"" + esc(avatarStyle(mailbox.address || mailbox.name)) + "\">" + esc(initials(mailbox.address || mailbox.name)) + "</span><span class=\"mailbox-details\"><span class=\"mailbox-name\">" + esc(mailbox.name || mailboxLabel(mailbox.address)) + "</span><span class=\"mailbox-address\">" + esc(mailbox.address || mailbox.id) + "</span></span><span class=\"mailbox-kind\" title=\"" + esc(mailbox.kind || "Shared") + "\">" + esc(mailbox.kind || "Shared") + "</span></button>"
      ).join("");
      renderSectionToggles();
      document.querySelectorAll("[data-mailbox]").forEach(item => item.addEventListener("click", async () => {
        state.activeMailbox = item.dataset.mailbox;
        state.folder = "inbox";
        state.selected = null;
        state.selectedIds.clear();
        state.selectionAnchor = null;
        setMobilePane("list");
        await loadMessages();
      }));
    }
    function renderSectionToggles() {
      document.querySelectorAll("[data-toggle-section]").forEach(title => title.classList.toggle("collapsed", state.collapsedSections.has(title.dataset.toggleSection)));
      document.querySelectorAll("[data-section]").forEach(body => body.classList.toggle("collapsed", state.collapsedSections.has(body.dataset.section)));
    }
    function renderFolders() {
      const folders = state.folders.length ? state.folders : [{id: "inbox", name: "Inbox", system: true, total: state.messages.length}];
      document.querySelector("#folder-list").innerHTML = folders.map(folder =>
        "<button class=\"folder " + (String(folder.id) === String(state.folder) ? "active" : "") + "\" data-folder=\"" + esc(folder.id) + "\"><span><span class=\"material-symbols-outlined\">" + folderIcon(folder) + "</span>" + esc(folder.name) + "</span>" + folderCountHTML(folder) + "</button>"
      ).join("");
      renderSectionToggles();
      document.querySelectorAll("[data-folder]").forEach(item => item.addEventListener("click", async () => {
        state.folder = item.dataset.folder;
        state.selected = null;
        state.selectedIds.clear();
        state.selectionAnchor = null;
        setMobilePane("list");
        await loadMessages();
      }));
      document.querySelectorAll("[data-folder]").forEach(item => {
        const folder = folders.find(row => String(row.id) === String(item.dataset.folder));
        item.addEventListener("dragover", event => {
          if (!state.dragging || !folder) return;
          event.preventDefault();
          item.classList.toggle("drop-allowed", canDropMessage(folder));
          item.classList.toggle("drop-denied", !canDropMessage(folder));
        });
        item.addEventListener("dragleave", () => item.classList.remove("drop-allowed", "drop-denied"));
        item.addEventListener("drop", event => {
          event.preventDefault();
          item.classList.remove("drop-allowed", "drop-denied");
          moveDraggedMessage(folder).catch(error => showToast(error.message, true)).finally(() => state.dragging = null);
        });
      });
    }
    function folderOptions(current) {
      return state.folders.filter(folder => folder.id !== "trash").map(folder => "<option value=\"" + esc(folder.id) + "\" " + (String(folder.id) === String(current) ? "selected" : "") + ">" + esc(folder.name) + "</option>").join("");
    }
    async function saveFolder(event) {
      event.preventDefault();
      const form = event.currentTarget;
      document.querySelector("#folder-error").textContent = "";
      try {
        const created = await api(withMailbox("/api/v1/folders"), {method: "POST", body: JSON.stringify({name: form.elements.name.value.trim()})});
        state.folder = created.id;
        form.reset();
        form.classList.add("hidden");
        await loadMessages();
      } catch (error) {
        document.querySelector("#folder-error").textContent = error.message;
      }
    }
    async function deleteFolder(name) {
      if (!confirm(t("Delete folder") + " " + name + "? " + t("Messages in it will be removed from this folder."))) return;
      await api(withMailbox("/api/v1/folders/" + encodeURIComponent(name)), {method: "DELETE"});
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
      translateUI(document.querySelector("#reader"));
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
      document.querySelector("#filter-title").textContent = filter.id ? t("Edit Filter") : t("Add Filter");
      document.querySelector("#filter-error").textContent = "";
      form.classList.remove("hidden");
      translateUI(form);
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
      translateUI(document.querySelector("#reader"));
    }
    function openContactModal(contact = {}) {
      const form = document.querySelector("#contact-modal");
      form.reset();
      form.elements.id.value = contact.id || "";
      form.elements.name.value = contact.name || "";
      form.elements.email.value = contact.email || "";
      document.querySelector("#contact-title").textContent = contact.id ? t("Edit Contact") : t("Add Contact");
      document.querySelector("#contact-error").textContent = "";
      form.classList.remove("hidden");
      translateUI(form);
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
      translateUI(document.querySelector("#reader"));
    }
    function openEventModal(item = {}) {
      const form = document.querySelector("#event-modal");
      form.reset();
      form.elements.id.value = item.id || "";
      form.elements.title.value = item.title || "";
      form.elements.starts_at.value = dateTimeLocal(item.starts_at);
      form.elements.ends_at.value = dateTimeLocal(item.ends_at || new Date(Date.now() + 7200000));
      document.querySelector("#event-title").textContent = item.id ? t("Edit Event") : t("Add Event");
      document.querySelector("#event-error").textContent = "";
      form.classList.remove("hidden");
      translateUI(form);
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
      const ids = activeSelectionIDs();
      if (!ids.length) {
        showToast("Select messages first", true);
        return;
      }
      for (const id of ids) {
        const response = await fetch(withMailbox("/api/v1/messages/" + encodeURIComponent(id) + "/report"), {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json", "X-CSRF-Token": state.csrf}, body: JSON.stringify({verdict})});
        if (!response.ok) throw await responseError(response, "Message report failed");
      }
      await loadMessages();
      showToast((ids.length === 1 ? "Message" : ids.length + " messages") + (verdict === "spam" ? " trained as spam and moved to Spam" : " trained as not spam and moved to Inbox"));
    }
    function closeMessageContextMenu() {
      document.querySelector("#message-context-menu").classList.add("hidden");
    }
    function contextMenuActions() {
      const single = activeSelectionIDs().length <= 1;
      const folder = folderKey(state.folder);
      const actions = [];
      if (folder === "drafts") {
        actions.push({id: "edit-draft", icon: "edit", label: "Edit draft"});
      } else if (single) {
        actions.push({id: "reply", icon: "reply", label: "Reply"});
        actions.push({id: "forward", icon: "forward", label: "Forward"});
      }
      if (canMarkSpamInFolder()) actions.push({id: "mark-spam", icon: "report", label: "Mark as spam"});
      if (canMarkNotSpamInFolder()) actions.push({id: "mark-ham", icon: "verified", label: "Mark as not spam"});
      if (folder === "inbox") actions.push({id: "archive-message", icon: "archive", label: "Archive"});
      if (folder === "inbox" || folder === "sent" || folder === "spam") actions.push({id: "trash-message", icon: "delete", label: "Move to trash", danger: true});
      if (folder === "trash") {
        actions.push({id: "restore-trash", icon: "restore_from_trash", label: "Restore"});
        actions.push({id: "delete-forever", icon: "delete_forever", label: "Delete forever", danger: true});
      }
      if (folder === "inbox") {
        state.folders.filter(item => !item.system).forEach(item => actions.push({id: "move-folder:" + item.id, icon: "drive_file_move", label: "Move to " + item.name}));
      }
      actions.push({id: "refresh", icon: "refresh", label: "Refresh"});
      return actions;
    }
    function openMessageContextMenu(event, id) {
      event.preventDefault();
      closeMessageContextMenu();
      if (!state.selectedIds.has(id)) {
        state.selectedIds.clear();
        state.selectedIds.add(id);
        state.selected = state.messages.find(item => item.id === id) || state.selected;
        state.selectionAnchor = id;
        document.querySelectorAll(".message").forEach(item => item.classList.toggle("selected", item.dataset.id === id));
      }
      const menu = document.querySelector("#message-context-menu");
      const actions = contextMenuActions();
      menu.innerHTML = actions.map(action => "<button class=\"context-menu-item " + (action.danger ? "danger" : "") + "\" type=\"button\" role=\"menuitem\" data-context-action=\"" + esc(action.id) + "\"><span class=\"material-symbols-outlined\">" + esc(action.icon) + "</span>" + esc(action.label) + "</button>").join("");
      menu.classList.remove("hidden");
      const width = 220;
      const height = Math.max(42, actions.length * 38 + 12);
      menu.style.left = Math.max(8, Math.min(event.clientX, window.innerWidth - width - 8)) + "px";
      menu.style.top = Math.max(8, Math.min(event.clientY, window.innerHeight - height - 8)) + "px";
    }
    async function runContextMenuAction(action) {
      if (action === "reply") return openResponse("reply");
      if (action === "forward") return openResponse("forward");
      if (action === "mark-spam") return reportSelected("spam");
      if (action === "mark-ham") return reportSelected("ham");
      if (action === "archive-message") return moveSelected("archive");
      if (action === "trash-message") return moveSelected("trash");
      if (action === "restore-trash") return restoreSelectedFromTrash();
      if (action === "delete-forever") return deleteSelectedForever();
      if (action === "edit-draft" && state.selected) return openDraftCompose(state.selected.id);
      if (action === "refresh") return loadMessages();
      if (action.startsWith("move-folder:")) return moveSelected(action.slice("move-folder:".length));
    }
    async function selectedDetail() {
      if (!state.selected) throw new Error("Select a message first");
      if (activeSelectionIDs().length > 1) throw new Error("Reply and forward work with one selected message");
      const response = await fetch(withMailbox("/api/v1/messages/" + encodeURIComponent(state.selected.id)), {credentials: "same-origin", cache: "no-store"});
      return response.ok ? response.json() : state.selected;
    }
    async function openResponse(mode) {
      const item = await selectedDetail();
      const form = document.querySelector("#compose-modal");
      const sender = emailOnly(item.from || state.selected.from);
      const originalBody = String(item.body || state.selected.preview || "");
      form.reset();
      clearRecipientChips();
      clearAttachments();
      populateFromSelector();
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
    async function openDraftCompose(id) {
      const detail = await api(withMailbox("/api/v1/messages/" + encodeURIComponent(id)));
      resetComposeState();
      state.currentDraftId = id;
      populateFromSelector(detail.from || "");
      if (detail.to) String(detail.to).split(",").map(item => item.trim()).filter(Boolean).forEach(addRecipientChip);
      document.querySelector("#compose-modal").elements.subject.value = detail.subject || "";
      document.querySelector("#compose-editor").innerHTML = detail.html || esc(detail.body || "").replace(/\n/g, "<br>");
      document.querySelector("#compose-error").textContent = "";
      openCompose();
    }
    function openCompose() {
      const select = document.querySelector("#compose-from");
      populateFromSelector(select ? select.value : "");
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
    function composeIsDirty() {
      const form = document.querySelector("#compose-modal");
      if (!form || form.classList.contains("hidden")) return false;
      normalizeRecipients();
      const hasRecipients = document.querySelectorAll(".recipient-chip").length > 0 || String(form.elements.to.value || "").trim() !== "";
      const hasFields = [form.elements.cc.value, form.elements.bcc.value, form.elements.subject.value].some(value => String(value || "").trim() !== "");
      const body = document.querySelector("#compose-editor").innerText.trim();
      const html = document.querySelector("#compose-editor").innerHTML.replace(/<br\s*\/?>/gi, "").trim();
      return hasRecipients || hasFields || body !== "" || html !== "" || state.attachments.length > 0;
    }
    function requestComposeClose(force = false) {
      if (!force && composeIsDirty() && !confirm(t("Discard this message draft?"))) return false;
      closeCompose();
      return true;
    }
    function resetComposeState() {
      const form = document.querySelector("#compose-modal");
      form.reset();
      clearRecipientChips();
      document.querySelector("#compose-editor").innerHTML = "";
      document.querySelector("#compose-error").textContent = "";
      clearAttachments();
      state.currentDraftId = "";
      localStorage.removeItem("proidentity-compose-draft");
    }
    function clearRecipientChips() {
      document.querySelectorAll(".recipient-chip").forEach(item => item.remove());
    }
    function populateFromSelector(preferred) {
      const select = document.querySelector("#compose-from");
      if (!select) return "";
      const options = sendableMailboxes();
      const selected = preferredSenderAddress(preferred || select.value);
      select.innerHTML = options.map(mailbox => {
        const address = mailboxAddress(mailbox);
        const name = mailbox.name || mailboxLabel(address);
        const kind = String(mailbox.kind || "").toLowerCase() === "shared" ? "Shared mailbox" : "Personal mailbox";
        const label = name && name.toLowerCase() !== address ? name + " <" + address + "> - " + kind : address + " - " + kind;
        return "<option value=\"" + esc(address) + "\" " + (address === selected ? "selected" : "") + ">" + esc(label) + "</option>";
      }).join("");
      select.value = selected;
      return selected;
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
        if (!draft || Object.keys(draft).length === 0) return;
        const form = document.querySelector("#compose-modal");
        state.currentDraftId = draft.draft_id || "";
        populateFromSelector(draft.from || "");
        if (draft.to) String(draft.to).split(",").map(item => item.trim()).filter(Boolean).forEach(addRecipientChip);
        form.elements.cc.value = draft.cc || "";
        form.elements.bcc.value = draft.bcc || "";
        form.elements.subject.value = draft.subject || "";
        document.querySelector("#compose-editor").innerHTML = draft.body_html || esc(draft.body || "").replace(/\n/g, "<br>");
      } catch {}
    }
    async function saveDraft() {
      const form = document.querySelector("#compose-modal");
      normalizeRecipients();
      const data = new FormData(form);
      const recipients = [normalizeRecipients().join(","), data.get("cc"), data.get("bcc")].flatMap(value => String(value || "").split(",").map(item => item.trim()).filter(Boolean));
      data.delete("to");
      recipients.forEach(recipient => data.append("to", recipient));
      data.set("from", String(data.get("from") || ""));
      data.set("subject", String(data.get("subject") || ""));
      data.set("body", document.querySelector("#compose-editor").innerText.trim());
      data.set("body_html", document.querySelector("#compose-editor").innerHTML.trim());
      data.set("draft_id", state.currentDraftId || "");
      data.delete("attachments");
      state.attachments.forEach(file => data.append("attachments", file, file.name));
      const response = await fetch("/api/v1/drafts", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"X-CSRF-Token": state.csrf}, body: data});
      const body = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(body.error || "Save draft failed");
      state.currentDraftId = body.id || "";
      localStorage.setItem("proidentity-compose-draft", JSON.stringify({draft_id: state.currentDraftId, from: data.get("from"), to: recipients.join(", "), cc: form.elements.cc.value, bcc: form.elements.bcc.value, subject: data.get("subject"), body: data.get("body"), body_html: data.get("body_html")}));
      document.querySelector("#compose-error").className = "error info";
      document.querySelector("#compose-error").textContent = t("Draft saved to mailbox");
      await loadFolders();
    }
    function filteredMessages() {
      const q = document.querySelector("#search").value.trim().toLowerCase();
      const tabbed = state.messageTab === "other" ? state.messages.filter(isOtherMessage) : state.messages.filter(item => !isOtherMessage(item) || state.folder !== "inbox");
      if (!q) return tabbed;
      return tabbed.filter(item => [item.from, item.to, item.subject, item.preview].some(value => String(value || "").toLowerCase().includes(q)));
    }
    function isOtherMessage(item) {
      const text = [item.from, item.subject, item.preview].join(" ").toLowerCase();
      return /no-reply|noreply|newsletter|notification|receipt|digest|update|marketing|auto-submitted|automatic|automatick|microsoft outlook|list-unsubscribe|bulk mail|system message/.test(text);
    }
    function messageGroupLabel(item) {
      const date = item.date ? new Date(item.date) : new Date();
      if (Number.isNaN(date.getTime())) return "Earlier";
      const today = new Date();
      const yesterday = new Date(Date.now() - 86400000);
      if (dateKey(date) === dateKey(today)) return "Today";
      if (dateKey(date) === dateKey(yesterday)) return "Yesterday";
      return date.toLocaleDateString([], {month: "short", day: "numeric", year: date.getFullYear() === today.getFullYear() ? undefined : "numeric"});
    }
    function renderMessageList(list, draggable) {
      if (!list.length) return "<div class=\"message-group\"><span class=\"material-symbols-outlined\">mail</span>" + esc(t("No messages")) + "</div>";
      let currentGroup = "";
      return list.map((item, index) => {
        const group = messageGroupLabel(item);
        const active = state.selected && state.selected.id === item.id ? " active" : "";
        const selected = state.selectedIds.has(item.id) ? " selected" : "";
        const unread = item.unread ? " unread" : "";
        const origin = item.trash_origin ? "<span class=\"tag\">FROM " + esc(item.trash_origin).toUpperCase() + "</span>" : "";
        const tag = state.folder === "trash" && origin ? origin : (/spam|security|dkim|spf|tls/i.test(item.subject || item.preview || "") ? "<span class=\"tag\">SECURITY</span>" : "<span class=\"tag\">MAIL</span>");
        const from = shortFrom(item.from);
        const groupHTML = group !== currentGroup ? "<div class=\"message-group\"><span class=\"material-symbols-outlined\">expand_more</span>" + esc(group) + "</div>" : "";
        currentGroup = group;
        return groupHTML + "<button class=\"message" + active + selected + unread + "\" data-id=\"" + esc(item.id) + "\" data-index=\"" + index + "\" draggable=\"" + (draggable ? "true" : "false") + "\" aria-selected=\"" + (state.selectedIds.has(item.id) ? "true" : "false") + "\"><span class=\"message-caret\"><span class=\"material-symbols-outlined\">chevron_right</span></span><span class=\"message-avatar\" style=\"" + esc(avatarStyle(item.from || item.subject)) + "\">" + esc(initials(emailOnly(item.from) || from)) + "</span><span class=\"message-body\"><span class=\"message-top\"><span class=\"from\">" + esc(from) + "</span><span class=\"time\">" + esc(messageTime(item)) + "</span></span><span class=\"subject\">" + esc(item.subject || "(no subject)") + "</span><span class=\"preview\">" + esc(item.preview || "") + "</span>" + tag + "</span></button>";
      }).join("");
    }
    function render() {
      if (state.view !== "mail") return;
      updateViewChrome("Search emails...");
      ensureMailboxes();
      document.querySelector("#avatar").textContent = initials(state.email);
      document.querySelector("#account-email").textContent = state.email || "--";
      document.querySelector("#trash-message").title = state.folder === "trash" ? "Delete forever" : "Delete";
      updateToolbarActions();
      document.querySelector(".pane-head h2").textContent = state.folder.charAt(0).toUpperCase() + state.folder.slice(1);
      const mailbox = currentMailbox();
      document.querySelector("#pane-mailbox").textContent = (mailbox.name || mailboxLabel(mailbox.address)) + " · " + (mailbox.address || mailbox.id || "");
      document.querySelectorAll("[data-message-tab]").forEach(tab => tab.classList.toggle("active", tab.dataset.messageTab === state.messageTab));
      renderMailboxes();
      renderFolders();
      const list = filteredMessages();
      const draggable = state.folder === "inbox" || state.folder === "sent" || state.folder === "spam" || state.folder === "trash";
      const messages = document.querySelector("#messages");
      messages.classList.toggle("compact", state.compactList);
      messages.innerHTML = renderMessageList(list, draggable);
      document.querySelectorAll(".message").forEach(item => item.addEventListener("contextmenu", event => openMessageContextMenu(event, item.dataset.id)));
      const densityButton = document.querySelector("#toggle-message-density");
      if (densityButton) {
        densityButton.classList.toggle("active", state.compactList);
        densityButton.setAttribute("aria-pressed", state.compactList ? "true" : "false");
        densityButton.title = state.compactList ? "Use comfortable message list" : "Use compact message list";
      }
      document.querySelectorAll(".message[draggable='true']").forEach(item => {
        item.addEventListener("dragstart", event => {
          const ids = state.selectedIds.has(item.dataset.id) ? activeSelectionIDs() : [item.dataset.id];
          state.dragging = {id: item.dataset.id, ids, sourceFolder: state.folder};
          event.dataTransfer.effectAllowed = "move";
          event.dataTransfer.setData("text/plain", ids.join(","));
          item.classList.add("dragging");
        });
        item.addEventListener("dragend", () => {
          item.classList.remove("dragging");
          document.querySelectorAll(".folder").forEach(folder => folder.classList.remove("drop-allowed", "drop-denied"));
          state.dragging = null;
        });
      });
      renderReader();
      translateUI(document.body);
    }
    function renderReader() {
      const item = state.selected;
      const reader = document.querySelector("#reader");
      if (!item) {
        reader.className = "reader-content";
        reader.innerHTML = "<h2>No messages yet</h2><div class=\"body\">New mail delivered by Postfix and Dovecot will appear here after refresh.</div>";
        translateUI(reader);
        return;
      }
      fetch(withMailbox("/api/v1/messages/" + encodeURIComponent(item.id)), { credentials: "same-origin", cache: "no-store" })
        .then(response => response.ok ? response.json() : item)
        .then(detail => {
          if (!state.selected || state.selected.id !== item.id) return;
          const originalID = item.id;
          if (detail.id && detail.id !== item.id) {
            item.id = detail.id;
            state.messages = state.messages.map(row => row.id === originalID ? {...row, id: detail.id, unread: !!detail.unread} : row);
            if (state.selectedIds.has(originalID)) {
              state.selectedIds.delete(originalID);
              state.selectedIds.add(detail.id);
            }
            if (state.selectionAnchor === originalID) state.selectionAnchor = detail.id;
            state.selected = item;
          }
          const wasUnread = !!item.unread;
          item.unread = !!detail.unread;
          if (wasUnread && !item.unread) {
            const folder = state.folders.find(row => String(row.id) === String(state.folder));
            if (folder && Number(folder.unread || 0) > 0) folder.unread = Number(folder.unread || 0) - 1;
            const messageButton = document.querySelector(".message[data-id=\"" + CSS.escape(originalID) + "\"]") || document.querySelector(".message[data-id=\"" + CSS.escape(item.id) + "\"]");
            if (messageButton) {
              messageButton.classList.remove("unread");
              messageButton.dataset.id = item.id;
            }
            renderFolders();
          }
          const subject = detail.subject || item.subject || "(no subject)";
          const from = detail.from || item.from || "";
          const to = detail.to || item.to || state.email || "";
          const senderName = shortFrom(from);
          const trustInfo = messageContentTrust(detail);
          const allowExternal = state.allowExternalSources.has(item.id) || trustInfo.allowed;
          const externalCount = Number(detail.external_source_count || 0);
          const trustReason = trustInfo.trustedSender ? "this trusted sender" : (trustInfo.trustedDomain ? "the trusted sender domain" : "this message");
          const trustDomainButton = trustInfo.domain
            ? (trustInfo.blockedPublicDomain
              ? "<button class=\"secondary-button\" type=\"button\" disabled title=\"Domain trust is disabled for public mail providers. Trust this sender instead.\"><span class=\"material-symbols-outlined\">domain_disabled</span>Always trust this domain</button>"
              : "<button class=\"secondary-button\" type=\"button\" data-trust-external-domain=\"" + esc(trustInfo.domain) + "\"><span class=\"material-symbols-outlined\">domain</span>Always trust this domain</button>")
            : "";
          const trustSenderButton = trustInfo.sender
            ? "<button class=\"secondary-button\" type=\"button\" data-trust-external-sender=\"" + esc(trustInfo.sender) + "\"><span class=\"material-symbols-outlined\">person_check</span>Always trust this sender</button>"
            : "";
          const sourceBanner = detail.html && externalCount > 0
            ? (allowExternal
              ? "<div class=\"external-source-banner allowed\"><div><strong>External content is loaded for " + esc(trustReason) + ".</strong><br>Remote sources are visible only for your account.</div></div>"
              : "<div class=\"external-source-banner\"><div><strong>External content is blocked.</strong><br>" + esc(externalCount) + " remote source" + (externalCount === 1 ? "" : "s") + " will not load until you allow it.</div><div class=\"external-source-actions\"><button class=\"secondary-button\" type=\"button\" data-load-external-sources=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">image</span>Load external content</button>" + trustSenderButton + trustDomainButton + "</div></div>")
            : "";
          const isDraft = state.folder === "drafts" || String(detail.mailbox || item.mailbox || "").toLowerCase().includes("draft");
          const draftTools = isDraft ? "<div class=\"external-source-banner allowed\"><strong>Draft message</strong><button class=\"secondary-button\" type=\"button\" data-edit-draft=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">edit</span>Edit draft</button></div>" : "";
          const messageBody = detail.html
            ? sourceBanner + "<iframe id=\"message-html-frame\" class=\"mail-html-frame\" sandbox=\"allow-same-origin allow-popups allow-popups-to-escape-sandbox\" title=\"Message body\"></iframe>"
            : "<p class=\"plain-body\">" + esc(detail.body || item.preview || "Message body is empty.") + "</p>";
          const authPanel = messageAuthStatus(detail);
          reader.className = "reader-content mail-reader";
          reader.innerHTML =
        "<h2>" + esc(subject) + "</h2>" +
        "<div class=\"sender-row\"><div class=\"sender\"><div class=\"sender-icon\" style=\"" + esc(avatarStyle(from || subject)) + "\">" + esc(initials(emailOnly(from) || senderName)) + "</div><div><strong>" + esc(senderName) + "</strong><div class=\"muted\">" + esc(from || "") + "</div><div class=\"muted\">To: " + esc(to) + "</div></div></div><div class=\"muted\">" + esc(messageTime(detail.date ? detail : item)) + "</div></div>" +
        authPanel + "<div class=\"body message-display-body\">" + draftTools + messageBody + "</div><div class=\"message-meta\"><div class=\"recommend message-summary\"><h3>MESSAGE SUMMARY</h3><ul><li>Mailbox: " + esc(detail.mailbox || item.mailbox) + "</li><li>Size: " + esc(detail.size_bytes || item.size_bytes) + " bytes</li><li>Message ID: " + esc(item.id) + "</li></ul></div></div>";
          translateUI(reader);
          const frame = document.querySelector("#message-html-frame");
          if (frame) {
            frame.srcdoc = mailFrameDocument(detail, allowExternal);
            frame.addEventListener("load", () => resizeMailFrame(frame));
          }
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
        if (body.mfa_required || body.mfa_setup_required) {
          await showMailboxMFAPanel(body);
        } else {
          await finishMailboxLogin(body);
        }
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
    document.querySelector("#mailbox-mfa-submit").addEventListener("click", () => verifyMailboxMFA().catch(error => {
      const errorBox = document.querySelector("#error");
      errorBox.className = "error";
      errorBox.textContent = error.message || "Two-factor verification failed";
    }));
    document.querySelector("#mailbox-push-manual").addEventListener("click", showMailboxPushManualCode);
    document.querySelector("#mailbox-push-check").addEventListener("click", () => {
      state.mailboxMFAPolling = false;
      pollMailboxProIdentityMFA(state.pendingMailboxMFA).catch(error => setMailboxPushStatus(error.message, true));
    });
    document.querySelector("#mailbox-mfa-code").addEventListener("keydown", event => {
      if (event.key === "Enter") {
        event.preventDefault();
        document.querySelector("#mailbox-mfa-submit").click();
      }
    });
    document.querySelector("#mailbox-mfa-cancel").addEventListener("click", () => {
      state.pendingMailboxMFA = null;
      hideMailboxMFAPanel();
    });
    document.querySelector(".compose").addEventListener("click", startCompose);
    document.querySelector("#mobile-compose").addEventListener("click", startCompose);
    document.querySelector("#close-compose").addEventListener("click", () => requestComposeClose());
    document.querySelector("#compose-backdrop").addEventListener("click", () => requestComposeClose());
    document.querySelector("#discard-compose").addEventListener("click", () => {
      resetComposeState();
      requestComposeClose(true);
    });
    document.querySelector("#save-draft").addEventListener("click", () => saveDraft().catch(error => {
      document.querySelector("#compose-error").className = "error";
      document.querySelector("#compose-error").textContent = error.message || "Save draft failed";
    }));
    document.querySelector("#expand-compose").addEventListener("click", () => document.querySelector("#compose-modal").classList.toggle("expanded"));
    document.querySelectorAll("[data-show-compose-field]").forEach(button => button.addEventListener("click", () => document.querySelector("#" + button.dataset.showComposeField).classList.remove("hidden")));
    document.querySelector("#to-chip-row input[name='to']").addEventListener("keydown", event => {
      if (event.key === "Enter" || event.key === ",") {
        event.preventDefault();
        addRecipientChip(event.currentTarget.value.replace(",", ""));
      }
    });
    document.querySelector("#attachment-input").addEventListener("change", event => {
      const next = [...state.attachments, ...event.currentTarget.files];
      const total = next.reduce((sum, file) => sum + file.size, 0);
      if (next.length > 10 || total > 25 * 1024 * 1024) {
        showToast("You can attach up to 10 files and 25 MB total", true);
        event.currentTarget.value = "";
        return;
      }
      state.attachments = next;
      event.currentTarget.value = "";
      renderAttachments();
    });
    document.querySelector("#profile-card").addEventListener("click", openProfileModal);
    document.querySelector("#close-profile").addEventListener("click", () => document.querySelector("#profile-modal").classList.add("hidden"));
    document.querySelector("#cancel-profile").addEventListener("click", () => document.querySelector("#profile-modal").classList.add("hidden"));
    document.querySelector("#profile-modal").addEventListener("submit", saveProfile);
    document.querySelector("#create-app-password").addEventListener("click", () => createAppPassword().catch(error => document.querySelector("#profile-error").textContent = error.message || "App password creation failed"));
    document.querySelector("#close-contact").addEventListener("click", () => document.querySelector("#contact-modal").classList.add("hidden"));
    document.querySelector("#cancel-contact").addEventListener("click", () => document.querySelector("#contact-modal").classList.add("hidden"));
    document.querySelector("#contact-modal").addEventListener("submit", saveContact);
    document.querySelector("#add-folder").addEventListener("click", () => document.querySelector("#folder-modal").classList.remove("hidden"));
    document.querySelector("#close-folder").addEventListener("click", () => document.querySelector("#folder-modal").classList.add("hidden"));
    document.querySelector("#cancel-folder").addEventListener("click", () => document.querySelector("#folder-modal").classList.add("hidden"));
    document.querySelector("#folder-modal").addEventListener("submit", saveFolder);
    document.querySelector("#open-filters-pane").addEventListener("click", () => loadFiltersView().catch(error => showToast(error.message, true)));
    document.querySelector("#open-filters").addEventListener("click", () => loadFiltersView().catch(error => showToast(error.message, true)));
    document.querySelector("#open-filters-rail").addEventListener("click", () => loadFiltersView().catch(error => showToast(error.message, true)));
    document.querySelector("#toggle-message-density").addEventListener("click", () => {
      state.compactList = !state.compactList;
      render();
    });
    document.querySelector("#open-folders-rail").addEventListener("click", async () => {
      try {
        state.view = "folders";
        updateViewChrome("Search folders...");
        await loadFolders();
        document.querySelector("#reader").innerHTML = "<div class=\"workspace-head\"><div><h2>Folders</h2><div class=\"muted\">Create custom folders and open them from the left rail.</div></div><button class=\"primary-button\" id=\"add-folder-inline\" type=\"button\">New Folder</button></div><div class=\"mini-grid\">" + state.folders.map(folder => "<div class=\"mini-row\"><div><strong>" + esc(folder.name) + "</strong><div class=\"muted\">" + esc(folder.total || 0) + " messages</div></div><div class=\"compact-actions\"><button class=\"secondary-button\" data-open-folder=\"" + esc(folder.id) + "\"><span class=\"material-symbols-outlined\">folder_open</span>Open</button>" + (folder.system ? "" : "<button class=\"danger-button\" data-delete-folder=\"" + esc(folder.id) + "\"><span class=\"material-symbols-outlined\">delete</span>Delete</button>") + "</div></div>").join("") + "</div>";
        document.querySelector("#add-folder-inline").addEventListener("click", () => document.querySelector("#folder-modal").classList.remove("hidden"));
        translateUI(document.querySelector("#reader"));
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
      form.elements.body_html.value = document.querySelector("#compose-editor").innerHTML.trim();
      const data = new FormData(event.currentTarget);
      const recipients = [normalizeRecipients().join(","), data.get("cc"), data.get("bcc")].flatMap(value => String(value || "").split(",").map(item => item.trim()).filter(Boolean));
      if (!recipients.length) {
        document.querySelector("#compose-error").textContent = t("Add at least one recipient");
        return;
      }
      data.delete("to");
      recipients.forEach(recipient => data.append("to", recipient));
      data.set("from", String(data.get("from") || ""));
      data.set("subject", String(data.get("subject") || ""));
      data.set("body", String(data.get("body") || ""));
      data.set("body_html", String(data.get("body_html") || ""));
      data.set("draft_id", state.currentDraftId || "");
      data.delete("attachments");
      state.attachments.forEach(file => data.append("attachments", file, file.name));
      const response = await fetch("/api/v1/send", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"X-CSRF-Token": state.csrf}, body: data});
      if (!response.ok) {
        document.querySelector("#compose-error").textContent = t("Send failed");
        return;
      }
      form.reset();
      clearRecipientChips();
      clearAttachments();
      localStorage.removeItem("proidentity-compose-draft");
      state.currentDraftId = "";
      document.querySelector("#compose-editor").innerHTML = "";
      closeCompose();
      await loadMessages();
    });
    document.querySelector("#refresh").addEventListener("click", () => loadMessages().then(() => showToast("Mailbox refreshed")).catch(error => showToast(error.message, true)));
    document.querySelectorAll("[data-mobile-pane]").forEach(button => button.addEventListener("click", () => setMobilePane(button.dataset.mobilePane)));
    if (mobileLayoutQuery.addEventListener) mobileLayoutQuery.addEventListener("change", syncMobilePane);
    else if (mobileLayoutQuery.addListener) mobileLayoutQuery.addListener(syncMobilePane);
    document.querySelectorAll("[data-toggle-section]").forEach(button => button.addEventListener("click", () => {
      const section = button.dataset.toggleSection;
      if (state.collapsedSections.has(section)) state.collapsedSections.delete(section);
      else state.collapsedSections.add(section);
      renderSectionToggles();
    }));
    document.querySelectorAll("[data-message-tab]").forEach(tab => tab.addEventListener("click", () => {
      state.messageTab = tab.dataset.messageTab || "focused";
      const list = filteredMessages();
      state.selected = list[0] || null;
      state.selectedIds = new Set(state.selected ? [state.selected.id] : []);
      state.selectionAnchor = state.selected ? state.selected.id : null;
      render();
    }));
    document.querySelector("#mark-spam").addEventListener("click", () => reportSelected("spam").catch(error => showToast(error.message, true)));
    document.querySelector("#mark-ham").addEventListener("click", () => reportSelected("ham").catch(error => showToast(error.message, true)));
    document.querySelector("#archive-message").addEventListener("click", () => moveSelected("archive").catch(error => showToast(error.message, true)));
    document.querySelector("#trash-message").addEventListener("click", () => (state.folder === "trash" ? deleteSelectedForever() : moveSelected("trash")).catch(error => showToast(error.message, true)));
    document.querySelector("#close-delete-confirm").addEventListener("click", hideDeleteConfirmation);
    document.querySelector("#cancel-delete-confirm").addEventListener("click", hideDeleteConfirmation);
    document.querySelector("#delete-confirm-modal").addEventListener("submit", confirmDeleteSelection);
    document.querySelector("#reply-message").addEventListener("click", () => openResponse("reply").catch(error => showToast(error.message, true)));
    document.querySelector("#reply-all-message").addEventListener("click", () => openResponse("reply").catch(error => showToast(error.message, true)));
    document.querySelector("#forward-message").addEventListener("click", () => openResponse("forward").catch(error => showToast(error.message, true)));
    document.querySelector("#open-contacts").addEventListener("click", () => loadContactsView().catch(error => showToast(error.message, true)));
    document.querySelector("#open-calendar").addEventListener("click", () => loadCalendarView().catch(error => showToast(error.message, true)));
    document.querySelector("#logout").addEventListener("click", async () => {
      await api("/api/v1/session", {method: "DELETE"});
      state.csrf = "";
      state.email = "";
      state.mailboxes = [];
      state.activeMailbox = "";
      state.profile = {};
      applyLanguage("en");
      state.contentTrust = [];
      state.allowExternalSources = new Set();
      setAuthenticated(false);
      document.querySelector("#login-panel").classList.remove("hidden");
    });
    document.querySelector("#avatar").addEventListener("click", () => document.querySelector("#account-menu").classList.toggle("hidden"));
    document.querySelector("#account-close").addEventListener("click", () => document.querySelector("#account-menu").classList.add("hidden"));
    document.querySelector("#account-logout").addEventListener("click", () => document.querySelector("#logout").click());
    document.querySelectorAll("button[data-editor-command]").forEach(button => button.addEventListener("click", () => {
      document.querySelector("#compose-editor").focus();
      document.execCommand(button.dataset.editorCommand, false, null);
    }));
    document.querySelector("select[data-editor-command='formatBlock']").addEventListener("change", event => {
      document.querySelector("#compose-editor").focus();
      document.execCommand("formatBlock", false, event.currentTarget.value);
    });
    document.querySelector("[data-editor-link]").addEventListener("click", () => {
      const url = prompt(t("Link URL"));
      if (!url) return;
      document.querySelector("#compose-editor").focus();
      document.execCommand("createLink", false, url);
    });
    document.querySelector("[data-editor-image]").addEventListener("click", () => {
      const url = prompt(t("Image URL"));
      if (!url) return;
      document.querySelector("#compose-editor").focus();
      document.execCommand("insertImage", false, url);
    });
    document.querySelector("[data-editor-clear]").addEventListener("click", () => {
      document.querySelector("#compose-editor").focus();
      document.execCommand("removeFormat", false, null);
    });
    document.addEventListener("scroll", closeMessageContextMenu, true);
    document.querySelector("#search").addEventListener("input", renderCurrentView);
    document.addEventListener("click", event => {
      const contextAction = event.target.closest("[data-context-action]");
      if (contextAction) {
        const action = contextAction.dataset.contextAction;
        closeMessageContextMenu();
        runContextMenuAction(action).catch(error => showToast(error.message, true));
        return;
      }
      if (!event.target.closest("#message-context-menu")) closeMessageContextMenu();
      const loadExternal = event.target.closest("[data-load-external-sources]");
      if (loadExternal) {
        state.allowExternalSources.add(loadExternal.dataset.loadExternalSources);
        renderReader();
        return;
      }
      const trustSender = event.target.closest("[data-trust-external-sender]");
      if (trustSender) {
        addContentTrust("sender", trustSender.dataset.trustExternalSender).catch(error => showToast(error.message, true));
        return;
      }
      const trustDomain = event.target.closest("[data-trust-external-domain]");
      if (trustDomain) {
        addContentTrust("domain", trustDomain.dataset.trustExternalDomain).catch(error => showToast(error.message, true));
        return;
      }
      const editDraft = event.target.closest("[data-edit-draft]");
      if (editDraft) {
        openDraftCompose(editDraft.dataset.editDraft).catch(error => showToast(error.message, true));
        return;
      }
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
      handleMessageSelection(button.dataset.id, event);
    });
    function handleMessageKeydown(event) {
      if (event.key === "Escape") {
        closeMessageContextMenu();
        hideDeleteConfirmation();
      }
      if (state.view !== "mail") return;
      const active = document.activeElement;
      if (active && (active.matches("input, textarea, select") || active.isContentEditable)) return;
      if (event.key === "Delete") {
        event.preventDefault();
        showDeleteConfirmation(state.folder === "trash" ? "permanent" : "trash");
        return;
      }
      if (event.shiftKey && event.key === "ArrowDown") {
        event.preventDefault();
        extendKeyboardSelection(1);
      } else if (event.shiftKey && event.key === "ArrowUp") {
        event.preventDefault();
        extendKeyboardSelection(-1);
      } else if (!event.shiftKey && event.key === "ArrowDown") {
        const list = filteredMessages();
        if (!list.length) return;
        const current = Math.max(0, list.findIndex(item => state.selected && item.id === state.selected.id));
        event.preventDefault();
        const nextID = list[Math.min(list.length - 1, current + 1)].id;
        selectMessageByID(nextID);
        render();
        focusMessage(nextID);
      } else if (!event.shiftKey && event.key === "ArrowUp") {
        const list = filteredMessages();
        if (!list.length) return;
        const current = Math.max(0, list.findIndex(item => state.selected && item.id === state.selected.id));
        event.preventDefault();
        const nextID = list[Math.max(0, current - 1)].id;
        selectMessageByID(nextID);
        render();
        focusMessage(nextID);
      }
    }
    document.addEventListener("keydown", handleMessageKeydown);
    bootstrapSession().catch(error => {
      setAuthenticated(false);
      document.querySelector("#error").textContent = error.message;
    });
  </script>
</body>
</html>
`
