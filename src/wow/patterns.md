# WoW addon code patterns

## Namespace pattern 1: globals only

Suitable for very small single-purpose addons (under ~200 lines).

```lua
function MyAddon_UpdateBags()
  -- ...
end

MyAddonFrame = CreateFrame("Frame", "MyAddonFrame", UIParent)
MyAddonFrame:RegisterEvent("BAG_UPDATE")
MyAddonFrame:SetScript("OnEvent", function()
  MyAddon_UpdateBags()
end)
```

Prefix every global with `MyAddon_` to prevent collisions. SavedVariable names follow the same convention.

## Namespace pattern 2: namespace table

Suitable for medium addons with multiple files. Recommended default.

```lua
local MyAddon = CreateFrame("Frame", "MyAddon", UIParent)

MyAddon.version = "1.0"
MyAddon.data    = {}
MyAddon.config  = {}

function MyAddon:Init()
  self.data = {}
end

function MyAddon:UpdateDisplay()
  if self.config.showUI then
    self:Show()
  end
end

MyAddon:RegisterEvent("ADDON_LOADED")
MyAddon:SetScript("OnEvent", function()
  if event == "ADDON_LOADED" and arg1 == "MyAddon" then
    this:Init()
  end
end)
```

Inside `SetScript` callbacks use `this` (not `self`) to reference the namespace frame. Across multiple files, all files share the same namespace by referencing the global:

```lua
-- MyAddon.lua (loads first)
MyAddon = CreateFrame("Frame", "MyAddon", UIParent)
MyAddon.modules = {}

-- modules/config.lua
function MyAddon:InitConfig() ... end

-- modules/ui.lua
function MyAddon:BuildUI() ... end
```

## Namespace pattern 3: environment sandboxing

Suitable for large frameworks where many module files need to share private state. Do not use this pattern for simple addons.

```lua
-- MyFramework.lua
MyFramework = CreateFrame("Frame", "MyFramework", UIParent)
MyFramework.env = {}
setmetatable(MyFramework.env, {__index = getfenv(0)})

MyFramework.env.L  = {}
MyFramework.env.db = {}

function MyFramework:GetEnvironment()
  return MyFramework.env
end

-- modules/unitframes.lua
setfenv(1, MyFramework:GetEnvironment())

function SetupUnitFrames()
  local text = L["Player"]
  local frame = CreateFrame("Frame", nil, UIParent)
end
```

Reading a global looks up the env table first, then `_G` via `__index`. Writing a global writes to the env table, not `_G`. WoW API remains available via `__index`.

## Accessing _G

In vanilla, `_G` may not be automatically defined. Assign it explicitly when needed:

```lua
local _G = _G or getfenv(0)
local frame = _G["MyAddon"]
```

## SavedVariables

Declare in the TOC:

```
## SavedVariables: MyAddon_config, MyAddon_cache
## SavedVariablesPerCharacter: MyAddon_charConfig
```

`SavedVariables` are shared across all characters on the account. `SavedVariablesPerCharacter` stores one copy per character and realm. SavedVariables are guaranteed populated by the time `ADDON_LOADED` fires. Do not assume they are available at file parse time.

Apply defaults with an explicit `== nil` check, never `x = x or default`. For a boolean field a saved `false` is falsy, so `x = x or true` silently resets the user's choice to `true` on every login. The `== nil` form only fills a field that was never set.

```lua
local frame = CreateFrame("Frame")
frame:RegisterEvent("ADDON_LOADED")
frame:SetScript("OnEvent", function()
  if event == "ADDON_LOADED" and arg1 == "MyAddon" then
    MyAddon_config = MyAddon_config or {}
    if MyAddon_config.volume   == nil then MyAddon_config.volume   = 1.0 end
    if MyAddon_config.showUI   == nil then MyAddon_config.showUI   = true end
    if MyAddon_config.fontsize == nil then MyAddon_config.fontsize = 12  end
  end
end)
```

## Defaults pattern

Apply a default table recursively over the loaded data:

```lua
local defaults = {
  volume   = 1.0,
  showUI   = true,
  fontsize = 12,
  color    = { r = 1, g = 1, b = 1 },
}

function MyAddon:ApplyDefaults(saved, defs)
  saved = saved or {}
  for k, v in pairs(defs) do
    if saved[k] == nil then
      if type(v) == "table" then
        saved[k] = {}
        self:ApplyDefaults(saved[k], v)
      else
        saved[k] = v
      end
    end
  end
  return saved
end

-- In ADDON_LOADED handler:
MyAddon_config = MyAddon:ApplyDefaults(MyAddon_config, defaults)
```

