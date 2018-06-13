package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
)

const TrayIconSize = BarHeight

var trayClients []TrayClient

type SysTray struct{}

func (SysTray) Init() error {
	selwin, err := xwindow.Generate(X)
	if err != nil {
		return err
	}
	pixel := X.Screen().BlackPixel
	visualId := X.Screen().RootVisual
	colormap := X.Screen().DefaultColormap
	selmask := xproto.CwBackPixel | xproto.CwBorderPixel | xproto.CwOverrideRedirect | xproto.CwColormap
	selwin.Create(X.RootWin(), -1, -1, 1, 1, selmask, pixel, pixel, 1, uint32(colormap))

	SetProp(selwin.Id, "_NET_SYSTEM_TRAY_ORIENTATION", xproto.AtomCardinal, uint32(Atom("_NET_SYSTEM_TRAY_ORIENTATION_HORIZ")))
	SetProp(selwin.Id, "_NET_SYSTEM_TRAY_VISUAL", xproto.AtomVisualid, uint32(visualId))
	SetProp(selwin.Id, "_NET_SYSTEM_TRAY_COLORS", xproto.AtomCardinal, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff)

	trayAtom := Atom("_NET_SYSTEM_TRAY_S0")

	xproto.SetSelectionOwner(X.Conn(), selwin.Id, trayAtom, X.TimeGet())
	selectionReply, err := xproto.GetSelectionOwner(X.Conn(), trayAtom).Reply()
	if err != nil {
		return err
	}
	if selectionReply.Owner != selwin.Id {
		return fmt.Errorf("Maybe another tray is running? %d %d", selectionReply.Owner, selwin.Id)
	}

	data := make([]uint32, 20)
	data[0] = uint32(X.TimeGet())
	data[1] = uint32(trayAtom)
	data[2] = uint32(selwin.Id)
	ev := xproto.ClientMessageEvent{
		Sequence: 0,
		Format:   32,
		Window:   X.RootWin(),
		Type:     Atom("MANAGER"),
		Data:     xproto.ClientMessageDataUnionData32New(data),
	}
	err = xevent.SendRootEvent(X, ev, 0xffffff)
	if err != nil {
		return err
	}

	go func() {
		ok := true
		for ok {
			event, err := X.Conn().WaitForEvent()
			if err != nil {
				fmt.Println("Weird error", err)
				continue
			}
			MainThread <- func() {
				HandleXgbEvent(event)
			}
		}
	}()

	return nil
}

func (SysTray) Draw(x int) {
	for _, client := range trayClients {
		//xproto.ClearArea(X.Conn(), true, client.Window, 0, 0, 60, 60)
		if client.CurrentX != x {
			xproto.ConfigureWindow(X.Conn(), client.Window, xproto.ConfigWindowX, []uint32{uint32(x)})
		}
		x += TrayIconSize
	}
}

func (SysTray) Width() int {
	return len(trayClients) * TrayIconSize
}

type TrayClient struct {
	Window   xproto.Window
	Map      bool
	CurrentX int
}

func HandleXgbEvent(event xgb.Event) {
	switch e := event.(type) {
	case xproto.EnterNotifyEvent:
		StartPumping()
	case xproto.LeaveNotifyEvent:
		StopPumping()
		sdl.PumpEvents()
	case xproto.ExposeEvent:
		sdl.PumpEvents()
	case xproto.ClientMessageEvent:
		HandleClientMessage(e)
	case xproto.ReparentNotifyEvent:
	case xproto.ConfigureNotifyEvent:
	case xproto.MapNotifyEvent:
	case xproto.UnmapNotifyEvent:
	case xproto.ResizeRequestEvent:
		HandleResizeRequest(e)
	case xproto.PropertyNotifyEvent:
		HandlePropertyNotify(e)
	case xproto.DestroyNotifyEvent:
		HandleDestroyNotify(e)
	default:
		fmt.Println("It's not a ClientMessage:", event.Bytes()[0])
	}
}

func KickTrayClients() {
	fmt.Println("Kicking tray clients")
	for _, client := range trayClients {
		xproto.UnmapWindow(X.Conn(), client.Window)
		xproto.ReparentWindow(X.Conn(), client.Window, X.RootWin(), 0, 0)
	}
}

func SetProp(window xproto.Window, property string, Type xproto.Atom, data ...uint32) error {
	propertyAtom := Atom(property)
	bs := make([]byte, len(data)*4)
	for i := 0; i < len(data); i++ {
		binary.LittleEndian.PutUint32(bs[i*4:], data[i])
	}
	xproto.ChangeProperty(X.Conn(), xproto.PropModeReplace, window, propertyAtom, Type, 32, uint32(len(data)), bs)
	return nil
}

func Atom(name string) xproto.Atom {
	atom, err := xprop.Atom(X, name, false)
	if err != nil {
		fmt.Println("No such atom:", name, err)
		return 0
	}
	return atom
}

