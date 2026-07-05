# Go Tooling

No third-party linters or formatters. Do not use `golangci-lint` or `govulncheck` under any circumstances.

| Tool | Purpose |
|---|---|
| `gofmt` | Format source files |
| `go vet` | Static analysis |
| `go test` | Run tests |
| `go mod tidy` | Keep go.mod/go.sum clean |

- Use the latest stable Go version. No `replace` directives in committed code. Run `go mod tidy` before committing.
- Third-party release tools (`goreleaser`, `cosign`) are permitted for projects that ship a compiled binary; a pure library needs neither, since it has nothing to build a release artifact from.
