package main

import "time"

type Clock struct{}

func (Clock) Draw(x int) {
	Draw(x, "clock", 0, Icon("clock"), time.Now().Format("15:04"))
}

func (Clock) Width() int {
	return Width("clock", Icon("clock"), "15:04")
}
