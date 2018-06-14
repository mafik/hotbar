// Probably the most customizable system bar in the world. Replacement
// for i3bar, xfce4-panel, etc.
//
// If you've ever been annoyed by quirks or limited customizability of
// default system bars, `hotbar` is for you.It's distinguishing
// feature is the ability to configure every aspect of drawing
// directly in the source code. So go ahead and tweak it to your
// liking!
//
// ![Dusk theme](screenshot1.png?raw=true "Dusk theme")
// ![Neon theme](screenshot2.png?raw=true "Neon theme")
//
// # Features
// * ability to tweak drawing routines directly in the source code
// * modules can draw custom GFX at 60fps with SDL2 & OpenGL
// * low power usage - only one redraw / minute
// * tweaking backlight or sound volume with mouse wheel
// * all modules are usable with touchscreens
//
// # Usage
// Run it with `hotbar`.
//
// Customize it by navigating to the source directory, and tweaking
// some of the files (knowledge of the go language is not really
// necessary):
//
//   cd ~/go/src/mrogalski.eu/go/hotbar/
//
// * `modules.go` - List of modules. Allows you to add, disable or reorder the modules.
// * `theme.go` - Margins, paddings, icons & fonts. This should cover 90% of customizations you might want.
// * `backlight.go`, `date.go`, `time.go`, `disk.go`, `power.go`, `pulseaudio.go`, `xkb.go`, `i3.go` - Default modules. Use this to tweak their behavior or as a base when creating your own modules.
// * `globals.go`, `stepper.go`, `widgets.go` - Variables and functions that may be helpful when adding new modules.
// * `main.go` - Logic for starting up hotbar
// * `tray.go` - Logic for displaying the system tray
// * `signals.go` - Logic for handling Ctrl+C gracefully
//
// After playing with the source, run it with:
//
//  go run *.go
//
// If you're happy with the results, save them as the `hotbar` command
// by running:
//
//   go install
//
// # Contributing
// Share your creations by *forking* this repository!
// 
// If you fixed something or added something that others may find
// useful, fire up a pull request.
//
// # Credits
// * Icons8.com - for the included icons
// * be5invis - for the included font
package main // import "mrogalski.eu/go/hotbar"

