# CLAUDE.md

Guidance for Claude Code when working in this repo.

## What this is

A tiny, single-purpose Go CLI: one Cloudflare DNS A record
(`tynet.us`'s apex), kept pointed at the home network's current public IP.
Deployed as a Kubernetes CronJob in `tya/tynet-infra`'s cluster (see that
repo's `gitops/ddns/application.yaml`) — this repo owns the binary, its
container image, and the Helm chart Argo renders; it does not own anything
cluster-side beyond that chart.

## Design choices worth knowing before changing this

- **One-shot, not a loop.** `main.go` does exactly one check-and-maybe-
  update pass and exits. The Kubernetes CronJob does the scheduling (see
  `deploy/charts/tynet-ddns/values.yaml`'s `schedule`, default every 5
  minutes) — there's no ticker/sleep loop to get wrong, and a stuck run
  just shows up as a slow Job rather than wedging a long-lived process.
- **No zone/record lookup by name.** `CF_ZONE_ID`/`CF_RECORD_ID` are passed
  in directly (plain config, not secret — see the chart's values.yaml).
  This tool only ever targets one record, so there's no reason to resolve
  zone/record names via the API on every run the way a multi-domain tool
  would.
- **`CF_API_TOKEN` is the only secret**, delivered via a 1Password
  `OnePasswordItem` → a dedicated vault (`tynet-ddns`) → a narrowly-scoped
  Service Account → its own 1Password Kubernetes Operator instance,
  isolated from every other app in that cluster (see `tya/tynet-infra`'s
  `gitops/ddns-onepassword-operator`). Don't widen that vault's scope or
  point this chart at a different (shared) vault without understanding why
  it was split out that way in the first place.
- **Partial PATCH.** `CloudflareClient.UpdateContent` sends only
  `{"content": ...}` — Cloudflare's API treats this as a partial update, so
  the record's `name`/`type`/`ttl`/`proxied` settings are left alone.
- **`HTTPClient` interface** (`cloudflare.go`, used by `publicip.go` too)
  exists purely for test injection — there is no real alternate
  implementation, don't add one speculatively.

## Release flow

1. Merge to `main` (CI: `go vet`/`gofmt -l`/`go test -race`, see
   `.github/workflows/ci.yml`).
2. Tag `vX.Y.Z` → `.github/workflows/image.yml` builds and pushes
   `ghcr.io/tya/tynet-ddns:vX.Y.Z` + `:latest` (`linux/arm64` only).
3. Bump `deploy/charts/tynet-ddns/values.yaml`'s `image.tag` to match,
   commit. Argo (manual sync, same convention as every app in
   `tynet-infra`) picks it up on the next sync.

No `.deb`/apt-packaging pipeline here, deliberately — this only ever runs
as a container in the cluster, never installed directly on a host.
