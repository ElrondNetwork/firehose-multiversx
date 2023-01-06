package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	mvxcore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/alteredAccount"
	"github.com/ElrondNetwork/elrond-go-core/data/firehose"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/tidwall/gjson"
)

func checkMetaBlock(apiTxResultBody string, txHash string) error {
	log.Info("checking meta block...")

	blockNum := gjson.Get(apiTxResultBody, "data.transaction.blockNonce").Uint()
	multiversxBlock, err := getBlockFromStorage(blockNum)
	if err != nil {
		return err
	}

	blockHash := gjson.Get(apiTxResultBody, "data.transaction.blockHash").String()
	err = checkHeader(
		multiversxBlock.MultiversxBlock,
		blockHash,
		map[mvxcore.HeaderType]struct{}{
			mvxcore.MetaHeader: {},
		})
	if err != nil {
		return err
	}

	scrs := gjson.Get(apiTxResultBody, "data.transaction.smartContractResults").Array()
	err = checkMetaSCRs(scrs, multiversxBlock.MultiversxBlock.SmartContractResult)
	if err != nil {
		return err
	}

	logs := gjson.Get(apiTxResultBody, "data.transaction.logs")
	err = checkMetaLogs(logs, multiversxBlock.MultiversxBlock.Logs, txHash)
	if err != nil {
		return err
	}

	return checkMetaAlteredAccounts(multiversxBlock.MultiversxBlock.AlteredAccounts)
}

func checkMetaSCRs(apiSCRs []gjson.Result, scrs map[string]*firehose.SCRWithFee) error {
	log.Info("checking meta scrs...")

	numApiSCRS := len(apiSCRs)
	numIndexedSCRS := len(scrs)
	if numApiSCRS != numIndexedSCRS {
		return fmt.Errorf("checkMetaSCRs: got %d scrs from api, but indexed %d scrs", numApiSCRS, numIndexedSCRS)
	}
	if numApiSCRS == 0 {
		return fmt.Errorf("checkMetaSCRs: expected non zero api scrs")
	}

	for _, apiSCR := range apiSCRs {
		hash := apiSCR.Get("hash").String()
		hashBytes, err := hex.DecodeString(hash)
		if err != nil {
			return err
		}

		scrFromProtocol, found := scrs[string(hashBytes)]
		if !found {
			return fmt.Errorf("checkMetaSCRs: api hash %s not found in indexed block", hash)
		}

		computedHash, err := mvxcore.CalculateHash(marshaller, hasher, scrFromProtocol.SmartContractResult)
		if err != nil {
			return err
		}

		if !bytes.Equal(computedHash, hashBytes) {
			return fmt.Errorf("checkMetaSCRs: computed scr hash: %s != received scr hash: %s",
				hex.EncodeToString(computedHash), hash)
		}
	}

	return nil
}

func checkMetaLogs(apiLog gjson.Result, logs map[string]*transaction.Log, txHash string) error {
	log.Info("checking meta logs...")

	numIndexedLogs := len(logs)
	if numIndexedLogs != 1 {
		return fmt.Errorf("checkMetaLogs: expected only one generated and indexed log, received %d", numIndexedLogs)
	}
	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return err
	}

	indexedLog, found := logs[string(txHashBytes)]
	if !found {
		return fmt.Errorf("checkMetaLogs: api tx hash %s not found in indexed logs", txHash)
	}

	apiEvents := apiLog.Get("events").Array()
	indexedEvents := indexedLog.Events
	numApiEvents := len(apiEvents)
	numIndexedAEvents := len(indexedEvents)
	if numApiEvents != numIndexedAEvents {
		return fmt.Errorf("checkMetaLogs: got %d events from api, but indexed %d events", numApiEvents, numIndexedAEvents)
	}
	if numIndexedAEvents == 0 {
		return fmt.Errorf("checkMetaLogs: expected non zero indexed events")
	}

	for _, apiEvent := range apiEvents {
		identifier := apiEvent.Get("identifier").String()
		if !contains(indexedEvents, identifier) {
			return fmt.Errorf("checkMetaLogs: api event identifier %s not found in indexed events", identifier)
		}
	}

	return nil
}

func contains(events []*transaction.Event, identifier string) bool {
	for _, event := range events {
		if string(event.Identifier) == identifier {
			return true
		}
	}

	return false
}

func checkMetaAlteredAccounts(alteredAccounts []*alteredAccount.AlteredAccount) error {
	log.Info("checking meta altered accounts...")

	numAlteredAccounts := len(alteredAccounts)
	if numAlteredAccounts != 1 {
		return fmt.Errorf("checkMetaAlteredAccounts: expected only one altered account: %s, got: %d", esdtIssueAddress, numAlteredAccounts)
	}

	acc := alteredAccounts[0]
	if acc.Address != esdtIssueAddress {
		return fmt.Errorf("checkMetaAlteredAccounts: expected altered account: %s, got: %s", esdtIssueAddress, acc.Address)
	}

	balance, castOk := big.NewInt(0).SetString(acc.Balance, 10)
	if !castOk {
		return fmt.Errorf("checkMetaAlteredAccounts: could not convert balance: %s to bigInt, address: %s", balance, esdtIssueAddress)
	}

	if balance.Cmp(big.NewInt(50000000000000000)) < 0 {
		return fmt.Errorf("checkMetaAlteredAccounts: expected %s address' balance increased after ESDT issue, but got %s", esdtIssueAddress, balance)
	}

	return nil
}
