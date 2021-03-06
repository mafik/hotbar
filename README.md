# hotbar
Probably the most customizable system bar in the world.

Download:
```shell
go get mrogalski.eu/go/hotbar
```

If you do not have the go command on your system, you need to [Install Go](http://golang.org/doc/install) first

* * *
Probably the most customizable system bar in the world. Replacement
for i3bar, xfce4-panel, etc.

If you've ever been annoyed by quirks or limited customizability of
default system bars, `hotbar` is for you.It's distinguishing
feature is the ability to configure every aspect of drawing
directly in the source code. So go ahead and tweak it to your
liking!

![Dusk theme](screenshot1.png?raw=true "Dusk theme")
![Neon theme](screenshot2.png?raw=true "Neon theme")

# Features
* ability to tweak drawing routines directly in the source code
* modules can draw custom GFX at 60fps with SDL2 & OpenGL
* low power usage - only one redraw / minute
* tweaking backlight or sound volume with mouse wheel
* all modules are usable with touchscreens

# Usage
Run it with `hotbar`.

Customize it by navigating to the source directory, and tweaking
some of the files (knowledge of the go language is not really
necessary):

```
cd ~/go/src/mrogalski.eu/go/hotbar/
```

* `modules.go` - List of modules. Allows you to add, disable or reorder the modules.
* `theme.go` - Margins, paddings, icons & fonts. This should cover 90% of customizations you might want.
* `backlight.go`, `date.go`, `time.go`, `disk.go`, `power.go`, `pulseaudio.go`, `xkb.go`, `i3.go` - Default modules. Use this to tweak their behavior or as a base when creating your own modules.
* `globals.go`, `stepper.go`, `widgets.go` - Variables and functions that may be helpful when adding new modules.
* `main.go` - Logic for starting up hotbar
* `tray.go` - Logic for displaying the system tray
* `signals.go` - Logic for handling Ctrl+C gracefully

After playing with the source, run it with:

```
go run *.go
```

If you're happy with the results, save them as the `hotbar` command
by running:

```
go install
```

# Contributing
Share your creations by *forking* this repository!

If you fixed something or added something that others may find
useful, fire up a pull request.

# Credits
* Icons8.com - for the included icons
* be5invis - for the included font



# Bugs
* PulseAudio devices don't include external sound cards (for example Bluetooth speakers)
* Tray icons don't refresh the background correctly.


* * *
Automatically generated by [autoreadme](https://github.com/jimmyfrasche/autoreadme) on 2018.06.14
