package main

import (
	"context"
	_ "embed"
	"github.com/floholz/chrome-app-launcher/calgo/calgo"
	"github.com/getlantern/systray"
	"log"
)

//go:embed assets/logo.ico
var logoData []byte

//go:embed assets/logout.ico
var iconLogoutData []byte

func main() {
	systray.Run(onReady, nil)
}

func onReady() {
	systray.SetIcon(logoData)
	systray.SetTitle("CalGo")
	systray.SetTooltip("Chrome App Launcher - Server")
	mQuit := systray.AddMenuItem("Quit", "Quit CalGo server")

	// Sets the icon of a menu item. Only available on Mac and Windows.
	mQuit.SetTemplateIcon(iconLogoutData, iconLogoutData)
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
		case <-mQuit.ClickedCh:
			cancel()
			systray.Quit()
		}
	}
}
