# toolctl — controls your tools

<img src="https://user-images.githubusercontent.com/547220/146074557-339fc1e4-f83e-4cbb-b885-74cb6b52fd46.png" width="200px" alt="A drawing of a cute gopher holding a wrench">

![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/toolctl/toolctl/CI/main) ![Codecov](https://img.shields.io/codecov/c/gh/toolctl/toolctl) ![Code Climate maintainability](https://img.shields.io/codeclimate/maintainability/toolctl/toolctl) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/toolctl/toolctl) ![GitHub](https://img.shields.io/github/license/toolctl/toolctl)

`toolctl` helps you manage your tools on Linux and macOS.

## How do I install `toolctl`?

For now, please head to the [release page](https://github.com/toolctl/toolctl/releases) and download the latest version for your platform.

In the not-too-distant future, we will also provide an easy-to-use setup script for `toolctl`.

## What can I do with `toolctl`?

### Get information about tools

```text
❯ toolctl info k9s
✨ k9s v0.25.8: Kubernetes CLI to manage your clusters in style
🏠 https://k9scli.io/
❌ Not installed

❯ toolctl info kubectl
✨ kubectl v1.23.0: The Kubernetes command-line tool
🔄 kubectl v1.21.2 is installed at /usr/local/bin/kubectl
```

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

## What will I soon be able to do with `toolctl`?

- Upgrade your tools

You have more ideas? Please [create an issue](https://github.com/toolctl/toolctl/issues/new) and let us know!

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
