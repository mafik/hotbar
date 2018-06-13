package main

import (
	"os/exec"
	"regexp"
)

type Disk struct{}

func (Disk) Draw(x int) {
	Draw(x, "ssd", 0, Icon("ssd"), FreeSpace+" GB")
}

func (Disk) Width() int {
	return Width("ssd", Icon("ssd"), FreeSpace+" GB")
}

func (Disk) Refresh() {
	out, err := exec.Command("btrfs", "filesystem", "usage", "/").Output()
	if err != nil {
		return
	}
	result := btrfsRegexp.FindSubmatch(out)[1]
	FreeSpace = string(result)
}

var btrfsRegexp, _ = regexp.Compile(`Free \(estimated\):\s+([0-9.]+)GiB`)
var FreeSpace string = "0.00"
