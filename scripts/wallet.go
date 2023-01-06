package main

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

func getAddressAndSK(pemPath string) (core.AddressHandler, []byte, error) {
	w := interactors.NewWallet()
	privateKey, err := w.LoadPrivateKeyFromPemFile(pemPath)
	if err != nil {
		log.Error("unable to load pem", "file", pemPath, "error", err)
		return nil, nil, err
	}

	address, err := w.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		log.Error("unable to load the address from the private key", "error", err)
		return nil, nil, err
	}

	log.Info("loaded", "address", address.AddressAsBech32String(), "file", pemPath)
	return address, privateKey, nil
}
