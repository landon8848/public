# When In…

An Alfred workflow that answers *"what time is it in &lt;place&gt;?"* — fuzzy
search over cities, countries, and timezones, with the result copied
chat-ready to your clipboard. A single static binary with no runtime
dependencies.

## Install

Download [`WhenIn.alfredworkflow`](WhenIn.alfredworkflow) and double-click to
install it in Alfred (requires the Powerpack). The bundled binary is universal
(Apple Silicon + Intel); nothing else to install.

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

Requires Go (see `go.mod`) and `lipo` (ships with the Xcode command-line tools)
for the universal build.

```sh
sh build/package.sh   # build the universal binary and repackage WhenIn.alfredworkflow
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

## License

GPL — see the repository root.
