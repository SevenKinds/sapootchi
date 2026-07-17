package ui

import "image/color"

// Theme is a full palette snapshot. SetTheme reassigns the package-level
// palette vars, so every draw call follows instantly — no plumbing.
type Theme struct {
	Name string

	BGTop, BGBottom  color.RGBA
	Panel, PanelHi   color.RGBA
	Text, TextDim    color.RGBA
	Accent, AccentHi color.RGBA
	Sec, SecHi       color.RGBA
	Disabled         color.RGBA
	Good, Warn, Bad  color.RGBA
	Energy, Gold     color.RGBA
	Track, Shadow    color.RGBA
	BtnInk           color.RGBA // button label color (buttons have their own fills)

	// Tab bar: background, active pill, active ink, inactive ink.
	NavBG, NavPill, NavInk, NavInkDim color.RGBA
}

// Themes, in toggle order. "Night" is the original look. "SAPO" and
// "SAPO Dark" follow the brand mockups (boxes: logo green #00D000, near-black,
// white type): light = white > green > black, dark = black > green > white.
var Themes = []Theme{
	{
		Name:  "Night",
		BGTop: rgba(0x3b, 0x46, 0x63), BGBottom: rgba(0x20, 0x25, 0x33),
		Panel: rgba(0x2a, 0x30, 0x40), PanelHi: rgba(0x33, 0x3a, 0x4d),
		Text: rgba(0xed, 0xef, 0xf5), TextDim: rgba(0x9a, 0xa3, 0xb2),
		Accent: rgba(0x4c, 0x8d, 0xff), AccentHi: rgba(0x6b, 0xa5, 0xff),
		Sec: rgba(0x7c, 0x5c, 0xff), SecHi: rgba(0x95, 0x7a, 0xff),
		Disabled: rgba(0x44, 0x4a, 0x59),
		Good:     rgba(0x4c, 0xc9, 0x6d), Warn: rgba(0xff, 0xb3, 0x3b), Bad: rgba(0xe6, 0x56, 0x4a),
		Energy: rgba(0x46, 0xc7, 0xe0), Gold: rgba(0xff, 0xd2, 0x4c),
		Track: rgba(0x1c, 0x20, 0x2b), Shadow: color.RGBA{0, 0, 0, 0x40},
		BtnInk: rgba(0xed, 0xef, 0xf5),
		NavBG:  rgba(0x1c, 0x20, 0x2b), NavPill: rgba(0x2a, 0x30, 0x40),
		NavInk: rgba(0xed, 0xef, 0xf5), NavInkDim: rgba(0x9a, 0xa3, 0xb2),
	},
	{
		Name:  "SAPO",
		BGTop: rgba(0xff, 0xff, 0xff), BGBottom: rgba(0xea, 0xf2, 0xe7),
		Panel: rgba(0xf3, 0xf6, 0xf1), PanelHi: rgba(0xdf, 0xf0, 0xdb), // mint tint
		Text: rgba(0x14, 0x14, 0x14), TextDim: rgba(0x5e, 0x6a, 0x60),
		Accent: rgba(0x00, 0xd0, 0x00), AccentHi: rgba(0x2a, 0xde, 0x2a),
		Sec: rgba(0x0f, 0xa0, 0x0f), SecHi: rgba(0x1c, 0xbd, 0x1c), // brand green cards
		Disabled: rgba(0xc9, 0xcf, 0xc6),
		Good:     rgba(0x00, 0xb3, 0x07), Warn: rgba(0xe3, 0x9a, 0x00), Bad: rgba(0xdd, 0x44, 0x33),
		Energy: rgba(0x00, 0xa8, 0xc6), Gold: rgba(0xdb, 0x9e, 0x00),
		Track: rgba(0xdd, 0xe3, 0xda), Shadow: color.RGBA{0, 0, 0, 0x26},
		BtnInk: rgba(0xff, 0xff, 0xff),
		NavBG:  rgba(0x00, 0xd0, 0x00), NavPill: rgba(0x00, 0xa8, 0x00), // green nav...
		NavInk: rgba(0x11, 0x11, 0x11), NavInkDim: rgba(0x06, 0x59, 0x06), // ...black icons
	},
	{
		Name:  "SAPO Dark",
		BGTop: rgba(0x1a, 0x1c, 0x1a), BGBottom: rgba(0x0b, 0x0c, 0x0b),
		Panel: rgba(0x1f, 0x22, 0x1f), PanelHi: rgba(0x2a, 0x2e, 0x2a),
		Text: rgba(0xf2, 0xf4, 0xf1), TextDim: rgba(0x98, 0xa2, 0x96),
		Accent: rgba(0x00, 0xd0, 0x00), AccentHi: rgba(0x2a, 0xe2, 0x2a),
		Sec: rgba(0x0e, 0x8f, 0x0e), SecHi: rgba(0x17, 0xad, 0x17), // brand green
		Disabled: rgba(0x3a, 0x40, 0x3a),
		Good:     rgba(0x35, 0xd0, 0x45), Warn: rgba(0xff, 0xb3, 0x3b), Bad: rgba(0xe6, 0x56, 0x4a),
		Energy: rgba(0x46, 0xc7, 0xe0), Gold: rgba(0xff, 0xd2, 0x4c),
		Track: rgba(0x14, 0x16, 0x14), Shadow: color.RGBA{0, 0, 0, 0x60},
		BtnInk: rgba(0xff, 0xff, 0xff),
		NavBG:  rgba(0x11, 0x13, 0x11), NavPill: rgba(0x22, 0x26, 0x22),
		NavInk: rgba(0x00, 0xd0, 0x00), NavInkDim: rgba(0x0b, 0x6e, 0x0b), // green icons
	},
}

func rgba(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 0xff} }

// SetTheme applies the named theme (unknown names fall back to the first).
func SetTheme(name string) {
	t := Themes[0]
	for _, th := range Themes {
		if th.Name == name {
			t = th
			break
		}
	}
	BGTop, BGBottom = t.BGTop, t.BGBottom
	Panel, PanelHi = t.Panel, t.PanelHi
	Text, TextDim = t.Text, t.TextDim
	Accent, AccentHover = t.Accent, t.AccentHi
	Secondary, SecondaryHi = t.Sec, t.SecHi
	Disabled = t.Disabled
	Good, Warn, Bad = t.Good, t.Warn, t.Bad
	Energy, Gold = t.Energy, t.Gold
	Track, Shadow = t.Track, t.Shadow
	ButtonInk = t.BtnInk
	NavBG, NavPill, NavInk, NavInkDim = t.NavBG, t.NavPill, t.NavInk, t.NavInkDim
}

// NextTheme returns the theme name after the given one (cycling).
func NextTheme(name string) string {
	for i, th := range Themes {
		if th.Name == name {
			return Themes[(i+1)%len(Themes)].Name
		}
	}
	return Themes[0].Name
}
