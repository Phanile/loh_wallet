package core

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"strconv"
	"time"
)

type GUI struct {
	address       fyne.CanvasObject
	window        fyne.Window
	history       *widget.List
	balance       *canvas.Text
	coinContainer *fyne.Container
	coinPrice     float64
}

func NewGUI(wallet *Wallet, window fyne.Window) *GUI {
	balanceText := canvas.NewText("$"+strconv.FormatUint(wallet.Balance*coinValueUSD, 10), color.White)
	balanceText.TextSize = 45
	balanceText.TextStyle = fyne.TextStyle{Bold: true}

	gui := &GUI{
		address: createAddressHeader(wallet.GetFormattedAddress(), wallet.Address.String()),
		balance: balanceText,
		history: widget.NewList(
			func() int { return 0 },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(id widget.ListItemID, item fyne.CanvasObject) {},
		),
		coinContainer: createCoinInfoBlock(wallet.Balance),
	}

	gui.window = window

	return gui
}

func (g *GUI) InitUI(wallet *Wallet) {
	form := container.NewVBox(
		g.address,
		g.balance,
		container.NewHBox(
			createSendButton(wallet, g.window),
		),
		g.coinContainer,
		g.history,
	)

	g.window.SetContent(form)
	g.window.Resize(fyne.NewSize(400, 600))
	g.window.Show()
}

func createAddressHeader(formattedAddr, address string) fyne.CanvasObject {
	addressBox := container.NewHBox()

	addressLabel := widget.NewLabel(formattedAddr)
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

func createSendButton(wallet *Wallet, window fyne.Window) fyne.CanvasObject {
	btn := widget.NewButtonWithIcon("", theme.MailSendIcon(), func() {
		showSendModal(wallet, window)
	})

	btn.Importance = widget.HighImportance

	return container.NewStack(
		container.NewCenter(
			container.NewVBox(
				btn,
				widget.NewLabel("Отправить"),
			),
		),
	)
}

func showSendModal(wallet *Wallet, parent fyne.Window) {
	addressEntry := widget.NewEntry()
	addressEntry.SetPlaceHolder("Recipient Address")

	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("Amount")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Address", Widget: addressEntry},
			{Text: "Amount", Widget: amountEntry},
		},
	}

	confirmDialog := dialog.NewCustomConfirm(
		"Confirm Transaction",
		"Sign",
		"Cancel",
		form,
		func(sign bool) {
			if sign {
				amEntry, _ := strconv.Atoi(amountEntry.Text)

				if addressEntry.Text == "" || amountEntry.Text == "" {
					dialog.ShowInformation("Error", "Please fill all fields", parent)
					return
				}

				showSigningDialog(wallet, addressEntry.Text, uint64(amEntry), parent)
			}
		},
		parent,
	)

	confirmDialog.Resize(fyne.NewSize(400, 200))
	confirmDialog.Show()
}

func showSigningDialog(wallet *Wallet, address string, amount uint64, parent fyne.Window) {
	progress := widget.NewProgressBarInfinite()
	content := container.NewVBox(
		widget.NewLabel("Please confirm transaction in your wallet"),
		progress,
	)

	progressDialog := dialog.NewCustomWithoutButtons("Signing Transaction", content, parent)

	progressDialog.SetButtons([]fyne.CanvasObject{
		widget.NewButton("Cancel", func() {
			progressDialog.Hide()
		}),
	})

	progressDialog.Show()

	fyne.Do(func() {
		err := wallet.SendTransaction(address, amount)

		progressDialog.Hide()

		if err != nil {
			dialog.ShowError(err, parent)
		} else {
			dialog.ShowInformation("Success", "Transaction sent!", parent)
		}
	})
}

func createCoinInfoBlock(coinAmount uint64) *fyne.Container {
	coinIcon := canvas.NewImageFromResource(nil)
	coinIcon.SetMinSize(fyne.NewSize(32, 32))

	coinName := widget.NewLabel("SCMN")
	coinName.TextStyle = fyne.TextStyle{Bold: true}

	priceText := "$" + strconv.FormatUint(coinValueUSD, 10)
	coinPrice := widget.NewLabel(priceText)
	coinPrice.Alignment = fyne.TextAlignTrailing

	amountText := strconv.FormatUint(coinAmount, 10) + " SCMN"
	coinAmountLabel := widget.NewLabel(amountText)
	coinAmountLabel.Alignment = fyne.TextAlignTrailing

	rightCol := container.NewGridWithRows(2,
		coinPrice,
		coinAmountLabel,
	)

	return container.NewBorder(nil, nil,
		container.NewHBox(coinIcon, coinName),
		rightCol,
	)
}
