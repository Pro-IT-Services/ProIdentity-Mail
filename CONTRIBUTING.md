# Contributing

Thanks for helping improve ProIdentity Mail.

## License of Contributions

By submitting a contribution, you agree that your contribution is licensed under
the same dual license as the project, as described in `LICENSE`.

Unless a separate written agreement says otherwise, you also grant ProIdentity
permission to include your contribution in commercial ProIdentity licenses.
This keeps the project able to offer a paid license for SaaS, hosting, MSP,
white-label, OEM, and cloud-provider use.

## Security Issues

Please do not publish exploit details for active vulnerabilities before a fix is
available. Report security issues privately to the project owner first.

## Development Checks

Run these before submitting changes:

```bash
go test ./...
```

On Windows development machines, also run:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\security-check.ps1
```
