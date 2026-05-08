# Security Policy

Thanks for helping keep `external-dns-technitium-webhook` and its users safe. This document explains how to report a vulnerability, what is in scope, and what to expect after you do.

## Supported Versions

This project follows semantic versioning. Only the latest minor release line receives security fixes; older minors are not back-patched.

| Version  | Supported          |
|----------|--------------------|
| `0.1.x`  | :white_check_mark: |
| `< 0.1`  | :x:                |

If you are running a build from `main` rather than a tagged release, please reproduce the issue against the latest published image (`ghcr.io/bugs5382/external-dns-technitium-webhook:<latest>`) before reporting.

## Reporting a Vulnerability

**Please do not open a public GitHub issue, pull request, or discussion for security problems.** Public disclosure before a fix is available puts every operator running this webhook at risk.

Instead, report privately via GitHub Security Advisories:

> https://github.com/Bugs5382/external-dns-technitium-webhook/security/advisories/new

A useful report typically includes:

- **Affected version(s)** — webhook image tag or commit SHA, plus ExternalDNS and Technitium DNS Server versions if relevant.
- **Impact** — what an attacker can do (e.g. read or modify DNS records, leak credentials, escalate within the cluster).
- **Reproduction** — minimal steps, configuration, or proof-of-concept. Redact any secrets from logs and config.
- **Suggested mitigation** — optional, but appreciated.

If the issue is in [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) or [Technitium DNS Server](https://github.com/TechnitiumSoftware/DnsServer) itself rather than in this webhook, please report it to those projects directly.

## Response Process

This is a small, volunteer-maintained project. Realistic timelines:

| Stage                          | Target                                |
|--------------------------------|---------------------------------------|
| Acknowledge report             | Within 7 days                         |
| Initial assessment & severity  | Within 14 days of acknowledgement     |
| Fix or mitigation              | Best-effort; depends on severity      |
| Coordinated public disclosure  | After a fix is released, or by mutual agreement if a fix is not feasible |

You will be kept in the loop on progress and credited in the published advisory unless you ask to remain anonymous.

## Scope

**In scope**

- The webhook code in this repository (`cmd/`, `internal/`).
- The published container image and its build pipeline (`Dockerfile`, GitHub Actions in `.github/workflows/`).
- The example deployment manifests and Helm values shown in `README.md`, where they could mislead an operator into an insecure default.

**Out of scope**

- Vulnerabilities in upstream ExternalDNS or Technitium DNS Server — report those upstream.
- Misconfiguration in your own cluster (e.g. leaking secrets via overly permissive RBAC, running with `TECHNITIUM_SSL_VERIFY=false` against an untrusted network).
- Denial-of-service attacks that require already-authenticated access to the Technitium API or to the Kubernetes cluster.
- Findings from automated scanners without a demonstrable exploit.

## Hall of Fame

No vulnerabilities have been reported yet. Researchers who report valid issues will be credited here (with their permission) once advisories are published.

## Operator Hardening Checklist

Not strictly part of this policy, but recommended for anyone running the webhook:

- Prefer `TECHNITIUM_TOKEN` over `TECHNITIUM_USER`/`TECHNITIUM_PASSWORD` and store it in a Kubernetes `Secret`, not a plain `env` value.
- Always set a `DOMAIN_FILTER` (or one of the regex filters); the webhook will refuse to start without one to avoid accidentally claiming every domain.
- Run with `TECHNITIUM_SSL_VERIFY=true` whenever the Technitium API is reachable over TLS.
- Keep `TECHNITIUM_DRY_RUN=true` until you have validated record changes in a non-production environment.
- Restrict network egress from the webhook pod so it can only reach the Technitium API host.
