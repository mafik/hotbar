package main

import (
	"math"

	"mrogalski.eu/go/xbacklight"
)

type Backlight struct{}

var backlight = 1.0
var backlighter xbacklight.Backlighter

var minBacklight = math.Nextafter(0, 1)
var steps = []float64{minBacklight, 0.1, 0.5, 1.0}

func (Backlight) Init() error {
	var err error
	backlighter, err = xbacklight.NewBacklighterPrimaryScreen()
	if err != nil {
		return err
	}
	backlight, err = backlighter.Get()
	if err != nil {
		return err
	}
	return nil
}

func (Backlight) LeftClick(x int) bool {
	backlight = NextStep(backlight, steps)
	backlighter.Set(backlight)
	return true
}

func (Backlight) RightClick(x int) bool {
	backlight = PrevStep(backlight, steps)
	backlighter.Set(backlight)
	return true
}

func (Backlight) Wheel(delta int) bool {
	start := backlight
	backlight += float64(delta) / 64
	if backlight < minBacklight {
		backlight = minBacklight
	}
	if backlight > 1 {
		backlight = 1
	}
	if start != backlight {
		backlighter.Set(backlight)
		return true
	}
	return false
}

func (Backlight) Width() int {
	return Width("backlight", Icon(""))
}

func (Backlight) Draw(x int) {
	Draw(x, "backlight", float32(backlight), Icon("backlight"))
}
