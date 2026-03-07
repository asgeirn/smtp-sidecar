# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
go build ./...                    # Build
go test ./...                     # Run all tests
go test -race ./...               # Run tests with race detector (used in CI)
go test -run TestMailHandler ./...  # Run a single test
go fmt ./...                      # Format code
go vet ./...                      # Lint
go tool golang.org/x/vuln/cmd/govulncheck ./...  # Vulnerability check
```

Cross-compile for Docker (scratch image, no CGO):
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o smtp-sidecar ./...
```

## Architecture

This is a single-package Go application (`package main`) that acts as an SMTP-to-Gmail-API bridge. It runs as a sidecar container, accepting SMTP connections and forwarding emails via the Gmail API.

**Flow:** SMTP client → `smtpd` listener (`main.go`) → `MailHandler` (`handler.go`) → sender/recipient filtering (`matcher.go`) → Gmail API send

Key components:
- **main.go** — Config parsing (env vars via `caarlos0/env`), OAuth2 client setup, SMTP server startup
- **handler.go** — `MailHandler` returns a closure that validates senders/recipients against regex patterns, then sends via Gmail API. Silently discards emails that fail validation (returns nil error)
- **matcher.go** — `MatchAnyPattern`: matches a string against a slice of regexps; empty pattern list permits all
- **patterns.go** — `ConvertToRegexPatterns`: parses comma-separated string into compiled `[]*regexp.Regexp`

## Configuration

Environment variables: `SMTP_LISTEN` (default `:2525`), `CREDENTIALS_JSON`, `TOKEN_JSON`, `SENDERS`, `RECIPIENTS`. The `SENDERS` and `RECIPIENTS` values are comma-separated regex patterns.

## CI/CD

GitLab CI pipeline (`.gitlab-ci.yml`) builds multi-arch Linux binaries (amd64/arm64), deploys as a scratch-based container image, and runs security scanning (govulncheck, semgrep, grype, trivy).
