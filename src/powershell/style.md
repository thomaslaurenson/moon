# PowerShell Style

PowerShell-specific style for `.ps1` scripts and sourced files. Assumes the core conventions.

## Script header

Open with the generated-file header (if applicable), then `#Requires`, then preferences:

```powershell
#Requires -Version 5.1

$ErrorActionPreference = 'Stop'
$ProgressPreference    = 'SilentlyContinue'
```

`#Requires -Version 5.1` refuses older hosts. `$ErrorActionPreference = 'Stop'` makes terminating errors propagate. `$ProgressPreference = 'SilentlyContinue'` suppresses the `Invoke-WebRequest` progress bar, greatly speeding non-interactive downloads.

## Formatting

- Indent with 4 spaces, never tabs. Max line length 120.
- Split long commands with a backtick continuation or use a here-string (`@' ... '@`).
- Align related assignments in a block using spaces.

## Naming

- Functions: `Verb-Noun` PascalCase from the approved verb list. Private output helpers may be verb-only (`Write-Info`, `Exit-Error`).
- Script-scope (module-level) variables: `$script:PascalCase`. Locals: `$camelCase`. Parameters: `PascalCase` in `param()`.
- No `readonly` equivalent exists; treat `$script:` variables set once at the top as constants by convention.

## Function comments

Any function that is not both obvious and short needs a header comment; all functions in a sourced script need one regardless of length. Describe behaviour, then labelled sections for parameters, globals read, outputs, and errors (omit those that do not apply). Use `Parameters:` (not `Arguments:`) to mirror `param()`.

```powershell
# Detect the current Windows architecture and return a platform identifier.
#
# Outputs:
#   string: platform identifier, e.g. windows_amd64 or windows_arm64
# Throws:
#   terminating error for unsupported architectures
function Get-Platform { ... }
```

## Error handling and output helpers

Define an `Exit-Error` helper (prints a red message, exits 1) and use it rather than inline `Write-Host ... ; exit 1`. Installer scripts also define `Write-Info` (green `[INFO]`), `Write-Warn` (yellow `[WARN]`), and `Write-Step` (cyan `>`). Respect `$env:NO_COLOR`: `$script:NoColor = ($null -ne $env:NO_COLOR)`.

## Entry point

Wrap the entry point in a named function (`Invoke-Installer` by convention) and call it via a guard that prevents execution when dot-sourced for testing:

```powershell
if ($MyInvocation.InvocationName -ne '.') {
    Invoke-Installer @PSBoundParameters
}
```
