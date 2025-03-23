# Mutating Webhook for Namespace Label Propagation

## Overview

This project implements a **mutating admission webhook** in Go, designed to enhance Kubernetes workload metadata management. The webhook ensures that labels assigned to a namespace are automatically copied to workloads (e.g. Pods) deployed within that respective namespace.

### Key Features

- **Label Synchronization**: Propagates namespace-level labels to workloads without manual intervention.
- **Customizable**: Easily configure which labels are copied via webhook settings.
- **Performance-Oriented**: Lightweight and optimized for production environments.
- **Secure**: Handles admission requests with authentication and RBAC.

---

## Installation

1. **Clone the Repository**
   ```bash
   git clone https://github.com/devops-yug/workload-label-mutating-webhook.git
   cd workload-label-mutating-webhook/helm
2. **Deploy the Webhook** Apply the provided Kubernetes manifests:
    ```bash
   helm install workload-label-mutating-webhook \
   --namespace workload-label-mutating-webhook \
   --create-namespace .
3. **Setup Namespace Labels** Add labels to namespaces using:
    ```bash
    kubectl label namespace <namespace-name> <key>=<value>
## Usage
Once deployed:

1. Labels assigned to namespaces are automatically appended to any workload created within those namespaces.
2. No changes are required in workload manifests—the webhook handles everything dynamically.

## How It Works
1. **Intercept Requests**: The webhook intercepts admission requests for workload creation.
2. **Mutate Metadata**: It evaluates the namespace's labels and applies them to the workload's `metadata.labels`.
3. **Return Response**: Kubernetes receives the modified workload and schedules it accordingly.

## Configuration
Modify the configuration file to specify:
- **Label Inclusion Rules**: Define which labels to propagate.
- **Namespace Selector**: Control which namespaces the webhook operates on.

## Development
### Prerequisites
- Go (>=1.20)
- Kubernetes (>=1.23)
- Docker (for containerized deployment)
- Helm

### Steps:
1. Build the project:
    ```bash
    go build -o webhook .
2. Run locally (for testing):
    ```bash
    ./webhook
## Contributing: 
We welcome contributions! Feel free to:
- Submit bug reports
- Open feature requests
- Create pull requests

## License
This project is licensed under the MIT License. See  for details.

You can now copy and paste this directly into a `README.md` file for your project. Let me know if there are any more adjustments you'd like! 🚀