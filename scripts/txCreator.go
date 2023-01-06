package main

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain/cryptoProvider"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

var (
	suite  = ed25519.NewEd25519()
	keyGen = signing.NewKeyGenerator(suite)
)

func sendIssueESDTTx(address core.AddressHandler, privateKey []byte) (string, error) {
	args := blockchain.ArgsElrondProxy{
		ProxyURL:            proxyUrl,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	proxy, err := blockchain.NewElrondProxy(args)
	if err != nil {
		return "", err
	}

	txBuilder, err := builders.NewTxBuilder(cryptoProvider.NewSigner())
	if err != nil {
		return "", err
	}

	ti, err := interactors.NewTransactionInteractor(proxy, txBuilder)
	if err != nil {
		return "", err
	}

	networkConfig, err := proxy.GetNetworkConfig(context.Background())
	if err != nil {
		return "", err
	}

	transactionArguments, err := proxy.GetDefaultTransactionArguments(context.Background(), address, networkConfig)
	if err != nil {
		return "", err
	}

	transactionArguments.Value = "50000000000000000"
	transactionArguments.GasLimit = txGasLimit
	transactionArguments.Data = []byte("issue@4141414141@41414141@6f@01@63616e55706772616465@74727565@63616e57697065@74727565@63616e467265657a65@74727565")
	transactionArguments.RcvAddr = esdtIssueAddress

	holder, err := cryptoProvider.NewCryptoComponentsHolder(keyGen, privateKey)
	if err != nil {
		return "", err
	}

	signedTx, err := ti.ApplySignatureAndGenerateTx(holder, transactionArguments)
	if err != nil {
		return "", err
	}

	hash, err := proxy.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	log.Info("sent transaction", "tx hash", hash)

	log.Info("waiting 5 rounds...")
	time.Sleep(time.Millisecond * time.Duration(networkConfig.RoundDuration*5))

	return hash, nil
}
