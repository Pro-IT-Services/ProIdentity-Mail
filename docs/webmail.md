# Webmail

## Purpose

The webmail service is the first browser-facing mail client for ProIdentity Mail. It authenticates mailbox users against MariaDB, reads delivered mail from Dovecot Maildir storage, and exposes a simple UI plus JSON API for recent messages.

Current live service:

- Default bind: `0.0.0.0:8082`
- DevMail URL: `http://192.168.254.125:8082/`
- Go command: `cmd/webmail`
- Main package: `internal/webmail`
- Mailbox storage root: `/var/vmail`

## Current UI Functions

### Login Form

Fields:

- Email
- Password

Behavior:

- The browser creates an HTTP Basic Auth header from the entered email and password.
- The UI calls `GET /api/v1/messages?limit=100`.
- If authentication succeeds, the recent message table is filled.
- If authentication fails, the current basic UI shows an empty table.

Current limitation:

- Credentials are held only in browser memory for the request.
- There is no session cookie, logout button, or persistent login yet.

### Recent Messages Table

Columns:

- From
- Subject
- Preview

Behavior:

- Displays recent messages returned by the webmail API.
- Shows message summary only, not full message body.
- Does not currently support opening a message.

## Current API Endpoints

### Health

`GET /healthz`

Returns:

```json
{"status":"ok"}
```

Used by service checks.

### Recent Messages

`GET /api/v1/messages?limit=100`

Authentication:

- Requires HTTP Basic Auth.
- Username is full email address.
- Password is the mailbox password.
- Password is verified against the MariaDB `users.password_hash` value.

Response:

```json
[
  {
    "id": "1778094364.M776907P12896.DevMail,S=786,W=806",
    "from": "sender@example.net",
    "to": "marko@example.com",
    "subject": "Webmail probe",
    "date": "2026-05-06T19:06:04Z",
    "preview": "hello webmail",
    "mailbox": "new",
    "size_bytes": 786
  }
]
```

Query parameters:

- `limit`: optional message count.
- Minimum/default behavior: if missing or invalid, uses `100`.
- Maximum: values over `300` are capped by the store behavior.

Error behavior:

- Missing credentials: `401 Unauthorized`.
- Invalid credentials: `401 Unauthorized`.
- Maildir or parsing failure for the mailbox listing: `500`.

## Current Maildir Functions

### Mailbox Path Resolution

For an authenticated email address:

```text
local_part@example.com
```

The service reads:

```text
/var/vmail/example.com/local_part/Maildir
```

It scans:

- `new`
- `cur`

### Message Ordering

Behavior:

- Reads regular files only.
- Sorts by filesystem modification time, newest first.
- Returns up to the requested limit.

### Message Parsing

For each message file:

- Parses headers with Go `net/mail`.
- Extracts `From`.
- Extracts `To`.
- Extracts `Subject`.
- Extracts `Date`.
- Uses file basename as message ID.
- Uses the first non-empty body line as preview.
- Truncates preview to 160 characters.
- Records mailbox source as `new` or `cur`.
- Records message size in bytes.

Current behavior when a message cannot be parsed:

- The service skips that file and continues listing other messages.

## Current Authentication

The webmail service uses a composite store:

- `SQLAuthStore` for password verification.
- `MaildirStore` for mailbox message listing.

SQL auth query:

- Joins `users` to `domains`.
- Matches `local_part@domain`.
- Requires user status `active`.
- Allows domain status `pending` or `active`.
- Verifies password with bcrypt.

Runtime service user:

- The systemd service runs as `vmail:vmail`.
- This is necessary because Dovecot creates mailbox directories with restrictive permissions owned by `vmail`.
- The service has read access to `/var/vmail`.

## Current Systemd Service

Service file:

```text
deploy/devmail/proidentity-webmail.service
```

Key settings:

- Runs `/opt/proidentity-mail/bin/webmail`.
- Uses `/etc/proidentity-mail/proidentity-mail.env`.
- Binds using `PROIDENTITY_WEBMAIL_ADDR`.
- Runs as `vmail`.
- Uses `ReadOnlyPaths=/var/vmail`.
- Restarts on failure.

## Current Configuration

