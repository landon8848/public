# bash Cal — Alfred Workflow

A tiny [Alfred](https://www.alfredapp.com/) workflow that shows a three-month
calendar in Alfred's Large Type overlay.

## What it does

Type `cal` into Alfred and it runs `cal -h -3`, displaying the previous,
current, and next month (`-3`) without highlighting today (`-h`). The output is
rendered in Large Type using a monospace font so the columns line up.

## Install

1. Download `bash-cal.alfredworkflow`.
2. Double-click it to import into Alfred (requires the [Powerpack](https://www.alfredapp.com/powerpack/)).

## Usage

| Keyword | Result |
| ------- | ------ |
| `cal`   | Shows a 3-month calendar in Large Type |

## How it works

`Keyword (cal)` → `Run Script (cal -h -3)` → `Large Type ({query})`

The Run Script action's stdout is passed downstream as `{query}`, which the
Large Type node renders on screen.
