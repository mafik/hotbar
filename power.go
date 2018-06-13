package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
)

var governor string = "powersave"
var battery int = 100
var batteryRegexp, _ = regexp.Compile(`percentage:\s+([0-9]+)%`)

type Power struct{}

func (Power) Draw(x int) {
	icon := "battery"
	if governor == "performance" {
		icon = "fast"
	}
	Draw(x, "battery", float32(battery)/100, Icon(icon))
}

func (Power) Width() int {
	return Width("battery", Icon("battery"))
}

func (Power) LeftClick(x int) bool {
	var newGovernor string
	if governor == "performance" {
		newGovernor = "powersave"
	} else {
		newGovernor = "performance"
	}
	if err := setGovernor(newGovernor); err != nil {
		fmt.Println("Error when changing CPU governor:", err)
		return false
	}
	governor = newGovernor
	return true
}

func (Power) Refresh() {
	battery = getBattery()
	governor = getGovernor()
}

func getBattery() int {
	out, err := exec.Command("upower", "-i", "/org/freedesktop/UPower/devices/battery_BAT0").Output()
	if err != nil {
		return 0
	}
	result := batteryRegexp.FindSubmatch(out)[1]
	var parsed int
	fmt.Sscan(string(result), &parsed)
	return parsed
}

func setGovernor(governor string) error {
	paths, err := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]/cpufreq/scaling_governor")
	if err != nil {
		return err
	}
	for _, path := range paths {
		err = exec.Command("sudo", "-n", "sh", "-c", "echo "+governor+" > "+path).Run()
		//err = ioutil.WriteFile(path, []byte(governor), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func getGovernor() string {
	bytes, _ := ioutil.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
	return string(bytes[:len(bytes)-1])
}
