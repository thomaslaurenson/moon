# WoW 1.12 Vanilla Lua Environment

The vanilla 1.12 client embeds Lua 5.0. Many Lua features an LLM produces by default are 5.1+ and will not work.

## Non-Negotiable Rules

- `## Interface: 11200` is the only valid interface version
- Path separators in TOC and XML files must be backslash `\` (written `\\` in Lua strings)
- `frame:SetScript("OnEvent", function()` takes **zero parameters**. Use globals `this`, `event`, `arg1`...`arg9`
- Never write `function(self, event, ...)` for event handlers (WoW 2.0+ API only)
- Do not use `os.*` or `io.*` (removed from the WoW sandbox)
- Do not use `_ENV`. Use `getfenv()` / `setfenv()`
- Vanilla has no `hooksecurefunc`; use a manual save-and-replace hook pattern instead
- Do not use `InCombatLockdown()` (the secure frame system does not exist in vanilla)
- Do not use `C_Timer.*`. Use `OnUpdate` with an elapsed counter instead
- Do not use `COMBAT_LOG_EVENT_UNFILTERED`. Use `CHAT_MSG_*` events and parse the message string
- Do not use `UnitGUID` (only available with the SuperWoW server mod; guard with `if UnitGUID then`)
- Do not use `GetItemInfoInstant`. Use `GetItemInfo` with a link
- Do not use `RegisterUnitWatch` or secure frame templates (`SecureActionButtonTemplate`, etc.)

## Lua 5.0 Incompatibilities

| Correct (Lua 5.0) | Wrong (Lua 5.1+) |
|---|---|
| `table.getn(t)` | `#t` |
| `string.gfind(s, p)` | `string.gmatch` |
| `unpack(t)` | `table.unpack` |
| `math.mod(a, b)` or `mod(a, b)` | `math.fmod` or `a % b` |
| `1/0` | `math.huge` |
| `gcinfo()` | `collectgarbage("count")` |
| `string.find` with captures | `string.match` |
| `getfenv()` / `setfenv()` | `_ENV` |

No `%` modulo operator. Use `math.mod(a, b)`. No `table.move`, `table.pack`, or `table.unpack`.

```lua
local count     = table.getn(myTable)
local remainder = math.mod(10, 3)
local infinity  = 1/0
local curMem, maxMem = gcinfo()

for word in string.gfind(str, "[^ ]+") do ... end

local _, _, cap = string.find(s, "(pattern)")

return unpack(myTable)
```

## Compatibility Polyfills

When writing code that must run on both 1.12 and later private server clients, declare these at the top of the first file:

```lua
gfind = string.gmatch or string.gfind
mod   = math.mod or mod
```

This lets the rest of the file use `gfind` and `mod` without per-call version checks. Do not add these to vanilla-only addons where `string.gfind` and `math.mod` always exist.

## Event Callback Globals

Inside every `SetScript` callback these implicit globals are set by the engine:

| Global | Value |
|---|---|
| `this` | The frame the script is attached to |
| `event` | The event name string (OnEvent only) |
| `arg1` to `arg9` | Event-specific arguments |

```lua
-- Correct: zero-parameter handler, globals for args
frame:SetScript("OnEvent", function()
  if event == "UNIT_HEALTH" and arg1 == "player" then
    -- arg1 is the unitID
  end
end)

-- Wrong: WoW 2.0+ signature
frame:SetScript("OnEvent", function(self, event, ...)
end)
```

`OnUpdate` uses `arg1` for elapsed seconds since the last frame:

```lua
frame:SetScript("OnUpdate", function()
  this.elapsed = (this.elapsed or 0) + arg1
  if this.elapsed < 0.5 then return end
  this.elapsed = 0
  -- runs every 0.5 s
end)
```

## WoW Globals

These WoW-provided globals replace removed stdlib functions:

| WoW global | Replaces |
|---|---|
| `GetTime()` | `os.time()`, float seconds since server start |
| `date("%H:%M:%S")` | `os.date()`, strftime-style |
| `strtrim(s)` | no Lua 5.0 equivalent |

## Client Version Detection

`GetBuildInfo()` returns the build version as a number. Use it to branch between vanilla and later clients when the same addon targets multiple server types:

```lua
local _, _, _, client = GetBuildInfo()
client = client or 11200

if client < 20000 then
  -- vanilla 1.12 client
elseif client < 30000 then
  -- TBC client
end
```

The fourth return value is the full build number (`11200` for patch 1.12). Only use this when the addon must support multiple private server versions; do not add it to vanilla-only addons.

## Performance: Local Caching

Cache frequently called globals as locals at the top of files used in hot paths (OnUpdate, combat events):

```lua
local _pairs  = pairs
local _floor  = math.floor
local _insert = table.insert
local _sort   = table.sort
local _find   = string.find
local _format = string.format
local _getn   = table.getn
```

## Manual Function Hooks

Vanilla has no `hooksecurefunc`. To wrap a function defined elsewhere, save the original and replace it:

```lua
local orig_GameTooltip_SetUnit = GameTooltip.SetUnit
GameTooltip.SetUnit = function(self, unit)
  orig_GameTooltip_SetUnit(self, unit)
  MyAddon:OnTooltipSetUnit(unit)
end
```

For global functions:

```lua
local orig_UseAction = UseAction
UseAction = function(slot, onSelf, onPet)
  orig_UseAction(slot, onSelf, onPet)
  MyAddon:OnActionUsed(slot)
end
```

Store originals in a local or namespaced table, not a global, to avoid collision with other addons hooking the same function.

