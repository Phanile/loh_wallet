package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"github.com/Phanile/uretra_network/crypto"
	"github.com/Phanile/uretra_network/types"
	"io"
	"net/http"
)

const (
	baseApiServerUrl = "http://192.168.3.2:3229"
	coinValueUSD     = 3
)

type GetBalanceResponse struct {
	Balance uint64 `json:"balance"`
	Error   string `json:"error,omitempty"`
}

type Wallet struct {
	PrivateKey crypto.PrivateKey
	PublicKey  crypto.PublicKey
	Address    types.Address
	Balance    uint64
	Count      uint32
}

func NewWallet(prefs fyne.Preferences) *Wallet {
	if prefs == nil {
		panic("preferences cannot be nil")
	}

	if addrHex := prefs.String("wallet_address"); addrHex != "" {
		addrBytes, errDecode := hex.DecodeString(addrHex)

		if errDecode != nil {
			panic(errDecode)
		}

		w := &Wallet{
			Address: types.AddressFromBytes(addrBytes),
		}

		balance, errBalance := w.GetFormattedBalance()

		if errBalance != nil {
			fmt.Println(errBalance)
		}

		w.Balance = balance

		return w
	}

	wallet := &Wallet{
		PrivateKey: crypto.GeneratePrivateKey(),
	}

	wallet.PublicKey = wallet.PrivateKey.PublicKey()
	wallet.Address = wallet.PublicKey.Address()
	prefs.SetString("wallet_address", wallet.Address.String())

	return wallet
}

func (w *Wallet) GetFormattedBalance() (uint64, error) {
	balance, err := w.fetchBalanceFromAPI()
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %v", err)
	}

	usdAmount := uint64(balance) * coinValueUSD

	return usdAmount, nil
}

func (w *Wallet) fetchBalanceFromAPI() (uint64, error) {
	url := fmt.Sprintf("%s/getBalance/%s", baseApiServerUrl, w.Address.String())

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status: %s", resp.Status)
	}

	body, errBody := io.ReadAll(resp.Body)
	if errBody != nil {
		return 0, fmt.Errorf("failed to read response: %v", errBody)
	}

	var apiResponse GetBalanceResponse
	if errUnmarshall := json.Unmarshal(body, &apiResponse); errUnmarshall != nil {
		return 0, fmt.Errorf("failed to parse response: %v", errUnmarshall)
	}

	if apiResponse.Error != "" {
		return 0, fmt.Errorf("API error: %s", apiResponse.Error)
	}

	return apiResponse.Balance, nil
}
