# Color Customization Guide

> Complete guide for customizing colors in SublimeGo — from built-in Tailwind palettes to custom hex/RGB colors, inspired by Filament PHP.

---

## Overview

SublimeGo provides flexible color customization:

- **Built-in Tailwind palettes** — 10 pre-defined color schemes
- **Custom hex colors** — Generate full palettes from `#3b82f6`
- **Custom RGB colors** — Generate palettes from `rgb(59, 130, 246)`
- **Dynamic color registration** — Register custom palettes at runtime
- **CSS variable generation** — Automatic Tailwind integration

---

## Quick Start

### Using Built-in Colors

```go
panel := engine.NewPanel("admin").
    WithPrimaryColor("blue")  // green, blue, red, purple, orange, pink, indigo, teal, amber, cyan
```

### Using Custom Hex Colors

```go
panel := engine.NewPanel("admin").
    WithCustomColor("#3b82f6")  // Any hex color
```

### Using Custom RGB Colors

```go
panel := engine.NewPanel("admin").
    WithCustomColor("rgb(59, 130, 246)")
```

---

## Built-in Color Palettes

SublimeGo includes all Tailwind CSS color palettes:

| Color | Example | Use Case |
|-------|---------|----------|
| `green` | 🟢 | Default, success states |
| `blue` | 🔵 | Info, links |
| `red` | 🔴 | Danger, errors |
| `purple` | 🟣 | Premium, creative |
| `orange` | 🟠 | Warnings, highlights |
| `pink` | 🩷 | Playful, feminine |
| `indigo` | 🟦 | Professional, tech |
| `teal` | 🩵 | Fresh, modern |
| `amber` | 🟡 | Warm, inviting |
| `cyan` | 🩵 | Cool, digital |

Each palette includes 11 shades (50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950).

---

## Custom Colors

### From Hex Code

Generate a full Tailwind-style palette from any hex color:

```go
import "github.com/bozz33/sublimego/color"

c := color.Color{}
palette := c.Hex("#3b82f6")  // Blue
palette := c.Hex("#10b981")  // Green
palette := c.Hex("#f59e0b")  // Amber
```

The generated palette includes:
- **50** — Lightest shade (97% lightness)
- **100-400** — Light shades
- **500** — Base color (your input)
- **600-900** — Dark shades
- **950** — Darkest shade (15% lightness)

### From RGB Values

```go
// String format
palette := c.RGB("rgb(59, 130, 246)")
palette := c.RGB("59, 130, 246")  // Also works

// Separate values
palette := c.FromRGB(59, 130, 246)
```

---

## Registering Custom Palettes

### Register a Custom Palette

```go
import "github.com/bozz33/sublimego/color"

manager := color.NewManager()

// From hex
customPalette := color.Color{}.Hex("#ff6b6b")
manager.Register("danger", customPalette)

// From RGB
brandPalette := color.Color{}.RGB("rgb(123, 45, 67)")
manager.Register("brand", brandPalette)
```

### Set as Primary Color

```go
manager.SetPrimary("brand")

// Generate CSS variables
css := manager.PrimaryCSSVars()
// Output:
//   --color-primary-50: #...;
//   --color-primary-100: #...;
//   ...
//   --color-primary-950: #...;
```

---

## Using Colors in Components

### In Badges

```go
import "github.com/bozz33/sublimego/table"

table.Badge("status").
    Label("Status").
    Colors(map[string]string{
        "active":   "green",
        "pending":  "yellow",
        "rejected": "red",
    })
```

### In Buttons

```go
import "github.com/bozz33/sublimego/actions"

actions.New("publish").
    Label("Publish").
    Color("primary")  // Uses panel's primary color

actions.New("delete").
    Label("Delete").
    Color("danger")   // Uses danger color
```

### In Notifications

```go
import "github.com/bozz33/sublimego/notifications"

notifications.Success("Published").WithBody("Post published successfully")
notifications.Danger("Error").WithBody("Failed to save")
notifications.Warning("Warning").WithBody("Unsaved changes")
notifications.Info("Info").WithBody("New version available")
```

---

## Advanced Usage

### Manual Palette Creation

