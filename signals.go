package main

import (
	"os"
	"os/signal"
)

func init() {
	signalChan := make(chan os.Signal)
	go func() {
		<-signalChan
		MainThread <- func() { KeepRunning = false }
	}()
	signal.Notify(signalChan, os.Interrupt)
}
