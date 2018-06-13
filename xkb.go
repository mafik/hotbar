package main

import (
	"os/exec"
)

type Xkb struct{}

func (Xkb) Init() error {
	xkbLayout = getXkbLayout()
	go monitorKeyboardBackground()
	return nil
}

func (Xkb) Draw(x int) {
	Draw(x, "keyboard", 0, Icon("keyboard"), xkbLayout)
}

func (Xkb) Width() int {
	return Width("keyboard", Icon("keyboard"), xkbLayout)
}

func (Xkb) LeftClick(x int) bool {
	exec.Command("xkb-switch", "-n").Run()
	return false // Update will come from the "monitor" goroutine
}

var xkbLayout string

func getXkbLayout() string {
	out, err := exec.Command("xkb-switch", "-p").Output()
	if err != nil {
		return "Install xkb-switch program!"
	}
	ret := string(out[:len(out)-1])
	return ret
}

func monitorKeyboardBackground() {
	var err error
	for err == nil {
		err = exec.Command("xkb-switch", "-w").Run()
		MainThread <- func() {
			xkbLayout = getXkbLayout()
			Redraw()
		}
	}
}
