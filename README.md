# toolctl — controls your tools

<img src="https://user-images.githubusercontent.com/547220/146074557-339fc1e4-f83e-4cbb-b885-74cb6b52fd46.png" width="200px" alt="A drawing of a cute gopher holding a wrench">

[![GitHub Workflow Status (main branch)](https://img.shields.io/github/workflow/status/toolctl/toolctl/CI/main)](https://github.com/toolctl/toolctl/actions?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/toolctl/toolctl)](https://goreportcard.com/report/github.com/toolctl/toolctl)
[![GitHub release (latest)](https://img.shields.io/github/v/release/toolctl/toolctl)](https://github.com/toolctl/toolctl/releases/latest)
[![GitHub](https://img.shields.io/github/license/toolctl/toolctl)](LICENSE)

`toolctl` helps you manage your tools on Linux and macOS.

## How do I install `toolctl`?

For now, please [download the latest version](https://github.com/toolctl/toolctl/releases/latest) for your platform.

In the not-too-distant future, we will also provide an easy-to-use setup script for `toolctl`.

## What can I do with `toolctl`?

### Install the latest versions of a tool

```text
❯ toolctl install k9s
👷 Installing v0.25.8 ...
🎉 Successfully installed
```

### Install a specific version of a tool

```text
❯ toolctl install kustomize@3.9.4
👷 Installing v3.9.4 ...
🎉 Successfully installed
```

### Upgrade a tool

```text
❯ toolctl upgrade yq
👷 Upgrading from v4.13.4 to v4.13.5 ...
👷 Removing v4.13.4 ...
👷 Installing v4.13.5 ...
🎉 Successfully installed
```

### Get information about tools

```text
❯ toolctl info k9s
✨ k9s v0.25.8: Kubernetes CLI to manage your clusters in style
✅ k9s v0.25.8 is installed at /usr/local/bin/k9s

❯ toolctl info kubectl
✨ kubectl v1.23.0: The Kubernetes command-line tool
🔄 kubectl v1.21.2 is installed at /usr/local/bin/kubectl

❯ toolctl info kuberlr
✨ kuberlr v0.4.1: Simple management of multiple kubectl versions
🏠 https://github.com/flavio/kuberlr
❌ Not installed
```

### Check if your tools are up-to-date

```text
❯ toolctl list | xargs toolctl info
[k9s      ] ✨ k9s v0.25.8: Kubernetes CLI to manage your clusters in style
[k9s      ] ✅ k9s v0.25.8 is installed at /usr/local/bin/k9s
[kubectl  ] ✨ kubectl v1.23.0: The Kubernetes command-line tool
[kubectl  ] 🔄 kubectl v1.21.2 is installed at /usr/local/bin/kubectl
[kustomize] ✨ kustomize v4.4.1: Template-free customization of Kubernetes configuration
[kustomize] 🔄 kustomize v3.9.4 is installed at /usr/local/bin/kustomize
```

## Which tools does `toolctl` support?

Currently it supports the following tools:

- age
- helm
- k9s
- kubectl
- kubectx
- kubefwd
- kubens
- kuberlr
- kustomize
- minikube
- yq

## Can I manage _my-favorite-tool_ with `toolctl`?

`toolctl` supports any tool that:

✔ consists of a single executable file\
✔ runs on Linux and/or macOS\
✔ includes a version command or flag\
✔ provides its source code and precompiled binaries online under a free and open source license

You know a tool that fits all of these criteria?
Please [create an issue](https://github.com/toolctl/toolctl/issues/new) and let us know!

## Credits

The `toolctl` logo was created with [gopherize.me](https://gopherize.me/).
Artwork by [Ashley McNamara](https://twitter.com/ashleymcnamara) based on original artwork by [Renee French](https://reneefrench.blogspot.co.uk/).

## License

[MIT](LICENSE)
