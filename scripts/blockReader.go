package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ElrondNetwork/firehose-multiversx/types"
	pbmultiversx "github.com/ElrondNetwork/firehose-multiversx/types/pb/sf/multiversx/type/v1"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/dstore"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
)

const storagePath = "../devel/standard/firehose-data/storage/one-blocks"

type blockMessage string

func (bm blockMessage) String() string {
	return string(bm)
}

func init() {
	bstream.GetBlockReaderFactory = bstream.BlockReaderFactoryFunc(blockReaderFactory)
	bstream.GetBlockDecoder = bstream.BlockDecoderFunc(types.BlockDecoder)
	bstream.GetBlockWriterHeaderLen = 10
	bstream.GetBlockPayloadSetter = bstream.MemoryBlockPayloadSetter
	bstream.GetMemoizeMaxAge = 20 * time.Second
}

func blockReaderFactory(reader io.Reader) (bstream.BlockReader, error) {
	return bstream.NewDBinBlockReader(reader, func(contentType string, version int32) error {
		protocol := pbbstream.Protocol(pbbstream.Protocol_value[contentType])
		if protocol != pbbstream.Protocol_ETH && version != 1 {
			return fmt.Errorf("reader only knows about %s block kind at version 1, got %s at version %d", protocol, contentType, version)
		}

		return nil
	})
}

func getBlockFromStorage(blockNum uint64) (*pbmultiversx.Block, error) {
	log.Info("getBlockFromStorage", "blockNum", blockNum)

	store, err := dstore.NewDBinStore(storagePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create store at path %q: %w", store, err)
	}

	ctx := context.Background()
	var files []string
	filePrefix := fmt.Sprintf("%010d", blockNum)
	err = store.Walk(ctx, filePrefix, func(filename string) (err error) {
		files = append(files, filename)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to find on block files: %w", err)
	}

	var nativeBlock *pbmultiversx.Block
	for _, filepath := range files {
		reader, err := store.OpenObject(ctx, filepath)
		if err != nil {
			return nil, fmt.Errorf("unable to read block filename %s: ; error: %s\n", filepath, err)
		}
		defer func() {
			errCloseReader := reader.Close()
			if errCloseReader != nil {
				log.Warn("could not close reader", "error", errCloseReader)
			}
		}()

		readerFactory, err := bstream.GetBlockReaderFactory.New(reader)
		if err != nil {
			return nil, fmt.Errorf("unable to create a reader factory, blocks filename %s: %s\n", filepath, err)
		}

		block, err := readerFactory.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("reading block: %w", err)
		}

		nativeBlock, err = getNativeBlock(block)
		if err != nil {
			return nil, err
		}

	}
	if nativeBlock == nil || nativeBlock.MultiversxBlock == nil {
		return nil, fmt.Errorf("go nil block from storage")
	}

	return nativeBlock, nil
}

func getNativeBlock(block *bstream.Block) (*pbmultiversx.Block, error) {
	nativeBlock := block.ToProtocol().(*pbmultiversx.Block)

	data, err := json.MarshalIndent(nativeBlock, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("json marshall: %w", err)
	}

	log.Info("loaded block", "block", blockMessage(data))
	return nativeBlock, nil
}
