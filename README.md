# 🌐 External DNS :: Technitium Webhook

A specialized webhook provider for [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) that enables automated record management for **Technitium DNS Server**.

## 🛠 Integration Overview

This project is designed to run exclusively as a **sidecar container** within the `external-dns` pod. It implements the ExternalDNS Webhook provider API to bridge Kubernetes resource discovery with Technitium's management API.

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

> Note: You have to either provide ``TECHNITIUM_USER`` and ``TECHNITIUM_PASSWORD`` or just  ``TECHNITIUM_TOKEN``

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

1.  **Navigate:** Go to the **Administration** tab and select the **Users** sub-tab.
2.  **Create User:** Click to add a new user.
    > **Note:** Ensure the **Username** and **Password** contain **no spaces**. The *Display Name* is purely cosmetic and can be formatted however you like.
3.  **Assign Permissions:** Add the new user to the **DNS Administrators** group.
4.  **Session Management:** You may set the *Session Timeout* to `0` for an indefinite session. However, the application is designed to automatically re-authenticate and refresh the API key before the timeout expires.

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
kubectl create secret generic technitium-credentials --from-literal=token='<YOUR_TECHNITIUM_TOKEN>'

# Or using username/password:
# kubectl create secret generic technitium-credentials \
#   --from-literal=username='<YOUR_USERNAME>' \
#   --from-literal=password='<YOUR_PASSWORD>'

cat <<EOF > external-dns-technitium-values.yaml
image:
  tag: v0.0.0  # replace with the desired version

# -- ExternalDNS log level.
logLevel: debug  # reduce in production

# -- if true, ExternalDNS will run in a namespaced scope (Role and Rolebinding will be namespaced too).
namespaced: false

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

To compile the project locally, install [go-task](https://taskfile.dev/docs/installation and then) execute:

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

## 🚀 Contribution

We welcome all Pull Requests! To ensure a smooth review process, please adhere to the following requirements:

* **✅ Validation:** Ensure your changes pass all checks. Running `task lint` will automatically verify code quality and inject the required license headers into required source files.
* **🧪 Testing:** All new functionality **must** include corresponding unit tests. A successful test pass is required for any merge to the `main` branch.
* **✍️ Security:** The final commit of your PR must be **signed** (e.g., GPG/SSH) before it can be merged for release.

## 🤝 Acknowledgments

### 📄 Acknowledgments

### 🛠️ The Technitium Team
A huge thank you to the [Technitium](https://technitium.com/) team for building such a robust, high-performance, and feature-rich open-source DNS server. This project is intended to make integrating their excellent software with Kubernetes seamless and efficient.

### ❤️ Personal Thanks
Building and maintaining open-source tools takes time and focus. I want to give a special thanks to **my wife, my daughter, and my son**. Your support and patience allow me the space to be a "geek" and contribute back to the community. You are my greatest motivation!

### 🏗️ Credits
A special thanks to the [external-dns-infoblox-webhook](https://github.com/AbsaOSS/external-dns-infoblox-webhook) team. This plugin is based on their excellent work—thank you for providing such a solid foundation for the community!

## 📄 License

This project is licensed under the **Apache License 2.0**. See the [LICENSE](LICENSE) file for details.
