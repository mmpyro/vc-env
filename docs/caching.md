# Caching strategy in `vc-env`

`vc-env` uses a three-layer caching strategy to manage the list of
available `vcluster` versions. This ensures that commands like
`list-remote` and `latest` remain fast, reliable, and respectful of
GitHub API rate limits.

```mermaid
flowchart TD
    cmd(["list-remote / latest"]) --> fresh{"disk cache<br/>younger than TTL?"}
    fresh -- yes --> serve(["serve from disk cache"])
    fresh -- no --> delta["delta fetch:<br/>only releases newer<br/>than the newest known"]
    delta -- ok --> merge["merge + dedupe + sort<br/>atomic write to disk"]
    merge --> serve
    delta -- network error --> stale{"stale cache<br/>on disk?"}
    stale -- yes --> warnStale(["warn + serve stale"])
    stale -- no --> baseline(["fall back to<br/>hardcoded baseline"])
```

## 1. The three layers

### Layer 1: Hardcoded baseline
A list of historically known stable and pre-release versions is baked directly into the `vc-env` binary (see `internal/cache/baseline.go`).

*   **Zero latency**: Provides a useful starting point even on the very first run.
*   **Offline fallback**: Acts as the ultimate fallback if both the disk cache and the network are unavailable.
*   **Anchor point**: The newest version in this list is used as the "anchor" for the first delta fetch.

### Layer 2: Disk cache
When `VCENV_ROOT` is set, `vc-env` persists the merged list of versions to a JSON file at:
`$VCENV_ROOT/cache/releases.json`

*   **Freshness**: If the cache file is younger than the TTL (Time To Live), it is served immediately without any network calls.
*   **Atomicity**: Writes use a "write-to-temp then rename" pattern to ensure that concurrent processes never read a partially written file.

### Layer 3: Delta fetch
If the disk cache is stale (older than TTL) or missing, `vc-env` performs a "delta fetch" from the GitHub API.

*   **Efficiency**: Instead of fetching all history, it only requests releases newer than the most recent version found in the stale cache (or the baseline).
*   **Auto-merge**: New releases are automatically merged with the existing known versions, deduplicated, and sorted.
*   **Graceful degradation**: If the network is unavailable during a delta fetch, `vc-env` will print a warning and fall back to the stale cache or the hardcoded baseline.

A delta fetch, from the client's point of view, looks like this:

```mermaid
sequenceDiagram
    autonumber
    participant U as user
    participant V as vc-env
    participant D as disk cache
    participant G as GitHub API
    U->>V: vc-env list-remote
    V->>D: read releases.json
    alt fresh (younger than TTL)
        D-->>V: cached list
        V-->>U: print list
    else stale or missing
        D-->>V: stale list (or baseline)
        V->>G: GET /releases?per_page=…&since=<anchor>
        alt network ok
            G-->>V: new releases
            V->>V: merge + dedupe + sort
            V->>D: atomic write releases.json
            V-->>U: print list
        else network error
            V-->>U: warn + print stale/baseline list
        end
    end
```

---

## 2. Configuration

You can customize the caching behavior using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `VCENV_ROOT` | Root directory for `vc-env`. If not set, caching is memory-only (no disk persistence). | N/A |
| `VCENV_CACHE_TTL` | How long a cache entry is considered fresh. Supports Go duration strings (e.g., `1h`, `30m`, `24h`, `0s`). | `1h` |

!!! tip "Disabling the cache"
    To force a fresh fetch every time, set the TTL to zero:

    ```bash
    export VCENV_CACHE_TTL=0s
    ```

    This still writes to disk after a successful fetch; it just never
    treats the on-disk copy as fresh.

---

## 3. Storage format

The cache file (`releases.json`) stores:

*   `fetched_at`: UTC timestamp of the last successful fetch.
*   `versions`: List of stable versions (newest-first).
*   `prerelease_versions`: List of all versions including pre-releases (newest-first).

---

## 4. Maintenance

The hardcoded baseline should be updated periodically (e.g., when releasing a new version of `vc-env`) to keep the "lower bound" reasonably close to the current state of the world. However, the system is designed to correct itself automatically via delta fetches even if the baseline is significantly out of date.
