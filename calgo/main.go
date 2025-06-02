package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/emersion/go-autostart"
	"github.com/floholz/chrome-app-launcher/calgo/calgo"
	"github.com/getlantern/systray"
	"log"
	"os"
)

//go:embed assets/logo.ico
var logoData []byte

//go:embed assets/logout.ico
var iconLogoutData []byte

//go:embed assets/square.ico
var iconDataSquare []byte

//go:embed assets/square-check.ico
var iconDataCheck []byte

var autostartApp *autostart.App

func main() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	autostartApp = &autostart.App{
		Name:        "calgo.startup",
		DisplayName: "CalGo",
		Exec:        []string{ex},
	}

	systray.Run(onReady, nil)
}

func onReady() {
	systray.SetIcon(logoData)
	systray.SetTitle("CalGo")
	systray.SetTooltip("Chrome App Launcher - Server")

	mStartup := systray.AddMenuItemCheckbox("Run on startup", "Launch CalGo server on startup", !autostartApp.IsEnabled())
	setStartupIcon(mStartup)

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit CalGo server")
	mQuit.SetIcon(iconLogoutData)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := calgo.Start(ctx)
		if err != nil {
			log.Printf("calgo server exited: %v", err)
		}
	}()

	for {
		select {
		case <-mStartup.ClickedCh:
			toggleStartup(mStartup)
		case <-mQuit.ClickedCh:
			cancel()
			systray.Quit()
		}
	}
}

func toggleStartup(item *systray.MenuItem) {
	setStartupIcon(item)
	fmt.Printf("Startup checked: %v\n", item.Checked())
	if autostartApp.IsEnabled() {
		err := autostartApp.Disable()
		if err != nil {
			panic(err)
		}
	} else {
		err := autostartApp.Enable()
		if err != nil {
			panic(err)
		}
	}
}

func setStartupIcon(item *systray.MenuItem) {
	if item.Checked() {
		item.Uncheck()
		item.SetIcon(iconDataSquare)
	} else {
		item.Check()
		item.SetIcon(iconDataCheck)
	}
}
