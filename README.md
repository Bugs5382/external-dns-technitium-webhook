# 🌐 External DNS :: Technitium Webhook

A specialized webhook provider for [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) that enables automated record management for **Technitium DNS Server**.

## 🛠 Integration Overview

This project is designed to run exclusively as a **sidecar container** within the `external-dns` pod. It implements the ExternalDNS Webhook provider API to bridge Kubernetes resource discovery with Technitium's management API.

> **Namespace layout:** This has been tested with Technitium and external-dns running in **different namespaces** as well as in the **same namespace**. Co-locating them in a single namespace is convenient for local homelab setups; separating them is the typical pattern for production clusters.

## 🚀 Quick Start

| Environment Variable       | Default value | Required |
|----------------------------|---------------|----------|
| TECHNITIUM_HOST            | localhost     | true     |
| TECHNITIUM_PORT            | 5380          | true     |
| TECHNITIUM_USER            |               | false    |
| TECHNITIUM_PASSWORD        |               | false    |
| TECHNITIUM_TOKEN           |               | false    |
| TECHNITIUM_SESSION_TTL     |               | false    |
| TECHNITIUM_SSL_VERIFY      | false         | false    |
| TECHNITIUM_DRY_RUN         | false         | false    |
| TECHNITIUM_CREATE_PTR      | false         | false    |
| TECHNITIUM_DEFAULT_TTL     | 300           | false    |
| TECHNITIUM_USE_TTL         | true          | false    |

> Note: You must provide either both `TECHNITIUM_USER` and `TECHNITIUM_PASSWORD`, or just `TECHNITIUM_TOKEN`.

external-dns environment variables:

| Environment Variable           | Default value | Required |
|--------------------------------|---------------|----------|
| SERVER_HOST                    | 127.0.0.1     | true     |
| SERVER_PORT                    | 8888          | true     |   
| HEALTH_CHECK_PORT              | 8080          | false    |
| SERVER_READ_TIMEOUT            |               | false    |
| SERVER_WRITE_TIMEOUT           |               | false    |
| DOMAIN_FILTER                  |               | true \*  |
| EXCLUDE_DOMAIN_FILTER          |               | true \*  |
| REGEXP_DOMAIN_FILTER           |               | true \*  |
| REGEXP_DOMAIN_FILTER_EXCLUSION |               | true \*  |
| REGEXP_NAME_FILTER             |               | true \*  |

> \* At least one of these must be set. While external-dns itself does not require a domain filter, this webhook will pause on startup if none are provided — running unscoped would let it claim every domain.

## 📄 Supported Records

| Record Type  | Status    |
|--------------|-----------|
| A     (IPv4) | Supported |
| AAAA  (IPv6) | Supported |
| CNAME        | Supported |
| TXT          | Supported |

You can manage these records types. PTR records and the zones will be created and/or deleted (Zones will not be deleted unless all records are gone and _TECHNITIUM_DELETE_PTR_ZONE_ is set to true.)

## ⚙️ Configuration

Follow these streamlined steps to configure your users and zones correctly.

### Phase 1: User Configuration
Before managing records, you need a user with the appropriate permissions.

#### Username and Password

1.  **Navigate:** Go to the **Administration** tab and select the **Users** sub-tab.
2.  **Create User:** Click to add a new user.
    > **Note:** Ensure the **Username** and **Password** contain **no spaces**. The *Display Name* is purely cosmetic and can be formatted however you like.
3.  **Assign Permissions:** Add the new user to the **DNS Administrators** group.
4.  **Session Management:** You may set the *Session Timeout* to `0` for an indefinite session. However, the application is designed to automatically re-authenticate and refresh the API key before the timeout expires.

#### Token Creation

1.  **Navigate:** After creating the user, head back to the **Sessions** tab and click **Create Token**.
2.  **Select User:** Choose the username you just created. Using the built-in `admin` account works but is not advised. You may optionally give the token a name.
3.  **Create & Copy:** Press **Create**, then copy the token immediately. As with any token, if you navigate away before copying it, you will have to start over.
4.  **Store:** Save the token as a Kubernetes secret (see the deployment example below).

### Phase 2: Zone Management
The API requires an existing zone to function. Please ensure your target zone is created manually before proceeding.

1.  **Navigate:** Open the **Zones** tab in the top navigation bar.
2.  **Initialize:** Click the **Add Zone** button.
3.  **Select Type:** Choose your required zone type.
    * **Primary Zone:** The standard choice for most setups.
    * **Other:** Consult your DNS Engineer if your infrastructure requires a Secondary or Forwarding zone.
4.  **Finalize:** Click **Add** to save the configuration.

## ☸️ Kubernetes Deployment

