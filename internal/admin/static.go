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
  <style>
    :root {
      --background: #f7f9fb;
      --surface: #ffffff;
      --surface-rail: #e0e3e5;
      --surface-soft: #f2f4f6;
      --surface-muted: #eceef0;
      --ink: #191c1e;
      --muted: #464554;
      --outline: #c7c4d7;
      --outline-strong: #767586;
      --primary: #4648d4;
      --primary-strong: #3032bd;
      --primary-soft: #e1e0ff;
      --success: #079455;
      --success-soft: #dcfae6;
      --warning: #b54708;
      --warning-soft: #fff4d6;
      --danger: #ba1a1a;
      --danger-soft: #ffdad6;
      --shadow: 0 4px 14px rgba(15, 23, 42, 0.06);
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
    .material-symbols-outlined {
      font-variation-settings: "FILL" 0, "wght" 400, "GRAD" 0, "opsz" 24;
      font-size: 22px;
      line-height: 1;
    }
    .layout { min-height: 100vh; padding-left: 240px; }
    aside {
      position: fixed;
      inset: 0 auto 0 0;
      width: 240px;
      background: var(--surface-rail);
      border-right: 1px solid rgba(118,117,134,.22);
      display: flex;
      flex-direction: column;
      padding: 18px 12px;
      z-index: 20;
    }
    .brand { display: flex; gap: 12px; align-items: center; margin-bottom: 36px; }
    .brand-mark {
      width: 36px;
      height: 36px;
      border-radius: 8px;
      background: var(--primary);
      color: white;
      display: grid;
      place-items: center;
      box-shadow: 0 8px 18px rgba(70,72,212,.25);
    }
    .brand h1 { margin: 0; color: var(--primary); font-size: 20px; line-height: 1; font-weight: 700; }
    .brand p { margin: 4px 0 0; color: var(--ink); font-size: 12px; font-weight: 600; letter-spacing: .08em; }
    nav { display: grid; gap: 4px; }
    .nav-item {
      display: flex;
      align-items: center;
      gap: 12px;
      width: 100%;
      border: 0;
      background: transparent;
      color: #252838;
      padding: 9px 12px;
      border-radius: 8px;
      font: inherit;
      font-size: 14px;
      cursor: pointer;
      text-align: left;
    }
    .nav-item:hover { background: rgba(255,255,255,.45); }
    .nav-item.active {
      color: var(--primary);
      background: rgba(255,255,255,.58);
      box-shadow: inset -4px 0 0 var(--primary);
      font-weight: 700;
    }
    .sidebar-bottom {
      margin-top: auto;
      border-top: 1px solid rgba(118,117,134,.22);
      padding-top: 16px;
      display: grid;
      gap: 4px;
    }
    header {
      position: sticky;
      top: 0;
      z-index: 10;
      height: 56px;
      background: rgba(247,249,251,.94);
      backdrop-filter: blur(12px);
      border-bottom: 1px solid var(--outline);
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 24px;
    }
    .top-left { display: flex; align-items: center; gap: 22px; }
    .page-label { margin: 0; font-size: 19px; line-height: 1; }
    .search {
      width: min(420px, 42vw);
      display: flex;
      align-items: center;
      gap: 10px;
      background: var(--surface-soft);
      border-radius: 8px;
      padding: 0 14px;
      min-height: 36px;
      color: var(--outline-strong);
    }
    .search input {
      border: 0;
      outline: 0;
      background: transparent;
      width: 100%;
      font: inherit;
      color: var(--ink);
    }
    .top-actions { display: flex; align-items: center; gap: 16px; }
    .status-pill {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      border-radius: 999px;
      background: var(--surface-muted);
      padding: 7px 14px;
      color: #2a2d3b;
      font-weight: 700;
      letter-spacing: .04em;
    }
    .avatar {
      width: 30px;
      height: 30px;
      border-radius: 50%;
      background: var(--primary);
      color: white;
      display: grid;
      place-items: center;
      font-weight: 700;
    }
    main { padding: 26px 28px 34px; max-width: 1360px; margin: 0 auto; }
    .hero {
      display: flex;
      justify-content: space-between;
      gap: 24px;
      align-items: end;
      margin-bottom: 22px;
    }
    .hero h2 { margin: 0; font-size: 28px; line-height: 1.15; }
    .hero p { margin: 6px 0 0; color: var(--muted); font-size: 14px; }
    button, input { font: inherit; }
    .primary-button {
      min-height: 42px;
      border: 0;
      border-radius: 8px;
      padding: 0 18px;
      display: inline-flex;
      align-items: center;
      gap: 12px;
      background: var(--primary);
      color: white;
      font-weight: 700;
      font-size: 14px;
      cursor: pointer;
      box-shadow: 0 8px 18px rgba(70,72,212,.22);
    }
    .primary-button:hover { background: var(--primary-strong); }
    .secondary-button {
      min-height: 36px;
      border: 1px solid var(--outline);
      border-radius: 8px;
      background: var(--surface);
      color: #272a38;
      padding: 0 14px;
      display: inline-flex;
      align-items: center;
      gap: 8px;
      font-weight: 600;
      cursor: pointer;
    }
    .stats {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 16px;
      margin-bottom: 22px;
    }
    .card {
      background: var(--surface);
      border: 1px solid var(--outline);
      border-radius: 12px;
      box-shadow: var(--shadow);
    }
    .stat-card { padding: 18px; min-height: 118px; }
    .stat-top { display: flex; justify-content: space-between; align-items: start; margin-bottom: 14px; }
    .stat-icon {
      width: 42px;
      height: 42px;
      border-radius: 8px;
      display: grid;
      place-items: center;
      background: var(--primary-soft);
      color: var(--primary);
    }
    .stat-icon.success { background: var(--success-soft); color: var(--success); }
    .stat-icon.danger { background: var(--danger-soft); color: var(--danger); }
    .trend { color: var(--success); background: #ecfdf3; border-radius: 999px; padding: 3px 10px; font-weight: 700; }
    .stat-value { font-size: 34px; line-height: 1; font-weight: 700; letter-spacing: 0; }
    .stat-label { margin-top: 6px; color: #202333; font-size: 13px; font-weight: 600; }
    .pulse {
      background: linear-gradient(135deg, #4648d4, #5558ee);
      color: white;
      padding: 18px;
      overflow: hidden;
      position: relative;
    }
    .pulse:after {
      content: "";
      position: absolute;
      right: -36px;
      bottom: -18px;
      width: 150px;
      height: 90px;
      border: 12px solid rgba(255,255,255,.12);
      transform: rotate(-42deg);
    }
    .pulse .small { letter-spacing: .18em; font-size: 12px; opacity: .9; }
    .pulse .money { margin-top: 8px; font-size: 24px; font-weight: 700; }
    .panel { overflow: hidden; }
    .panel-head {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 16px;
      padding: 18px 20px;
      border-bottom: 1px solid var(--outline);
    }
    .panel-head h3 { margin: 0; font-size: 18px; }
    .panel-actions { display: flex; gap: 10px; flex-wrap: wrap; }
    .table-wrap { overflow-x: auto; }
    table { width: 100%; border-collapse: collapse; }
    th, td { padding: 12px 18px; border-bottom: 1px solid #e6e8ef; text-align: left; vertical-align: middle; }
    th {
      background: var(--surface-soft);
      color: #232638;
      font-size: 12px;
      letter-spacing: .12em;
      text-transform: uppercase;
    }
    tbody tr:hover { background: #f8f9ff; }
    .identity { display: flex; align-items: center; gap: 14px; }
    .initials {
      width: 34px;
      height: 34px;
      border-radius: 50%;
      display: grid;
      place-items: center;
      background: #f0f1fa;
      border: 1px solid var(--outline);
      color: var(--primary);
      font-weight: 700;
    }
    .muted { color: #69677d; }
    .badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      border-radius: 999px;
      padding: 4px 11px;
      font-weight: 700;
      font-size: 12px;
      letter-spacing: .08em;
    }
    .badge.active { color: #067647; background: var(--success-soft); border: 1px solid #abefc6; }
    .badge.pending { color: #b54708; background: var(--warning-soft); border: 1px solid #fedf89; }
    .badge.disabled, .badge.suspended { color: var(--danger); background: var(--danger-soft); border: 1px solid #fecdca; }
    .forms {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 14px;
      margin-bottom: 20px;
    }
    form.card { padding: 16px; display: grid; gap: 10px; }
    form h3 { margin: 0 0 4px; font-size: 16px; }
    label { display: grid; gap: 6px; color: var(--muted); font-size: 12px; font-weight: 700; letter-spacing: .04em; text-transform: uppercase; }
    input {
      width: 100%;
      min-height: 36px;
      border: 1px solid #cbd5e1;
      border-radius: 8px;
      padding: 8px 12px;
      background: white;
      color: var(--ink);
    }
    input:focus { outline: 2px solid rgba(70,72,212,.18); border-color: var(--primary); }
    select {
      width: 100%;
      min-height: 36px;
      border: 1px solid #cbd5e1;
      border-radius: 8px;
      padding: 8px 12px;
      background: white;
      color: var(--ink);
      font: inherit;
    }
    .dns-grid { display: grid; gap: 8px; padding: 16px 20px 20px; }
    .dns-record {
      display: grid;
      grid-template-columns: 80px 1fr 2fr;
      gap: 12px;
      align-items: start;
      padding: 10px;
      background: #fbfcff;
      border: 1px solid #e5e7ef;
      border-radius: 8px;
    }
    code { overflow-wrap: anywhere; font: 12px/1.5 ui-monospace, SFMono-Regular, Consolas, monospace; }
    .toast { min-height: 20px; color: var(--muted); }
    .toast.error { color: var(--danger); }
    footer {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: center;
      color: #6d6a80;
      margin-top: 44px;
    }
    .compliance { display: flex; align-items: center; gap: 12px; }
    .compliance .material-symbols-outlined { color: var(--success); }
    .hidden { display: none !important; }
    @media (max-width: 1100px) {
      .layout { padding-left: 0; }
      aside { position: static; width: auto; height: auto; }
      .stats, .forms { grid-template-columns: 1fr; }
      header { position: static; }
      .hero { align-items: stretch; flex-direction: column; }
    }
  </style>
</head>
<body>
  <aside>
    <div class="brand">
      <div class="brand-mark"><span class="material-symbols-outlined">business</span></div>
      <div>
        <h1>ProIdentity Mail</h1>
        <p>Enterprise Admin</p>
      </div>
    </div>
    <nav>
      <button class="nav-item active" data-view="tenants"><span class="material-symbols-outlined">corporate_fare</span><span>Tenants</span></button>
      <button class="nav-item" data-view="domains"><span class="material-symbols-outlined">domain</span><span>Domains</span></button>
      <button class="nav-item" data-view="users"><span class="material-symbols-outlined">group</span><span>Users</span></button>
      <button class="nav-item" data-view="dns"><span class="material-symbols-outlined">dns</span><span>DNS Records</span></button>
      <button class="nav-item" data-view="quarantine"><span class="material-symbols-outlined">gpp_maybe</span><span>Quarantine</span></button>
      <button class="nav-item" data-view="audit"><span class="material-symbols-outlined">receipt_long</span><span>Audit Logs</span></button>
      <button class="nav-item" data-view="settings"><span class="material-symbols-outlined">settings</span><span>Settings</span></button>
    </nav>
    <div class="sidebar-bottom">
      <button class="nav-item"><span class="material-symbols-outlined">account_circle</span><span>Profile</span></button>
      <button class="nav-item"><span class="material-symbols-outlined">logout</span><span>Logout</span></button>
    </div>
  </aside>

  <div class="layout">
    <header>
      <div class="top-left">
        <h2 class="page-label" id="top-title">Tenants</h2>
        <div class="search"><span class="material-symbols-outlined">search</span><input id="search" placeholder="Search tenants by name or ID..."></div>
      </div>
      <div class="top-actions">
        <span class="status-pill"><span class="material-symbols-outlined">security</span> System Status</span>
        <span class="material-symbols-outlined">notifications</span>
        <div class="avatar">AD</div>
      </div>
    </header>

    <main>
      <section class="hero">
        <div>
          <h2 id="hero-title">Manage Tenants</h2>
          <p id="hero-subtitle">Organization-level provisioning and compliance monitoring.</p>
        </div>
        <button class="primary-button" id="primary-action" type="button"><span class="material-symbols-outlined">add</span><span>Create Tenant</span></button>
      </section>

      <section class="stats">
        <div class="card stat-card">
          <div class="stat-top"><div class="stat-icon"><span class="material-symbols-outlined">corporate_fare</span></div><span class="trend">+ live</span></div>
          <div class="stat-value" id="stat-tenants">0</div>
          <div class="stat-label">Total Tenants</div>
        </div>
        <div class="card stat-card">
          <div class="stat-top"><div class="stat-icon success"><span class="material-symbols-outlined">check_circle</span></div></div>
          <div class="stat-value" id="stat-domains">0</div>
          <div class="stat-label">Configured Domains</div>
        </div>
        <div class="card stat-card">
          <div class="stat-top"><div class="stat-icon"><span class="material-symbols-outlined">group</span></div></div>
          <div class="stat-value" id="stat-users">0</div>
          <div class="stat-label">Mailbox Users</div>
        </div>
        <div class="card pulse">
          <div class="small">SECURITY POSTURE</div>
          <div class="money">DKIM + AV</div>
          <div>Rspamd, ClamAV, SPF/DKIM/DMARC guidance active</div>
        </div>
      </section>

      <section class="forms" id="forms">
        <form class="card" id="tenant-form">
          <h3>Create Tenant</h3>
          <label>Name<input name="name" autocomplete="organization" required></label>
          <label>Slug<input name="slug" autocomplete="off" required></label>
          <button class="primary-button" type="submit"><span class="material-symbols-outlined">add</span>Create</button>
        </form>
        <form class="card" id="domain-form">
          <h3>Create Domain</h3>
          <label>Tenant ID<input name="tenant_id" inputmode="numeric" required></label>
          <label>Domain<input name="name" autocomplete="off" required></label>
          <button class="primary-button" type="submit"><span class="material-symbols-outlined">domain_add</span>Create</button>
        </form>
        <form class="card" id="user-form">
          <h3>Create User</h3>
          <label>Tenant ID<input name="tenant_id" inputmode="numeric" required></label>
          <label>Domain ID<input name="primary_domain_id" inputmode="numeric" required></label>
          <label>Local Part<input name="local_part" autocomplete="username" required></label>
          <label>Display Name<input name="display_name" autocomplete="name"></label>
          <label>Password<input name="password" type="password" autocomplete="new-password" required></label>
          <button class="primary-button" type="submit"><span class="material-symbols-outlined">person_add</span>Create</button>
        </form>
      </section>

      <p class="toast" id="status"></p>

      <section class="card panel" id="tenants-panel">
        <div class="panel-head"><h3>Registered Organizations</h3><div class="panel-actions"><button class="secondary-button"><span class="material-symbols-outlined">filter_list</span>Filters</button><button class="secondary-button"><span class="material-symbols-outlined">download</span>Export</button></div></div>
        <div class="table-wrap"><table><thead><tr><th>Name</th><th>Slug</th><th>Status</th><th>Created Date</th></tr></thead><tbody id="tenants"></tbody></table></div>
      </section>

      <section class="card panel hidden" id="domains-panel">
        <div class="panel-head"><h3>Domains Management</h3><div class="panel-actions"><button class="secondary-button"><span class="material-symbols-outlined">verified</span>DNS Verified</button></div></div>
        <div class="table-wrap"><table><thead><tr><th>Domain</th><th>Tenant</th><th>Status</th><th>DKIM</th><th>Actions</th></tr></thead><tbody id="domains"></tbody></table></div>
      </section>

      <section class="card panel hidden" id="users-panel">
        <div class="panel-head"><h3>Mailbox Users</h3><div class="panel-actions"><button class="secondary-button"><span class="material-symbols-outlined">admin_panel_settings</span>Policy</button></div></div>
        <div class="table-wrap"><table><thead><tr><th>User</th><th>Tenant</th><th>Domain</th><th>Status</th><th>Quota</th></tr></thead><tbody id="users"></tbody></table></div>
      </section>

      <section class="card panel hidden" id="dns-panel">
        <div class="panel-head"><h3>DNS Records Configuration</h3><div class="panel-actions"><button class="secondary-button" id="refresh-dns"><span class="material-symbols-outlined">refresh</span>Refresh DNS</button></div></div>
        <div class="dns-grid" id="dns"><div class="muted">Select DNS from a domain row to view MX, SPF, DKIM, DMARC, MTA-STS, and TLS reporting records.</div></div>
      </section>

      <section class="card panel hidden" id="quarantine-panel">
        <div class="panel-head"><h3>Quarantine Events</h3><div class="panel-actions"><button class="secondary-button" id="refresh-quarantine"><span class="material-symbols-outlined">refresh</span>Refresh</button></div></div>
        <div class="table-wrap"><table><thead><tr><th>Verdict</th><th>Recipient</th><th>Sender</th><th>Scanner</th><th>Action</th><th>Symbols</th><th>Date</th></tr></thead><tbody id="quarantine"></tbody></table></div>
      </section>

      <section class="card panel hidden" id="audit-panel">
        <div class="panel-head"><h3>Audit Logs</h3><div class="panel-actions"><button class="secondary-button" id="refresh-audit"><span class="material-symbols-outlined">refresh</span>Refresh</button></div></div>
        <div class="table-wrap"><table><thead><tr><th>Action</th><th>Actor</th><th>Target</th><th>Tenant</th><th>Metadata</th><th>Date</th></tr></thead><tbody id="audit"></tbody></table></div>
      </section>

      <section class="card panel hidden" id="settings-panel">
        <div class="panel-head"><h3>Tenant Mail Policies</h3><div class="panel-actions"><button class="secondary-button" id="refresh-policies"><span class="material-symbols-outlined">refresh</span>Refresh</button></div></div>
        <div class="table-wrap"><table><thead><tr><th>Tenant</th><th>Spam</th><th>Malware</th><th>TLS Auth</th><th>Actions</th></tr></thead><tbody id="policies"></tbody></table></div>
      </section>

      <section class="card panel hidden" id="placeholder-panel">
        <div class="panel-head"><h3 id="placeholder-title">Coming Next</h3></div>
        <div class="dns-grid"><div class="muted">This section is reserved for the next admin module. The visual shell is ready; backend functions will be attached as we implement each service capability.</div></div>
      </section>

      <footer>
        <div class="compliance"><span class="material-symbols-outlined">verified_user</span><span>All data processed under SOC2 compliance standards</span></div>
        <div>Legal Registry &nbsp;&nbsp; API Documentation &nbsp;&nbsp; Admin Support</div>
      </footer>
    </main>
  </div>

  <script>
    const state = { tenants: [], domains: [], users: [], quarantine: [], audit: [], policies: [], view: "tenants", dnsDomainId: null };
    const statusEl = document.querySelector("#status");
    const searchEl = document.querySelector("#search");
    const showStatus = (text, error) => { statusEl.textContent = text || ""; statusEl.className = error ? "toast error" : "toast"; };
    const api = async (path, options = {}) => {
      const response = await fetch(path, { headers: {"Content-Type": "application/json"}, ...options });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || response.statusText);
      return data;
    };
    const esc = value => String(value ?? "").replace(/[&<>"']/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"}[char]));
    const initials = value => String(value || "?").split(/[\s.-]+/).filter(Boolean).slice(0, 2).map(part => part[0]).join("").toUpperCase();
    const badge = status => "<span class=\"badge " + esc(status) + "\"><span>●</span>" + esc(status || "active") + "</span>";
    const dateText = value => value ? new Date(value).toLocaleDateString() : "Today";
    const quotaText = bytes => bytes ? Math.round(bytes / 1024 / 1024 / 1024) + " GB" : "10 GB";

    function filtered(items, fields) {
      const q = searchEl.value.trim().toLowerCase();
      if (!q) return items;
      return items.filter(item => fields.some(field => String(item[field] || "").toLowerCase().includes(q)));
    }
    function render() {
      document.querySelector("#stat-tenants").textContent = state.tenants.length;
      document.querySelector("#stat-domains").textContent = state.domains.length;
      document.querySelector("#stat-users").textContent = state.users.length;
      document.querySelector("#tenants").innerHTML = filtered(state.tenants, ["name","slug","id"]).map(item =>
        "<tr><td><div class=\"identity\"><span class=\"initials\">" + esc(initials(item.name)) + "</span><div><strong>" + esc(item.name) + "</strong><div class=\"muted\">Enterprise Plan</div></div></div></td><td class=\"muted\">" + esc(item.slug) + "</td><td>" + badge(item.status) + "</td><td>" + esc(dateText(item.created_at)) + "</td></tr>"
      ).join("");
      document.querySelector("#domains").innerHTML = filtered(state.domains, ["name","status","id"]).map(item =>
        "<tr><td><div class=\"identity\"><span class=\"initials\">" + esc(initials(item.name)) + "</span><div><strong>" + esc(item.name) + "</strong><div class=\"muted\">Domain ID " + esc(item.id) + "</div></div></div></td><td>" + esc(item.tenant_id) + "</td><td>" + badge(item.status) + "</td><td><code>" + esc(item.dkim_selector || "mail") + "</code></td><td><button class=\"secondary-button\" data-dns=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">dns</span>DNS</button></td></tr>"
      ).join("");
      document.querySelector("#users").innerHTML = filtered(state.users, ["local_part","display_name","status","id"]).map(item =>
        "<tr><td><div class=\"identity\"><span class=\"initials\">" + esc(initials(item.display_name || item.local_part)) + "</span><div><strong>" + esc(item.display_name || item.local_part) + "</strong><div class=\"muted\">" + esc(item.local_part) + "</div></div></div></td><td>" + esc(item.tenant_id) + "</td><td>" + esc(item.primary_domain_id) + "</td><td>" + badge(item.status) + "</td><td>" + esc(quotaText(item.quota_bytes)) + "</td></tr>"
      ).join("");
      document.querySelector("#quarantine").innerHTML = filtered(state.quarantine, ["recipient","sender","verdict","action","scanner","symbols_json"]).map(item =>
        "<tr><td>" + badge(item.verdict) + "</td><td><strong>" + esc(item.recipient) + "</strong><div class=\"muted\">Tenant " + esc(item.tenant_id) + "</div></td><td class=\"muted\">" + esc(item.sender || "-") + "</td><td>" + esc(item.scanner) + "</td><td>" + esc(item.action) + "</td><td><code>" + esc(item.symbols_json || "{}") + "</code></td><td>" + esc(dateText(item.created_at)) + "</td></tr>"
      ).join("");
      document.querySelector("#audit").innerHTML = filtered(state.audit, ["actor_type","action","target_type","target_id","metadata_json"]).map(item =>
        "<tr><td><strong>" + esc(item.action) + "</strong></td><td>" + esc(item.actor_type) + "<div class=\"muted\">" + esc(item.actor_id || "-") + "</div></td><td>" + esc(item.target_type) + "<div class=\"muted\">" + esc(item.target_id) + "</div></td><td>" + esc(item.tenant_id || "-") + "</td><td><code>" + esc(item.metadata_json || "{}") + "</code></td><td>" + esc(dateText(item.created_at)) + "</td></tr>"
      ).join("");
      document.querySelector("#policies").innerHTML = filtered(state.policies, ["tenant_id","spam_action","malware_action"]).map(item =>
        "<tr><td><strong>Tenant " + esc(item.tenant_id) + "</strong></td><td><select data-policy-field=\"spam_action\" data-policy=\"" + esc(item.tenant_id) + "\"><option " + selected(item.spam_action, "mark") + ">mark</option><option " + selected(item.spam_action, "quarantine") + ">quarantine</option><option " + selected(item.spam_action, "reject") + ">reject</option></select></td><td><select data-policy-field=\"malware_action\" data-policy=\"" + esc(item.tenant_id) + "\"><option " + selected(item.malware_action, "quarantine") + ">quarantine</option><option " + selected(item.malware_action, "reject") + ">reject</option></select></td><td><input type=\"checkbox\" data-policy-field=\"require_tls_for_auth\" data-policy=\"" + esc(item.tenant_id) + "\" " + (item.require_tls_for_auth ? "checked" : "") + "></td><td><button class=\"secondary-button\" data-save-policy=\"" + esc(item.tenant_id) + "\"><span class=\"material-symbols-outlined\">save</span>Save</button></td></tr>"
      ).join("");
    }
    async function refresh() {
      const [tenants, domains, users, quarantine, audit, policies] = await Promise.all([api("/api/v1/tenants"), api("/api/v1/domains"), api("/api/v1/users"), api("/api/v1/quarantine"), api("/api/v1/audit"), api("/api/v1/policies")]);
      state.tenants = tenants || [];
      state.domains = domains || [];
      state.users = users || [];
      state.quarantine = quarantine || [];
      state.audit = audit || [];
      state.policies = policies || [];
      render();
      showStatus("Loaded live platform data");
    }
    function setView(view) {
      state.view = view;
      document.querySelectorAll(".nav-item[data-view]").forEach(item => item.classList.toggle("active", item.dataset.view === view));
      ["tenants","domains","users","dns","quarantine","audit","settings"].forEach(id => document.querySelector("#" + id + "-panel").classList.toggle("hidden", view !== id));
      document.querySelector("#placeholder-panel").classList.add("hidden");
      document.querySelector("#forms").classList.toggle("hidden", !["tenants","domains","users"].includes(view));
      const copy = {
        tenants: ["Tenants", "Manage Tenants", "Organization-level provisioning and compliance monitoring.", "Search tenants by name or ID..."],
        domains: ["Domains", "Manage Domains", "Domain onboarding, verification, and deliverability records.", "Search domains by name or ID..."],
        users: ["Users", "Manage Users", "Mailbox provisioning, status review, and quota visibility.", "Search users by name or ID..."],
        dns: ["DNS Records", "DNS Records Configuration", "MX, SPF, DKIM, DMARC, MTA-STS, and TLS reporting guidance.", "Search domains..."],
        quarantine: ["Quarantine", "Quarantine Events", "Spam, malware, phishing, and policy holds before delivery.", "Search quarantine events..."],
        audit: ["Audit Logs", "Audit Logs", "Security event and administrative activity timeline.", "Search audit events..."],
        settings: ["Settings", "Settings", "Platform security and service configuration.", "Search settings..."]
      }[view];
      document.querySelector("#top-title").textContent = copy[0];
      document.querySelector("#hero-title").textContent = copy[1];
      document.querySelector("#hero-subtitle").textContent = copy[2];
      searchEl.placeholder = copy[3];
      document.querySelector("#placeholder-title").textContent = copy[1];
      render();
    }
    async function submitForm(event, path, numericFields = []) {
      event.preventDefault();
      const form = event.currentTarget;
      const body = Object.fromEntries(new FormData(form).entries());
      numericFields.forEach(field => body[field] = Number(body[field]));
      try {
        await api(path, { method: "POST", body: JSON.stringify(body) });
        form.reset();
        await refresh();
        showStatus("Saved successfully");
      } catch (error) {
        showStatus(error.message, true);
      }
    }
    const selected = (current, value) => current === value ? "selected" : "";
    async function savePolicy(tenantID) {
      const fields = document.querySelectorAll("[data-policy='" + CSS.escape(String(tenantID)) + "']");
      const body = {tenant_id: Number(tenantID)};
      fields.forEach(field => {
        if (field.type === "checkbox") body[field.dataset.policyField] = field.checked;
        else body[field.dataset.policyField] = field.value;
      });
      await api("/api/v1/policies/" + tenantID, {method: "PUT", body: JSON.stringify(body)});
      await refresh();
      showStatus("Policy saved");
    }
    async function loadDNS(id) {
      try {
        const dns = await api("/api/v1/domains/" + id + "/dns");
        setView("dns");
        document.querySelector("#dns").innerHTML = dns.records.map(record =>
          "<div class=\"dns-record\"><strong>" + esc(record.type) + "</strong><code>" + esc(record.name) + "</code><code>" + esc(record.priority ? record.priority + " " : "") + esc(record.value) + "</code></div>"
        ).join("");
        showStatus("DNS records loaded for " + dns.domain);
      } catch (error) {
        showStatus(error.message, true);
      }
    }
    document.querySelector("#tenant-form").addEventListener("submit", event => submitForm(event, "/api/v1/tenants"));
    document.querySelector("#domain-form").addEventListener("submit", event => submitForm(event, "/api/v1/domains", ["tenant_id"]));
    document.querySelector("#user-form").addEventListener("submit", event => submitForm(event, "/api/v1/users", ["tenant_id","primary_domain_id"]));
    document.querySelectorAll(".nav-item[data-view]").forEach(item => item.addEventListener("click", () => setView(item.dataset.view)));
    document.querySelector("#primary-action").addEventListener("click", () => document.querySelector("#forms").scrollIntoView({behavior: "smooth"}));
    document.querySelector("#refresh-quarantine").addEventListener("click", () => refresh().catch(error => showStatus(error.message, true)));
    document.querySelector("#refresh-audit").addEventListener("click", () => refresh().catch(error => showStatus(error.message, true)));
    document.querySelector("#refresh-policies").addEventListener("click", () => refresh().catch(error => showStatus(error.message, true)));
    document.addEventListener("click", event => { const id = event.target.closest("[data-dns]")?.dataset.dns; if (id) loadDNS(id); });
    document.addEventListener("click", event => { const id = event.target.closest("[data-save-policy]")?.dataset.savePolicy; if (id) savePolicy(id).catch(error => showStatus(error.message, true)); });
    searchEl.addEventListener("input", render);
    refresh().catch(error => showStatus(error.message, true));
  </script>
</body>
</html>
`
