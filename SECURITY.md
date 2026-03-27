# Security

## Supported versions

Security fixes are applied to the default branch (`main`). Use the latest commit or tagged release.

## Reporting a vulnerability

Please **do not** open a public issue for security vulnerabilities.

Instead, report them through one of these channels:

1. **GitHub Security Advisories** (preferred): open a [private security advisory](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability) for this repository if the feature is enabled.
2. **Maintainer contact**: if advisories are unavailable, contact the repository owner via the email shown on their GitHub profile, with subject line containing `[Timo security]`.

Include enough detail to reproduce or understand the issue (steps, affected version or commit, impact). We aim to acknowledge reports within a few business days.

## Scope notes

- Timo imports **Moodle course backups** (`.mbz`). Treat backup files as sensitive: they may contain course content, user identifiers, or grades depending on what Moodle included.
- The HTTP server is intended for **local or trusted networks** unless you add your own authentication, TLS, and hardening. Do not expose it to the public internet without a reverse proxy and appropriate controls.
