# Security Policy

## Supported Versions

We currently support the production GreenVue API service only.  
Security patches and updates are applied continuously.

| Deployment                | Supported |
| ------------------------- | --------- |
| GreenVue API (production) | ✅        |

## Reporting a Vulnerability

If you discover a security vulnerability in GreenVue API, **please do not open a public issue**.

Instead, report it privately by emailing:

**greenvue.security@protonmail.com**

Please include:

- Detailed description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Optional: Suggestions for a fix or mitigation

We aim to acknowledge all security reports within **48 hours**, and provide a remediation timeline after validation.

---

## Security Practices

- All credentials and secrets are managed **externally through environment variables**.
- Development fallbacks exist only for **local development** and are **never used in production**.
- The codebase is routinely scanned with **[gosec](https://github.com/securego/gosec)** for known security issues.
- Sensitive credentials previously exposed have been **rotated** and **permanently purged** from Git history.
- `.env` and other sensitive config files are **.gitignored**.

---

## Platform Integrity

GreenVue API is a **centralized service** operated by the GreenVue team.  
Users **cannot self-host** the backend.  
We maintain full control over the production environment to ensure service integrity, reliability, and security.

---

## Acknowledgments

We thank all individuals who report security vulnerabilities responsibly.

Together, we make GreenVue stronger and safer.
