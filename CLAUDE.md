# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests (unit + lint + fmt + vet)
make test

# Run a single test (requires CONFIG env var)
CONFIG=testdata/config_test.yaml go test -race -run TestName ./pkg/...

# Run with coverage report
make coverage

# Run backend locally (port 9000)
make run

# Run frontend dev server (port 3000)
cd front && yarn install && yarn dev

# Build everything (frontend + Go binary + Docker image)
make build

# E2E tests against k3s
make e2e

# Lint Go
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run -v

# Lint frontend
cd front && yarn lint

# Validate Helm chart
make testChart
```

All unit tests require `CONFIG=testdata/config_test.yaml`. The test script is `./scripts/test-pkg.sh`.

## Architecture

This is a **feature-branch namespace manager** for Kubernetes: it lists, monitors, scales, and deletes namespaces that represent feature branch deployments, with integrations to GitLab, AWS, Azure, and Hetzner Cloud.

### Request path

HTTP request → `pkg/web` (gorilla/mux router) → `pkg/api` (business logic) → `pkg/client` (Kubernetes/cloud SDKs) → JSON response

- `pkg/web/web.go` sets up routes; `handlerAPI.go`, `handlerEnvironment.go` are the handler files.
- `pkg/api/` contains ~60 files, each focused on a specific operation (e.g., `GetEnvironments.go`, `ScaleNamespace.go`). The core entity is `Environment`, defined in `pkg/types/`.
- `pkg/client/` wraps all external SDKs (Kubernetes `client-go`, GitLab, AWS, Azure, Hetzner, Sentry).
- `pkg/cache/` provides Redis or in-memory caching, transparent to callers.

### Background batch

`pkg/batch/batch.go` runs as a Kubernetes leader-elected process. When elected, it periodically (every 30 min) triggers scale-down of stale namespaces and token cleanup. Enabled via `--batch.enabled=true` flag.

### Configuration

`pkg/config/config.go` defines all configuration as a single struct parsed from a YAML file pointed to by the `CONFIG` env var. Key fields: Kubernetes client config, GitLab URL/token, cloud credentials, scale-down delay rules per label selector.

### Frontend

`front/` is a Nuxt.js 2 (Vue 2) app. In production, `yarn generate` produces `front/dist/`, which is served as a SPA by `pkg/web/handlerSPA.go`. In development, the frontend proxies API calls to `localhost:9000`.

### Webhooks

`pkg/webhook/` receives events from AWS SNS, Azure Event Grid, or plain HTTP callbacks to trigger namespace operations without polling.

### Observability

- Prometheus metrics: `pkg/metrics/`
- OpenTelemetry tracing: `pkg/telemetry/`
- Sentry error reporting: wired in `pkg/client/`

## Linter notes

Max line length is 180. Disabled linters include: `noinlineerr`, `funcorder`, `godoclint`, `err113`, `cyclop`, `depguard`, `exhaustruct`, `funlen`, `gochecknoglobals`, `gocognit`, `ireturn`, `mnd`, `musttag`, `nestif`, `testpackage`, `varnamelen`, `wrapcheck`, `goconst`. Formatters enabled: `gci`, `gofmt`, `gofumpt`, `goimports`. Run `golangci-lint run` before submitting PRs — CI enforces it.
