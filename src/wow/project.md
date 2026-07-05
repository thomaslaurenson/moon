# WoW Addon Project Structure

## TOC File

Every addon requires a `.toc` file at the root of the addon folder.

```
## Interface: 11200
## Title: |cff33ffccMyAddon|r
## Author: YourName
## Notes: A short description.
## Version: 1.0.0
## OptionalDeps: pfUI
## SavedVariables: MyAddon_config
## SavedVariablesPerCharacter: MyAddon_charConfig

Locales\enUS.lua
Locales\deDE.lua
MyAddon.lua
modules\config.lua
modules\ui.lua
```

## Header Fields

| Field | Required | Notes |
|---|---|---|
| `## Interface: 11200` | yes | Always `11200` for vanilla 1.12 |
| `## Title:` | yes | Displayed in addon list; supports colour codes |
| `## Author:` | no | Displayed in addon list |
| `## Notes:` | no | Tooltip in addon list |
| `## Notes-deDE:` | no | Localised notes (deDE, frFR, ruRU, koKR, zhCN, zhTW, esES, ptBR) |
| `## Version:` | no | Accessible via `GetAddOnMetadata` |
| `## SavedVariables:` | no | Global variables saved per-account |
| `## SavedVariablesPerCharacter:` | no | Variables saved per character |
| `## OptionalDeps:` | no | Addons to load before this one when present |
| `## Dependencies:` | no | Addons that must be loaded first (hard dependency) |
| `## LoadOnDemand: 1` | no | Only load when explicitly requested |
| `## DefaultState: disabled` | no | Disabled in new installs |
| `## X-Website:` | no | Custom metadata (use `X-` prefix) |

## File Loading Rules

- Files are listed one per line, relative to the addon folder
- Path separator is backslash `\` (a WoW engine requirement even on Linux/macOS)
- Files execute in the exact order listed; load dependencies before the files that use them
- Lines beginning with `#` (single hash, not `##`) are comments

## XML Include Files

For addons with many files, XML manifests replace listing each file in the TOC:

```
# TOC entry:
init\env.xml
```

```xml
<!-- init\env.xml -->
<Ui xmlns="http://www.blizzard.com/wow/ui/">
  <Include file="..\Locales\enUS.lua"/>
  <Include file="..\Locales\deDE.lua"/>
  <Include file="..\core.lua"/>
</Ui>
```

XML paths also use backslashes. Relative paths are from the XML file's own directory.

Prefer pure Lua frame creation over XML. It avoids parsing ambiguity and is easier to maintain.

## Project Structure Patterns

### Pattern 1: Single-File Addon

```
MyAddon/
  MyAddon.toc
  MyAddon.lua
```

Suitable for simple utilities and single-purpose tools.

### Pattern 2: Small Multi-File Addon

```
MyAddon/
  MyAddon.toc
  Locales\
    Locale_enUS.lua
    Locale_deDE.lua
  Localization.lua
  data.lua
  ui.lua
  MyAddon.lua
```

### Pattern 3: Medium Addon with XML Loading

```
MyAddon/
  MyAddon.toc
  init\
    env.xml
    modules.xml
  compat\
    client.lua
  config.lua
  core.lua
  ui.lua
  slashcmd.lua
```

### Pattern 4: Large Framework

```
MyFramework/
  MyFramework.toc
  MyFramework.lua
  init\
    env.xml
    compat.xml
    api.xml
    modules.xml
  env\
    locales_enUS.lua
    tables.lua
    profiles.lua
  compat\
    vanilla.lua
  api\
    api.lua
  modules\
    unitframes.lua
    actionbar.lua
    minimap.lua
```

### Pattern 5: Split-Module TOC

Each module is its own LoadOnDemand addon with its own `.toc`:

```
DPSMate/
  DPSMate/
    DPSMate.toc
    DPSMate.lua
  DPSMate_Healing/
    DPSMate_Healing.toc
    DPSMate_Healing.lua
  DPSMate_DamageTaken/
    DPSMate_DamageTaken.toc
    DPSMate_DamageTaken.lua
```

Suitable for large combat meters where users load only the modules they need.

## Runtime Addon Introspection

`GetAddOnInfo(name)` queries another addon's status at runtime. Use it to detect optional dependencies or to read the current addon's own name:

```lua
-- returns: name, title, notes, enabled, loadable, reason, security
local name, title, notes, enabled = GetAddOnInfo("pfUI")
if enabled then
  -- pfUI is installed and enabled
end
```

`GetAddOnMetadata` is NOT available in vanilla 1.12. To read your own version at runtime, store it manually in your namespace:

```lua
-- MyAddon.lua
MyAddon.version = "1.2.0"
```

To detect the current addon's folder name dynamically (needed when the user may have renamed the folder):

```lua
-- The TOC file name must match the folder name.
-- GetAddOnInfo(n) iterates by index; self-identify by matching the title field
-- against your known title string, or simply hardcode the folder name.
```