func QueryXembedInfo(w xproto.Window) {
	i := FindTrayClient(w)
	if i < 0 {
		i = len(trayClients)
		trayClients = append(trayClients, TrayClient{Window: w, Map: true, CurrentX: -1})
	}
	xi, err := xprop.GetProperty(X, w, "_XEMBED_INFO")
	if err == nil {
		data := xi.Value
		//xeVersion = binary.LittleEndian.Uint32(data)
		trayClients[i].Map = (binary.LittleEndian.Uint32(data[4:]) & 1) != 0
	}
}

func FindTrayClient(window xproto.Window) int {
	for i, client := range trayClients {
		if client.Window == window {
			return i
		}
	}
	return -1
}

func HandleResizeRequest(e xproto.ResizeRequestEvent) {
	i := FindTrayClient(e.Window)
	if i < 0 {
		return
	}
	WindowWidth, _ := W.GetSize()
	update := xproto.ConfigureNotifyEvent{
		Event:            e.Window,
		Window:           e.Window,
		X:                int16(WindowWidth) - int16((i+1)*TrayIconSize),
		Y:                0,
		Width:            TrayIconSize,
		Height:           TrayIconSize,
		BorderWidth:      0,
		AboveSibling:     xproto.WindowNone,
		OverrideRedirect: false,
	}
	xproto.SendEventChecked(X.Conn(), false, e.Window, xproto.EventMaskStructureNotify, string(update.Bytes()))
}

func HandleClientMessage(msg xproto.ClientMessageEvent) {
	fmt.Println("Handling ClientMessage")
	if msg.Format != 32 {
		fmt.Println("Wrong format of ClientMessage")
		return
	}
	if msg.Type != Atom("_NET_SYSTEM_TRAY_OPCODE") {
		fmt.Println("Not a _NET_SYSTEM_TRAY_OPCODE!")
		return
	}
	op := msg.Data.Data32[1]
	const SYSTEM_TRAY_REQUEST_DOCK = 0
	//const SYSTEM_TRAY_BEGIN_MESSAGE = 1
	//const SYSTEM_TRAY_CANCEL_MESSAGE = 2
	if op != SYSTEM_TRAY_REQUEST_DOCK {
		fmt.Println("Not a REQUEST_DOCK message!")
		return
	}
	client := xproto.Window(msg.Data.Data32[2])
	mask := uint32(xproto.CwEventMask)
	values := []uint32{xproto.EventMaskPropertyChange | xproto.EventMaskStructureNotify | xproto.EventMaskResizeRedirect}
	xproto.ChangeWindowAttributes(X.Conn(), client, mask, values)

	xproto.ReparentWindow(X.Conn(), client, XW, 20, 0)

	mask16 := uint16(xproto.ConfigWindowWidth | xproto.ConfigWindowHeight)
	values = []uint32{TrayIconSize, TrayIconSize}
	xproto.ConfigureWindow(X.Conn(), client, mask16, values)
	//SetProp(client, "_NET_WM_BYPASS_COMPOSITOR", xproto.AtomCardinal, 1)

	QueryXembedInfo(client)

	const XEMBED_EMBEDDED_NOTIFY = 0
	data := make([]uint32, 20)
	data[0] = uint32(X.TimeGet())
	data[1] = uint32(XEMBED_EMBEDDED_NOTIFY)
	data[2] = 0
	data[3] = uint32(XW)
	data[4] = 0 // xembed version
	ev := xproto.ClientMessageEvent{
		Sequence: 0,
		Format:   32,
		Window:   client,
		Type:     Atom("_XEMBED"),
		Data:     xproto.ClientMessageDataUnionData32New(data),
	}
	xproto.SendEvent(X.Conn(), false, client, xproto.EventMaskNoEvent, string(ev.Bytes()))

	xproto.ChangeSaveSet(X.Conn(), xproto.SetModeInsert, client)

	ConfigureTrayClients()
}

func HandleDestroyNotify(e xproto.DestroyNotifyEvent) {
	fmt.Println("Handling DestroyNotify")
	i := FindTrayClient(e.Window)
	if i < 0 {
		fmt.Println("DestroyNotify for an unknown tray client", e.Window)
		return
	}
	trayClients = append(trayClients[:i], trayClients[i+1:]...)
	ConfigureTrayClients()
}

// BUG(mrogalski): Tray icons don't refresh the background correctly.
func ConfigureTrayClients() {
	for _, client := range trayClients {
		xproto.UnmapWindow(X.Conn(), client.Window)
	}
	Redraw()
	time.Sleep(time.Millisecond * 17)
	for _, client := range trayClients {
		if client.Map {
			xproto.MapWindow(X.Conn(), client.Window)
		}
	}
}

func HandlePropertyNotify(e xproto.PropertyNotifyEvent) {
	//atomName, _ := xproto.GetAtomName(X.Conn(), e.Atom).Reply().Name
	if e.Atom != Atom("_XEMBED_INFO") {
		//
		//fmt.Println("PropertyNotify for unknown atom:", atomName.Name)
		return
	}
	if e.State != xproto.PropertyNewValue {
		fmt.Println("PropertyNotify can't handle property deletions!")
		return
	}
	QueryXembedInfo(e.Window)
	ConfigureTrayClients()
}
