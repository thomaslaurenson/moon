# WoW Vanilla Frame API

## CreateFrame

```lua
local frame = CreateFrame(frameType, frameName, parent)
```

`frameName` is optional — pass `nil` for anonymous frames. Named frames are accessible as globals via `_G["name"]`. Use `UIParent` as the parent for top-level frames.

## Frame Types

| Type | Purpose |
|---|---|
| `"Frame"` | Base container; no built-in interactivity |
| `"Button"` | Clickable frame; has `OnClick`, `OnMouseDown`, `OnMouseUp` |
| `"StatusBar"` | Fill bar (health bars, cast bars, XP bars) |
| `"Slider"` | Draggable slider handle |
| `"EditBox"` | Text input |
| `"ScrollFrame"` | Scrollable content container |
| `"GameTooltip"` | Tooltip popup |
| `"Model"` | 3D model display; also the correct type for cooldown frames in vanilla |
| `"MessageFrame"` | Scrolling message area |

`"Cooldown"` is not a valid vanilla `CreateFrame` type. Use `"Model"` instead.

## SetPoint (Anchors)

```lua
frame:SetPoint(point, relativeTo, relativePoint, offsetX, offsetY)
```

| Anchor | Position |
|---|---|
| `"CENTER"` | Centre |
| `"TOP"` | Top centre |
| `"BOTTOM"` | Bottom centre |
| `"LEFT"` | Left centre |
| `"RIGHT"` | Right centre |
| `"TOPLEFT"` | Top-left corner |
| `"TOPRIGHT"` | Top-right corner |
| `"BOTTOMLEFT"` | Bottom-left corner |
| `"BOTTOMRIGHT"` | Bottom-right corner |

```lua
frame:SetPoint("CENTER", UIParent, "CENTER", 0, 0)
frame:SetPoint("TOPLEFT", parent, "BOTTOMRIGHT", 5, -10)
```

A frame can have up to two anchors simultaneously.

## Size

```lua
frame:SetWidth(200)
frame:SetHeight(100)
```

`SetSize` is not available in all vanilla builds — prefer `SetWidth`/`SetHeight`.

## Visibility

```lua
frame:Show()
frame:Hide()
frame:IsShown()     -- own visibility only
frame:IsVisible()   -- accounts for parent visibility
```

## Strata and Frame Level

Valid strata (back to front): `BACKGROUND`, `LOW`, `MEDIUM`, `HIGH`, `DIALOG`, `FULLSCREEN`, `FULLSCREEN_DIALOG`, `TOOLTIP`.

```lua
frame:SetFrameStrata("MEDIUM")
frame:SetFrameLevel(5)   -- 0-127 within the strata
```

## Backdrop

```lua
frame:SetBackdrop({
  bgFile   = "Interface\\BUTTONS\\WHITE8X8",
  edgeFile = "Interface\\BUTTONS\\WHITE8X8",
  tile     = false,
  tileSize = 0,
  edgeSize = 1,
  insets   = { left = -1, right = -1, top = -1, bottom = -1 },
})
frame:SetBackdropColor(0, 0, 0, 0.75)
frame:SetBackdropBorderColor(0.1, 0.1, 0.1, 1)
```

## Font Strings

```lua
local text = frame:CreateFontString(nil, "OVERLAY", "GameFontWhite")
text:SetPoint("TOPLEFT", frame, "TOPLEFT", 5, -5)
text:SetText("Hello World")
text:SetTextColor(1, 1, 0)
```

Layers (back to front): `"BACKGROUND"`, `"BORDER"`, `"ARTWORK"`, `"OVERLAY"`, `"HIGHLIGHT"`.

Common font templates: `"GameFontNormal"`, `"GameFontWhite"`, `"GameFontHighlight"`, `"GameFontSmall"`.

## Textures

```lua
local tex = frame:CreateTexture(nil, "ARTWORK")
tex:SetTexture("Interface\\AddOns\\MyAddon\\Textures\\icon.tga")
tex:SetAllPoints(frame)
tex:SetWidth(32)
tex:SetHeight(32)
tex:SetVertexColor(r, g, b)
tex:SetAlpha(0.5)

-- Solid colour without an external file
tex:SetTexture(1, 0, 0, 0.5)
```

## StatusBar

```lua
local bar = CreateFrame("StatusBar", nil, parent)
bar:SetStatusBarTexture("Interface\\BUTTONS\\WHITE8X8")
bar:SetStatusBarColor(0.2, 0.8, 0.2, 1)
bar:SetMinMaxValues(0, 100)
bar:SetValue(75)

local current      = bar:GetValue()
local min, max     = bar:GetMinMaxValues()
```

## Button

```lua
local btn = CreateFrame("Button", "MyButton", parent)
btn:SetWidth(80)
btn:SetHeight(20)
btn:SetNormalTexture("Interface\\...")
btn:SetHighlightTexture("Interface\\...")
btn:SetPushedTexture("Interface\\...")
btn:EnableMouse(true)

btn:SetScript("OnClick", function()
  if arg1 == "LeftButton" then
    MyAddon:DoSomething()
  end
end)
```

## Movable Frames

`RegisterForDrag` is required for the drag scripts to fire. Without it, `OnDragStart` never triggers.

```lua
frame:SetMovable(true)
frame:EnableMouse(true)
frame:RegisterForDrag("LeftButton")
frame:SetClampedToScreen(true)

frame:SetScript("OnDragStart", function()
  this:StartMoving()
end)
frame:SetScript("OnDragStop", function()
  this:StopMovingOrSizing()
end)
```

## Querying Frame Properties

```lua
frame:GetWidth()
frame:GetHeight()
frame:GetLeft()
frame:GetTop()
frame:GetRight()
frame:GetBottom()
frame:GetCenter()      -- returns cx, cy
frame:GetName()
frame:GetParent()
frame:GetEffectiveScale()    -- accounts for parent scale chain
frame:IsObjectType("Frame")  -- type check: "Frame", "Button", "Model", etc.

local f = _G["MyAddonFrame"]
```

`GetEffectiveScale()` returns the cumulative scale of the frame and all its parents. Use it when converting pixel offsets to absolute screen coordinates in vanilla, where `UIParent:GetScale()` may not equal 1.

`SetHitRectInsets(left, right, top, bottom)` extends or shrinks the clickable area relative to the frame's visual bounds. Use negative values to expand beyond the visual edge:

```lua
frame:SetHitRectInsets(-4, -4, -4, -4)   -- 4px expansion on all sides
```

## Mouse Wheel

```lua
frame:EnableMouseWheel(true)
frame:SetScript("OnMouseWheel", function()
  if arg1 > 0 then
    -- scrolled up
  else
    -- scrolled down
  end
end)
```

`arg1` is `1` for scroll up and `-1` for scroll down.