```go
import "github.com/bozz33/sublimego/color"

palette := &color.Palette{
    Name: "custom",
    Shades: []color.Shade{
        {Number: 50,  Hex: "#fef2f2"},
        {Number: 100, Hex: "#fee2e2"},
        {Number: 200, Hex: "#fecaca"},
        {Number: 300, Hex: "#fca5a5"},
        {Number: 400, Hex: "#f87171"},
        {Number: 500, Hex: "#ef4444"},  // Base color
        {Number: 600, Hex: "#dc2626"},
        {Number: 700, Hex: "#b91c1c"},
        {Number: 800, Hex: "#991b1b"},
        {Number: 900, Hex: "#7f1d1d"},
        {Number: 950, Hex: "#450a0a"},
    },
}

manager := color.NewManager()
manager.Register("custom-red", palette)
```

### Generate CSS Variables for All Palettes

```go
manager := color.NewManager()
manager.Register("brand", color.Color{}.Hex("#3b82f6"))
manager.Register("accent", color.Color{}.Hex("#10b981"))

// Generate CSS for all registered palettes
css := manager.AllCSSVars()
// Output:
//   --color-brand-50: #...;
//   --color-brand-100: #...;
//   ...
//   --color-accent-50: #...;
//   --color-accent-100: #...;
//   ...
```

### Use in Templates

```go
// In your custom CSS
<style>
    .my-button {
        background-color: var(--color-primary-500);
        color: white;
    }
    .my-button:hover {
        background-color: var(--color-primary-600);
    }
</style>
```

---

## Color Theory Tips

### Choosing a Primary Color

1. **Brand alignment** — Use your brand's primary color
2. **Contrast** — Ensure good contrast with white/dark backgrounds
3. **Accessibility** — Test with WCAG contrast checkers
4. **Emotion** — Blue = trust, Green = success, Red = urgency

### Generating Harmonious Palettes

```go
// Complementary colors (opposite on color wheel)
primary := color.Color{}.Hex("#3b82f6")   // Blue
accent := color.Color{}.Hex("#f59e0b")    // Orange

// Analogous colors (adjacent on color wheel)
primary := color.Color{}.Hex("#3b82f6")   // Blue
secondary := color.Color{}.Hex("#8b5cf6") // Purple
tertiary := color.Color{}.Hex("#06b6d4")  // Cyan
```

---

## Complete Example

```go
package main

import (
    "github.com/bozz33/sublimego/color"
    "github.com/bozz33/sublimego/engine"
)

func main() {
    // 1. Create custom brand palette
    brandColor := color.Color{}.Hex("#ff6b6b")
    
    // 2. Register it
    colorManager := color.NewManager()
    colorManager.Register("brand", brandColor)
    colorManager.SetPrimary("brand")
    
    // 3. Use in panel
    panel := engine.NewPanel("admin").
        WithCustomColor("#ff6b6b").  // Same as brand color
        SetBrandName("My App").
        SetPath("/admin")
    
    // 4. The panel will automatically use your custom color for:
    //    - Primary buttons
    //    - Active navigation items
    //    - Focus states
    //    - Badges with "primary" color
    //    - Success notifications
}
```

---

## Comparison with Filament

| Feature | Filament PHP | SublimeGo |
|---------|-------------|-----------|
| Built-in palettes | ✅ All Tailwind | ✅ All Tailwind |
| Hex colors | ✅ `Color::hex()` | ✅ `Color{}.Hex()` |
| RGB colors | ✅ `Color::rgb()` | ✅ `Color{}.RGB()` |
| Custom shades | ✅ Array of RGB | ✅ Manual `Palette` |
| CSS variables | ✅ Auto-generated | ✅ Auto-generated |
| Runtime registration | ✅ `FilamentColor::register()` | ✅ `Manager.Register()` |

SublimeGo's color system is **directly inspired by Filament** and provides the same level of flexibility and ease of use.

---

## Best Practices

1. **Stick to one primary color** — Consistency is key
2. **Use semantic names** — `danger`, `success`, `warning`, `info`
3. **Test in dark mode** — Ensure colors work in both themes
4. **Limit custom palettes** — Too many colors = visual chaos
5. **Use the 500 shade** — As your base color reference
6. **Generate, don't hardcode** — Let the system create shades for you
7. **Accessibility first** — Always check contrast ratios
