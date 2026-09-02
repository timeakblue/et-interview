# ExploreTech Design Language Reference

This document captures the visual and structural design language of ExploreTech apps. Use it as a reference when building apps that should feel like a cohesive pair.

---

## Color Palette

### Brand / Primary

| Token | Hex | Usage |
|---|---|---|
| `--color-primary` | `#072947` | Primary brand color, GlobalRail background, buttons |
| `--color-primary-hover` | `#0a3a5c` | Hover state for primary elements |
| `--color-primary-dark` | `#051a2e` | Dark variant, gradients |

### Neutrals

| Token | Hex | Usage |
|---|---|---|
| `--color-background` | `#ffffff` | Page/panel backgrounds |
| `--color-surface` | `#f8f9fa` | List rows, input backgrounds, section surfaces |
| `--color-border` | `#e2e4e7` | Dividers, input borders, row separators |
| `--color-text-primary` | `#1a1a1a` | Primary readable text |
| `--color-text-secondary` | `#6b7280` | Captions, subtitles, helper text |
| `--color-text-muted` | `#9ca3af` | Placeholder text, disabled labels |

### Semantic / Status

| Color | Hex | Usage |
|---|---|---|
| Success | `#16a34a` | Completed states |
| Success (bright) | `#10b981` | Complete status badges |
| Warning | `#d97706` | Warning states |
| Canceled | `#f59e0b` | Canceled status |
| Error | `#dc2626` | Error states |
| Error (bright) | `#ef4444` | Failed status badges |
| Info | `#2563eb` | Informational accents |

### Data / Accent

These colors are used to distinguish data categories in 3D views and map layers:

| Color | Hex | Usage |
|---|---|---|
| Blue accent | `#3b82f6` | Visibility toggles, general accents |

### Overlays & Scene

Used in 3D/map contexts where UI sits over a dark canvas:

| Value | Usage |
|---|---|
| `rgba(0, 0, 0, 0.6)` | Overlay panel backgrounds |
| `rgba(255, 255, 255, 0.1)` | Overlay buttons (default) |
| `rgba(255, 255, 255, 0.2)` | Overlay buttons (hover) |
| `#1a1a1a` | 3D scene background |

---

## Typography

**Font stack:** `Inter, system-ui, -apple-system, sans-serif`  
**Monospace:** `ui-monospace, SFMono-Regular, 'SF Mono', Menlo, monospace`

### Scale

| Size | Weight | Usage |
|---|---|---|
| `18px` | `600` | Page titles (h1) |
| `15px` | `500` | Emphasized body, empty-state titles |
| `14px` | `400–500` | Primary body text, buttons |
| `13px` | `400–500` | Secondary body, form fields, row content |
| `12px` | `400` | Captions, secondary labels |
| `11px` | `600` | Section headers (uppercase) |
| `10px` | `400–500` | Smallest labels |

### Conventions

- **Section headers** are `11px`, `fontWeight: 600`, `text-transform: uppercase`, `letter-spacing: 0.05em` — this is a strong visual pattern throughout the app.
- **Numeric data** uses `font-variant-numeric: tabular-nums` to keep columns aligned.
- Long text is truncated with `text-overflow: ellipsis` + `white-space: nowrap` rather than wrapping.

---

## Spacing

The app uses a consistent base-4 spacing system. Most measurements are multiples of 4px.

### Common Values

| Value | Context |
|---|---|
| `2–4px` | Icon gaps, tight internal padding |
| `6–8px` | Small component padding, button padding |
| `12px` | Section gaps, row internal spacing |
| `14–16px` | Row padding (vertical), standard horizontal padding |
| `20–24px` | Page header padding, larger component gutters |
| `40–60px` | Empty/error state padding |

### Fixed Layout Dimensions

| Measurement | Value |
|---|---|
| GlobalRail width | `52px` |
| Sidebar width | `320px` |
| Inspector width | `380px` |
| Main content left offset | `52px` (matches rail) |

