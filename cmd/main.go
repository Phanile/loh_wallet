package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/Phanile/loh_wallet/core"
)

func main() {
	application := app.NewWithID("loh.wallet")
	window := application.NewWindow("Loh Wallet")

	prefs := application.Preferences()
	wallet := core.NewWallet(prefs)

	gui := core.NewGUI(wallet, window)
	gui.InitUI(wallet)

	application.Run()
}