Environment variable:

```dotenv
PROIDENTITY_WEBMAIL_ADDR=0.0.0.0:8082
```

Default in Go config:

```text
0.0.0.0:8082
```

## Current Test Probe

Probe script:

```text
deploy/devmail/webmail-probe.sh
```

Behavior:

- Creates a temporary random mailbox user through the admin API.
- Generates a random password on the server.
- Delivers a test message through SMTP.
- Calls the webmail API with Basic Auth.
- Verifies the delivered subject appears in the returned JSON.
- Does not print the generated password.

## Current User-Facing Limitations

The current webmail is a foundation, not a complete mail client yet.

Not implemented yet:

- Full message view.
- HTML message rendering.
- MIME multipart parsing.
- Attachments.
- Compose.
- Reply.
- Reply all.
- Forward.
- Send mail.
- Drafts.
- Sent folder.
- Trash.
- Move message.
- Delete message.
- Mark read/unread.
- Search.
- Pagination.
- Threading/conversations.
- Folder tree.
- User settings.
- Contacts integration.
- Calendar integration in the webmail UI.

## Planned Webmail Functions

### Authentication and Sessions

Planned:

- Login page with server-side session.
- Secure cookies.
- Logout.
- Session expiry.
- Optional two-factor authentication.
- Rate limiting.
- Login audit events.

### Mailbox Navigation

Planned:

- Folder list.
- Inbox.
- Sent.
- Drafts.
- Trash.
- Spam.
- Custom folders.
- Message counts per folder.
- Unread counts.

### Message Reading

Planned:

- Full message body endpoint.
- Plain text rendering.
- HTML rendering with sanitization.
- Attachment list.
- Attachment download.
- Inline image handling.
- Header details.
- DKIM/SPF/DMARC/security result display.

### Message Actions

Planned:

- Mark as read.
- Mark as unread.
- Delete.
- Move to folder.
- Restore from trash.
- Mark as spam.
- Mark as not spam.
- Download raw message.
- Print message.

### Compose and Send

Planned:

- Compose UI.
- To/Cc/Bcc fields.
- Subject/body editor.
- Attachments.
- Save draft.
- Send through authenticated submission.
- Store sent copy.
- Reply.
- Reply all.
- Forward.

### Search and Filtering

Planned:

- Search by sender.
- Search by recipient.
- Search by subject.
- Search by body text.
- Date range filters.
- Attachment filters.
- Spam/security filters.
- Saved searches.

### Spam and Malware UX

Planned:

- Spam folder view.
- Malware quarantine folder view.
- Release quarantined message.
- Delete quarantined message.
- Train as spam.
- Train as ham/not-spam.
- Sender allow/block controls.

### Calendar and Contacts Integration

Planned:

- Contacts picker while composing.
- Contact quick add from sender.
- Calendar invite detection.
- Accept/decline meeting invites.
- Show user calendar side panel.
- Link to CardDAV/CalDAV collections.

### Performance and Caching

Planned:

- Redis cache for the most recent 200 to 300 messages per mailbox.
- Cached message summaries.
- Lazy full-body loading.
- Pagination cursor.
- Background mailbox indexer.
- Search index.

### Security Hardening

Planned:

- HTML sanitizer for message bodies.
- Attachment content-type validation.
- Download headers to prevent script execution.
- CSRF protection for session-based actions.
- Per-user rate limits.
- Audit logging.
- Strict security headers.
- Reverse proxy TLS enforcement.

## Testing Notes

Useful current test account:

- Email: `marko@external-1778096561.local`
- Password: `secret123456`

Useful live checks:

```powershell
Invoke-WebRequest -UseBasicParsing http://192.168.254.125:8082/
```

For API testing, use Basic Auth with a mailbox account:

```powershell
$pair = "marko@external-1778096561.local:secret123456"
$token = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes($pair))
Invoke-RestMethod `
  -Uri "http://192.168.254.125:8082/api/v1/messages?limit=20" `
  -Headers @{ Authorization = "Basic $token" }
```

Safer automated live probe:

```powershell
ssh root@192.168.254.125 "bash /tmp/proidentity-devmail/webmail-probe.sh"
```