---

## Border Radius

| Token | Value | Usage |
|---|---|---|
| `--radius-sm` | `4px` | Buttons, inputs, most UI elements |
| `--radius-md` | `6px` | Medium components |
| `--radius-lg` | `8–10px` | Cards, containers |
| — | `12px` | Icon containers, empty state boxes |
| — | `24px` | Tooltip backgrounds |
| — | `50%` | Circular elements (color pickers, avatars) |

The default radius is `4px`. Rounder elements (12px, 24px) are used sparingly for decorative or floating UI.

---

## Shadows

| Token | Value | Usage |
|---|---|---|
| `--shadow-subtle` | `0 2px 6px rgba(0,0,0,0.08)` | Hover elevation on buttons |
| Subtle | `0 1px 2px rgba(0,0,0,0.05)` | Very light surface lift |
| Medium | `0 4px 12px rgba(0,0,0,0.10)` | Dropdowns, standard float |
| Elevated | `0 8px 24px rgba(0,0,0,0.15)` | Modals, prominent overlays |

Shadows are used conservatively — only for floating elements (dropdowns, tooltips, modals). Flat surfaces (cards in a list, page panels) rely on borders rather than shadow for separation.

---

## Layout & Structure

### App Shell

```
┌────────────────────────────────────────────────┐
│  GlobalRail │  Main Content Area               │
│   52px      │  flex: 1 — overflow: hidden      │
│   fixed     │                                  │
│   z:40      │  ┌──────────────────────────┐    │
│             │  │ Page Header              │    │
│             │  │ padding: 20px 24px       │    │
│             │  │ border-bottom: 1px       │    │
│             │  ├──────────────────────────┤    │
│             │  │ Content Body             │    │
│             │  │ flex: 1, overflow: auto  │    │
│             │  │ background: color-surface│    │
│             │  └──────────────────────────┘    │
└────────────────────────────────────────────────┘
```

- Root: `height: 100vh`, `width: 100vw`, `overflow: hidden` — no page scroll.
- The GlobalRail is `position: fixed` and always visible.
- All scrolling happens inside bounded content regions.

### Page Headers

Every primary view opens with a header zone:
- `padding: 20px 24px`
- `border-bottom: 1px solid var(--color-border)`
- Title: `18px / 600`
- Subtitle: `13px / color-text-secondary`, `margin-top: 4px`
- Action buttons right-aligned via `justify-content: space-between`

### Sidebars & Inspectors

- Sidebars are collapsible panels with sections.
- Section headers: `11px / uppercase / 600`, `padding: 12px 16px`.
- Section rows: `padding: 14px 20px`, `border-bottom: 1px solid var(--color-border)`.
- Hover state: background shifts to `var(--color-surface)`, transition `100ms ease-out`.

---

## Component Patterns

### GlobalRail Navigation

- Background: `var(--color-primary)` (`#072947`)
- Icon buttons: transparent background, white icon at `0.6` opacity
- Active item: `opacity: 1`, left border `2px solid #ffffff`
- Hover: left border `2px solid rgba(255,255,255,0.2)`
- Icons: Lucide, `20px`, `strokeWidth: 1.5`
- Tooltips: dark bg `#1a1a1a`, white text `12px`, appear to the right

### Buttons

- Border-radius: `4–6px`
- Font: `13–14px / 500–600`
- Transition: `all 150ms ease-out`
- **Primary:** bg `var(--color-primary)`, hover `var(--color-primary-hover)`, white text
- **Secondary / Ghost:** `border: 1px solid var(--color-border)`, hover bg `var(--color-surface)`
- Disabled: `opacity: 0.5`

### List Rows

- Container bg: `var(--color-background)` or `var(--color-surface)`
- Row padding: `14px 20px`
- Separator: `1px solid var(--color-border)`
- Hover: bg → `var(--color-surface)`, `transition: background 100ms ease-out`
- Text truncation pattern: `overflow: hidden`, `white-space: nowrap`, `text-overflow: ellipsis`

