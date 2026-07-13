# Security Policy

## Reporting a vulnerability

If you discover a security issue, please do not open a public issue.
Report it via the Security tab on GitHub:
https://github.com/dengmengmian/swcr-go/security

Expect an acknowledgement within 48 hours and a resolution timeline
within 7 days.

## Scope

swcr-go is a local CLI tool that reads source code from disk and writes
a .docx file. There is no network I/O and no server component.

Security concerns that are in scope:

- Path traversal when resolving --indir or --exclude arguments.
- Maliciously crafted filenames that escape the base directory.
- Any avenue where processing untrusted source code could trigger
  unexpected behaviour (very large files, long lines, binary content
  misidentified as source).

Concerns that are out of scope:

- Anything requiring the attacker to already have arbitrary code
  execution on the local machine.
- Issues in third-party development tools (Go compiler, golangci-lint,
  goreleaser) -- report those upstream.

## Supported versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |

No backport branches are maintained for older releases.
