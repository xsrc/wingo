package main

import (
    "log"
    "os"
    "runtime/pprof"
)

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/ewmh"
    "github.com/BurntSushi/xgbutil/keybind"
    "github.com/BurntSushi/xgbutil/mousebind"
    "github.com/BurntSushi/xgbutil/xevent"
)

// global variables!
var X *xgbutil.XUtil
var WM *state
var ROOT *window
var CONF *conf
var THEME *theme

func main() {
    var err error

    f, err := os.Create("zzz.prof")
    if err != nil {
        log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    X, err = xgbutil.Dial("")
    if err != nil {
        logError.Println(err)
        logError.Println("Error connecting to X, quitting...")
        return
    }
    defer X.Conn().Close()

    // Create a root window abstraction and load its geometry
    ROOT = newWindow(X.RootWin())
    _, err = ROOT.geometry()
    if err != nil {
        logError.Println("Could not get ROOT window geometry because: %v", err)
        logError.Println("Cannot continue. Quitting...")
        return
    }

    // Load configuration
    err = loadConfig()
    if err != nil {
        logError.Println(err)
        logError.Println("No configuration found. Quitting...")
        return
    }

    // Load theme
    err = loadTheme()
    if err != nil {
        logError.Println(err)
        logError.Println("No theme configuration found. Quitting...")
        return
    }

    // Create WM state
    WM = newState()

    // Set supported atoms
    ewmh.SupportedSet(X, []string{"_NET_WM_ICON"})

    // Allow key and mouse bindings to do their thang
    keybind.Initialize(X)
    mousebind.Initialize(X)

    // Attach all global key bindings
    attachAllKeys()

    // Attach all root mouse bindings
    rootMouseConfig()

    // Setup some cursors we use
    setupCursors()

    // Listen to Root. It is all-important.
    ROOT.listen(xgb.EventMaskPropertyChange |
                xgb.EventMaskSubstructureNotify |
                xgb.EventMaskSubstructureRedirect |
                xgb.EventMaskButtonPress)

    // Oblige map request events
    xevent.MapRequestFun(clientMapRequest).Connect(X, X.RootWin())

    // Oblige configure requests from windows we don't manage.
    xevent.ConfigureRequestFun(configureRequest).Connect(X, X.RootWin())

    xevent.Main(X)
}