## Per-Character variables

```lua
-- TOC: ## SavedVariablesPerCharacter: MyAddon_charConfig

frame:SetScript("OnEvent", function()
  if event == "ADDON_LOADED" and arg1 == "MyAddon" then
    MyAddon_charConfig         = MyAddon_charConfig or {}
    MyAddon_charConfig.windowX = MyAddon_charConfig.windowX or 0
    MyAddon_charConfig.windowY = MyAddon_charConfig.windowY or 0
  end
end)
```

## Saving and restoring frame position

```lua
frame:SetScript("OnMouseUp", function()
  this:StopMovingOrSizing()
  local point, _, relativePoint, x, y = this:GetPoint()
  MyAddon_charConfig.anchor = {
    point         = point,
    relativePoint = relativePoint,
    x             = x,
    y             = y,
  }
end)

function MyAddon:RestorePosition(frame, saved)
  if saved and saved.anchor then
    frame:ClearAllPoints()
    frame:SetPoint(
      saved.anchor.point,
      UIParent,
      saved.anchor.relativePoint,
      saved.anchor.x,
      saved.anchor.y
    )
  end
end
```

## SavedVariables constraints

- Only serialisable types can be saved: tables, strings, numbers, booleans
- Functions, frames, and userdata cannot be saved
- Variables set to `nil` are removed from the saved file on next logout
- `ReloadUI()` forces a re-save and reload of all addons

## Profile system

For settings that differ per character, key a profile table by character and realm:

```lua
function MyAddon:GetProfile()
  local key = GetRealmName() .. "-" .. UnitName("player")
  MyAddon_profiles       = MyAddon_profiles or {}
  MyAddon_profiles[key]  = MyAddon_profiles[key] or {}
  return MyAddon_profiles[key]
end
```

## Slash commands

Register slash commands at file parse time (not inside an event handler):

```lua
SLASH_MYADDON1 = "/myaddon"
SLASH_MYADDON2 = "/ma"

SlashCmdList["MYADDON"] = function(input)
  if not input or input == "" then
    MyAddon:ShowHelp()
    return
  end

  local parts = {}
  for word in string.gfind(input, "[^ ]+") do
    table.insert(parts, word)
  end

  local cmd  = parts[1]
  local args = ""
  for i = 2, table.getn(parts) do
    args = args .. parts[i]
    if parts[i + 1] then args = args .. " " end
  end

  if cmd == "show" then
    MyAddon:Show()
  elseif cmd == "hide" then
    MyAddon:Hide()
  else
    MyAddon:ShowHelp()
  end
end
```

Naming rules: `SLASH_<KEY><N>` must match `SlashCmdList["KEY"]` exactly. `N` starts at 1 and increments for each alias. The slash string must be lowercase.

Do not re-register reserved commands: `/reload`, `/cast`, `/use`, `/target`, `/w`, `/run`.

## Localisation

Use a global locale table populated by locale-specific files, loaded before the main addon files.

```lua
-- Localization.lua (loaded first in TOC)
MyAddon_Locale = {}

-- Locales\Locale_enUS.lua
local L = MyAddon_Locale
if GetLocale() == "enUS" or not next(L) then
  L["Settings"]          = "Settings"
  L["Reset to defaults"] = "Reset to defaults"
end

-- Locales\Locale_deDE.lua
local L = MyAddon_Locale
if GetLocale() == "deDE" then
  L["Settings"]          = "Einstellungen"
  L["Reset to defaults"] = "Auf Standard zurücksetzen"
end

-- MyAddon.lua
local L = MyAddon_Locale
frame.titleText:SetText(L["Settings"])
```

The enUS file uses `not next(L)` as a fallback so English loads if no other locale matched.

`GetLocale()` returns: `"enUS"`, `"enGB"`, `"deDE"`, `"frFR"`, `"esES"`, `"ptBR"`, `"ruRU"`, `"koKR"`, `"zhCN"`, `"zhTW"`.

## Colour escape codes

WoW uses inline colour sequences in display strings. Format: `|c` + 8-digit hex (alpha then RGB) + text + `|r` to reset.

