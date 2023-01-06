package main

import (
	"encoding/hex"
	"fmt"

	mvxcore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/firehose"
)

func checkHeader(
	multiversxBlock *firehose.FirehoseBlock,
	expectedHash string,
	expectedHeaderTypes map[mvxcore.HeaderType]struct{},
) error {
	log.Info("checking header", "hash", expectedHash)

	hashBytes := hasher.Compute(string(multiversxBlock.HeaderBytes))
	hashStr := hex.EncodeToString(hashBytes)
	if hashStr != expectedHash {
		return fmt.Errorf("checkHeader: invalid header hash after computation, expected: %s, got: %s", expectedHash, hashStr)
	}

	indexedHeaderHash := hex.EncodeToString(multiversxBlock.HeaderHash)
	if indexedHeaderHash != expectedHash {
		return fmt.Errorf("checkHeader: indexed invalid header hash %s , expected %v", indexedHeaderHash, expectedHeaderTypes)
	}

	headerType := mvxcore.HeaderType(multiversxBlock.HeaderType)
	_, found := expectedHeaderTypes[headerType]
	if !found {
		return fmt.Errorf("checkHeader: indexed invalid header type %s , expected %v", headerType, expectedHeaderTypes)
	}

	return nil
}
