# toolctl

<img src="https://user-images.githubusercontent.com/547220/146074557-339fc1e4-f83e-4cbb-b885-74cb6b52fd46.png" width="200px" alt="A drawing of a cute gopher holding a wrench">

[![GitHub Workflow Status (main branch)](https://img.shields.io/github/actions/workflow/status/toolctl/toolctl/ci.yml?branch=main)](https://github.com/toolctl/toolctl/actions/workflows/ci.yml?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/toolctl/toolctl)](https://goreportcard.com/report/github.com/toolctl/toolctl)
[![GitHub release (latest)](https://img.shields.io/github/v/release/toolctl/toolctl)](https://github.com/toolctl/toolctl/releases/latest)
[![GitHub](https://img.shields.io/github/license/toolctl/toolctl)](LICENSE)

`toolctl` helps you manage your tools on Linux and macOS.

## Installation

### Automatic

```shell
curl -fsSL https://raw.githubusercontent.com/toolctl/install/main/install | sh
```

### Manual

You can [download the latest version of toolctl](https://github.com/toolctl/toolctl/releases/latest) and run it from any directory you like.

## Getting Started

### Get information about tools

#### Info about all installed and supported tools

```text
❯ toolctl info
[k9s      ] ✨ k9s v0.25.8: Kubernetes CLI to manage your clusters in style
[k9s      ] ✅ k9s v0.25.8 is installed at /home/adent/.local/bin/k9s
[kubectl  ] ✨ kubectl v1.23.0: The Kubernetes command-line tool
[kubectl  ] 🔄 kubectl v1.21.2 is installed at /home/adent/.local/bin/kubectl
```

#### Info about a specific tool

```text
❯ toolctl info gh
✨ gh v2.4.0: GitHub's official command line tool
🏠 https://cli.github.com/
❌ Not installed
```

### Install tools

#### Install the latest version of a tool

```text
❯ toolctl install k9s
👷 Installing v0.25.8 ...
🎉 Successfully installed
```

#### Install a specific version of a tool

```text
❯ toolctl install terraform@0.11.15
👷 Installing v0.11.15 ...
🎉 Successfully installed
```

### Upgrade tools

```text
❯ toolctl upgrade
[gh     ] ✅ Already up to date (v2.34.0)
[toolctl] ✅ Already up to date (v0.4.11)
[yq     ] 👷 Upgrading from v4.13.4 to v4.13.5 ...
[yq     ] 👷 Removing v4.13.4 ...
[yq     ] 👷 Installing v4.13.5 ...
[yq     ] 🎉 Successfully installed
```

## Supported Tools

Currently, `toolctl` supports the following tools:

- [age](https://age-encryption.org/): A simple, modern and secure encryption tool
- [age-keygen](https://age-encryption.org/): A simple, modern and secure encryption tool
- [air](https://github.com/cosmtrek/air): Live reload for Go apps
- [chezmoi](https://chezmoi.io/): Manage your dotfiles across multiple diverse machines, securely
- [cloudflared](https://github.com/cloudflare/cloudflared): Cloudflare Tunnel client
- [dive](https://github.com/wagoodman/dive): A tool for exploring each layer in a docker image
- [dockerfilegraph](https://github.com/patrickhoefler/dockerfilegraph): Visualize your multi-stage Dockerfiles
- [eksctl](https://eksctl.io/): The official CLI for Amazon EKS
- [gdu](https://github.com/dundee/gdu): Fast disk usage analyzer with console interface written in Go
- [gh](https://cli.github.com/): GitHub's official command line tool
- [godolint](https://github.com/zabio3/godolint): Dockerfile linter, written in Golang 🐳
- [golangci-lint](https://golangci-lint.run/): Fast linters runner for Go
- [gping](https://github.com/orf/gping): Ping, but with a graph
- [helm](https://helm.sh/): The Kubernetes package manager
- [hugo](https://gohugo.io/): The world's fastest framework for building websites
- [k9s](https://k9scli.io/): Kubernetes CLI to manage your clusters in style
- [kind](https://kind.sigs.k8s.io/): Kubernetes in Docker - local clusters for testing Kubernetes
- [kompose](https://kompose.io/): Convert Compose to Kubernetes
- [kops](https://kops.sigs.k8s.io/): Production grade K8s installation, upgrades, and management
- [kubectl](https://kubernetes.io/docs/reference/kubectl/): The Kubernetes command-line tool
- [kubectx](https://github.com/ahmetb/kubectx): Faster way to switch between Kubernetes contexts
- [kubefwd](https://github.com/txn2/kubefwd): Bulk port forwarding Kubernetes services for local development
- [kubens](https://github.com/ahmetb/kubectx): Faster way to switch between Kubernetes namespaces
- [kuberlr](https://github.com/flavio/kuberlr): Simple management of multiple kubectl versions
- [kustomize](https://kustomize.io/): Template-free customization of Kubernetes configuration
- [minikube](https://minikube.sigs.k8s.io/): Run Kubernetes locally
- [pulumi](https://www.pulumi.com/): Developer-first infrastructure as code
- [skaffold](https://skaffold.dev/): Easy and repeatable Kubernetes development
- [sops](https://github.com/getsops/sops): Simple and flexible tool for managing secrets
- [stern](https://github.com/stern/stern): Multi pod and container log tailing for Kubernetes
- [task](https://taskfile.dev/): A task runner / simpler Make alternative written in Go
- [terraform](https://www.terraform.io/): Infrastructure as code software tool
- [tkn](https://github.com/tektoncd/cli): A CLI for interacting with Tekton
- [toolctl](https://github.com/toolctl/toolctl): The tool to control your tools
- [yq](https://mikefarah.gitbook.io/yq/): Portable command-line YAML processor

Our goal is to support as many tools as possible, so expect this list to grow significantly over time.

In general, `toolctl` currently supports any tool that:

✔ consists of a single executable file\
✔ runs on Linux and/or macOS\
✔ includes a command or flag to print its [semantic version](https://semver.org/)\
✔ provides its source code and precompiled binaries online under a free and open source license

If you know a tool that fits all of these criteria, please [open an issue](https://github.com/toolctl/api/issues/new?title=Tool%20request:%20) and let us know!

## Credits

The `toolctl` logo was created with [gopherize.me](https://gopherize.me/).
Artwork by [Ashley McNamara](https://twitter.com/ashleymcnamara) based on original artwork by [Renee French](https://reneefrench.blogspot.co.uk/).

## License

[MIT](LICENSE)
