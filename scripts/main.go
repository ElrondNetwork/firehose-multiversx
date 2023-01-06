package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli"
)

var (
	log        = logger.GetOrCreate("checker")
	marshaller = &marshal.GogoProtoMarshalizer{}
	hasher     = blake2b.NewBlake2b()
)

var checkMetaFlag = cli.BoolFlag{
	Name:  "check-meta",
	Usage: "Boolean flag to specify if checker should be used for meta blocks or shard blocks",
}

const (
	proxyPem         = "testnet/testnet-local/sandbox/proxy/config/walletKey.pem"
	esdtIssueAddress = "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"
	proxyUrl         = "http://127.0.0.1:7950"
	txGasLimit       = 55141500
)

func main() {
	app := cli.NewApp()
	app.Name = "Tool to check data integration validity into firehose ingestion process"
	app.Usage = "This tool only works if a local testnet and a firehose node are started. See integration-test.sh and local-testnet.sh scripts"
	app.Flags = []cli.Flag{checkMetaFlag}
	app.Action = func(c *cli.Context) error {
		checkMeta := c.GlobalBool(checkMetaFlag.Name)
		return startProcess(checkMeta)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
		return
	}
}

func startProcess(checkMeta bool) error {
	address, privateKey, err := getAddressAndSK(proxyPem)
	if err != nil {
		return err
	}

	txHash, err := sendIssueESDTTx(address, privateKey)
	if err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf("%s/transaction/%s?withResults=true", proxyUrl, txHash))
	if err != nil {
		return err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Warn("could not close response body", "error", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if checkMeta {
		return checkMetaBlock(string(body), txHash)
	}

	hyperBlockNonce := gjson.Get(string(body), "data.transaction.hyperblockNonce").Uint()
	return checkShardBlock(hyperBlockNonce, address.AddressAsBech32String(), txHash)
}
