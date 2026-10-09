---
hide:
  - navigation
  - toc
---

<div class="vc-hero" markdown>

# vc-env

**A version manager for the [vcluster](https://www.vcluster.com/) CLI** —
install many `vcluster` versions side by side and let a shim pick the right
one per shell, per directory, or per machine.

[Get started](#install){ .md-button .md-button--primary }
[Installation guide](installation-and-configuration.md){ .md-button }
[GitHub](https://github.com/mmpyro/vc-env){ .md-button }

</div>

Think of `vc-env` as [tfenv](https://github.com/tfutils/tfenv) /
[pyenv](https://github.com/pyenv/pyenv) for
[vcluster](https://www.vcluster.com/). It installs each `vcluster` binary
under `$VCENV_ROOT/versions/<v>/vcluster`, drops a shim on your `PATH`, and
resolves the right version at every call — so `vcluster create`,
`vcluster connect` and friends keep working unchanged, they just hit the
version you actually want.

```sh
vc-env install 0.21.1      # download, checksum-verify, install
vc-env global  0.21.1      # or `vc-env local` / `vc-env shell`
vcluster version           # → 0.21.1
```

## Why vc-env

<div class="grid cards" markdown>

-   :material-swap-horizontal:{ .lg .middle } **Transparent shim**

    ---

    `$VCENV_ROOT/shims/vcluster` is prepended to `PATH` and dispatches to
    the resolved binary — scripts, IDEs and CI call `vcluster` as usual.

    [:octicons-arrow-right-24: How selection works](installation-and-configuration.md#how-configuration-is-discoveredloaded)

-   :material-layers-triple:{ .lg .middle } **Three-level version resolution**

    ---

    Shell (`VCENV_VERSION`) > local (`.vcluster-version`, walked up) >
    global (`$VCENV_ROOT/version`). First match wins.

    [:octicons-arrow-right-24: CLI reference](cli-reference.md#commands)

-   :material-database-check:{ .lg .middle } **Cached release list**

    ---

    `list-remote` and `latest` go through a three-layer cache (baked-in
    baseline → disk → delta fetch), so they stay fast and respect the
    GitHub rate limit.

    [:octicons-arrow-right-24: Caching strategy](caching.md)

-   :material-shield-check:{ .lg .middle } **Atomic install + checksum**

    ---

    `vc-env install` streams with a progress bar, verifies the SHA256
    from the release, and writes atomically (temp + rename). An
    interrupted install never leaves a half-written binary.

    [:octicons-arrow-right-24: install command](cli-reference.md#install)

-   :material-laptop:{ .lg .middle } **Cross-platform**

    ---

    Prebuilt binaries for `darwin/amd64`, `darwin/arm64`, `linux/amd64`
    and `linux/arm64`; OS and CPU architecture are detected automatically.

    [:octicons-arrow-right-24: Supported platforms](installation-and-configuration.md#supported-platforms)

-   :material-update:{ .lg .middle } **Self-upgrade**

    ---

    `vc-env upgrade` replaces the current binary in place with the latest
    stable release of `vc-env` — no package manager required.

    [:octicons-arrow-right-24: upgrade command](cli-reference.md#upgrade)

</div>

## Install

Pick how you want to install the `vc-env` binary. The next step is always
the same: export `VCENV_ROOT` and run `eval "$(vc-env init)"` in your
shell profile.

=== "Binary (macOS / Linux)"

    ```sh
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m); case $arch in x86_64) arch=amd64 ;; aarch64) arch=arm64 ;; esac
    curl -fsSL -o vc-env "https://github.com/mmpyro/vc-env/releases/download/v1.0.0/vc-env-$os-$arch"
    chmod +x vc-env && sudo mv vc-env /usr/local/bin/vc-env
    vc-env version
    ```

=== "From source"

    ```sh
    git clone https://github.com/mmpyro/vc-env.git
    cd vc-env
    make build
    sudo mv build/vc-env /usr/local/bin/vc-env
    vc-env version
    ```

=== "Cross-compile all platforms"

    ```sh
    make build-all
    ls build/
    ```

Then wire the shim into your shell:

```sh
# in ~/.bashrc or ~/.zshrc
export VCENV_ROOT="$HOME/.vcenv"
eval "$(vc-env init)"
source <(vc-env completion bash)  # or: zsh / fish / powershell
```

Full install details, platform notes and troubleshooting:
[Installation and configuration](installation-and-configuration.md).

## What's next

| Guide | Read it when you want to… |
|---|---|
| [Installation and configuration](installation-and-configuration.md) | install `vc-env`, set `VCENV_ROOT`, wire the shim, and resolve common errors |
| [CLI reference](cli-reference.md) | know every command, flag, environment variable and exit code |
| [Caching strategy](caching.md) | understand how `list-remote` / `latest` are cached and how to tune the TTL |
