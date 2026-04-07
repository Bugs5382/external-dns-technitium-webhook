# 🌐 External DNS :: Technitium Webhook

A specialized webhook provider for [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) that enables automated record management for **Technitium DNS Server**.

## 🛠 Integration Overview

This project is designed to run exclusively as a **sidecar container** within the `external-dns` pod. It implements the ExternalDNS Webhook provider API to bridge Kubernetes resource discovery with Technitium's management API.

## 🚀 Quick Start

| Environment Variable   | Default value | Required |
|------------------------|---------------|----------|
| TECHNITIUM_HOST        | localhost     | true     |
| TECHNITIUM_PORT        | 53443         | true     |
| TECHNITIUM_APIKEY      |               | true     |
| TECHNITIUM_SSL_VERIFY  | true          | false    |
| TECHNITIUM_DRY_RUN     | false         | false    |
| TECHNITIUM_CREATE_PTR  | false         | false    |
| TECHNITIUM_DEFAULT_TTL | 300           | false    |
| TECHNITIUM_USE_TTL     | true          | false    |

## 📄 Supported Records

## ⚙️ Configuration

Setting up Technitium doesn’t have to be a chore. I’ve polished your instructions to make them more professional, readable, and authoritative while keeping that helpful peer-to-peer vibe.

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

## 🏗 Development

### 🛠 Build

To compile the project locally, execute:

```bash
make build
```

To remove build artifacts and clean your workspace:

```bash
make clean
```

If you are **contributing** to this project, you must first initialize the linting environment:
 ```bash
 make lint-init
 ```
 This command installs all necessary dependencies and tools for code analysis.

Once initialized, you can analyze the codebase by running:

```bash
make lint
```

To verify only the project licenses, use:

```bash
make license
```

### 🧪 Test

To execute the unit testing suite, run:

```bash
make test
```

## 🚀 Contribution

We welcome all Pull Requests! To ensure a smooth review process, please adhere to the following requirements:

* **✅ Validation:** Ensure your changes pass all checks. Running `make lint` will automatically verify code quality and inject the required license headers into required source files.
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