package main

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var X *xgbutil.XUtil
var XW xproto.Window
var W *sdl.Window
var R *sdl.Renderer
var Font *ttf.Font
var LineHeight int
var KeepRunning = true
var Dir string
var MainThread = make(chan func())
