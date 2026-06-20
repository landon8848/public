#!/usr/bin/env python3
"""Build internal/whenin/index.json from GeoNames cities + countries.

Run from anywhere:  python3 build/build_index.py [--min-pop N]

This is the only Python left in the project: a dev-time data tool. The
runtime is Go and embeds the generated index.json via go:embed.

Only cities and countries are searchable. Timezone abbreviation, UTC
offset and IANA name are display-only metadata the runtime derives
from each entry's `iana` field — they are not their own index entries
(searching "EDT" / "America/Chicago" was pure noise in real use).

Downloads GeoNames cities15000.zip once into build/cache/ (gitignored).
Output index.json is committed so a fresh clone can build without network.

The norm() below MUST stay byte-identical in behaviour to Go's
internal/whenin.Norm, or runtime matches against the baked search_names
silently fail.
"""
import argparse
import io
import json
import os
import unicodedata
import urllib.request
import zipfile


def _strip_diacritics(s):
    nfkd = unicodedata.normalize("NFKD", s)
    return "".join(c for c in nfkd if not unicodedata.combining(c))


def norm(s):
    s = _strip_diacritics(s).casefold()
    for ch in "/_-":
        s = s.replace(ch, " ")
    return " ".join(s.split())


HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(HERE)
CACHE = os.path.join(HERE, "cache")
OUT = os.path.join(ROOT, "internal", "whenin", "index.json")
COUNTRIES = os.path.join(HERE, "countries.json")
GEONAMES_URL = "https://download.geonames.org/export/dump/cities15000.zip"


def fetch_geonames() -> bytes:
    os.makedirs(CACHE, exist_ok=True)
    cached = os.path.join(CACHE, "cities15000.zip")
    if os.path.exists(cached):
        print(f"Using cached {cached}", file=sys.stderr)
        with open(cached, "rb") as f:
            return f.read()
    print(f"Downloading {GEONAMES_URL} ...", file=sys.stderr)
    req = urllib.request.Request(GEONAMES_URL, headers={"User-Agent": "whenin-alfred-build"})
    with urllib.request.urlopen(req, timeout=60) as resp:
        data = resp.read()
    with open(cached, "wb") as f:
        f.write(data)
    print(f"Cached {len(data)} bytes -> {cached}", file=sys.stderr)
    return data


def load_countries():
    with open(COUNTRIES, encoding="utf-8") as f:
        rows = json.load(f)
    cc_to_name = {r["cc"]: r["name"] for r in rows}
    return rows, cc_to_name


def build_city_entries(zip_bytes: bytes, cc_to_name,
                        min_pop: int, us_min_pop: int):
    entries = []
    with zipfile.ZipFile(io.BytesIO(zip_bytes)) as zf:
        raw = zf.read("cities15000.txt").decode("utf-8")
    for line in raw.splitlines():
        col = line.split("\t")
        if len(col) < 18:
            continue
        name = col[1]
        ascii_name = col[2]
        cc = col[8]
        try:
            pop = int(col[14] or 0)
        except ValueError:
            pop = 0
        tz = col[17]
        # US gets a lower floor (US-heavy user base needs city-level
        # granularity); the rest of the world stays at min_pop.
        floor = us_min_pop if cc == "US" else min_pop
        if pop < floor or not tz:
            continue
        country = cc_to_name.get(cc, cc)
        # GeoNames alternatenames are unreliable (multilingual, arbitrary
        # order) and add noise + size. The matcher's diacritic-stripped
        # search_name + difflib fallback handle variants well enough.
        # Keep asciiname only if it differs (covers odd display spellings).
        aliases = []
        an = norm(ascii_name)
        if an and an != norm(name):
            aliases.append(an)
        entries.append({
            "name": name,
            "search_name": norm(name),
            "kind": "city",
            "iana": tz,
            "country": country,
            "country_code": cc or None,
            "aliases": aliases,
            "population": pop,
        })
    return entries


def build_country_entries(country_rows):
    out = []
    for r in country_rows:
        out.append({
            "name": r["name"],
            "search_name": norm(r["name"]),
            "kind": "country",
            "iana": r["iana"],
            "country": r["capital"],
            "country_code": r["cc"],
            "aliases": [norm(a) for a in r.get("aliases", [])],
            "population": 0,
        })
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--min-pop", type=int, default=100000,
                    help="Min population, rest of world (default: 100000). "
                         "Source is cities15000, so <15000 has no effect.")
    ap.add_argument("--us-min-pop", type=int, default=25000,
                    help="Min population for US cities (default: 25000). "
                         "Lower than --min-pop: the user base is US-heavy "
                         "and the US spans 6+ zones, so it needs finer "
                         "city granularity.")
    args = ap.parse_args()

    country_rows, cc_to_name = load_countries()
    zip_bytes = fetch_geonames()

    cities = build_city_entries(zip_bytes, cc_to_name,
                                args.min_pop, args.us_min_pop)
    countries = build_country_entries(country_rows)

    index = cities + countries
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(index, f, ensure_ascii=False, separators=(",", ":"))

    size = os.path.getsize(OUT)
    print(f"Wrote {len(index)} entries "
          f"(cities={len(cities)} countries={len(countries)}) "
          f"-> {OUT} ({size/1_000_000:.2f} MB)", file=sys.stderr)


if __name__ == "__main__":
    main()
