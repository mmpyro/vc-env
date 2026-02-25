# vc-env Documentation

`vc-env` is a version manager for the [`vcluster`](https://www.vcluster.com/) CLI.

It installs multiple `vcluster` versions under a single root directory, generates a `vcluster` shim, and selects the correct `vcluster` binary at runtime based on your configured version.

## Guides

- [Installation and configuration](installation-and-configuration.md)
- [CLI reference](cli-reference.md) (including `exec` and `status`)
- [Caching strategy](caching.md)

## How version selection works

When you run `vcluster ...`, the shim selects a version using the following priority:

1. **Shell version** via `VCENV_VERSION` (typically set by `vc-env shell` after `eval "$(vc-env init)"`).
2. **Local version** via a `.vcluster-version` file in the current directory or any parent directory.
3. **Global version** via `$VCENV_ROOT/version`.

If no version is configured, the shim fails with an actionable error message.

## Project layout under `VCENV_ROOT`

By default, `VCENV_ROOT` is set to `~/.vcenv`.

```text
$VCENV_ROOT/
├── versions/
│   ├── <vcluster-version>/
│   │   └── vcluster
│   └── ...
├── shims/
│   └── vcluster
└── version
```