```lua
-- alpha is always ff (fully opaque) in practice
DEFAULT_CHAT_FRAME:AddMessage("|cff33ffccMyAddon|r: loaded")

-- utility function for r/g/b values in 0-1 range
local function rgbhex(r, g, b)
  return string.format("|cff%02x%02x%02x", r * 255, g * 255, b * 255)
end

local gold   = rgbhex(1,    0.82, 0)
local red    = rgbhex(0.9,  0.1,  0.1)
local green  = rgbhex(0.1,  0.9,  0.1)

frame.text:SetText(rgbhex(1, 0.82, 0) .. "100g|r")
```

## RAID_CLASS_COLORS

The global `RAID_CLASS_COLORS` table maps class tokens to `{r, g, b, colorStr}` entries. `r`, `g`, `b` are in the 0-1 range. `colorStr` is a pre-formatted `aarrggbb` hex string (without the `|c` prefix).

```lua
local localName, classToken = UnitClass("target")
if classToken and RAID_CLASS_COLORS[classToken] then
  local c = RAID_CLASS_COLORS[classToken]
  frame.text:SetTextColor(c.r, c.g, c.b)
end
```

Valid vanilla class tokens: `"WARRIOR"`, `"PALADIN"`, `"HUNTER"`, `"ROGUE"`, `"PRIEST"`, `"SHAMAN"`, `"MAGE"`, `"WARLOCK"`, `"DRUID"`. There is no `"DEATHKNIGHT"` in 1.12.

## String utilities

Vanilla provides no `strsplit` or `table.wipe`. Implement them locally when needed.

### strsplit

```lua
local function strsplit(delimiter, subject)
  if not subject then return nil end
  local fields = {}
  local pattern = string.format("([^%s]+)", delimiter)
  string.gsub(subject, pattern, function(c)
    fields[table.getn(fields) + 1] = c
  end)
  return unpack(fields)
end
```

### wipe (empty a table in place)

After wiping, use `t[table.getn(t)+1] = v` rather than `table.insert(t, v)` because `table.insert` has undefined behaviour on tables emptied this way.

```lua
local function wipe(src)
  for k in pairs(src) do
    src[k] = nil
  end
  return src
end
```

### CopyTable (deep copy)

Required when copying SavedVariables defaults so that nested tables are not shared references.

```lua
local function CopyTable(src)
  if type(src) ~= "table" then return src end
  local copy = {}
  for k, v in pairs(src) do
    copy[CopyTable(k)] = CopyTable(v)
  end
  return setmetatable(copy, getmetatable(src))
end
```

## Money formatting

`GetMoney()` returns the player's total copper as an integer. Split it into denominations with `math.mod`:

```lua
local function FormatMoney(copper)
  local gold   = math.floor(copper / 100 / 100)
  local silver = math.floor(math.mod(copper / 100, 100))
  local cop    = math.floor(math.mod(copper, 100))
  return gold .. "g " .. silver .. "s " .. cop .. "c"
end

frame.text:SetText(FormatMoney(GetMoney()))
```

With colour codes:

```lua
local function CreateGoldString(money)
  local gold   = math.floor(money / 100 / 100)
  local silver = math.floor(math.mod(money / 100, 100))
  local copper = math.floor(math.mod(money, 100))

  local s = ""
  if gold > 0 then
    s = s .. "|cffffffff" .. gold .. "|cffffd700g"
  end
  if silver > 0 or gold > 0 then
    s = s .. "|cffffffff " .. silver .. "|cffc7c7cfs"
  end
  s = s .. "|cffffffff " .. copper .. "|cffeda55fc"
  return s
end
```

## Deferred execution (QueueFunction)

Use a FIFO queue with an `OnUpdate` driver when you need to spread initialisation work across frames rather than blocking on a single `ADDON_LOADED` handler:

```lua
local queue = {}
local runner = CreateFrame("Frame")
runner:Hide()
runner:SetScript("OnUpdate", function()
  local item = table.remove(queue, 1)
  if item then
    local func = item[1]
    func(item[2], item[3], item[4])
  end
  if table.getn(queue) == 0 then
    this:Hide()
  end
end)

local function QueueFunction(func, a, b, c)
  queue[table.getn(queue) + 1] = { func, a, b, c }
  runner:Show()
end
```

Each `OnUpdate` call processes one item. The frame hides itself when the queue empties so it does not burn CPU when idle.