import (
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func ForEachModule(cb func(x int, drawer Drawer)) {
	x := 0
	for _, m := range Modules {
		drawer, ok := m.(Drawer)
		if !ok {
			continue
		}
		cb(x, drawer)
		x += drawer.Width()
	}
}

func ModuleAt(x int) (int, Drawer) {
	left := 0
	for _, m := range Modules {
		drawer, ok := m.(Drawer)
		if !ok {
			continue
		}
		right := left + drawer.Width()
		if x >= left && x < right {
			return x - left, drawer
		}
		left = right
	}
	return 0, nil
}

func Redraw() {
	R.SetDrawColorArray(T.Default.Background...)
	R.Clear()

	ForEachModule(func(x int, drawer Drawer) {
		drawer.Draw(x)
	})

	R.Present()
}

type SdlEventWatcher struct{}

func (SdlEventWatcher) FilterEvent(e sdl.Event, userdata interface{}) bool {
	shouldRedraw := false
	x32, _, _ := sdl.GetMouseState()
	x := int(x32)

	switch event := e.(type) {
	case *sdl.MouseWheelEvent:
		_, drawer := ModuleAt(x)
		if drawer == nil {
			break
		}
		wheeler, ok := drawer.(Wheeler)
		if !ok {
			break
		}
		shouldRedraw = wheeler.Wheel(int(event.Y))
	case *sdl.TouchFingerEvent:
		w, _ := W.GetSize()
		x = int(event.X * float32(w))
		if event.Type != sdl.FINGERDOWN {
			break
		}
		relX, drawer := ModuleAt(x)
		if drawer == nil {
			break
		}
		clicker, ok := drawer.(LeftClicker)
		if !ok {
			break
		}
		shouldRedraw = clicker.LeftClick(relX)
	case *sdl.MouseButtonEvent:
		if event.Button == 1 {
			relX, drawer := ModuleAt(x)
			if drawer == nil {
				break
			}
			clicker, ok := drawer.(LeftClicker)
			if !ok {
				break
			}
			shouldRedraw = clicker.LeftClick(relX)
		} else if event.Button == 3 {
			relX, drawer := ModuleAt(x)
			if drawer == nil {
				break
			}
			clicker, ok := drawer.(RightClicker)
			if !ok {
				break
			}
			shouldRedraw = clicker.RightClick(relX)
		}
	case *sdl.WindowEvent:
		switch event.Event {
		case sdl.WINDOWEVENT_SHOWN:
			shouldRedraw = true
		case sdl.WINDOWEVENT_EXPOSED:
			shouldRedraw = true
		case sdl.WINDOWEVENT_SIZE_CHANGED:
			shouldRedraw = true
		default:
		}
	default:
		fmt.Println("Got event:", reflect.TypeOf(e))
		return true
	}
	if shouldRedraw {
		Redraw()
	}
	return true
}

var pumpTicker *time.Ticker

func StartPumping() {
	StopPumping()
	pumpTicker = time.NewTicker(time.Millisecond * 16)
	go func(t *time.Ticker) {
		ok := true
		for ok {
			MainThread <- sdl.PumpEvents
			_, ok = <-t.C
		}
	}(pumpTicker)
}

func StopPumping() {
	if pumpTicker != nil {
		pumpTicker.Stop()
		pumpTicker = nil
	}
}

type Initializer interface {
	Init() error
}

type Refresher interface {
	Refresh()
}

type LeftClicker interface {
	Drawer
	LeftClick(x int) bool
}

type RightClicker interface {
	Drawer
	RightClick(x int) bool
}

type Wheeler interface {
	Drawer
	Wheel(delta int) bool
}

type Drawer interface {
	Width() int
	Draw(x int)
}

func initGlobals() error {
	if img.Init(img.INIT_PNG) != img.INIT_PNG {
		return fmt.Errorf("Couldn't initialize SDL IMG")
	}
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return err
	}
	if err := ttf.Init(); err != nil {
		return err
	}
	mode, err := sdl.GetCurrentDisplayMode(0)
	if err != nil {
		return err
	}
	flags := uint32(sdl.WINDOW_HIDDEN | sdl.WINDOW_BORDERLESS)
	W, err = sdl.CreateWindow("mbar", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, mode.W, BarHeight, flags)
	if err != nil {
		return err
	}
	R, err = sdl.CreateRenderer(W, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		return err
	}
	X, err = xgbutil.NewConn()
	if err != nil {
		return err
	}
	info, err := W.GetWMInfo()
	if err != nil {
		return err
	}
	XW = xproto.Window(info.GetX11Info().Window)
	SetProp(XW, "_NET_WM_WINDOW_TYPE", xproto.AtomAtom, uint32(Atom("_NET_WM_WINDOW_TYPE_DOCK")))
	SetProp(XW, "_NET_WM_STRUT_PARTIAL", xproto.AtomCardinal, 0, 0, 0, BarHeight, 0, 0, 0, 0, 0, 0, uint32(mode.W-1))
	SetProp(XW, "_NET_WM_BYPASS_COMPOSITOR", xproto.AtomCardinal, 2)
	xproto.ChangeWindowAttributes(X.Conn(), XW, xproto.CwEventMask,
		[]uint32{xproto.EventMaskEnterWindow | xproto.EventMaskLeaveWindow | xproto.EventMaskExposure})
	goenv, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return err
	}
	Dir = strings.TrimSpace(string(goenv)) + "/src/mrogalski.eu/go/hotbar"
	Font, err = ttf.OpenFont(Dir+"/"+T.FontPath, T.FontHeight)
	if err != nil {
		return err
	}
	_, LineHeight, _ = Font.SizeUTF8(" ")
	sdl.AddEventWatch(SdlEventWatcher{}, nil)
	sdl.EventState(sdl.MOUSEBUTTONDOWN, sdl.ENABLE)
	sdl.EventState(sdl.MOUSEWHEEL, sdl.ENABLE)
	sdl.EventState(sdl.MOUSEBUTTONUP, sdl.DISABLE)
	sdl.EventState(sdl.MOUSEMOTION, sdl.DISABLE)
	return sdl.GetError()
}

func RefreshModules() {
	for _, m := range Modules {
		if refresher, ok := m.(Refresher); ok {
			refresher.Refresh()
		}
	}
}

func main() {
	if err := initGlobals(); err != nil {
		fmt.Println("Couldn't initialize bar:", err)
		return
	}

	for _, m := range Modules {
		if initializer, ok := m.(Initializer); ok {
			if err := initializer.Init(); err != nil {
				fmt.Println("Couldn't initialize module ", reflect.TypeOf(m), ":", err)
				return
			}
		}
	}

	RefreshModules()

	W.Show()

	go func() {
		for {
			now := time.Now()
			nextMinute := now.Add(time.Minute)
			nextMinuteStart := time.Date(
				nextMinute.Year(),
				nextMinute.Month(),
				nextMinute.Day(),
				nextMinute.Hour(),
				nextMinute.Minute(),
				0,
				0,
				nextMinute.Location())
			time.Sleep(nextMinuteStart.Sub(now))
			MainThread <- func() {
				RefreshModules()
				Redraw()
			}
		}
	}()

	for KeepRunning {
		(<-MainThread)()
	}
	KickTrayClients()
	W.Destroy()
}
