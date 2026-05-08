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
      grid-template-columns: 74px minmax(150px,.45fr) minmax(260px,1fr);
      gap: 10px;
      align-items: start;
      background: #fbfcfd;
    }
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
    .login-card { width: min(430px, 100%); background: white; border: 1px solid var(--outline); border-radius: 8px; box-shadow: var(--shadow); padding: 22px; }
    .login-card h2 { margin: 0; font-size: 22px; }
    .login-card p { margin: 6px 0 18px; color: var(--muted); }
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
    </form>
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
  <script>
    const views = [
      ["dashboard", "Dashboard", "Operational overview", "space_dashboard"],
      ["onboarding", "Onboarding", "Tenant to first user setup", "fact_check"],
      ["tenants", "Tenants", "Organizations and customer boundaries", "apartment"],
      ["domains", "Domains", "Hosted domains, DNS, aliases, and catch-all", "dns"],
      ["users", "Users", "People, shared mailboxes, quota, and permissions", "group"],
      ["security", "Security", "Tenant spam, malware, and TLS policy", "shield_lock"],
      ["quarantine", "Quarantine", "Held spam and malware messages", "gpp_maybe"],
      ["audit", "Audit", "Admin and security activity", "receipt_long"],
      ["system", "System", "Service health and integration endpoints", "settings"]
    ];
    const state = {
      tenants: [], domains: [], users: [], aliases: [], catchAll: [], sharedPermissions: [], quarantine: [], audit: [], policies: [],
      view: "dashboard", domainTab: "domains", userTab: "people", selectedTenantId: "", selectedDomainId: "", dns: null, csrf: "", query: "", health: "checking"
    };

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
    const visible = items => {
      const q = state.query.trim().toLowerCase();
      if (!q) return items;
      return items.filter(item => JSON.stringify(item).toLowerCase().includes(q));
    };
    const domainsForTenant = (tenantID = state.selectedTenantId) => state.domains.filter(item => tenantID && String(item.tenant_id) === String(tenantID));
    const selectedTenantDomains = () => domainsForTenant();
    const selectedTenant = () => byID(state.tenants, state.selectedTenantId);
    const selectedDomain = () => byID(state.domains, state.selectedDomainId);
    const setupIncomplete = () => !state.tenants.length || !state.domains.length || !state.users.length;
    const activeViews = () => setupIncomplete() ? views : views.filter(item => item[0] !== "onboarding");
    const scopedUsers = () => state.users.filter(item =>
      state.selectedTenantId && String(item.tenant_id) === String(state.selectedTenantId) &&
      state.selectedDomainId && String(item.primary_domain_id) === String(state.selectedDomainId)
    );
    const filteredUsers = () => visible(state.users.filter(item =>
      (!state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId)) &&
      (!state.selectedDomainId || String(item.primary_domain_id) === String(state.selectedDomainId))
    ));
    const usersInTenant = () => state.users.filter(item => !state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId));
    const sharedMailboxes = () => usersInTenant().filter(item => (item.mailbox_type || "user") === "shared");
    const normalUsers = () => usersInTenant().filter(item => (item.mailbox_type || "user") === "user");
    const filteredNormalUsers = () => visible(normalUsers().filter(item => !state.selectedDomainId || String(item.primary_domain_id) === String(state.selectedDomainId)));
    const filteredSharedMailboxes = () => visible(sharedMailboxes().filter(item => !state.selectedDomainId || String(item.primary_domain_id) === String(state.selectedDomainId)));
    function badge(value) {
      const text = String(value || "unknown");
      const low = text.toLowerCase();
      const cls = /active|ok|good|released|mark/.test(low) ? "good" : /held|pending|quarantine|spam|warn/.test(low) ? "warn" : /reject|delete|malware|failed|bad/.test(low) ? "bad" : "";
      return "<span class=\"badge " + cls + "\">" + esc(text) + "</span>";
    }
    function showStatus(message, isError) {
      const toast = $("#toast");
      toast.textContent = message;
      toast.className = "toast" + (isError ? " error" : "");
      clearTimeout(showStatus.timer);
      showStatus.timer = setTimeout(() => toast.classList.add("hidden"), 3800);
    }
    async function api(path, options) {
      const init = Object.assign({credentials: "same-origin", cache: "no-store", headers: {}}, options || {});
      if (init.body && !init.headers["Content-Type"]) init.headers["Content-Type"] = "application/json";
      if (state.csrf && init.method && init.method !== "GET") init.headers["X-CSRF-Token"] = state.csrf;
      const response = await fetch(path, init);
      if (response.status === 204) return null;
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || response.statusText || "Request failed");
      return data;
    }
    function syncScope() {
      if (!state.tenants.some(item => String(item.id) === String(state.selectedTenantId))) {
        state.selectedTenantId = state.tenants[0] ? String(state.tenants[0].id) : "";
      }
      const domains = domainsForTenant();
      if (!domains.some(item => String(item.id) === String(state.selectedDomainId))) {
        state.selectedDomainId = domains[0] ? String(domains[0].id) : "";
      }
      if (state.dns && String(state.dns.domain_id) !== String(state.selectedDomainId)) state.dns = null;
    }
    function setTenantScope(tenantID) {
      state.selectedTenantId = String(tenantID || "");
      const nextDomain = domainsForTenant()[0];
      state.selectedDomainId = nextDomain ? String(nextDomain.id) : "";
      state.dns = null;
      state.query = "";
      $("#search").value = "";
      render();
    }
    function setDomainScope(domainID) {
      state.selectedDomainId = String(domainID || "");
      state.dns = null;
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
      const [tenants, domains, users, aliases, catchAll, sharedPermissions, quarantine, audit, policies] = await Promise.all([
        api("/api/v1/tenants"), api("/api/v1/domains"), api("/api/v1/users"),
        api("/api/v1/aliases"), api("/api/v1/catch-all"), api("/api/v1/shared-permissions"),
        api("/api/v1/quarantine"), api("/api/v1/audit"), api("/api/v1/policies")
      ]);
      state.tenants = tenants || [];
      state.domains = domains || [];
      state.users = users || [];
      state.aliases = aliases || [];
      state.catchAll = catchAll || [];
      state.sharedPermissions = sharedPermissions || [];
      state.quarantine = quarantine || [];
      state.audit = audit || [];
      state.policies = policies || [];
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
      const badges = ["Tenant: " + (tenant ? tenant.name : "none")];
      if (includeAllDomain || domain) badges.push("Domain: " + (domain ? domain.name : "none"));
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
      const domainSelect = domains.length ? "<label class=\"scope-picker\">Domain<select id=\"global-domain\">" + domainOptions(state.selectedTenantId, state.selectedDomainId) + "</select></label>" : "<div class=\"scope-empty\">No domains in this tenant yet</div>";
      const users = scopedUsers().length;
      const shared = scopedUsers().filter(item => (item.mailbox_type || "user") === "shared").length;
      const aliases = state.aliases.filter(item => domain && String(item.domain_id) === String(domain.id)).length;
      const catchAll = state.catchAll.some(item => domain && String(item.domain_id) === String(domain.id));
      bar.innerHTML = "<div class=\"scope-main\"><label class=\"scope-picker\">Tenant<select id=\"global-tenant\">" + tenantOptions(state.selectedTenantId) + "</select></label>" + domainSelect + "</div><div class=\"scope-summary\">" +
        "<span class=\"badge\">Users " + esc(users) + "</span><span class=\"badge\">Shared " + esc(shared) + "</span><span class=\"badge\">Aliases " + esc(aliases) + "</span><span class=\"badge " + (catchAll ? "good" : "") + "\">Catch-all " + esc(catchAll ? "set" : "none") + "</span>" +
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
    }
    function stat(label, value, icon, cls) {
      return "<article class=\"card stat\"><div class=\"stat-top\"><span>" + label + "</span><span class=\"stat-icon " + (cls || "") + "\"><span class=\"material-symbols-outlined\">" + icon + "</span></span></div><div><div class=\"stat-value\">" + esc(value) + "</div><div class=\"muted small\">Current live count</div></div></article>";
    }
    function renderDashboard() {
      const tenantID = state.selectedTenantId;
      const domainID = state.selectedDomainId;
      const held = state.quarantine.filter(item => (!tenantID || String(item.tenant_id) === String(tenantID)) && (!item.status || item.status === "held")).length;
      const domainCount = domainsForTenant(tenantID).length;
      const userCount = domainID ? scopedUsers().length : state.users.filter(item => tenantID && String(item.tenant_id) === String(tenantID)).length;
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
        "<div class=\"step-title\"><span>Tenant</span><strong>" + esc(tenantName(tenantID)) + "</strong></div><div class=\"step-title\"><span>Domain</span><strong>" + esc(domainName(domainID)) + "</strong></div><div class=\"step-title\"><span>Users in domain</span>" + badge(userCount) + "</div></div></section>";
      return (incomplete ? "<section class=\"card\" style=\"margin-bottom:16px\"><div class=\"panel-head\"><div><h4>Setup needs attention</h4><p>Create the first tenant, domain, and user before using advanced mail routing.</p></div><button class=\"button primary\" data-view=\"onboarding\"><span class=\"material-symbols-outlined\">rocket_launch</span>Continue setup</button></div></section>" : "") +
        "<div class=\"grid stats\">" +
        stat("Tenants", state.tenants.length, "apartment") + stat("Domains", domainCount, "dns") +
        stat("Users", userCount, "group") + stat("Held messages", held, "gpp_maybe") +
        "</div><section class=\"card\" style=\"margin-bottom:16px\"><div class=\"panel-head\"><div><h4>Common work</h4><p>Jump straight into the task without hunting through a long page.</p></div></div><div class=\"card-body quick-actions\">" +
        "<button class=\"quick-action\" data-view=\"users\" data-set-user-tab=\"people\"><strong><span class=\"material-symbols-outlined\">person_add</span>Create user</strong><span class=\"muted small\">Normal mailbox with quota and login.</span></button>" +
        "<button class=\"quick-action\" data-view=\"domains\" data-set-domain-tab=\"aliases\"><strong><span class=\"material-symbols-outlined\">alternate_email</span>Add alias</strong><span class=\"muted small\">Alternate address under a tenant domain.</span></button>" +
        "<button class=\"quick-action\" data-view=\"domains\" data-set-domain-tab=\"catchall\"><strong><span class=\"material-symbols-outlined\">all_inbox</span>Set catch-all</strong><span class=\"muted small\">Unknown recipients route to one mailbox.</span></button>" +
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
      return tenantID ? contextField("Tenant", tenantName(tenantID), "tenant_id", tenantID) : "<label>Tenant<select name=\"tenant_id\" required>" + tenantOptions(tenantID) + "</select></label>";
    }
    function domainField(tenantID, domainID, name = "primary_domain_id") {
      return domainID ? contextField("Domain", domainName(domainID), name, domainID) : "<label>Domain<select name=\"" + esc(name) + "\" required>" + domainOptions(tenantID, domainID) + "</select></label>";
    }
    function typeField(mailboxType) {
      return hiddenInput("mailbox_type", mailboxType) + "<label>Type<span class=\"context-field\">" + esc(mailboxType === "shared" ? "Shared mailbox" : "User mailbox") + "</span></label>";
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
        tenantField(tenantID) + domainField(tenantID, domainID) + typeField(mailboxType) + "<label>Quota<input name=\"quota_gb\" type=\"number\" min=\"1\" step=\"1\" value=\"10\"></label>" +
        "<label>Local part<input name=\"local_part\" placeholder=\"marko\" required></label><label>Display name<input name=\"display_name\" placeholder=\"Marko Admin\"></label>" +
        (mailboxType === "shared" ? "" : "<label class=\"full\">Password<input name=\"password\" type=\"password\" autocomplete=\"new-password\" required></label>") +
        "<button class=\"button primary full\" type=\"submit\"><span class=\"material-symbols-outlined\">person_add</span>Create " + esc(mailboxType === "shared" ? "shared mailbox" : "user") + "</button></form></div>";
    }
    function renderTenants() {
      const actions = setupIncomplete() ? "<button class=\"button\" data-view=\"onboarding\"><span class=\"material-symbols-outlined\">add</span>Guided create</button>" : "";
      return "<div class=\"grid two-col\"><section class=\"card\"><div class=\"panel-head\"><div><h4>Tenants</h4><p>Each tenant is an isolated organization boundary.</p></div>" + actions + "</div>" +
        table(["Tenant", "Slug", "Status", "Created", "Actions"], visible(state.tenants).map(item => "<tr><td><div class=\"identity\"><span class=\"initials\">" + esc(initials(item.name)) + "</span><div><strong>" + esc(item.name) + "</strong><div class=\"muted small\">ID " + esc(item.id) + "</div></div></div></td><td><code>" + esc(item.slug) + "</code></td><td>" + badge(item.status) + "</td><td>" + esc(dateText(item.created_at)) + "</td><td><button class=\"button\" data-select-tenant=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">check_circle</span>Select</button></td></tr>"), "No tenants match this view.") +
        "</section><section class=\"card\"><div class=\"panel-head\"><div><h4>Create tenant</h4><p>First step for every customer, site, or organization.</p></div></div><div class=\"card-body\">" + tenantStep() + "</div></section></div>";
    }
    function renderDomains() {
      const tenantID = state.selectedTenantId;
      const domainID = state.selectedDomainId;
      const domain = selectedDomain();
      const domains = visible(domainsForTenant(tenantID));
      const aliasRows = state.aliases.filter(item => tenantID && domainID && String(item.tenant_id) === String(tenantID) && String(item.domain_id) === String(domainID));
      const catchRows = state.catchAll.filter(item => tenantID && domainID && String(item.tenant_id) === String(tenantID) && String(item.domain_id) === String(domainID));
      const dnsRecords = state.dns && String(state.dns.domain_id) === String(domainID) ? state.dns.records : [];
      const dnsTitle = domain ? "DNS records for " + domain.name : "DNS records";
      const tabBar = tabs("domain", [["domains", "Domains & DNS", "dns"], ["aliases", "Aliases", "alternate_email"], ["catchall", "Catch-all", "all_inbox"]], state.domainTab);
      if (!tenantID) {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Choose tenant first</h4><p>Create or select a tenant before adding domains.</p></div><button class=\"button primary\" data-view=\"onboarding\"><span class=\"material-symbols-outlined\">rocket_launch</span>Start setup</button></div></section>";
      }
      if (state.domainTab === "aliases") {
        const body = domainID ? "<div class=\"card-body\"><form class=\"form-grid\" data-form=\"alias\">" + contextField("Tenant", tenantName(tenantID), "tenant_id", tenantID) + contextField("Alias domain", domainName(domainID), "domain_id", domainID) + "<label>Alias local part<input name=\"source_local_part\" placeholder=\"sales\" required></label><label>Destination<select name=\"destination\" required>" + userEmailOptions(tenantID) + "</select></label><button class=\"button primary full\" type=\"submit\"><span class=\"material-symbols-outlined\">alternate_email</span>Create alias</button></form></div>" : "<div class=\"card-body\"><div class=\"scope-empty\">Add a domain in this tenant before creating aliases.</div></div>";
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Domain aliases</h4><p>Only addresses for the selected domain are shown here; destinations can be any mailbox in the tenant.</p></div>" + tenantDomainActions(true) + "</div>" + tabBar + body +
          table(["Alias", "Destination", "Created"], visible(aliasRows).map(item => "<tr><td><strong>" + esc(item.source_local_part + "@" + domainName(item.domain_id)) + "</strong></td><td>" + esc(item.destination) + "</td><td>" + esc(dateText(item.created_at)) + "</td></tr>"), "No aliases for this domain.") + "</section>";
      }
      if (state.domainTab === "catchall") {
        const body = domainID ? "<div class=\"card-body\"><form class=\"form-grid\" data-form=\"catch-all\">" + contextField("Tenant", tenantName(tenantID), "tenant_id", tenantID) + contextField("Domain", domainName(domainID), "domain_id", domainID) + "<label class=\"full\">Catch-all mailbox<select name=\"destination\" required>" + userEmailOptions(tenantID) + "</select></label><button class=\"button primary full\" type=\"submit\"><span class=\"material-symbols-outlined\">all_inbox</span>Set catch-all mailbox</button></form></div>" : "<div class=\"card-body\"><div class=\"scope-empty\">Add a domain in this tenant before setting catch-all routing.</div></div>";
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Catch-all mailbox</h4><p>Unknown recipients for the selected domain route to one mailbox after users and aliases are checked.</p></div>" + tenantDomainActions(true) + "</div>" + tabBar + body +
          table(["Domain", "Catch-all mailbox", "Status"], visible(catchRows).map(item => "<tr><td>" + esc(domainName(item.domain_id)) + "</td><td>" + esc(item.destination) + "</td><td>" + badge(item.status) + "</td></tr>"), "No catch-all mailbox configured for this domain.") + "</section>";
      }
      return "<div class=\"grid two-col\"><section class=\"card\"><div class=\"panel-head\"><div><h4>Domains in " + esc(tenantName(tenantID)) + "</h4><p>Select one domain, then use the tabs for aliases and catch-all.</p></div>" + tenantDomainActions(false) + "</div>" + tabBar +
        table(["Domain", "Status", "DKIM", "Actions"], domains.map(item => "<tr><td><strong>" + esc(item.name) + "</strong><div class=\"muted small\">ID " + esc(item.id) + "</div></td><td>" + badge(String(item.id) === String(domainID) ? "current" : item.status) + "</td><td><code>" + esc(item.dkim_selector || "mail") + "</code></td><td><div class=\"actions\"><button class=\"button\" data-select-domain=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">check_circle</span>Open</button><button class=\"button\" data-load-dns=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">dns</span>DNS</button></div></td></tr>"), "No domains in this tenant yet.") +
        "</section><section class=\"card\"><div class=\"panel-head\"><div><h4>Add domain</h4><p>New domains are added to the selected tenant.</p></div></div><div class=\"card-body\">" + domainStep(tenantID) + "</div><div class=\"panel-head\"><div><h4>" + esc(dnsTitle) + "</h4><p>Publish these records with your DNS provider.</p></div></div><div class=\"card-body record-grid\">" + renderDNSRecords(dnsRecords) + "</div></section></div>";
    }
    function renderUsers() {
      const tenantID = state.selectedTenantId;
      const domainID = state.selectedDomainId;
      const scopedIDs = new Set(scopedUsers().map(item => String(item.id)));
      const permissionRows = state.sharedPermissions.filter(item => tenantID && String(item.tenant_id) === String(tenantID) && (scopedIDs.has(String(item.shared_mailbox_id)) || scopedIDs.has(String(item.user_id))));
      const tabBar = tabs("user", [["people", "People", "person"], ["shared", "Shared mailboxes", "groups"], ["permissions", "Permissions", "key"]], state.userTab);
      const userRows = item => "<tr><td><div class=\"identity\"><span class=\"initials\">" + esc(initials(item.display_name || item.local_part)) + "</span><div><strong>" + emailFor(item) + "</strong><div class=\"muted small\">" + esc(item.display_name || "-") + "</div></div></div></td><td>" + esc(domainName(item.primary_domain_id)) + "</td><td>" + badge(item.status) + "</td><td>" + esc(formatBytes(item.quota_bytes)) + "</td><td>" + esc(dateText(item.created_at)) + "</td></tr>";
      if (!tenantID || !domainID) {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Choose a domain first</h4><p>Users, shared mailboxes, and permissions are managed inside one selected domain.</p></div><button class=\"button\" data-view=\"domains\"><span class=\"material-symbols-outlined\">dns</span>Open domains</button></div>" + tabBar + "<div class=\"card-body\"><div class=\"scope-empty\">Select or create a domain in the global bar before managing users.</div></div></section>";
      }
      if (state.userTab === "shared") {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Shared mailboxes for " + esc(domainName(domainID)) + "</h4><p>Group inboxes are created here; access is granted on the Permissions tab.</p></div>" + tenantDomainActions(true) + "</div>" + tabBar + "<div class=\"card-body\">" + userStep(tenantID, domainID, "shared") + "</div>" +
          table(["Shared mailbox", "Domain", "Status", "Quota", "Created"], filteredSharedMailboxes().map(userRows), "No shared mailboxes in this domain.") + "</section>";
      }
      if (state.userTab === "permissions") {
        return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Shared mailbox permissions</h4><p>Grant read, send as, send on behalf, and manage rights inside the selected domain.</p></div>" + tenantDomainActions(true) + "</div>" + tabBar + "<div class=\"card-body\"><form class=\"form-grid\" data-form=\"shared-permission\">" + contextField("Tenant", tenantName(tenantID), "tenant_id", tenantID) + "<label>Shared mailbox<select name=\"shared_mailbox_id\" required>" + sharedMailboxOptions(tenantID, domainID) + "</select></label><label>User<select name=\"user_id\" required>" + normalUserOptions(tenantID, domainID) + "</select></label><label><span><input name=\"can_read\" type=\"checkbox\" checked> Read</span></label><label><span><input name=\"can_send_as\" type=\"checkbox\"> Send as</span></label><label><span><input name=\"can_send_on_behalf\" type=\"checkbox\"> Send on behalf</span></label><label><span><input name=\"can_manage\" type=\"checkbox\"> Manage</span></label><button class=\"button primary full\" type=\"submit\"><span class=\"material-symbols-outlined\">key</span>Grant permission</button></form></div>" +
          table(["Shared", "User", "Rights"], visible(permissionRows).map(item => "<tr><td>" + esc(userLabel(item.shared_mailbox_id)) + "</td><td>" + esc(userLabel(item.user_id)) + "</td><td>" + esc(permissionText(item)) + "</td></tr>"), "No shared permissions in this domain.") + "</section>";
      }
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>People for " + esc(domainName(domainID)) + "</h4><p>Normal user mailboxes with login, quota, webmail, IMAP, SMTP, CalDAV, and CardDAV access.</p></div>" + tenantDomainActions(true) + "</div>" + tabBar + "<div class=\"card-body\">" + userStep(tenantID, domainID, "user") + "</div>" +
        table(["User", "Domain", "Status", "Quota", "Created"], filteredNormalUsers().map(userRows), "No users in this domain.") + "</section>";
    }
    function renderSecurity() {
      const rows = state.policies.filter(item => !state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId));
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Security policy</h4><p>Spam action, malware action, and TLS requirements for the selected tenant.</p></div><button class=\"button\" data-refresh><span class=\"material-symbols-outlined\">refresh</span>Refresh</button></div>" +
        table(["Tenant", "Spam", "Malware", "Require TLS auth", "Actions"], visible(rows).map(item => "<tr><td><strong>" + esc(tenantName(item.tenant_id)) + "</strong><div class=\"muted small\">Tenant " + esc(item.tenant_id) + "</div></td><td><select data-policy-field=\"spam_action\" data-policy=\"" + esc(item.tenant_id) + "\"><option " + selected(item.spam_action, "mark") + ">mark</option><option " + selected(item.spam_action, "quarantine") + ">quarantine</option><option " + selected(item.spam_action, "reject") + ">reject</option></select></td><td><select data-policy-field=\"malware_action\" data-policy=\"" + esc(item.tenant_id) + "\"><option " + selected(item.malware_action, "quarantine") + ">quarantine</option><option " + selected(item.malware_action, "reject") + ">reject</option></select></td><td><input type=\"checkbox\" data-policy-field=\"require_tls_for_auth\" data-policy=\"" + esc(item.tenant_id) + "\" " + checked(item.require_tls_for_auth) + "></td><td><button class=\"button primary\" data-save-policy=\"" + esc(item.tenant_id) + "\"><span class=\"material-symbols-outlined\">save</span>Save</button></td></tr>"), "No policy found for this tenant.") + "</section>";
    }
    function renderQuarantine() {
      const rows = state.quarantine.filter(item => !state.selectedTenantId || String(item.tenant_id) === String(state.selectedTenantId));
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Quarantine</h4><p>Release false positives or delete messages that should not be delivered.</p></div><button class=\"button\" data-refresh><span class=\"material-symbols-outlined\">refresh</span>Refresh</button></div>" +
        table(["Verdict", "Recipient", "Sender", "Scanner", "Status", "Date", "Actions"], visible(rows).map(item => "<tr><td>" + badge(item.verdict) + "</td><td><strong>" + esc(item.recipient) + "</strong><div class=\"muted small\">Tenant " + esc(item.tenant_id) + "</div></td><td class=\"muted\">" + esc(item.sender || "-") + "</td><td>" + esc(item.scanner || "-") + "</td><td>" + badge(item.status || "held") + "</td><td>" + esc(dateText(item.created_at)) + "</td><td>" + quarantineActions(item) + "</td></tr>"), "No quarantine events for this tenant.") + "</section>";
    }
    function renderAudit() {
      const rows = state.audit.filter(item => !state.selectedTenantId || !item.tenant_id || String(item.tenant_id) === String(state.selectedTenantId));
      return "<section class=\"card\"><div class=\"panel-head\"><div><h4>Audit log</h4><p>Security-significant admin and message events.</p></div><button class=\"button\" data-refresh><span class=\"material-symbols-outlined\">refresh</span>Refresh</button></div>" +
        table(["Action", "Actor", "Target", "Tenant", "Metadata", "Date"], visible(rows).map(auditFullRow), "No audit events for this tenant.") + "</section>";
    }
    function renderSystem() {
      return "<div class=\"grid two-col\"><section class=\"card\"><div class=\"panel-head\"><div><h4>Service health</h4><p>Live browser checks against the admin service.</p></div><button class=\"button\" data-check-health><span class=\"material-symbols-outlined\">monitor_heart</span>Check health</button></div><div class=\"card-body grid three-col\">" +
        stat("Admin API", state.health, "api") + stat("Session", state.csrf ? "active" : "login", "cookie") + stat("Data reload", "ready", "sync") +
        "</div></section><section class=\"card\"><div class=\"panel-head\"><div><h4>Client endpoints</h4><p>Use these for autodiscovery, CalDAV, and CardDAV integration checks.</p></div></div><div class=\"card-body step-list\">" +
        endpoint("/.well-known/proidentity-mail/config.json?emailaddress=user@example.com") + endpoint("/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=user@example.com") + endpoint("/.well-known/caldav") + endpoint("/.well-known/carddav") +
        "</div></section></div>";
    }
    function table(headings, rows, emptyText) {
      return "<div class=\"table-wrap\"><table><thead><tr>" + headings.map(h => "<th>" + esc(h) + "</th>").join("") + "</tr></thead><tbody>" + rows.join("") + emptyRows(rows, emptyText) + "</tbody></table></div>";
    }
    function emptyRows(rows, text) {
      return rows.length ? "" : "<tr><td class=\"muted\" colspan=\"9\">" + esc(text) + "</td></tr>";
    }
    function tenantOptions(current) {
      const rows = state.tenants.map(item => "<option value=\"" + esc(item.id) + "\" " + selected(item.id, current) + ">" + esc(item.name) + "</option>");
      return rows.length ? rows.join("") : "<option value=\"\">Create a tenant first</option>";
    }
    function domainOptions(tenantID, current) {
      const domains = state.domains.filter(item => !tenantID || String(item.tenant_id) === String(tenantID));
      const rows = domains.map(item => "<option value=\"" + esc(item.id) + "\" " + selected(item.id, current) + ">" + esc(item.name) + "</option>");
      return rows.length ? rows.join("") : "<option value=\"\">Create a domain first</option>";
    }
    function userEmailOptions(tenantID) {
      const rows = state.users.filter(item => tenantID && String(item.tenant_id) === String(tenantID)).map(item => {
        const email = emailAddress(item);
        return "<option value=\"" + esc(email) + "\">" + esc(email) + " (" + esc(item.mailbox_type || "user") + ")</option>";
      });
      return rows.length ? rows.join("") : "<option value=\"\">Create a user first</option>";
    }
    function normalUserOptions(tenantID, domainID = "") {
      const rows = normalUsers().filter(item => tenantID && String(item.tenant_id) === String(tenantID) && (!domainID || String(item.primary_domain_id) === String(domainID))).map(item => "<option value=\"" + esc(item.id) + "\">" + emailFor(item) + "</option>");
      return rows.length ? rows.join("") : "<option value=\"\">Create a user first</option>";
    }
    function sharedMailboxOptions(tenantID, domainID = "") {
      const rows = sharedMailboxes().filter(item => tenantID && String(item.tenant_id) === String(tenantID) && (!domainID || String(item.primary_domain_id) === String(domainID))).map(item => "<option value=\"" + esc(item.id) + "\">" + emailFor(item) + "</option>");
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
      return records.map(record => "<div class=\"dns-record\"><strong>" + esc(record.type) + "</strong><code>" + esc(record.name) + "</code><code>" + esc(record.priority ? record.priority + " " : "") + esc(record.value) + "</code></div>").join("");
    }
    function quarantineActions(item) {
      if (item.status && item.status !== "held") return "<span class=\"muted small\">" + esc(item.resolution_note || "resolved") + "</span>";
      return "<div class=\"actions\"><button class=\"button\" data-quarantine-action=\"release\" data-quarantine-id=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">outbox</span>Release</button><button class=\"button danger\" data-quarantine-action=\"delete\" data-quarantine-id=\"" + esc(item.id) + "\"><span class=\"material-symbols-outlined\">delete</span>Delete</button></div>";
    }
    function auditRow(item) {
      return "<tr><td><strong>" + esc(item.action) + "</strong><div class=\"muted small\">" + esc(dateText(item.created_at)) + "</div></td></tr>";
    }
    function auditFullRow(item) {
      return "<tr><td><strong>" + esc(item.action) + "</strong></td><td>" + esc(item.actor_type) + "</td><td>" + esc(item.target_type) + " <code>" + esc(item.target_id) + "</code></td><td>" + esc(item.tenant_id || "-") + "</td><td><code>" + esc(item.metadata_json || "{}") + "</code></td><td>" + esc(dateText(item.created_at)) + "</td></tr>";
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
    async function submitForm(form) {
      const type = form.dataset.form;
      const data = Object.fromEntries(new FormData(form).entries());
      ["tenant_id", "primary_domain_id", "domain_id", "shared_mailbox_id", "user_id"].forEach(key => { if (data[key]) data[key] = Number(data[key]); });
      if (data.quota_gb) {
        data.quota_bytes = Number(data.quota_gb) * 1073741824;
        delete data.quota_gb;
      }
      ["can_read", "can_send_as", "can_send_on_behalf", "can_manage"].forEach(key => data[key] = data[key] === "on");
      const path = type === "tenant" ? "/api/v1/tenants" : type === "domain" ? "/api/v1/domains" : type === "alias" ? "/api/v1/aliases" : type === "catch-all" ? "/api/v1/catch-all" : type === "shared-permission" ? "/api/v1/shared-permissions" : "/api/v1/users";
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
      form.reset();
      await refresh();
      showStatus((type === "user" ? "User" : type === "catch-all" ? "Catch-all" : type === "shared-permission" ? "Shared permission" : type[0].toUpperCase() + type.slice(1)) + " created");
    }
    async function loadDNS(id) {
      const selectedID = id === "selected" ? ($("#dns-domain")?.value || state.selectedDomainId) : id;
      if (!selectedID) throw new Error("Select a domain first");
      state.dns = await api("/api/v1/domains/" + selectedID + "/dns");
      state.selectedDomainId = String(selectedID);
      setView(state.view === "onboarding" ? "onboarding" : "domains");
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
    async function resolveQuarantine(id, action) {
      const note = prompt("Resolution note", action === "release" ? "false positive" : "malware/spam removed") || "";
      await api("/api/v1/quarantine/" + id + "/" + action, {method: "POST", body: JSON.stringify({resolution_note: note})});
      await refresh();
      showStatus("Quarantine event " + (action === "delete" ? "deleted" : "released"));
    }
    async function logout() {
      await api("/api/v1/session", {method: "DELETE"});
      state.csrf = "";
      $("#login-cover").classList.remove("hidden");
      showStatus("Logged out");
    }
    document.addEventListener("click", event => {
      const tab = event.target.closest("[data-tab]");
      if (tab) {
        if (tab.dataset.tabScope === "domain") state.domainTab = tab.dataset.tab;
        if (tab.dataset.tabScope === "user") state.userTab = tab.dataset.tab;
        render();
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
      const dns = event.target.closest("[data-load-dns]")?.dataset.loadDns;
      if (dns) loadDNS(dns).catch(error => showStatus(error.message, true));
      const save = event.target.closest("[data-save-policy]")?.dataset.savePolicy;
      if (save) savePolicy(save).catch(error => showStatus(error.message, true));
      const qButton = event.target.closest("[data-quarantine-action]");
      if (qButton) resolveQuarantine(qButton.dataset.quarantineId, qButton.dataset.quarantineAction).catch(error => showStatus(error.message, true));
      const selectTenant = event.target.closest("[data-select-tenant]")?.dataset.selectTenant;
      if (selectTenant) {
        state.selectedTenantId = String(selectTenant);
        const nextDomain = domainsForTenant()[0];
        state.selectedDomainId = nextDomain ? String(nextDomain.id) : "";
        state.dns = null;
        setView("domains");
      }
      const selectDomain = event.target.closest("[data-select-domain]")?.dataset.selectDomain;
      if (selectDomain) { state.selectedDomainId = String(selectDomain); loadDNS(selectDomain).catch(error => showStatus(error.message, true)); }
      const copy = event.target.closest("[data-copy]")?.dataset.copy;
      if (copy) navigator.clipboard.writeText(copy).then(() => showStatus("Copied")).catch(error => showStatus("Copy failed: " + error.message, true));
    });
    document.addEventListener("submit", event => {
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
    });
    $("#search").addEventListener("input", event => { state.query = event.target.value; render(); });
    $("#reload-nav").addEventListener("click", () => refresh().then(() => showStatus("Data refreshed")).catch(error => showStatus(error.message, true)));
    $("#reload-top").addEventListener("click", () => refresh().then(() => showStatus("Data refreshed")).catch(error => showStatus(error.message, true)));
    $("#logout-nav").addEventListener("click", () => logout().catch(error => showStatus(error.message, true)));
    $("#start-onboarding").addEventListener("click", () => setView("onboarding"));
    $("#copy-discovery").addEventListener("click", () => navigator.clipboard.writeText(location.origin + "/.well-known/proidentity-mail/config.json?emailaddress=user@example.com").then(() => showStatus("Discovery URL copied")).catch(error => showStatus("Copy failed: " + error.message, true)));
    $("#login-form").addEventListener("submit", async event => {
      event.preventDefault();
      const form = event.currentTarget;
      const data = Object.fromEntries(new FormData(form).entries());
      try {
        const response = await fetch("/api/v1/session", {method: "POST", credentials: "same-origin", cache: "no-store", headers: {"Content-Type": "application/json"}, body: JSON.stringify(data)});
        const body = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(body.error || "Login failed");
        state.csrf = body.csrf_token || "";
        $("#login-cover").classList.add("hidden");
        form.reset();
        await refresh();
        showStatus("Logged in");
      } catch (error) {
        showStatus(error.message, true);
      }
    });
    render();
    bootstrapSession().catch(error => showStatus(error.message, true));
  </script>
</body>
</html>`
