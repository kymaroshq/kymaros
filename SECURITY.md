# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Kymaros, please report it
responsibly by emailing **security@kymaros.io**.

Please do **NOT** open a public GitHub issue for security vulnerabilities.

We will acknowledge your report within 48 hours and provide a timeline
for a fix. We aim to release patches within 7 days of confirmation.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.6.x   | Current release    |
| < 0.6   | Not supported      |

## Security Model

Kymaros follows the principle of least privilege:

- The Operator only requests the RBAC permissions it needs (see `config/rbac/role.yaml`)
- Sandbox namespaces are isolated with NetworkPolicy deny-all by default
- No data leaves your cluster (no telemetry, no phone-home)
- The API server does not store credentials (uses Kubernetes Secrets)
- All container images run as non-root with read-only root filesystems
- Exec health checks use the Kubernetes API (no shell injection surface)
