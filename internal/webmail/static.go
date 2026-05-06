package webmail

const webmailIndexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>ProIdentity Webmail</title>
  <style>
    body { margin: 0; background: #f6f7f9; color: #18202b; font: 14px/1.45 system-ui, sans-serif; }
    header { padding: 18px 24px; background: #111827; color: white; }
    main { max-width: 1100px; margin: 0 auto; padding: 18px; display: grid; gap: 14px; }
    form, section { background: white; border: 1px solid #dce2ea; border-radius: 8px; padding: 16px; }
    label { display: grid; gap: 6px; color: #647184; font-weight: 650; }
    input, button { min-height: 38px; border-radius: 6px; font: inherit; }
    input { border: 1px solid #dce2ea; padding: 8px 10px; }
    button { border: 0; background: #0f766e; color: white; padding: 8px 12px; font-weight: 700; }
    table { width: 100%; border-collapse: collapse; }
    th, td { padding: 10px 12px; border-bottom: 1px solid #dce2ea; text-align: left; vertical-align: top; }
    th { color: #647184; font-size: 11px; text-transform: uppercase; }
  </style>
</head>
<body>
  <header><h1>ProIdentity Webmail</h1></header>
  <main>
    <form id="login">
      <label>Email<input name="email" autocomplete="username" required></label>
      <label>Password<input name="password" type="password" autocomplete="current-password" required></label>
      <button type="submit">Load Mailbox</button>
    </form>
    <section>
      <table><thead><tr><th>From</th><th>Subject</th><th>Preview</th></tr></thead><tbody id="messages"></tbody></table>
    </section>
  </main>
  <script>
    document.querySelector("#login").addEventListener("submit", async event => {
      event.preventDefault();
      const data = new FormData(event.currentTarget);
      const token = btoa(data.get("email") + ":" + data.get("password"));
      const response = await fetch("/api/v1/messages?limit=100", {headers: {Authorization: "Basic " + token}});
      const messages = response.ok ? await response.json() : [];
      document.querySelector("#messages").innerHTML = messages.map(item => "<tr><td>" + item.from + "</td><td>" + item.subject + "</td><td>" + item.preview + "</td></tr>").join("");
    });
  </script>
</body>
</html>
`
