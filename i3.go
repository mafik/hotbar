package main

import (
	"fmt"

	"github.com/mdirkse/i3ipc-go"
)

var i3Socket *i3ipc.IPCSocket
var i3Workspaces []i3ipc.Workspace
var i3Menu = false

type I3Workspaces struct{}

func (I3Workspaces) Init() (err error) {
	i3Socket, err = i3ipc.GetIPCSocket()
	if err != nil {
		return err
	}
	i3ipc.StartEventListener()
	updateWorkspaces()
	go listenI3()
	return nil
}

func (I3Workspaces) Draw(x int) {
	for _, w := range i3Workspaces {
		var fill float32 = 0
		if w.Focused {
			fill = 1
		}
		Draw(x, "workspace", fill, Icon(fmt.Sprint(w.Num%10)))
		x += Width("workspace", Icon(""))
	}
}

func (I3Workspaces) Width() int {
	return Width("workspace", Icon("")) * len(i3Workspaces)
}

func (I3Workspaces) LeftClick(x int) bool {
	w := i3Workspaces[x/Width("workspace", Icon(""))]
	i3Socket.Command("workspace " + fmt.Sprint(w.Num))
	return false
}

type i3Command struct {
	icon    string
	command string
}

var i3Commands = []i3Command{
	i3Command{"left", "move left"},
	i3Command{"up", "move up"},
	i3Command{"down", "move down"},
	i3Command{"right", "move right"},
	i3Command{"horizontal", "split h"},
	i3Command{"vertical", "split v"},
	i3Command{"fullscreen", "fullscreen toggle"},
	i3Command{"floating", "floating toggle"},
	i3Command{"parent", "focus parent"},
	i3Command{"kill", "kill"},
}

type I3Menu struct{}

func (I3Menu) blocks() []interface{} {
	blocks := []interface{}{Icon("menu")}
	if i3Menu {
		for _, command := range i3Commands {
			blocks = append(blocks, Icon(command.icon))
		}
	}
	return blocks
}

func (m I3Menu) Draw(x int) {
	Draw(x, "menu", 0, m.blocks()...)
}

func (m I3Menu) Width() int {
	return Width("menu", m.blocks()...)
}

func (I3Menu) LeftClick(x int) bool {
	if x < T.IconSize {
		i3Menu = !i3Menu
		return true
	} else {
		x -= T.IconSize
		i3Commands[x/T.IconSize].Run()
		return false
	}
}

func (c i3Command) Run() {
	i3Socket.Command(c.command)
}

func updateWorkspaces() {
	ret, err := i3Socket.GetWorkspaces()
	if err != nil {
		fmt.Println("Error when querying i3 workspaces:", err)
	}
	i3Workspaces = ret
	Redraw()
}

func listenI3() {
	updates, err := i3ipc.Subscribe(i3ipc.I3WorkspaceEvent)
	if err != nil {
		fmt.Println("Error when listening to i3 updates")
		return
	}
	for {
		_, ok := <-updates
		if !ok {
			fmt.Println("i3 update channel closed!")
			break
		}
		MainThread <- updateWorkspaces
	}
}