### Empty / Error States

- Layout: centered flex column
- Icon container: `64×64px`, `border-radius: 12px`, bg `var(--color-surface)`
- Title: `15px / 500`
- Description: `13px / color-text-secondary`
- Padding: `60px 24px` (empty), `40px 24px` (error)

### Loading States

- Inline spinner: Lucide `Loader2` icon, CSS `spin 0.8s linear infinite`
- Skeleton rows: `background: var(--color-surface)`, `animation: pulse 1.5s ease-in-out infinite` (opacity 1 → 0.4 → 1)

### Toggles / Switches

- Track: `14px` tall, `border-radius: 7px`
- Active color: `#3b82f6`, inactive: `#d1d5db`
- Thumb: white circle, `transition: left 0.2s`
- Used for layer visibility, boolean settings

### Status Badges

Inline pill-style indicators:
- Small dot + text pattern, or text-only with semantic color
- Font: `12px / 500`
- No border, text color carries the semantic meaning

---

## Icons

### Lucide React (primary)

Used for all navigation, action, and state icons throughout the app.

- Standard size: `14px` (small UI), `20px` (navigation), `28px` (larger contexts)
- Standard stroke: `strokeWidth={1.5}` — slightly thinner than default
- Color via `color` prop or CSS inheritance

Common icons: `Map`, `Box`, `FolderOpen`, `LogOut`, `Eye`, `EyeOff`, `ChevronDown`, `ChevronRight`, `AlertCircle`, `CheckCircle`, `XCircle`, `Loader2`

### Material-UI Icons (secondary)

Used in authentication and select 3D viewer controls only. Not the primary icon system.

- Size via `sx={{ fontSize: 20 }}`
- Examples: `PersonOutline`, `LockOutline`, `Visibility`, `VisibilityOff`

---

## Animation & Transitions

### Timing

| Variable | Value | Usage |
|---|---|---|
| `--transition-default` | `200ms ease-out` | Standard UI transitions |
| Fast | `150ms ease-out` | Button hover, quick feedback |
| Very fast | `100ms ease-out` | Background swaps on row hover |
| Toggle | `0.2s` | Switch/toggle thumb movement |

### Easing

- `ease-out` — default for all interactive state changes (feels snappy)
- `ease-in-out` — used sparingly for larger motion
- `linear` — spinning/looping animations only

### Named Animations

```css
/* Loading spinner */
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Component fade-in */
@keyframes fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}

/* Skeleton pulse */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50%       { opacity: 0.4; }
}
```

### Transform Patterns

- `scale(1.1)` — hover emphasis (color pickers, small action icons)
- `translateX(-50%) / translateY(-50%)` — centering absolutely-positioned elements
- `rotate()` — map compass, chevron open/close states

---

## Opacity Scale

| Value | Usage |
|---|---|
| `1.0` | Active, fully visible |
| `0.9` | Near-full, slight de-emphasis |
| `0.6` | Default icon opacity in nav rail |
| `0.5` | Disabled state |
| `0.3 – 0.2` | Secondary overlay elements |
| `0.1 – 0.05` | Subtle background tints |

---

## Quick Reference Cheat Sheet

```
Primary brand:    #072947
Primary hover:    #0a3a5c
Background:       #ffffff
Surface:          #f8f9fa
Border:           #e2e4e7
Text primary:     #1a1a1a
Text secondary:   #6b7280
Text muted:       #9ca3af

Font:             Inter (system-ui fallback)
Body text:        13–14px / 400
Labels:           11px / 600 / uppercase
Titles:           18px / 600

Radius:           4px default, 6–8px medium, 12px decorative
Border:           1px solid #e2e4e7
Shadow:           0 4px 12px rgba(0,0,0,0.10)
Transition:       150–200ms ease-out

Rail width:       52px
Sidebar width:    320px
```
