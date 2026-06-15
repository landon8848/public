# kc

A 1Password-backed kubeconfig manager: kubeconfigs live as 1Password Documents
and are materialised on demand, never stored cleartext in `~/.kube/`. Single Go
binary, macOS + Linux, zsh/bash/fish.

## Requirements

- **[1Password CLI (`op`)](https://developer.1password.com/docs/cli/)** —
  installed and signed in. This is the only external runtime dependency.
- **A 1Password vault named `Private`** — `kc add` stores documents there (this
  is the default vault on personal accounts). The target vault is not yet
  configurable.
- **Shell hook** — `set` (and `remove`/`rm`) update your shell's `KUBECONFIG`, so
  they need the `kc` shell wrapper installed (see [Setup](#setup)). `run` and
  `add`/`list`/`show` work without it.
- `kubectl` — only for `kc run`.

## Install

### `go install` (needs a Go toolchain)

```sh
go install github.com/landon8848/public/kc@latest
```

The binary lands in `$(go env GOPATH)/bin/kc` — make sure that's on your `PATH`.

### Build from source

```sh
git clone https://github.com/landon8848/public
cd public/kc
go build -o kc .
```

## Setup

Add the shell hook to your rc file so `kc set` can update the current shell. Pick
your shell:

```sh
# ~/.zshrc
eval "$(kc shell-init zsh)"

# ~/.bashrc
eval "$(kc shell-init bash)"

# ~/.config/fish/config.fish
kc shell-init fish | source
```

Open a new shell (or re-source the rc file) afterward.

## Usage

```sh
kc add -n prod-eu ~/.kube/config  # register a kubeconfig under a name (stores it in 1Password)
kc add ~/.kube/config             # no -n → interactive form prompts for the name
kc add -f - -n prod-eu            # read kubeconfig from stdin
kc list                           # list registered configs (relative dates)
kc list -v                        # verbose: 1Password ref + absolute timestamps
kc show                           # show the active config (-v for detail)
kc set                            # interactive picker → activates in this shell
kc set -n prod-eu                 # activate a named config in this shell
kc run -n prod-eu -- get pods     # one-shot kubectl; nothing touches disk
kc edit -n prod-eu                # edit a config in $EDITOR (or replace from file/stdin)
kc remove -n prod-eu              # delete a config from 1Password and the registry
```

`-V` prints the version; `-v`/`--verbose` is available on `list`, `show`, `set`,
and `run`.

## How it works

Kubeconfigs are stored as 1Password Documents in your Private vault (tagged
`kc`). A small TOML registry maps friendly names → document refs. `kc set` writes
the config to a `0400` tempfile and exports `KUBECONFIG` to it (cleaned on shell
exit); `kc run` skips disk entirely, piping the config to `kubectl` over a file
descriptor (`KUBECONFIG=/dev/fd/3`).
