# Admin Panel

## Purpose

The admin panel is the first control-plane UI for the ProIdentity Mail platform. It manages platform objects used by the mail services: tenants, domains, users, and DNS records. It also exposes mail client discovery endpoints used by desktop and mobile clients.

Current live service:

- Default bind: `0.0.0.0:8080`
- DevMail URL: `http://192.168.254.125:8080/`
- Go command: `cmd/webadmin`
- Main package: `internal/admin`

## Current UI Functions

### Dashboard Shell

The root page `/` serves the embedded HTML admin UI.

Current screen areas:

- Create forms panel
- Tenants table
- Domains table
- Users table
- DNS records table
- Refresh button
- Inline status/error text

### Create Tenant

Form fields:

- `Name`
- `Slug`

Behavior:

- Submits to `POST /api/v1/tenants`.
- Creates an active tenant.
- Creates a default tenant policy in the database.
- Refreshes tenant/domain/user tables after success.

### Create Domain

Form fields:

- `Tenant ID`
- `Domain`

Behavior:

- Submits to `POST /api/v1/domains`.
- Creates a pending domain.
- Uses default DKIM selector `mail`.
- Domain becomes usable by Postfix/Dovecot maps while status is `pending` or `active`.
- DNS records can be viewed from the domain table.

### Create User

Form fields:

- `Tenant ID`
- `Domain ID`
- `Local Part`
- `Display Name`
- `Password`

Behavior:

- Submits to `POST /api/v1/users`.
- Hashes the password using bcrypt before storage.
- Creates an active mailbox user.
- Does not return password hash in API response.
- Mailbox becomes usable for SMTP AUTH, IMAP, POP3, DAV auth, and webmail auth.

### List Tenants

Behavior:

- Loads from `GET /api/v1/tenants`.
- Displays tenant ID, name, slug, and status.
- Ordered newest first by the SQL store.

### List Domains

Behavior:

- Loads from `GET /api/v1/domains`.
- Displays domain ID, domain name, tenant ID, and DNS action.
- Ordered newest first by the SQL store.

### List Users

Behavior:

- Loads from `GET /api/v1/users`.
- Displays user ID, local part, primary domain ID, and status.
- Password hashes are removed from API responses.

### View DNS Records

Behavior:

- Domain table DNS button calls `GET /api/v1/domains/{domainID}/dns`.
- Displays generated DNS records for the selected domain.

Current generated record types:

- MX
- SPF TXT
- DMARC TXT
- MTA-STS TXT
- SMTP TLS reporting TXT
- DKIM TXT, when an active DKIM key exists

## Current API Endpoints

### Health

`GET /healthz`

Returns:

```json
{"status":"ok"}
```

Used by service checks and uptime probes.

### Tenants

`GET /api/v1/tenants`

Returns a JSON array of tenants.

`POST /api/v1/tenants`

Request:

```json
{
  "name": "Example Org",
  "slug": "example"
}
```

Response:

```json
{
  "id": 1,
  "name": "Example Org",
  "slug": "example",
  "status": "active"
}
```

Validation:

- `name` is required.
- `slug` is required.

### Domains

`GET /api/v1/domains`

Returns a JSON array of domains.

`POST /api/v1/domains`

Request:

```json
{
  "tenant_id": 1,
  "name": "example.com"
}
```

Response:

```json
{
  "id": 1,
  "tenant_id": 1,
  "name": "example.com",
  "status": "pending",
  "dkim_selector": "mail"
}
```

Validation:

- `tenant_id` is required.
- `name` is required.

### Domain DNS

`GET /api/v1/domains/{domainID}/dns`

Response:

```json
{
  "domain_id": 1,
  "domain": "example.com",
  "records": [
    {
      "type": "MX",
      "name": "example.com",
      "value": "mail.example.com",
      "priority": 10
    }
  ]
}
```

Validation:

- `domainID` must be a positive integer.

### Users

`GET /api/v1/users`

Returns a JSON array of users with password hashes removed.

`POST /api/v1/users`

Request:

```json
{
  "tenant_id": 1,
  "primary_domain_id": 1,
  "local_part": "marko",
  "display_name": "Marko",
  "password": "secret123456"
}
```

Response:

```json
{
  "id": 1,
  "tenant_id": 1,
  "primary_domain_id": 1,
  "local_part": "marko",
  "display_name": "Marko",
  "status": "active",
  "quota_bytes": 10737418240
}
```

