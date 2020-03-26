package v2

import (
	"fmt"
	"github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated"
	"github.com/algorand/go-algorand/logging"
	"github.com/labstack/echo/v4"
	"strings"

	"github.com/algorand/go-algorand/data"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/node"
	"github.com/algorand/go-algorand/protocol"
	"github.com/algorand/go-codec/codec"
)

// returnError logs an internal message while returning the encoded response.
func returnError(ctx echo.Context, code int, internal error, external string, logger logging.Logger) error {
	logger.Info(internal)
	return ctx.JSON(code, generated.Error{Error:external})
}

func addrOrNil(addr basics.Address) *string {
	if addr.IsZero() {
		return nil
	}
	ret := addr.String()
	return &ret
}

func strOrNil(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

func numOrNil(num uint64) *uint64 {
	if num == 0 {
		return nil
	}
	return &num
}

func byteOrNil(data []byte) *[]byte {
	if len(data) == 0 {
		return nil
	}
	return &data
}


func computeAssetIndexInPayset(tx node.TxnWithStatus, txnCounter uint64, payset []transactions.SignedTxnWithAD) (aidx *uint64) {
	// Compute transaction index in block
	offset := -1
	for idx, stxnib := range payset {
		if tx.Txn.Txn.ID() == stxnib.Txn.ID() {
			offset = idx
			break
		}
	}

	// Sanity check that txn was in fetched block
	if offset < 0 {
		return nil
	}

	// Count into block to get created asset index
	idx := txnCounter - uint64(len(payset)) + uint64(offset) + 1
	return &idx
}

// computeAssetIndexFromTxn returns the created asset index given a confirmed
// transaction whose confirmation block is available in the ledger. Note that
// 0 is an invalid asset index (they start at 1).
func computeAssetIndexFromTxn(tx node.TxnWithStatus, l *data.Ledger) (aidx *uint64) {
	// Must have ledger
	if l == nil {
		return nil
	}
	// Transaction must be confirmed
	if tx.ConfirmedRound == 0 {
		return nil
	}
	// Transaction must be AssetConfig transaction
	if tx.Txn.Txn.AssetConfigTxnFields == (transactions.AssetConfigTxnFields{}) {
		return nil
	}
	// Transaction must be creating an asset
	if tx.Txn.Txn.AssetConfigTxnFields.ConfigAsset != 0 {
		return nil
	}

	// Look up block where transaction was confirmed
	blk, err := l.Block(tx.ConfirmedRound)
	if err != nil {
		return nil
	}

	payset, err := blk.DecodePaysetFlat()
	if err != nil {
		return nil
	}

	return computeAssetIndexInPayset(tx, blk.BlockHeader.TxnCounter, payset)
}

func getCodecHandle(formatPtr *string) (codec.Handle, error) {
	format := "json"
	if formatPtr != nil {
		format = strings.ToLower(*formatPtr)
	}

	var handle codec.Handle = protocol.JSONHandle
	if format == "json" {
		handle = protocol.JSONHandle
	} else if format == "msgpack" || format == "msgp" {
		handle = protocol.CodecHandle
	} else {
		fmt.Sprintf("invalid format: %s", format)
	}

	return handle, nil
}

func encode(handle codec.Handle, obj interface{}) (string, error) {
	var output []byte
	enc := codec.NewEncoderBytes(&output, handle)

	err := enc.Encode(obj)
	if err != nil {
		return "", fmt.Errorf("failed to encode object: %v", err)
	}
	return string(output), nil
}
