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
      font: 14px/1.5 "Public Sans", system-ui, sans-serif;
      letter-spacing: 0;
    }
    .material-symbols-outlined {
      font-variation-settings: "FILL" 0, "wght" 400, "GRAD" 0, "opsz" 24;
      font-size: 24px;
      line-height: 1;
    }
    header {
      height: 64px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 24px;
      border-bottom: 1px solid var(--outline);
      background: var(--background);
    }
    .brand { display: flex; align-items: center; gap: 16px; color: var(--primary); }
    .brand h1 { margin: 0; font-size: 26px; line-height: 1; font-weight: 700; }
    .top-actions { display: flex; align-items: center; gap: 24px; }
    .search {
      width: min(390px, 34vw);
      min-height: 46px;
      display: flex;
      align-items: center;
      gap: 14px;
      background: #e3e5e7;
      border-radius: 999px;
      padding: 0 20px;
      color: var(--outline-strong);
      font-weight: 600;
      letter-spacing: .08em;
    }
    .search input { width: 100%; border: 0; outline: 0; background: transparent; color: var(--ink); font: inherit; }
    .avatar {
      width: 40px;
      height: 40px;
      border-radius: 50%;
      background: #5d60f0;
      color: white;
      display: grid;
      place-items: center;
      font-weight: 700;
    }
    .app {
      height: calc(100vh - 64px);
      display: grid;
      grid-template-columns: 280px 400px minmax(420px, 1fr) 64px;
      overflow: hidden;
    }
    aside {
      background: var(--surface-soft);
      border-right: 1px solid var(--outline);
      display: flex;
      flex-direction: column;
      padding: 24px 16px;
      min-width: 0;
    }
    .compose {
      min-height: 60px;
      border: 0;
      border-radius: 12px;
      background: #5b5ef1;
      color: white;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 12px;
      font: inherit;
      font-weight: 700;
      font-size: 20px;
      cursor: pointer;
      box-shadow: 0 10px 22px rgba(70,72,212,.22);
      margin-bottom: 28px;
    }
    .folder-list { display: grid; gap: 8px; }
    .folder {
      min-height: 48px;
      border: 0;
      border-radius: 999px;
      background: transparent;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      padding: 0 18px;
      color: #272a3a;
      font: inherit;
      font-size: 17px;
      cursor: pointer;
    }
    .folder span:first-child { display: flex; align-items: center; gap: 16px; }
    .folder.active { background: var(--secondary-soft); color: #3f465c; font-weight: 700; }
    .count {
      min-width: 32px;
      border-radius: 999px;
      background: var(--primary);
      color: white;
      padding: 2px 10px;
      font-size: 12px;
      font-weight: 700;
    }
    .labels {
      margin-top: auto;
      border-top: 1px solid var(--outline);
      padding: 24px 10px 0;
      display: grid;
      gap: 12px;
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
      min-height: 62px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 18px;
      border-bottom: 1px solid var(--outline);
      background: var(--surface-soft);
    }
    .pane-head h2 { margin: 0; font-size: 22px; }
    .message-list { overflow: auto; flex: 1; }
    .message {
      border: 0;
      border-bottom: 1px solid var(--outline);
      background: var(--surface);
      width: 100%;
      text-align: left;
      padding: 20px 18px;
      cursor: pointer;
      border-left: 4px solid transparent;
      display: grid;
      gap: 6px;
      color: var(--ink);
    }
    .message:hover { background: #fafaff; }
    .message.active { background: rgba(218,226,253,.35); border-left-color: var(--primary); }
    .message-top { display: flex; justify-content: space-between; gap: 12px; }
    .from { font-weight: 700; font-size: 16px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .time { color: #6b6880; font-size: 13px; font-weight: 700; letter-spacing: .06em; white-space: nowrap; }
    .subject { font-weight: 700; font-size: 17px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .preview { color: var(--muted); font-size: 15px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
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
      min-height: 58px;
      border-bottom: 1px solid var(--outline);
      display: flex;
      align-items: center;
      gap: 24px;
      padding: 0 28px;
      color: #292b3c;
      background: var(--surface);
    }
    .security-strip {
      margin: 32px 34px 24px;
      border: 1px solid var(--outline);
      border-radius: 12px;
      background: #f8f9ff;
      min-height: 94px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 24px;
      padding: 24px;
      color: #4b536b;
      font-weight: 700;
      letter-spacing: .04em;
      overflow: hidden;
    }
    .security-items { display: flex; align-items: center; gap: 22px; flex-wrap: wrap; }
    .security-items span { display: inline-flex; align-items: center; gap: 8px; }
    .security-items .material-symbols-outlined { color: var(--primary); }
    .reader-content { padding: 8px 34px 48px; max-width: 820px; }
    .reader h2 { margin: 0 0 26px; font-size: 34px; line-height: 1.15; }
    .sender-row { display: flex; align-items: center; justify-content: space-between; gap: 24px; margin-bottom: 36px; }
    .sender { display: flex; align-items: center; gap: 16px; }
    .sender-icon {
      width: 48px;
      height: 48px;
      border-radius: 50%;
      background: var(--surface-muted);
      color: var(--primary);
      display: grid;
      place-items: center;
    }
    .sender strong { font-size: 21px; }
    .body { font-size: 19px; line-height: 1.55; }
    .recommend {
      margin: 24px 0;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: var(--surface-muted);
      padding: 24px;
    }
    .recommend h3 { margin: 0 0 12px; color: var(--primary); font-size: 15px; letter-spacing: .08em; }
    .rail {
      border-left: 1px solid var(--outline);
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 32px;
      padding-top: 34px;
      background: var(--surface-soft);
    }
    .rail .bottom { margin-top: auto; display: grid; gap: 30px; padding-bottom: 28px; }
    .login {
      position: fixed;
      inset: 64px 0 0;
      background: rgba(247,249,251,.92);
      backdrop-filter: blur(12px);
      z-index: 30;
      display: grid;
      place-items: center;
      padding: 24px;
    }
    .login.hidden { display: none; }
    .login-card {
      width: min(460px, 100%);
      border: 1px solid var(--outline);
      border-radius: 16px;
      background: white;
      box-shadow: 0 20px 48px rgba(15,23,42,.12);
      padding: 30px;
      display: grid;
      gap: 16px;
    }
    .login-card h2 { margin: 0; font-size: 28px; }
    label { display: grid; gap: 7px; color: var(--muted); font-size: 12px; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; }
    input {
      min-height: 44px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      padding: 8px 12px;
      font: inherit;
    }
    input:focus { outline: 2px solid rgba(70,72,212,.18); border-color: var(--primary); }
    .primary-button {
      min-height: 46px;
      border: 0;
      border-radius: 8px;
      background: var(--primary);
      color: white;
      font: inherit;
      font-weight: 700;
      cursor: pointer;
    }
    .error { color: var(--danger); min-height: 20px; }
    @media (max-width: 1100px) {
      body { overflow: auto; height: auto; }
      .app { height: auto; grid-template-columns: 1fr; }
      aside, .list-pane, .reader, .rail { min-height: auto; }
      .rail { display: none; }
      .search { width: 44vw; }
    }
  </style>
</head>
<body>
  <header>
    <div class="brand"><span class="material-symbols-outlined">shield</span><h1>ProIdentity Mail</h1></div>
    <div class="top-actions">
      <div class="search"><span class="material-symbols-outlined">search</span><input id="search" placeholder="Search emails..."></div>
      <span class="material-symbols-outlined" id="refresh" title="Refresh">refresh</span>
      <div class="avatar" id="avatar">--</div>
    </div>
  </header>

  <div class="app">
    <aside>
      <button class="compose" type="button"><span class="material-symbols-outlined">edit</span>Compose</button>
      <nav class="folder-list">
        <button class="folder active"><span><span class="material-symbols-outlined">inbox</span>Inbox</span><span class="count" id="inbox-count">0</span></button>
        <button class="folder"><span><span class="material-symbols-outlined">send</span>Sent</span></button>
        <button class="folder"><span><span class="material-symbols-outlined">drafts</span>Drafts</span></button>
        <button class="folder"><span><span class="material-symbols-outlined">delete</span>Trash</span></button>
        <button class="folder"><span><span class="material-symbols-outlined">report</span>Spam</span><span>0</span></button>
      </nav>
      <div class="labels">
        <h3>Labels</h3>
        <div class="label"><span class="dot danger"></span>High Security</div>
        <div class="label"><span class="dot"></span>Internal</div>
        <div class="label"><span class="dot dark"></span>Partners</div>
      </div>
    </aside>

    <section class="list-pane">
      <div class="pane-head"><h2>Inbox</h2><span class="material-symbols-outlined">filter_list</span></div>
      <div class="message-list" id="messages"></div>
    </section>

    <section class="reader">
      <div class="toolbar">
        <span class="material-symbols-outlined">archive</span>
        <span class="material-symbols-outlined">report</span>
        <span class="material-symbols-outlined">delete</span>
        <span style="height:24px;width:1px;background:var(--outline)"></span>
        <span class="material-symbols-outlined">reply</span>
        <span class="material-symbols-outlined">reply_all</span>
        <span class="material-symbols-outlined">forward</span>
      </div>
      <div class="security-strip">
        <div class="security-items">
          <span><span class="material-symbols-outlined">verified_user</span>SPF: CHECK</span>
          <span><span class="material-symbols-outlined">verified_user</span>DKIM: CHECK</span>
          <span><span class="material-symbols-outlined">lock</span>TLS: ENCRYPTED</span>
        </div>
        <div>TRUSTED SENDER IDENTITY VERIFIED</div>
      </div>
      <article class="reader-content" id="reader">
        <h2>Select a message</h2>
        <div class="body">Load your mailbox, then choose a message from the inbox list to inspect the sender, subject, and preview.</div>
      </article>
    </section>

    <aside class="rail">
      <span class="material-symbols-outlined">calendar_month</span>
      <span class="material-symbols-outlined">task_alt</span>
      <span class="material-symbols-outlined">contacts</span>
      <div class="bottom">
        <span class="material-symbols-outlined">settings</span>
        <span class="material-symbols-outlined">help</span>
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

  <script>
    const state = { token: "", email: "", messages: [], selected: null };
    const esc = value => String(value ?? "").replace(/[&<>"']/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"}[char]));
    const initials = email => String(email || "--").split("@")[0].split(/[._-]+/).filter(Boolean).slice(0, 2).map(part => part[0]).join("").toUpperCase() || "--";
    const messageTime = item => item.date ? new Date(item.date).toLocaleString([], {month: "short", day: "numeric", hour: "2-digit", minute: "2-digit"}) : "";
    const shortFrom = value => String(value || "Unknown").replace(/<.*>/, "").replace(/"/g, "").trim() || "Unknown";
    async function loadMessages() {
      const response = await fetch("/api/v1/messages?limit=100", { headers: { Authorization: "Basic " + state.token } });
      if (!response.ok) throw new Error("Mailbox authentication failed");
      state.messages = await response.json();
      state.selected = state.messages[0] || null;
      render();
    }
    function filteredMessages() {
      const q = document.querySelector("#search").value.trim().toLowerCase();
      if (!q) return state.messages;
      return state.messages.filter(item => [item.from, item.to, item.subject, item.preview].some(value => String(value || "").toLowerCase().includes(q)));
    }
    function render() {
      document.querySelector("#avatar").textContent = initials(state.email);
      document.querySelector("#inbox-count").textContent = state.messages.length;
      const list = filteredMessages();
      document.querySelector("#messages").innerHTML = list.map((item, index) => {
        const active = state.selected && state.selected.id === item.id ? " active" : "";
        const tag = /spam|security|dkim|spf|tls/i.test(item.subject || item.preview || "") ? "<span class=\"tag\">SECURITY</span>" : "<span class=\"tag\">MAIL</span>";
        return "<button class=\"message" + active + "\" data-id=\"" + esc(item.id) + "\"><div class=\"message-top\"><span class=\"from\">" + esc(shortFrom(item.from)) + "</span><span class=\"time\">" + esc(messageTime(item)) + "</span></div><div class=\"subject\">" + esc(item.subject || "(no subject)") + "</div><div class=\"preview\">" + esc(item.preview || "") + "</div>" + tag + "</button>";
      }).join("");
      renderReader();
    }
    function renderReader() {
      const item = state.selected;
      if (!item) {
        document.querySelector("#reader").innerHTML = "<h2>No messages yet</h2><div class=\"body\">New mail delivered by Postfix and Dovecot will appear here after refresh.</div>";
        return;
      }
      document.querySelector("#reader").innerHTML =
        "<h2>" + esc(item.subject || "(no subject)") + "</h2>" +
        "<div class=\"sender-row\"><div class=\"sender\"><div class=\"sender-icon\"><span class=\"material-symbols-outlined\">business</span></div><div><strong>" + esc(shortFrom(item.from)) + "</strong><div class=\"muted\">" + esc(item.from || "") + "</div><div class=\"muted\">To: " + esc(item.to || state.email) + "</div></div></div><div class=\"muted\">" + esc(messageTime(item)) + "</div></div>" +
        "<div class=\"body\"><p>" + esc(item.preview || "Message preview is empty.") + "</p><div class=\"recommend\"><h3>MESSAGE SUMMARY</h3><ul><li>Mailbox: " + esc(item.mailbox) + "</li><li>Size: " + esc(item.size_bytes) + " bytes</li><li>Message ID: " + esc(item.id) + "</li></ul></div><p>Full MIME rendering, attachments, reply, and compose actions are the next webmail backend slice.</p></div>";
    }
    document.querySelector("#login").addEventListener("submit", async event => {
      event.preventDefault();
      const data = new FormData(event.currentTarget);
      state.email = String(data.get("email") || "");
      state.token = btoa(state.email + ":" + String(data.get("password") || ""));
      document.querySelector("#error").textContent = "";
      try {
        await loadMessages();
        document.querySelector("#login-panel").classList.add("hidden");
      } catch (error) {
        document.querySelector("#error").textContent = error.message;
      }
    });
    document.querySelector("#refresh").addEventListener("click", () => loadMessages().catch(error => document.querySelector("#error").textContent = error.message));
    document.querySelector("#search").addEventListener("input", render);
    document.addEventListener("click", event => {
      const button = event.target.closest("[data-id]");
      if (!button) return;
      state.selected = state.messages.find(item => item.id === button.dataset.id) || null;
      render();
    });
  </script>
</body>
</html>
`
