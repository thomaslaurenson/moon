# WoW Vanilla Event System

## Core Pattern

```lua
local frame = CreateFrame("Frame", "MyAddonFrame", UIParent)

frame:RegisterEvent("PLAYER_LOGIN")
frame:RegisterEvent("UNIT_HEALTH")

frame:SetScript("OnEvent", function()
  if event == "PLAYER_LOGIN" then
    DEFAULT_CHAT_FRAME:AddMessage("Hello, " .. UnitName("player") .. "!")
  elseif event == "UNIT_HEALTH" and arg1 == "player" then
    -- player health changed
  end
end)
```

Always guard with `event ==` before checking `arg1`. Checking `arg1` alone is unreliable when multiple events are registered on the same frame.

## Addon Lifecycle Events

| Event | `arg1` | When |
|---|---|---|
| `VARIABLES_LOADED` | -- | SavedVariables loaded from disk |
| `ADDON_LOADED` | addon name | A specific addon finished loading |
| `PLAYER_LOGIN` | -- | Player enters the world for the first time |
| `PLAYER_ENTERING_WORLD` | -- | Player loads into any zone or instance |
| `PLAYER_LOGOUT` | -- | Player is logging out |

Use `ADDON_LOADED` with a name check for initialisation — SavedVariables are guaranteed available:

```lua
frame:RegisterEvent("ADDON_LOADED")
frame:SetScript("OnEvent", function()
  if event == "ADDON_LOADED" and arg1 == "MyAddon" then
    MyAddon:Init()
  end
end)
```

## Player and Unit Events

| Event | Key args | Notes |
|---|---|---|
| `PLAYER_LEVEL_UP` | `arg1` = new level | |
| `UNIT_HEALTH` | `arg1` = unitID | Fires for any tracked unit |
| `UNIT_MANA` | `arg1` = unitID | Power change (mana/rage/energy) |
| `UNIT_MAXHEALTH` | `arg1` = unitID | |
| `PLAYER_TARGET_CHANGED` | -- | Use `UnitExists("target")` to check |
| `PLAYER_ENTER_COMBAT` | -- | |
| `PLAYER_LEAVE_COMBAT` | -- | |
| `UNIT_FLAGS` | `arg1` = unitID | Faction, PvP status changed |

## Combat Log Events

Vanilla uses `CHAT_MSG_*` events for all combat information. Each event delivers a pre-formatted message string in `arg1` that must be parsed with `string.find`. There is no `COMBAT_LOG_EVENT_UNFILTERED` in vanilla.

| Event | Message example |
|---|---|
| `CHAT_MSG_COMBAT_SELF_HITS` | `"You hit Ragnaros for 234 damage."` |
| `CHAT_MSG_COMBAT_SELF_MISSES` | `"You miss Ragnaros."` |
| `CHAT_MSG_SPELL_SELF_DAMAGE` | `"Your Fireball hits Ragnaros for 1234 Fire damage."` |
| `CHAT_MSG_SPELL_PERIODIC_SELF_DAMAGE` | DoT tick messages |
| `CHAT_MSG_COMBAT_CREATURE_VS_SELF_HITS` | `"Ragnaros hits you for 2345 damage."` |

## Bag and UI Events

| Event | Key args | Notes |
|---|---|---|
| `BAG_UPDATE` | `arg1` = bagID | A bag's contents changed |
| `BANKFRAME_OPENED` | -- | |
| `ITEM_LOCK_CHANGED` | -- | |
| `MINIMAP_ZONE_CHANGED` | -- | |
| `WORLD_MAP_UPDATE` | -- | |

## Group Events

| Event | Notes |
|---|---|
| `PARTY_MEMBERS_CHANGED` | Party size or members changed |
| `RAID_ROSTER_UPDATE` | Raid roster changed |
| `UNIT_PET` | A unit's pet changed |

## Chat Events

| Event | `arg1` | `arg2` | Notes |
|---|---|---|---|
| `CHAT_MSG_SAY` | message | sender | |
| `CHAT_MSG_YELL` | message | sender | |
| `CHAT_MSG_WHISPER` | message | sender | |
| `CHAT_MSG_PARTY` | message | sender | |
| `CHAT_MSG_RAID` | message | sender | |
| `CHAT_MSG_GUILD` | message | sender | |
| `CHAT_MSG_ADDON` | message | sender | `arg3` = channel prefix |

