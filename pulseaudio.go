package main

import (
	"mrogalski.eu/go/pulseaudio"
)

type PulseAudioVolume struct{}

func (PulseAudioVolume) Init() error {
	c, err := pulseaudio.NewClient("mbar")
	if err != nil {
		return err
	}
	paClient = c
	sink := paClient.Sinks()[0]
	paVolume = int(sink.GetVolume()*120+1) / 2
	paDevice = getPaDevice()
	paClient.Subscribe(pulseaudioCallback)
	return nil
}

func (PulseAudioVolume) Draw(x int) {
	Draw(x, "sound-volume", float32(paVolume)/100, volumeIcon())
}

func (PulseAudioVolume) Width() int {
	return Width("sound-volume", volumeIcon())
}

func (PulseAudioVolume) Wheel(delta int) bool {
	setPaVolume(paVolume + delta)
	return false // update will come from paCallback
}

func (PulseAudioVolume) LeftClick(x int) bool {
	setPaVolume(((paVolume+10)/20 + 5) % 6 * 20)
	return false // update will come from paCallback
}

func (PulseAudioVolume) RightClick(x int) bool {
	setPaVolume(((paVolume+10)/20 + 1) % 6 * 20)
	return false // update will come from paCallback
}

func volumeIcon() Icon {
	if paVolume > 60 {
		return "sound-3"
	}
	if paVolume > 40 {
		return "sound-2"
	}
	if paVolume > 20 {
		return "sound-1"
	}
	if paVolume > 0 {
		return "sound-0"
	}
	return "sound-muted"
}

type PulseAudioDevice struct{}

func (PulseAudioDevice) Draw(x int) {
	Draw(x, "sound-device", 0, Icon(paDevice))
}

func (PulseAudioDevice) Width() int {
	return Width("sound-device", Icon(paDevice))
}

func (PulseAudioDevice) LeftClick(x int) bool {
	togglePaDevice()
	return false
}

var paClient *pulseaudio.Client
var paVolume = 50
var paDevice = "analog-output-speaker"

/*
Block 1:

Always unmute

Left-click: change volume by += 20% (mod 100%)
Right-click: change volume by -= 20% (mod 100%)
Scroll: change volume slowly (clamp 100%)

Block 2:

Left-click: toggle sound card
*/

func pulseaudioCallback() {
	go func() {
		MainThread <- func() {
			sink := paClient.Sinks()[0]
			paVolume = int(sink.GetVolume()*120+1) / 2
			paDevice = getPaDevice()
			Redraw()
		}
	}()
}

func setPaVolume(volume int) {
	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}
	paVolume = volume
	paClient.Sinks()[0].SetVolume(float32(paVolume) / 60)
}

func getPaDevice() string {
	cards := paClient.Cards()
	if len(cards) == 0 {
		return paDevice
	}
	card := cards[0]
	for _, port := range card.Ports {
		if port.Available == pulseaudio.PortAvailableNo {
			continue
		}
		for _, profile := range port.Profiles {
			if profile.Name == card.ActiveProfile.Name {
				return port.Name
			}
		}
	}
	return card.ActiveProfile.Name
}

// BUG(mrogalski): PulseAudio devices don't include external sound cards (for example Bluetooth speakers)
func togglePaDevice() {
	card := paClient.Cards()[0]
	var niceProfiles []*pulseaudio.ProfileInfo
	for _, profile := range card.Profiles {
		if !profile.Available {
			continue
		}
		if profile.SourceCount > 0 {
			continue
		}
		niceProfiles = append(niceProfiles, profile)
	}
	// filter secondary profiles for ports
	for _, port := range card.Ports {
		present := false
		for _, profile := range port.Profiles {
			for i := 0; i < len(niceProfiles); i++ {
				current := niceProfiles[i]
				if current.Name == profile.Name {
					if present {
						niceProfiles = append(niceProfiles[:i], niceProfiles[i+1:]...)
					} else {
						present = true
					}
				}
			}
		}
	}
	for i, profile := range niceProfiles {
		if profile.Name == card.ActiveProfile.Name {
			next := niceProfiles[(i+1)%len(niceProfiles)]
			next.Activate()
			return
		}
	}
	niceProfiles[0].Activate()
}
