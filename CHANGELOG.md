# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `.golangci.yml` lint configuration covering errcheck, staticcheck, gosec,
  govet, ineffassign, gocritic, revive, and more.
- CI `lint` job using `golangci/golangci-lint-action@v7`, with the `build` job
  gated behind it.
- `CONTRIBUTING.md` with development setup, code style, PR process, and commit
  conventions.
- `SECURITY.md` with vulnerability reporting process and scope.
- `.github/dependabot.yml` for weekly Go module and monthly GitHub Actions
  dependency updates.

### Fixed

- Variable shadowing in `marshalEncode` (`docx.go`) — the inner `err` now uses
  a distinct `writeErr` binding.
- Deferred `rc.Close()` inside a loop in `readZipEntry` (`docx_test.go`) —
  replaced with an explicit close, avoiding a potential resource leak.
- Test file permissions changed from `0o644` to `0o600` to satisfy gosec G306.

## [0.2.0] — 2026-07-13

### Added

- `--version` flag reporting build version, Go runtime, OS/arch, commit hash,
  and build date (injected via ldflags).
- Smart auto-exclude: common dependency/build directories (`node_modules`,
  `vendor`, `__pycache__`, `.venv`, `dist`, `build`, `target`, etc.) and binary
  file extensions (`.pyc`, `.class`, `.o`, `.so`, `.dll`, `.exe`, etc.) are
  skipped by default. Disable with `--no-auto-exclude`.
- `--dry-run` preview mode: prints file count, code line count, and estimated
  page count to stdout without generating a `.docx`.
- `--max-pages N` page limit with `--page-mode first|last|front30back30`
  truncation strategies — covers the common "front 30 + back 30 pages"
  requirement for software copyright submissions.
- Block comment support via a stateful `CommentStripper` that recognises
  `/* */`, `<!-- -->`, `""" """`, and `''' '''` out of the box.
  `--block-comment/-b OPEN:CLOSE` allows custom pairs; `--no-block-comment`
  disables the defaults.
- Shell completion subcommand: `swcr completion bash|zsh|fish|powershell`.
- Deterministic output: discovered source files are sorted alphabetically.
- `Makefile` with `build`, `test`, `vet`, `lint`, `clean`, and `install` targets.
- `.goreleaser.yml` for cross-platform release automation (linux/darwin/windows
  on amd64/arm64).

### Changed

- `CodeWriter` was refactored to separate `CollectLines` (in-memory) from
  `WriteLines` (to disk), enabling dry-run and pagination without generating
  a `.docx`.

### Removed

- Standalone `isBlankLine` and `isCommentLine` helpers — superseded by
  `CommentStripper.ProcessLine`.

## [0.1.0] — 2026-07-13

### Added

- Initial Go rewrite of [kenley2021/swcr](https://github.com/kenley2021/swcr)
  (Python, MIT).
- CLI built on `spf13/cobra` with flags matching the original `click` interface
  (`--title`, `--indir`, `--ext`, `--comment-char`, `--exclude`, `--outfile`,
  `--verbose`, and formatting options).
- Pure-Go `.docx` writer using only the standard library (`archive/zip` +
  `encoding/xml`) — no external docx dependency.
- `CodeFinder`: recursive directory walker with extension filtering and hidden-
  file skipping.
- `CodeWriter`: line-by-line comment and blank-line removal.
- Tuned default formatting (宋体 10.5pt, fixed 10.5pt line spacing, 0pt before,
  2.3pt after) that produces exactly 50 lines per A4 page.
- Page header with centred title and right-aligned PAGE field.
- CI matrix (`go build`, `go vet`, `go test -race`) across Go 1.21–1.23 and
  stable.

[Unreleased]: https://github.com/dengmengmian/swcr-go/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/dengmengmian/swcr-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/dengmengmian/swcr-go/compare/9eda77c...v0.1.0
