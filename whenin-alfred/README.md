# When In…

An Alfred workflow that answers *"what time is it in &lt;place&gt;?"* — fuzzy
search over cities, countries, and timezones, with the result copied
chat-ready to your clipboard. A single static binary with no runtime
dependencies.

## Install

The workflow ships **without** a binary — you compile the engine yourself with
one command. This keeps the download tiny and, because a binary you build
locally is never quarantined, sidesteps macOS Gatekeeper entirely (no "developer
cannot be verified" prompt, no `xattr` dance).

1. **Install the engine** (needs a [Go toolchain](https://go.dev/dl/)):

   ```sh
   go install github.com/landon8848/public/whenin-alfred@latest
   ```

   This drops a `whenin-alfred` binary in `$(go env GOPATH)/bin` (usually
   `~/go/bin`). You do **not** need that directory on your `PATH` — the workflow
   looks there directly.

2. **Install the workflow:** download
   [`WhenIn.alfredworkflow`](WhenIn.alfredworkflow) and double-click to import it
   into Alfred (requires the Powerpack).

That's it. The workflow's script filter finds the engine at runtime, checking (in
order): a binary bundled next to the workflow, `~/go/bin/whenin-alfred`, your
`PATH`, and the Homebrew locations. If it can't find one, the first result tells
you to run the `go install` command above.

To **update**, re-run the `go install` line; `@latest` pulls the newest engine.

## Usage

Type the keyword `when`, then one of:

- **`when in <place>`** — current time in a city or country, fuzzy-matched.
  A bare `when <place>` does the same.
- **`when tz <zone>`** — a timezone code (`PST`, `JST`, `IST`, …) or an IANA
  name (`Europe/Berlin`).
- **`when is <country>`** — public holidays *(coming soon)*.

Press <kbd>Enter</kbd> on a result to copy a chat-ready line like
`2:32 PM Fri in Rome (CEST, UTC+2)`.

The clock format (12h / 24h) is set in the workflow's configuration.

## How it works

- Bundles a prebuilt index of ~25k cities (GeoNames `cities15000`) plus
  countries. Timezone abbreviation, UTC offset, and IANA name are derived at
  runtime from each entry's zone — they aren't searchable noise in the index.
- Greedy fuzzy matcher: exact → prefix → substring → typo-correction, weighted
  by population, so the big, likely city wins.
- Compiled to a static binary with the index and the IANA timezone database
  embedded. It starts in a few milliseconds, which matters because Alfred
  re-runs the Script Filter on every keystroke.

## Build from source

The engine is an ordinary Go program — `go install` (see [Install](#install)) is
the supported build, and `go build -o whenin .` produces a local binary if you
want one.

`build/package.sh` just (re)zips the workflow bundle (`info.plist` + icon) into
`WhenIn.alfredworkflow`. It no longer compiles or embeds a binary, so it needs
neither a cross-compile nor `lipo`:

```sh
sh build/package.sh   # repackage WhenIn.alfredworkflow (no binary)
```

### Rebuilding the index

`build/build_index.py` regenerates `internal/whenin/index.json` from GeoNames
(downloaded once into `build/cache/`):

```sh
python3 build/build_index.py            # defaults: --min-pop 100000, --us-min-pop 25000
```

The US floor is lower because US cities span many zones and benefit from finer
granularity. The `norm()` in that script must stay behaviourally identical to
`internal/whenin.Norm`, or runtime matches against the baked search names fail.

### Icon

`sh build/make_icon.sh` re-renders `workflow/icon.png` from the source SVG
(requires `librsvg`).

## Troubleshooting

- **Results say "When In… engine not installed"** — the engine binary isn't where
  the workflow looks. Run `go install github.com/landon8848/public/whenin-alfred@latest`
  and confirm it landed in `~/go/bin` (or `$(go env GOPATH)/bin`).
- **`./whenin: No such file or directory`** — you're running an older build that
  expects a bundled binary, but none is present. Reinstall using the two steps in
  [Install](#install).
- **"…cannot be opened because the developer cannot be verified" / killed** —
  Gatekeeper quarantined a *downloaded* binary. The current install path avoids
  this by building locally. If you have a stale downloaded copy, clear it with
  `xattr -dr com.apple.quarantine <path>`.

## License

GPL — see the repository root.
