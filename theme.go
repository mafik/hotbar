package main

const BarHeight = 60

type Color []uint8

type Class struct {
	Foreground Color
	Background Color
	Fill       Color
	Padding    int
}

type Theme struct {
	IconDir    string
	FontPath   string
	FontHeight int
	TintIcons  bool
	IconSize   int
	Default    Class
	Classes    map[string]Class
}

var DuskTheme = Theme{
	IconDir:    "dusk",
	FontPath:   "iosevka-term-ss05-regular.ttf",
	FontHeight: BarHeight / 2,
	TintIcons:  false,
	IconSize:   BarHeight,
	Default: Class{
		Fill:       Color{97, 212, 253},
		Foreground: Color{255, 255, 255},
		Background: Color{48, 175, 215},
		Padding:    BarHeight / 6,
	},
	Classes: map[string]Class{
		"ssd": Class{
			Background: Color{167, 177, 197},
			Padding:    BarHeight / 6,
		},
		"clock": Class{
			Background: Color{121, 190, 122},
			Padding:    BarHeight / 6,
		},
		"keyboard": Class{
			Background: Color{159, 141, 108},
			Padding:    BarHeight / 6,
		},
		"calendar": Class{
			Background: Color{51, 153, 214},
			Padding:    BarHeight / 6,
		},
		"backlight": Class{
			Background: Color{233, 163, 41},
			Fill:       Color{249, 227, 174},
			Padding:    BarHeight / 6,
		},
		"battery": Class{
			Fill:       Color{237, 120, 153},
			Background: Color{70, 59, 81},
			Padding:    BarHeight / 6,
		},
		"sound-device": Class{
			Background: Color{233, 94, 70},
			Padding:    BarHeight / 6,
		},
		"sound-volume": Class{
			Background: Color{233, 163, 41},
			Fill:       Color{249, 227, 174},
			Padding:    BarHeight / 6,
		},
		"workspace": Class{
			Padding: 0,
		},
		"menu": Class{
			Padding: 0,
		},
	},
}

var NeonTheme = Theme{
	IconDir:    "1em",
	FontPath:   "iosevka-term-ss05-regular.ttf",
	FontHeight: BarHeight * 4 / 6,
	TintIcons:  true,
	IconSize:   BarHeight * 5 / 6,
	Default: Class{
		Fill:       Color{50, 50, 50},
		Foreground: Color{128, 128, 128},
		Background: Color{0, 0, 0},
	},
	Classes: map[string]Class{
		"ssd": Class{
			Foreground: Color{167, 177, 197},
			Padding:    BarHeight / 6,
		},
		"clock": Class{
			Foreground: Color{121, 190, 122},
			Padding:    BarHeight / 6,
		},
		"keyboard": Class{
			Foreground: Color{159, 141, 108},
			Padding:    BarHeight / 6,
		},
		"calendar": Class{
			Foreground: Color{51, 153, 214},
			Padding:    BarHeight / 6,
		},
		"backlight": Class{
			Fill:       Color{233, 163, 41},
			Foreground: Color{249, 227, 174},
			Padding:    BarHeight / 6,
		},
		"battery": Class{
			Foreground: Color{237, 120, 153},
			Fill:       Color{70, 59, 81},
			Padding:    BarHeight / 6,
		},
		"sound-device": Class{
			Foreground: Color{233, 94, 70},
			Padding:    BarHeight / 6,
		},
		"sound-volume": Class{
			Fill:       Color{103, 63, 0},
			Foreground: Color{249, 227, 174},
			Padding:    BarHeight / 6,
		},
	},
}

var T = NeonTheme
