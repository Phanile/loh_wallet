package core

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"github.com/Phanile/uretra_network/core"
	"github.com/Phanile/uretra_network/crypto"
	"github.com/Phanile/uretra_network/types"
	"io"
	"net/http"
)

const (
	baseApiServerUrl = "http://192.168.3.2:3229"
	coinValueUSD     = 4
)

var currentNonce uint64 = 1

type GetBalanceResponse struct {
	Balance uint64 `json:"balance"`
	Error   string `json:"error,omitempty"`
}

type Wallet struct {
	PrivateKey crypto.PrivateKey
	PublicKey  crypto.PublicKey
	Address    types.Address
	Balance    uint64
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

		privKeyPEM := prefs.String("wallet_private_key")
		pubKeyPEM := prefs.String("wallet_public_key")

		if privKeyPEM == "" {
			panic("wallet_private_key is empty")
		}

		if pubKeyPEM == "" {
			panic("wallet_public_key is empty")
		}

		privKey, serializePrivateKeyErr := crypto.PrivateKeyFromBytes([]byte(privKeyPEM))
		pubKey, serializePublicKeyErr := crypto.PublicKeyFromBytes([]byte(pubKeyPEM))

		if serializePrivateKeyErr != nil {
			panic(serializePrivateKeyErr)
		}

		if serializePublicKeyErr != nil {
			panic(serializePublicKeyErr)
		}

		w := &Wallet{
			Address:    types.AddressFromBytes(addrBytes),
			PrivateKey: privKey,
			PublicKey:  pubKey,
		}

		balance, errBalance := w.fetchBalanceFromAPI()

		if errBalance != nil {
			w.Balance = 0
		}

		w.Balance = balance

		return w
	}

	privKey := crypto.GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	privKeyBytes, _ := privKey.Bytes()
	pubKeyBytes, _ := pubKey.Bytes()

	prefs.SetString("wallet_address", pubKey.Address().String())
	prefs.SetString("wallet_private_key", string(privKeyBytes))
	prefs.SetString("wallet_public_key", string(pubKeyBytes))

	wallet := &Wallet{
		PrivateKey: privKey,
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		Balance:    0,
	}

	return wallet
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

func (w *Wallet) GetFormattedAddress() string {
	return w.Address.String()[:5] + "..." + w.Address.String()[len(w.Address.String())-5:]
}

func (w *Wallet) SendTransaction(address string, amount uint64) error {
	if w.Balance < amount {
		return fmt.Errorf("wallet balance too small")
	}

	addrBytes, _ := hex.DecodeString(address)

	tx := core.NewTransaction(
		nil,
		w.PublicKey,
		types.AddressFromBytes(addrBytes),
		amount,
		currentNonce,
	)

	_ = tx.Sign(w.PrivateKey)

	if !tx.Verify() {
		return fmt.Errorf("invalid signature")
	}

	currentNonce++

	buf := &bytes.Buffer{}

	err := gob.NewEncoder(buf).Encode(tx)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/tx", baseApiServerUrl)
	resp, errResp := http.Post(url, "application/octet-stream", buf)
	if errResp != nil {
		return errResp
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status: %s", resp.Status)
	}

	return nil
}
