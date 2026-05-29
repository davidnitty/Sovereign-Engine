# Sovereign Engine

Sovereign Engine is a Go-based multi-cloud infrastructure management platform with:

- A Cobra CLI in `cmd/cli`
- A dashboard server in `cmd/ui`
- A standalone encryption helper in `cmd/encrypt`
- AES-256-GCM secret encryption
- SQLite-backed resource state
- A provider abstraction for AWS, Azure, GCP, DigitalOcean, Hetzner, and OVH
- YAML compositions in `compositions/`

## Phase 1 Status

Phase 1 provides a functional local control plane:

- `engine deploy <composition> --provider <name> --name <instance>`
- `engine list [--provider <name>]`
- `engine status <id>`
- `engine destroy <id>`
- `engine cleanup enable|disable|status`
- `engine encrypt <string>`
- `engine decrypt <string>`

Deployments load a composition YAML file, create a resource through the selected provider adapter, and persist state in SQLite. The provider adapters are intentionally narrow and side-effect free until provider-specific creation calls are wired in with account-safe parameters.

Crossplane does not currently ship as an embeddable in-process control plane with a SQLite backend. This repository therefore isolates composition execution behind `internal/composition.Engine` so a real Crossplane-backed implementation can replace the local executor without changing the CLI surface.

## Security

Set `ENGINE_MASTER_KEY` in production. If unset, Sovereign Engine creates a local development key under the data directory and creates `.env.enc` with `0600` permissions. In production, serve the dashboard/API behind a TLS-terminating reverse proxy.

## Build

```bash
go mod tidy
go test ./...
go build -o engine ./cmd/cli
go build -o engine-ui ./cmd/ui
go build -o engine-encrypt ./cmd/encrypt
```

`github.com/mattn/go-sqlite3` requires CGO. Static builds are possible with a suitable C toolchain and platform-specific linker flags.
