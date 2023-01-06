package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"

	mvxcore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/alteredAccount"
	"github.com/ElrondNetwork/elrond-go-core/data/firehose"
	"github.com/tidwall/gjson"
)

func checkShardBlock(hyperBlockNonce uint64, address string, txHash string) error {
	log.Info("checking shard block...")

	resp, err := http.Get(fmt.Sprintf("%s/hyperblock/by-nonce/%d", proxyUrl, hyperBlockNonce))
	if err != nil {
		return err
	}

	log.Info("checking data from hyperblock", "hyperBlockNonce", hyperBlockNonce)
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

	shardBlocks := gjson.Get(string(body), "data.hyperblock.shardBlocks").Array()
	if len(shardBlocks) != 1 {
		return fmt.Errorf("checkShardHeader: should only have one shard, but got %d", len(shardBlocks))
	}

	shardBlockNonce := shardBlocks[0].Get("nonce").Uint()
	multiversxBlock, err := getBlockFromStorage(shardBlockNonce)
	if err != nil {
		return err
	}

	err = checkShardHeader(multiversxBlock.MultiversxBlock, shardBlocks)
	if err != nil {
		return err
	}

	apiTxs := gjson.Get(string(body), "data.hyperblock.transactions").Array()
	err = checkShardTxs(apiTxs, multiversxBlock.MultiversxBlock.Transactions, txHash)
	if err != nil {
		return err
	}

	return checkShardAlteredAccounts(multiversxBlock.MultiversxBlock.AlteredAccounts, address)
}

func checkShardHeader(multiversxBlock *firehose.FirehoseBlock, shardBlocks []gjson.Result) error {
	shardBlockHash := shardBlocks[0].Get("hash").String()
	return checkHeader(
		multiversxBlock,
		shardBlockHash,
		map[mvxcore.HeaderType]struct{}{
			mvxcore.ShardHeaderV1: {},
			mvxcore.ShardHeaderV2: {},
		})
}

func checkShardTxs(apiTxs []gjson.Result, transactions map[string]*firehose.TxWithFee, txHash string) error {
	log.Info("checking shard txs...")

	err := checkApiTxExists(apiTxs, txHash)
	if err != nil {
		return err
	}

	numIndexedTxs := len(transactions)
	if numIndexedTxs != 1 {
		return fmt.Errorf("checkShardTxs: expected only one sent tx, got %d", numIndexedTxs)
	}

	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return err
	}

	protocolTx, found := transactions[string(txHashBytes)]
	if !found {
		return fmt.Errorf("checkShardTxs: could not find expected indexed tx hash: %s", apiTxs)
	}

	txProtocolHash, err := mvxcore.CalculateHash(marshaller, hasher, protocolTx.Transaction)
	if err != nil {
		return err
	}

	txProtocolHexHash := hex.EncodeToString(txProtocolHash)
	if txProtocolHexHash != txHash {
		return fmt.Errorf("checkShardTxs: invalid computed tx hash, expected: %s, got %s", txHash, txProtocolHexHash)
	}

	initialPaidFeed := protocolTx.FeeInfo.GetInitialPaidFee()
	expectedInitialPaidFee := big.NewInt(txGasLimit * 1000000000)
	if initialPaidFeed.Cmp(expectedInitialPaidFee) != 0 {
		return fmt.Errorf("checkShardTxs: invalid initial tx paid fee, expected: %s, got %s",
			expectedInitialPaidFee.String(),
			initialPaidFeed.String(),
		)
	}

	return nil
}

func checkApiTxExists(apiTxs []gjson.Result, txHash string) error {
	for _, apiTx := range apiTxs {
		apiTxHash := apiTx.Get("hash").String()
		if apiTxHash == txHash {
			return nil
		}
	}

	return fmt.Errorf("could not find generated tx hash: %s in api hyperBlock", txHash)
}

func checkShardAlteredAccounts(alteredAccounts []*alteredAccount.AlteredAccount, expectedAddress string) error {
	log.Info("checking shard altered accounts...")

	numAlteredAccounts := len(alteredAccounts)
	if numAlteredAccounts != 1 {
		return fmt.Errorf("checkShardAlteredAccounts: expected only one altered account: %s, got: %d", expectedAddress, numAlteredAccounts)
	}

	acc := alteredAccounts[0]
	if acc.Address != expectedAddress {
		return fmt.Errorf("checkShardAlteredAccounts: expected altered account: %s, got: %s", expectedAddress, acc.Address)
	}

	if !(acc.Nonce > 0) {
		return fmt.Errorf("checkShardAlteredAccounts: expected %s address' nonce > 0", expectedAddress)
	}

	return nil
}
