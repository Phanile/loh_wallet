package core

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"time"
)

type GUI struct {
	wallet      *Wallet
	address     fyne.CanvasObject
	App         fyne.App
	window      fyne.Window
	balance     *canvas.Text
	history     *widget.List
	toEntry     *widget.Entry
	amountEntry *widget.Entry
}

func NewGUI() *GUI {
	myApp := app.NewWithID("loh.wallet")
	prefs := myApp.Preferences()
	window := myApp.NewWindow("Loh Wallet")

	wallet := NewWallet(prefs)

	balanceText := canvas.NewText(fmt.Sprintf("$%.2f USD", float64(wallet.Balance)/1.0), color.White)
	balanceText.TextSize = 45
	balanceText.TextStyle = fyne.TextStyle{Bold: true}

	gui := &GUI{
		wallet:  wallet,
		address: createAddressHeader(wallet.Address.String()),
		App:     myApp,
		window:  window,
		balance: balanceText,
		history: widget.NewList(
			func() int { return 0 },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(id widget.ListItemID, item fyne.CanvasObject) {},
		),
	}

	gui.initUI()
	return gui
}

func (g *GUI) initUI() {
	g.toEntry = widget.NewEntry()
	g.toEntry.SetPlaceHolder("Recipient Address (hex)")
	g.amountEntry = widget.NewEntry()
	g.amountEntry.SetPlaceHolder("Amount")

	form := container.NewVBox(
		g.address,
		g.balance,
		widget.NewLabel("Send Transaction"),
		g.toEntry,
		g.amountEntry,
		widget.NewLabel("Transaction History"),
		g.history,
	)

	g.window.SetContent(form)
	g.window.Resize(fyne.NewSize(400, 600))
	g.window.Show()
}

func createAddressHeader(address string) fyne.CanvasObject {
	addressBox := container.NewHBox()

	addressLabel := widget.NewLabel(address)
	addressLabel.Alignment = fyne.TextAlignCenter
	addressLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), nil)

	copyBtn.OnTapped = func() {
		clipboard := fyne.CurrentApp().Clipboard()
		clipboard.SetContent(address)

		copyBtn.SetIcon(theme.ConfirmIcon())
		copyBtn.Importance = widget.MediumImportance
		copyBtn.Refresh()

		go func() {
			time.Sleep(1 * time.Second)
			fyne.DoAndWait(func() {
				copyBtn.SetIcon(theme.ContentCopyIcon())
				canvas.Refresh(copyBtn)
			})
		}()
	}

	addressBox.Add(container.NewCenter(addressLabel))
	addressBox.Add(copyBtn)

	return container.NewVBox(
		container.NewCenter(addressBox),
	)
}
