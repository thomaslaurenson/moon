# WoW addon project structure

## TOC file

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

## Header fields

| Field | Required | Notes |
|---|---|---|
| `## Interface: 11200` | yes | Always `11200` for vanilla 1.12 |
| `## Title:` | yes | Displayed in addon list; supports colour codes |
| `## Author:` | no | Displayed in addon list |
| `## Notes:` | no | Tooltip in addon list |
| `## Notes-deDE:` | no | Localised notes (deDE, frFR, ruRU, koKR, zhCN, zhTW, esES, ptBR) |
| `## Version:` | no | Metadata only; `GetAddOnMetadata` does not exist in 1.12, so store the version in your namespace (see Runtime Addon Introspection) |
| `## SavedVariables:` | no | Global variables saved per-account |
| `## SavedVariablesPerCharacter:` | no | Variables saved per character |
| `## OptionalDeps:` | no | Addons to load before this one when present |
| `## Dependencies:` | no | Addons that must be loaded first (hard dependency) |
| `## LoadOnDemand: 1` | no | Only load when explicitly requested |
| `## DefaultState: disabled` | no | Disabled in new installs |
| `## X-Website:` | no | Custom metadata (use `X-` prefix) |

## File loading rules

- Files are listed one per line, relative to the addon folder
- Path separator is backslash `\` (a WoW engine requirement even on Linux/macOS)
- Files execute in the exact order listed; load dependencies before the files that use them
- Lines beginning with `#` (single hash, not `##`) are comments

## XML include files

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

## Project structure patterns

### Pattern 1: single-file addon

```
MyAddon/
  MyAddon.toc
  MyAddon.lua
```

Suitable for simple utilities and single-purpose tools.

### Pattern 2: small multi-file addon

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

### Pattern 3: medium addon with XML loading

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

### Pattern 4: large framework

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

### Pattern 5: split-module TOC

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

## Runtime addon introspection

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

Do not try to detect your own folder name dynamically. A TOC file's name must match its folder name, so the name is fixed at author time: hardcode it as a constant in your namespace. The `ADDON_LOADED` handler already receives the loading addon's name in `arg1`, which is the reliable way to know when your own addon has finished loading.
