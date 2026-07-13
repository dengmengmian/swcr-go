# Contributing to swcr-go

Thank you for considering contributing to swcr-go!

## Development setup

    git clone https://github.com/dengmengmian/swcr-go.git
    cd swcr-go

Requires Go 1.21+ and optionally golangci-lint.

## Quick checks before submitting

    make build   # compiles ./cmd/swcr
    make vet     # go vet ./...
    make test    # go test -race -count=1 ./...
    make lint    # go vet + gofmt check (also golangci-lint run)

## Code style

- Standard Go formatting (gofmt / goimports).
- All public symbols must have doc comments.
- Tests accompany every new package or significant function.
- golangci-lint (config in .golangci.yml) must pass with zero issues.

## Pull request process

1. Fork and create a feature branch.
2. Make your changes, including tests.
3. Ensure all checks pass before opening a PR.
4. Open a PR against main.

## Commit conventions

This project uses Conventional Commits:

- feat: new feature
- fix: bug fix
- refactor: code change (neither bug fix nor feature)
- docs: documentation only
- test: test additions or corrections
- chore: build scripts, CI, dependency updates

## License

By contributing, you agree that your contributions will be licensed under
the MIT License that covers this project.