Validation:

- `tenant_id` is required.
- `primary_domain_id` is required.
- `local_part` is required.
- `password` is required.

## Client Discovery Functions

### Mail Autoconfig

Endpoints:

- `GET /.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=user@example.com`
- `GET /mail/config-v1.1.xml?emailaddress=user@example.com`

Purpose:

- Provides Thunderbird-style email client autoconfiguration.
- Builds the mail host as `mail.<domain>`.

Current generated settings:

- IMAP SSL on port `993`
- POP3 SSL on port `995`
- SMTP submission STARTTLS on port `587`
- Username is `%EMAILADDRESS%`

### CalDAV/CardDAV Discovery Redirects

Endpoints:

- `GET /.well-known/caldav`
- `HEAD /.well-known/caldav`
- `GET /.well-known/carddav`
- `HEAD /.well-known/carddav`

Behavior:

- Returns `307 Temporary Redirect`.
- Redirects to `/dav/`.

Current note:

- The DAV implementation lives in the separate `groupware` daemon on port `8081`; this admin service only provides discovery redirects.

## Database Objects Managed

Current admin functions touch these tables:

- `tenants`
- `tenant_policies`
- `domains`
- `users`
- `dkim_keys`, read for DNS output

Related service-generated tables:

- `audit_events`
- `aliases`
- `quarantine_events`
- `calendars`
- `calendar_objects`
- `address_books`
- `contact_objects`

## Security Status

Current implemented protections:

- User passwords are stored as bcrypt hashes.
- API responses do not expose password hashes.
- Mail service authentication uses the same users table through Dovecot.
- DNS DKIM records are generated from active DKIM keys.

Important current gap:

- Admin UI/API authentication is not implemented yet in the committed code at this point. It should be added before exposing `8080` beyond a trusted test network.

Recommended next admin security functions:

- Admin login/session support.
- Role-based access control.
- Tenant-scoped admin permissions.
- CSRF protection for browser forms.
- Audit events for every create/update/delete.
- Optional IP allowlist or reverse-proxy auth.
- Rate limiting for login and write APIs.

## Planned Admin Panel Functions

### Tenant Management

Planned:

- Edit tenant name/slug.
- Suspend/reactivate tenant.
- Tenant quotas and limits.
- Tenant security policy editor.
- Tenant-specific branding.

### Domain Management

Planned:

- Domain verification workflow.
- DNS status checks for MX/SPF/DKIM/DMARC/MTA-STS/TLS-RPT.
- Activate/disable domain.
- DKIM rotate/retire/regenerate keys.
- Per-domain catch-all settings.
- Per-domain outbound signing policy.

### User Management

Planned:

- Edit display name.
- Reset password.
- Lock/unlock user.
- Disable/delete user.
- Mailbox quota editor.
- Force password change.
- Show mailbox storage usage.
- Show recent login/auth events.

### Aliases and Routing

Planned:

- Create aliases.
- Create distribution lists.
- Configure forwarding.
- Configure catch-all mailboxes.
- Tenant/domain route overrides.

### Spam and Malware Administration

Planned:

- Quarantine browser.
- Release/delete quarantined messages.
- Mark message as spam.
- Mark message as not spam.
- Per-user allow/block lists.
- Per-domain allow/block lists.
- Rspamd symbol view.
- Spam score threshold editor.
- Malware event dashboard.

### Mail Flow Operations

Planned:

- Queue viewer.
- Retry queued message.
- Delete queued message.
- Message trace by sender/recipient/message ID.
- SMTP delivery logs.
- DKIM signing status.
- TLS delivery status.

### Groupware Administration

Planned:

- Show user calendars.
- Show user address books.
- Reset DAV collections.
- Export/import contacts.
- Export/import calendars.
- Device/session list.

### Platform Operations

Planned:

- Service health dashboard.
- Version/build display.
- Config render status.
- Migration status.
- Backup/restore controls.
- Log viewer.
- Certificate/ACME management.

## Testing Notes

Useful current test account:

- Email: `marko@external-1778096561.local`
- Password: `secret123456`

Useful live checks:

```powershell
Invoke-RestMethod http://192.168.254.125:8080/healthz
Invoke-RestMethod http://192.168.254.125:8080/api/v1/domains
Invoke-RestMethod http://192.168.254.125:8080/api/v1/domains/2/dns
```