## OnUpdate

`OnUpdate` fires every rendered frame. `arg1` is elapsed seconds since the last call. Always throttle:

```lua
frame:SetScript("OnUpdate", function()
  this.elapsed = (this.elapsed or 0) + arg1
  if this.elapsed < 0.5 then return end
  this.elapsed = 0
  MyAddon:UpdateDisplay()
end)
```

## Script Handlers

| Handler | `arg1` | Notes |
|---|---|---|
| `OnShow` | -- | Frame became visible |
| `OnHide` | -- | Frame was hidden |
| `OnClick` | button name | `"LeftButton"`, `"RightButton"` |
| `OnMouseDown` | button name | |
| `OnMouseUp` | button name | |
| `OnMouseWheel` | 1 or -1 | Scroll direction |
| `OnEnter` | -- | Mouse entered frame |
| `OnLeave` | -- | Mouse left frame |
| `OnValueChanged` | new value | Slider or StatusBar |
| `OnTextChanged` | -- | EditBox text changed |
| `OnEscapePressed` | -- | Escape key in EditBox |
| `OnDragStart` | button | |
| `OnDragStop` | -- | |
| `OnReceiveDrag` | -- | Item dropped on frame |

## Registering and Unregistering

```lua
frame:UnregisterEvent("UNIT_HEALTH")
frame:UnregisterAllEvents()
frame:RegisterEvent("UNIT_HEALTH")
```

## Addon Messaging (Cross-Client)

Use `SendAddonMessage` for network communication between clients. There is no custom event dispatcher for within-client addon communication — call functions directly.

```lua
local PREFIX = "MyAddon"

local comm = CreateFrame("Frame")
comm:RegisterEvent("CHAT_MSG_ADDON")
comm:SetScript("OnEvent", function()
  if event == "CHAT_MSG_ADDON" and arg1 == PREFIX then
    local sender  = arg4
    local message = arg2
    local _, _, key, value = string.find(message, "^([^:]+):(.+)")
    if key == "VERSION" then
      MyAddon:HandleVersion(sender, value)
    end
  end
end)

local function Broadcast(key, value)
  SendAddonMessage(PREFIX, key .. ":" .. tostring(value), "RAID")
end
```

`SendAddonMessage(prefix, message, channel)` — `channel` is `"PARTY"`, `"RAID"`, `"GUILD"`, `"BATTLEGROUND"`, or `"WHISPER"`. Max message length is 255 bytes. Messages are received by all clients in the channel including the sender. Throttle to at most once per 0.5 s.

## Buff and Debuff API

Vanilla uses a slot-based buff API, not the 2.0+ `UnitAura`. `GetPlayerBuff` only works for the player; use `UnitBuff`/`UnitDebuff` for any unit.

```lua
-- player buffs: iterate slots 0-31
for i = 0, 31 do
  local buffIndex = GetPlayerBuff(i, "HELPFUL")
  if buffIndex < 0 then break end   -- -1 means empty slot
  local texture  = GetPlayerBuffTexture(buffIndex)
  local timeleft = GetPlayerBuffTimeLeft(buffIndex)
  local count    = GetPlayerBuffApplications(buffIndex)
end

-- any unit buffs (target, party1, etc.)
-- returns: texture, count, debuffType, duration, timeleft
local texture, count = UnitBuff(unit, index)
local texture, count = UnitDebuff(unit, index)
```

`GetPlayerBuff(slotIndex, filter)` accepts `"HELPFUL"` (buffs) or `"HARMFUL"` (debuffs). Slots are not necessarily contiguous; iterate all 32 and stop when the return is `< 0`.

## Unit Queries

`UnitClass(unit)` returns two values. Use the second (the class token) for colour lookups, not the first (localised name):

```lua
local localName, classToken = UnitClass("player")
-- classToken: "WARRIOR", "PALADIN", "HUNTER", "ROGUE", "PRIEST",
--             "SHAMAN", "MAGE", "WARLOCK", "DRUID"
```
