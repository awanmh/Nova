# Security Policy

## Reporting a Vulnerability
If you discover a security vulnerability within NOVA, please report it privately to the maintainers. Do not disclose security vulnerabilities publicly until they have been addressed.

## Secret Redaction & Safety Boundaries
NOVA is designed with local-first security as a primary requirement:
- Sensitive files (`.env`, `*.pem`, `*.key`, tokens) are redacted or ignored by default.
- Filesystem operations are strictly bounded to the workspace directory.
- Shell executions undergo safety risk classification and require explicit user approval for destructive commands.
