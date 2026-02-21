package color

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Color provides methods to generate color palettes from various formats.
// Inspired by Filament's Color class.
type Color struct{}

// Hex generates a full color palette (50-950) from a hex color code.
// Example: Color{}.Hex("#3b82f6") returns a blue palette.
func (Color) Hex(hex string) *Palette {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		// Fallback to gray if invalid
		return generatePaletteFromRGB(156, 163, 175)
	}

	r, err := strconv.ParseInt(hex[0:2], 16, 64)
	if err != nil {
		return generatePaletteFromRGB(156, 163, 175) // gray-400 fallback
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 64)
	if err != nil {
		return generatePaletteFromRGB(156, 163, 175)
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 64)
	if err != nil {
		return generatePaletteFromRGB(156, 163, 175)
	}

	return generatePaletteFromRGB(int(r), int(g), int(b))
}

// RGB generates a full color palette from an RGB string.
// Example: Color{}.RGB("rgb(59, 130, 246)") or Color{}.RGB("59, 130, 246")
func (Color) RGB(rgb string) *Palette {
	// Clean the string
	rgb = strings.TrimPrefix(rgb, "rgb(")
	rgb = strings.TrimPrefix(rgb, "rgba(")
	rgb = strings.TrimSuffix(rgb, ")")
	rgb = strings.ReplaceAll(rgb, " ", "")

	parts := strings.Split(rgb, ",")
	if len(parts) < 3 {
		return generatePaletteFromRGB(156, 163, 175)
	}

	r, _ := strconv.Atoi(parts[0])
	g, _ := strconv.Atoi(parts[1])
	b, _ := strconv.Atoi(parts[2])

	return generatePaletteFromRGB(r, g, b)
}

// FromRGB generates a palette from separate R, G, B values (0-255).
func (Color) FromRGB(r, g, b int) *Palette {
	return generatePaletteFromRGB(r, g, b)
}

// generatePaletteFromRGB creates a full Tailwind-style palette (50-950) from a base RGB color.
// Uses HSL color space to generate lighter and darker shades.
func generatePaletteFromRGB(r, g, b int) *Palette {
	h, s, l := rgbToHSL(r, g, b)

	shades := []Shade{
		{Number: 50, Hex: hslToHex(h, s, 0.97)},
		{Number: 100, Hex: hslToHex(h, s, 0.94)},
		{Number: 200, Hex: hslToHex(h, s, 0.86)},
		{Number: 300, Hex: hslToHex(h, s, 0.77)},
		{Number: 400, Hex: hslToHex(h, s, 0.66)},
		{Number: 500, Hex: hslToHex(h, s, l)}, // base color
		{Number: 600, Hex: hslToHex(h, s, l*0.83)},
		{Number: 700, Hex: hslToHex(h, s, l*0.67)},
		{Number: 800, Hex: hslToHex(h, s, l*0.52)},
		{Number: 900, Hex: hslToHex(h, s, l*0.35)},
		{Number: 950, Hex: hslToHex(h, s, l*0.15)},
	}

	return &Palette{
		Name:   "custom",
		Shades: shades,
	}
}

// rgbToHSL converts RGB (0-255) to HSL (0-360, 0-1, 0-1).
func rgbToHSL(r, g, b int) (h, s, l float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	l = (max + min) / 2.0

	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6
			}
		case gf:
			h = (bf-rf)/d + 2
		case bf:
			h = (rf-gf)/d + 4
		}
		h *= 60
	}

	return h, s, l
}

// hslToHex converts HSL to hex color string.
func hslToHex(h, s, l float64) string {
	r, g, b := hslToRGB(h, s, l)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// hslToRGB converts HSL to RGB (0-255).
func hslToRGB(h, s, l float64) (r, g, b int) {
	var rf, gf, bf float64

	if s == 0 {
		rf = l
		gf = l
		bf = l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		rf = hueToRGB(p, q, h/360.0+1.0/3.0)
		gf = hueToRGB(p, q, h/360.0)
		bf = hueToRGB(p, q, h/360.0-1.0/3.0)
	}

	r = int(math.Round(rf * 255))
	g = int(math.Round(gf * 255))
	b = int(math.Round(bf * 255))
	return
}

// hueToRGB is a helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}
