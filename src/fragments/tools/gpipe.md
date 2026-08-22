# gpipe

`thomaslaurenson/gpipe` turns built binaries into a publishable release: install scripts, a checksum file, and optionally a cosign signing bundle. It is language-agnostic. It consumes binaries from the paths named in `.gpipe.yml` and neither knows nor cares what produced them, so the same configuration serves a goreleaser build, a CMake build, or anything else that leaves files on disk.

## Where it sits

A release is three steps, in order:

1. **Build.** Whatever the language uses produces binaries into `dist/`.
2. **gpipe.** Generates the install scripts, checksums and signing bundle from those binaries.
3. **Publish.** `gh release create` attaches the binaries and the generated files.

gpipe neither builds nor publishes. Keeping the three separate is what lets the build step change language without the release step changing at all.

## What it writes

gpipe writes into the repository root, not `dist/`:

| File | Written |
|---|---|
| `install.sh` | always |
| `install.ps1` | always |
| `checksums.txt` | always |
| `checksums.txt.sigstore.json` | only with `cosign_sign: true` |

All of them are build output, so all of them belong in the `clean` target alongside `dist/`. The signing bundle is the one most often missed, because it appears only once signing is switched on and its name does not match a `checksums.txt` entry. List the files rather than reaching for a `checksums.txt*` glob: a glob quietly widens as gpipe gains outputs, and `clean` should only ever remove what this project can rebuild.

Do not hand-roll `sha256sum` alongside it. gpipe's `checksums.txt` covers the installer scripts themselves as well as the platform binaries, which is what lets a cautious user verify `install.sh` before piping it to a shell.

## .gpipe.yml

Lives at the project root. `binary`, `platforms` and `hooks` are the whole config surface.

```yaml
binary: myapp

platforms:
  linux_amd64:
    path: ./dist/myapp-linux-x86_64
    name: myapp-linux-x86_64
  linux_arm64:
    path: ./dist/myapp-linux-aarch64
    name: myapp-linux-aarch64
  darwin_amd64:
    path: ./dist/myapp-darwin-x86_64
    name: myapp-darwin-x86_64
  darwin_arm64:
    path: ./dist/myapp-darwin-aarch64
    name: myapp-darwin-aarch64
  windows_amd64:
    path: ./dist/myapp-windows-x86_64.exe
    name: myapp-windows-x86_64.exe
```

- Every `path` must match where the build step actually left the binary. These platform keys are gpipe's vocabulary, so the build has to name its output to suit; the language fragment says how.
- One platform key maps to exactly one binary. There is no way to express "either of these builds, reader's choice", so a platform that could ship more than one flavour has to pick the one that runs everywhere.
- Shell completions are not gpipe's job, and a `post-sh` hook is not the place for them either. An installer that writes outside the install directory does more than the user asked for: it has to guess the shell, pick between several completion directories, and leaves nothing behind to undo it. The binary prints its own script (`<binary> completion bash`) and installing it is the user's call. The hook is for work that belongs to the install itself, and it has the location in `INSTALL_DIR` and the name in `BINARY`.

Validate the config before relying on it in CI. `gpipe validate --repo <owner/repo> --version v0.0.0` checks the schema, the platform identifiers and any hooks, without needing the binaries present.

Add `.gpipe.yml` to the `pr.yml` and `main.yml` paths filters: a change to it is a change to the release.

## The action

```yaml
- uses: thomaslaurenson/gpipe@vN
  with:
    cosign_sign: true
```

- `version` and `repo` default to `github.ref_name` and `github.repository`. A `tag.yml` that fires only on `v*` therefore needs neither, since `ref_name` is already the semantic version gpipe expects.
- The action builds gpipe from its own checkout, so the ref pinned in `uses:` is the gpipe that runs and there is no separate version input to keep current.
- It installs its own Go from its own `go.mod`. A caller needs no `actions/setup-go` for its benefit: the runner image ships Go, but not necessarily a version new enough, and a project that never invokes Go directly has no reason to carry the setup step for a tool.
- `cosign_sign: true` needs `id-token: write` on this job **and** on the caller job in `tag.yml`. Granting it in only one of the two fails at signing time.

## Not for prereleases

gpipe cannot run on a rolling prerelease channel, and that is a property of the tool rather than a preference. It validates `--version` as a semantic version, so the literal string `dev` is rejected outright. A semver-shaped stand-in such as `v1.2.4-dev` is worse: the installers it generates hardcode `releases/download/<version>/<asset>`, so every download URL in the script 404s against a release actually tagged `dev`.

Install scripts are a release-only artefact. A rolling channel deleted and recreated on every push to main is the wrong thing to hang a stable `curl | bash` URL off in any case, so a prerelease publishes raw binaries and nothing else.
