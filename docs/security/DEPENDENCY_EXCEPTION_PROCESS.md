# Dependency Vulnerability Exception Process

This process is used when `scripts/security-check.ps1` or CI reports a dependency vulnerability that cannot be immediately removed or upgraded.

## Rules

1. No exception is allowed for a vulnerability that `govulncheck` reports as reachable from ProIdentity Mail code unless the affected feature is disabled and a compensating control is documented.
2. Exceptions must be time-limited. The maximum initial expiry is 30 days for High/Critical and 90 days for Medium/Low.
3. The exception must name the module, version, vulnerability identifier, affected services, reachability status, and planned remediation.
4. The exception must describe why upgrading or removing the dependency is not possible immediately.
5. The exception must be reviewed before release. Expired exceptions block release.

## Required Record

Use this format in the release/security notes:

```text
Dependency exception:
  Module:
  Version:
  Vulnerability:
  Severity:
  Reachable by govulncheck: yes/no
  Affected service(s):
  Reason upgrade is delayed:
  Compensating control:
  Owner:
  Opened:
  Expires:
  Remediation plan:
```

## Current Policy

The default release decision is: no reachable High or Critical dependency vulnerabilities may ship. Unreachable findings may ship only with a current exception record and a planned upgrade/removal.
