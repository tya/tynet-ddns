# tynet-ddns

Keeps one Cloudflare DNS A record (`tynet.us`'s apex) pointed at the home
network's current public IP. No static IP from the ISP, so this has to run
on a schedule.

## How it works

One check-and-maybe-update pass per invocation — no internal loop. Designed
to run as a Kubernetes CronJob (see `deploy/charts/tynet-ddns`), which does
the scheduling:

1. GET the current public IP (`https://api.ipify.org`).
2. GET the Cloudflare record's current `content`.
3. Equal: log "no change", exit 0. Different: PATCH the record, log
   old → new, exit 0.

A non-zero exit surfaces as a failed Job — no notification plumbing beyond
that; the next scheduled run just tries again.

## Running locally

```sh
go build -o bin/tynet-ddns .
CF_API_TOKEN=... CF_ZONE_ID=... CF_RECORD_ID=... ./bin/tynet-ddns -dry-run
```

`-dry-run` logs the diff without PATCHing — use it to sanity-check
credentials/IDs before a real run.

## Deployment

See `deploy/charts/tynet-ddns` (Helm chart, rendered by ArgoCD — see
`gitops/ddns/application.yaml` in `tya/tynet-infra`). The chart's
`OnePasswordItem` resolves `CF_API_TOKEN` via a dedicated 1Password vault
(`tynet-ddns`) and a narrowly-scoped Service Account + operator instance,
isolated from any other app's secrets — see
`gitops/ddns-onepassword-operator` in `tya/tynet-infra` for why.

Image: `ghcr.io/tya/tynet-ddns`, built by `.github/workflows/image.yml` on
`v*` tags (`linux/arm64` only — this only ever runs on Pi nodes).