The Technitium webhook is provided as a regular OCI image released in the [GitHub container registry](https://github.com/Bugs5382/external-dns-technitium-webhook/pkgs/container/external-dns-technitium-webhook). The deployment can be performed in every way Kubernetes supports. The following example shows the deployment as a [sidecar container](https://kubernetes.io/docs/concepts/workloads/pods/#workload-resources-for-managing-pods) in the ExternalDNS pod using the [charts for ExternalDNS](https://github.com/kubernetes-sigs/external-dns/tree/master/charts/external-dns).

```shell
helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/

# Using a static API token (recommended):
# The secret must live in the same namespace as the external-dns / webhook pod.
kubectl create secret generic technitium-credentials --from-literal=token='<YOUR_TECHNITIUM_TOKEN>'

# Or using username/password:
# kubectl create secret generic technitium-credentials \
#   --from-literal=username='<YOUR_USERNAME>' \
#   --from-literal=password='<YOUR_PASSWORD>'

cat <<EOF > external-dns-technitium-values.yaml
# image:
#   tag: v0.0.0  # replace with the desired version of external-dns

# -- ExternalDNS log level.
logLevel: debug  # reduce in production

# -- if true, ExternalDNS will run in a namespaced scope (Role and Rolebinding will be namespaced too).
namespaced: false

# policy: sync  # sync will update the DNS records to match the desired state of the resources; default is "upsert" and should remain so until your confident everything is operational

# -- Kubernetes resources to monitor for DNS entries.
sources:
  - ingress
  - service
  - crd

provider:
  name: webhook
  webhook:
    image:
      repository: ghcr.io/bugs5382/external-dns-technitium-webhook
      tag: v0.0.0  # replace with the desired version
      pullPolicy: IfNotPresent
    env:
    - name: DOMAIN_FILTER
      value: "example.com"  # replace with your domain name
    - name: LOG_LEVEL
      value: debug  # reduce in production
    - name: TECHNITIUM_HOST
      value: "http://your-technitium-server"  # replace with your Technitium host URL
    - name: TECHNITIUM_PORT
      value: "5380"
    # Token-based auth (recommended):
    - name: TECHNITIUM_TOKEN
      valueFrom:
        secretKeyRef:
          name: technitium-credentials
          key: token
    # Username/password auth (alternative to token):
    # - name: TECHNITIUM_USER
    #   valueFrom:
    #     secretKeyRef:
    #       name: technitium-credentials
    #       key: username
    # - name: TECHNITIUM_PASSWORD
    #   valueFrom:
    #     secretKeyRef:
    #       name: technitium-credentials
    #       key: password
    - name: SERVER_PORT
      value: "8888"  # default and recommended port for the webhook provider
    - name: HEALTH_CHECK_PORT
      value: "8080"  # default and recommended port for metrics and health endpoints
    - name: TECHNITIUM_DRY_RUN
      value: "true"  # set to "false" to allow changes to your DNS records
EOF

helm upgrade external-dns-technitium external-dns/external-dns \
  --version 1.19.0 \
  -f external-dns-technitium-values.yaml \
  --install
```

## 🏗 Development

### 🛠 Build

To compile the project locally, install [go-task](https://taskfile.dev/docs/installation) and then execute:

```bash
task build
```

To remove build artifacts and clean your workspace:

```bash
task clean
```

If you are **contributing** to this project, you must first initialize the linting environment:
 ```bash
 task lint-init
 ```
 This command installs all necessary dependencies and tools for code analysis.

Once initialized, you can analyze the codebase by running:

```bash
task lint
```

To verify only the project licenses, use:

```bash
task license
```

### 🧪 Test

To execute the unit testing suite, run:

```bash
task test
```

### 🔁 End-to-End Tests

A live integration test stands up a [`kind`](https://kind.sigs.k8s.io/) cluster and runs the full pipeline — Technitium DNS Server, this webhook, and ExternalDNS — to confirm that a Kubernetes Service annotation produces a real record inside a Technitium zone. **The E2E job is a required check on `main`; PRs cannot merge until it is green.**

**What it does**

1. Builds the webhook image locally and loads it into the `kind` cluster as `external-dns-technitium-webhook:e2e`.
2. Installs Technitium via the [`Bugs5382/helm-technitium-chart`](https://github.com/Bugs5382/helm-technitium-chart) chart with `admin` / `admin` bootstrapped through `config.adminPassword`.
3. Logs in to the Technitium API and creates a Primary zone (`example.test`).
4. Installs ExternalDNS (chart `1.19.0`) with this webhook as a sidecar, pointed at the in-cluster Technitium Service.
5. Applies a test `Service` annotated with `external-dns.alpha.kubernetes.io/hostname=e2e.example.test` and `target=10.0.0.42`.
6. Polls the Technitium API until the expected `A` record appears in the zone (or fails with full diagnostics).

**Where the files live**

| Path                                       | Purpose                                                       |
|--------------------------------------------|---------------------------------------------------------------|
| `.github/workflows/job-e2e.yaml`           | GitHub Actions workflow — runs on PRs and pushes to `main`.   |
| `__test__/e2e/run.sh`                      | Orchestration script (idempotent; safe to re-run).            |
| `__test__/e2e/values-technitium.yaml`      | Helm values for the Technitium chart.                         |
| `__test__/e2e/values-external-dns.yaml`    | Helm values for ExternalDNS + this webhook as a sidecar.      |
| `__test__/e2e/workload.yaml`               | The Service whose annotations drive the record.               |

**Helm chart version policy**

The workflow checks out [`Bugs5382/helm-technitium-chart`](https://github.com/Bugs5382/helm-technitium-chart) at the **default branch (`main`) — i.e. the latest commit** — on every run. This intentionally exposes any chart-side breakage early. If a chart change breaks the E2E:

1. Confirm the failure reproduces locally (see below).
2. If the chart change is the cause, decide whether to update the values in `__test__/e2e/` or to pin the workflow to a known-good chart ref. Either change goes in the same PR as the fix.
3. The `workflow_dispatch` trigger accepts a `chart_ref` input (branch, tag, or SHA) so you can re-run the job against a specific chart version without editing the workflow.

When you cut a release of this webhook, also note the chart commit SHA you validated against in the release notes — that is what was actually exercised by CI.

**Running it locally**

You need `kind`, `helm`, `kubectl`, `docker`, `curl`, and `jq` on your `$PATH`.

```bash
# 1. Build and load the webhook image into a kind cluster named "e2e".
mkdir -p dist/external-dns-technitium-webhook_linux_amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath \
  -o dist/external-dns-technitium-webhook_linux_amd64/external-dns-technitium-webhook \
  ./cmd/webhook
docker build --build-arg TARGETARCH=amd64 -t external-dns-technitium-webhook:e2e .

kind create cluster --name e2e
kind load docker-image external-dns-technitium-webhook:e2e --name e2e

# 2. Clone the chart next to this repo (or anywhere; pass the path through env).
git clone https://github.com/Bugs5382/helm-technitium-chart.git ../helm-technitium-chart

# 3. Run the orchestrator.
TECHNITIUM_CHART_PATH=../helm-technitium-chart/technitium ./__test__/e2e/run.sh

# 4. Tear down when finished.
kind delete cluster --name e2e
```

The script accepts a few overrides via environment variables — see the comments at the top of `__test__/e2e/run.sh` for the full list (`ZONE`, `RECORD`, `RECORD_TARGET`, `ADMIN_USER`, `ADMIN_PASS`, `WEBHOOK_IMAGE`, `EXTERNAL_DNS_CHART_VERSION`).

## 🚀 Contribution

We welcome all Pull Requests! To ensure a smooth review process, please adhere to the following requirements:

* **✅ Validation:** Ensure your changes pass all checks. Running `task lint` will automatically verify code quality and inject the required license headers into required source files.
* **🧪 Unit Tests:** All new functionality **must** include corresponding unit tests. A successful test pass is required for any merge to the `main` branch.
* **🔁 E2E Tests:** The end-to-end job (`.github/workflows/job-e2e.yaml`) must be green before merge. If your change touches the Technitium API surface, the webhook server, or the deployment shape, run it locally first — see [End-to-End Tests](#-end-to-end-tests) above.
* **✍️ Security:** The final commit of your PR must be **signed** (e.g., GPG/SSH) before it can be merged for release.

## 🤝 Acknowledgments

### 🛠️ The Technitium Team
A huge thank you to the [Technitium](https://technitium.com/) team for building such a robust, high-performance, and feature-rich open-source DNS server. This project is intended to make integrating their excellent software with Kubernetes seamless and efficient.

### ❤️ Personal Thanks
Building and maintaining open-source tools takes time and focus. I want to give a special thanks to **my wife, my daughter, and my son**. Your support and patience allow me the space to be a "geek" and contribute back to the community. You are my greatest motivation!

### 🏗️ Credits
A special thanks to the [external-dns-infoblox-webhook](https://github.com/AbsaOSS/external-dns-infoblox-webhook) team. This plugin is based on their excellent work—thank you for providing such a solid foundation for the community!

## 📄 License

This project is licensed under the **Apache License 2.0**. See the [LICENSE](LICENSE) file for details.
