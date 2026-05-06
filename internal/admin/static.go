package admin

const adminIndexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>ProIdentity Mail Admin</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f6f7f9;
      --panel: #ffffff;
      --ink: #18202b;
      --muted: #647184;
      --line: #dce2ea;
      --accent: #0f766e;
      --accent-dark: #115e59;
      --danger: #b42318;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      background: var(--bg);
      color: var(--ink);
      font: 14px/1.45 system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      padding: 18px 24px;
      background: #111827;
      color: white;
      border-bottom: 1px solid #0b1020;
    }
    h1, h2 { margin: 0; letter-spacing: 0; }
    h1 { font-size: 20px; font-weight: 700; }
    h2 { font-size: 15px; font-weight: 700; }
    main {
      display: grid;
      grid-template-columns: minmax(280px, 380px) minmax(0, 1fr);
      gap: 18px;
      padding: 18px;
      max-width: 1440px;
      margin: 0 auto;
    }
    section {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      overflow: hidden;
    }
    .section-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      padding: 14px 16px;
      border-bottom: 1px solid var(--line);
    }
    .stack { display: grid; gap: 12px; padding: 16px; }
    label { display: grid; gap: 6px; color: var(--muted); font-size: 12px; font-weight: 650; }
    input, select {
      width: 100%;
      min-height: 38px;
      border: 1px solid var(--line);
      border-radius: 6px;
      padding: 8px 10px;
      background: white;
      color: var(--ink);
      font: inherit;
    }
    button {
      min-height: 38px;
      border: 0;
      border-radius: 6px;
      padding: 8px 12px;
      background: var(--accent);
      color: white;
      font: inherit;
      font-weight: 700;
      cursor: pointer;
    }
    button:hover { background: var(--accent-dark); }
    button.secondary {
      border: 1px solid var(--line);
      background: white;
      color: var(--ink);
    }
    button.secondary:hover { background: #eef2f6; }
    .grid {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 18px;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      font-size: 13px;
    }
    th, td {
      padding: 10px 12px;
      text-align: left;
      border-bottom: 1px solid var(--line);
      vertical-align: top;
    }
    th { color: var(--muted); font-size: 11px; text-transform: uppercase; }
    code {
      display: inline-block;
      max-width: 100%;
      overflow-wrap: anywhere;
      font: 12px/1.4 ui-monospace, SFMono-Regular, Consolas, monospace;
    }
    .status {
      color: var(--muted);
      font-size: 13px;
      min-height: 20px;
    }
    .error { color: var(--danger); }
    @media (max-width: 980px) {
      main { grid-template-columns: 1fr; }
      .grid { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <header>
    <h1>ProIdentity Mail Admin</h1>
    <button class="secondary" id="refresh" type="button">Refresh</button>
  </header>
  <main>
    <section>
      <div class="section-head"><h2>Create</h2><span class="status" id="status"></span></div>
      <div class="stack">
        <form class="stack" id="tenant-form">
          <h2>Tenant</h2>
          <label>Name<input name="name" autocomplete="organization" required></label>
          <label>Slug<input name="slug" autocomplete="off" required></label>
          <button type="submit">Create Tenant</button>
        </form>
        <form class="stack" id="domain-form">
          <h2>Domain</h2>
          <label>Tenant ID<input name="tenant_id" inputmode="numeric" required></label>
          <label>Domain<input name="name" autocomplete="off" required></label>
          <button type="submit">Create Domain</button>
        </form>
        <form class="stack" id="user-form">
          <h2>User</h2>
          <label>Tenant ID<input name="tenant_id" inputmode="numeric" required></label>
          <label>Domain ID<input name="primary_domain_id" inputmode="numeric" required></label>
          <label>Local Part<input name="local_part" autocomplete="username" required></label>
          <label>Display Name<input name="display_name" autocomplete="name"></label>
          <label>Password<input name="password" type="password" autocomplete="new-password" required></label>
          <button type="submit">Create User</button>
        </form>
      </div>
    </section>
    <div class="grid">
      <section>
        <div class="section-head"><h2>Tenants</h2></div>
        <table><thead><tr><th>ID</th><th>Name</th><th>Status</th></tr></thead><tbody id="tenants"></tbody></table>
      </section>
      <section>
        <div class="section-head"><h2>Domains</h2></div>
        <table><thead><tr><th>ID</th><th>Domain</th><th>DNS</th></tr></thead><tbody id="domains"></tbody></table>
      </section>
      <section>
        <div class="section-head"><h2>Users</h2></div>
        <table><thead><tr><th>ID</th><th>User</th><th>Status</th></tr></thead><tbody id="users"></tbody></table>
      </section>
      <section style="grid-column: 1 / -1;">
        <div class="section-head"><h2>DNS Records</h2></div>
        <table><thead><tr><th>Type</th><th>Name</th><th>Value</th></tr></thead><tbody id="dns"></tbody></table>
      </section>
    </div>
  </main>
  <script>
    const statusEl = document.querySelector("#status");
    const showStatus = (text, error = false) => {
      statusEl.textContent = text;
      statusEl.className = error ? "status error" : "status";
    };
    const api = async (path, options = {}) => {
      const response = await fetch(path, {
        headers: {"Content-Type": "application/json"},
        ...options
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || response.statusText);
      return data;
    };
    const rows = (target, items, render) => {
      document.querySelector(target).innerHTML = items.map(render).join("");
    };
    async function refresh() {
      const [tenants, domains, users] = await Promise.all([
        api("/api/v1/tenants"),
        api("/api/v1/domains"),
        api("/api/v1/users")
      ]);
      rows("#tenants", tenants, item => "<tr><td>" + item.id + "</td><td>" + item.name + "<br><code>" + item.slug + "</code></td><td>" + item.status + "</td></tr>");
      rows("#domains", domains, item => "<tr><td>" + item.id + "</td><td>" + item.name + "<br><code>tenant " + item.tenant_id + "</code></td><td><button class=\"secondary\" type=\"button\" data-dns=\"" + item.id + "\">DNS</button></td></tr>");
      rows("#users", users, item => "<tr><td>" + item.id + "</td><td>" + item.local_part + "<br><code>domain " + item.primary_domain_id + "</code></td><td>" + item.status + "</td></tr>");
      showStatus("Loaded");
    }
    async function submitForm(event, path, numericFields = []) {
      event.preventDefault();
      const form = event.currentTarget;
      const body = Object.fromEntries(new FormData(form).entries());
      for (const field of numericFields) body[field] = Number(body[field]);
      try {
        await api(path, {method: "POST", body: JSON.stringify(body)});
        form.reset();
        await refresh();
        showStatus("Saved");
      } catch (error) {
        showStatus(error.message, true);
      }
    }
    document.querySelector("#tenant-form").addEventListener("submit", event => submitForm(event, "/api/v1/tenants"));
    document.querySelector("#domain-form").addEventListener("submit", event => submitForm(event, "/api/v1/domains", ["tenant_id"]));
    document.querySelector("#user-form").addEventListener("submit", event => submitForm(event, "/api/v1/users", ["tenant_id", "primary_domain_id"]));
    document.querySelector("#refresh").addEventListener("click", refresh);
    document.addEventListener("click", async event => {
      const id = event.target.dataset && event.target.dataset.dns;
      if (!id) return;
      try {
        const dns = await api("/api/v1/domains/" + id + "/dns");
        rows("#dns", dns.records, record => "<tr><td>" + record.type + "</td><td><code>" + record.name + "</code></td><td><code>" + (record.priority ? record.priority + " " : "") + record.value + "</code></td></tr>");
        showStatus("DNS for " + dns.domain);
      } catch (error) {
        showStatus(error.message, true);
      }
    });
    refresh().catch(error => showStatus(error.message, true));
  </script>
</body>
</html>
`
