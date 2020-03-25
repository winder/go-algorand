package v2

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/algorand/go-codec/codec"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated"
	"github.com/algorand/go-algorand/data"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/logging"
	"github.com/algorand/go-algorand/node"
	"github.com/algorand/go-algorand/protocol"
)

type V2Handlers struct {
	Node     *node.AlgorandFullNode
	Log      logging.Logger
	Shutdown <-chan struct{}
}

func returnErrorLogInternal(ctx echo.Context, code int, err error, logger logging.Logger, internal string) error {
	logger.Info(internal)
	return ctx.JSON(code, generated.Error{Error:err.Error()})
}

func returnError(ctx echo.Context, code int, err error, logger logging.Logger) error {
	logger.Info(err)
	return ctx.JSON(code, generated.Error{Error:err.Error()})
}


// (POST /v2/register-participation-keys/{account-id})
func (v2 *V2Handlers) PostV2RegisterParticipationKeysAccountId(ctx echo.Context, accountId string, params generated.PostV2RegisterParticipationKeysAccountIdParams) error {
	// TODO
	return nil
}

// (POST /v2/shutdown)
func (v2 *V2Handlers) PostV2Shutdown(ctx echo.Context, params generated.PostV2ShutdownParams) error {
	// TODO
	return nil
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

// Get account information.
// (GET /v2/accounts/{address})
func (v2 *V2Handlers) AccountInformation(ctx echo.Context, address string) error {
	addr, err := basics.UnmarshalChecksumAddress(address)
	if err != nil {
		return returnError(ctx, http.StatusBadRequest, err, v2.Log)
	}

	myLedger := v2.Node.Ledger()
	lastRound := myLedger.Latest()
	record, err := myLedger.Lookup(lastRound, basics.Address(addr))
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, err, v2.Log)
	}
	recordWithoutPendingRewards, err := myLedger.LookupWithoutRewards(lastRound, basics.Address(addr))
	if err != nil {
		return returnErrorLogInternal(ctx, http.StatusInternalServerError, err, v2.Log, errFailedLookingUpLedger)
	}

	amount := record.MicroAlgos
	amountWithoutPendingRewards := recordWithoutPendingRewards.MicroAlgos
	pendingRewards, overflowed := basics.OSubA(amount, amountWithoutPendingRewards)
	if overflowed {
		return returnErrorLogInternal(ctx, http.StatusInternalServerError, err, v2.Log, errInternalFailure)
	}

	assets := make([]generated.AssetHolding, 0)
	if len(record.Assets) > 0 {
		//assets = make(map[uint64]v1.AssetHolding)
		for curid, holding := range record.Assets {
			var creator string
			creatorAddr, err := myLedger.GetAssetCreator(curid)
			if err == nil {
				creator = creatorAddr.String()
			} else {
				// Asset may have been deleted, so we can no
				// longer fetch the creator
				creator = ""
			}

			holding := generated.AssetHolding{
				Amount:   holding.Amount,
				AssetId:  uint64(curid),
				Creator:  creator,
				IsFrozen: holding.Frozen,
			}

			assets = append(assets, holding)
		}
	}

	createdAssets := make([]generated.Asset, 0)
	if len(record.AssetParams) > 0 {
		for idx, params := range record.AssetParams {
			assetParams := generated.AssetParams{
				Creator:       address,
				Total:         params.Total,
				Decimals:      uint64(params.Decimals),
				DefaultFrozen: &params.DefaultFrozen,
				MetadataHash:  byteOrNil(params.MetadataHash[:]),
				Name:          strOrNil(params.AssetName),
				UnitName:      strOrNil(params.UnitName),
				Url:           strOrNil(params.URL),
				Clawback:      addrOrNil(params.Clawback),
				Freeze:        addrOrNil(params.Freeze),
				Manager:       addrOrNil(params.Manager),
				Reserve:       addrOrNil(params.Reserve),
			}
			asset := generated.Asset{
				Index:  uint64(idx),
				Params: assetParams,
			}
			createdAssets = append(createdAssets, asset)
		}
	}

	var apiParticipation *generated.AccountParticipation
	if record.VoteID != (crypto.OneTimeSignatureVerifier{}) {
		apiParticipation = &generated.AccountParticipation {
			VoteParticipationKey:      byteOrNil(record.VoteID[:]),
			SelectionParticipationKey: byteOrNil(record.SelectionID[:]),
			VoteFirstValid:            numOrNil(uint64(record.VoteFirstValid)),
			VoteLastValid:             numOrNil(uint64(record.VoteLastValid)),
			VoteKeyDilution:           numOrNil(uint64(record.VoteKeyDilution)),
		}
	}

	response := generated.AccountResponse{
		Type:                        nil,
		Round:                       uint64(lastRound),
		Address:                     addr.String(),
		Amount:                      amount.Raw,
		PendingRewards:              pendingRewards.Raw,
		AmountWithoutPendingRewards: amountWithoutPendingRewards.Raw,
		Rewards:                     record.RewardedMicroAlgos.Raw,
		Status:                      record.Status.String(),
		RewardBase:                  &record.RewardsBase,
		Participation:               apiParticipation,
		CreatedAssets:               &createdAssets,
		Assets:                      &assets,
	}

	return ctx.JSON(http.StatusOK, response)
}

// Get the block for the given round.
// (GET /v2/blocks/{round})
func (v2 *V2Handlers) GetBlock(ctx echo.Context, round uint64, params generated.GetBlockParams) error {
	handle, err := getCodecHandle(params.Format)
	if err != nil {
		return returnError(ctx, http.StatusBadRequest, err, v2.Log)
	}

	// TODO: What is raw block bytes, should I use that instead?
	//blockbytes, err := rpcs.RawBlockBytes(v2.Node.Ledger(), basics.Round(round))

	ledger := v2.Node.Ledger()
	b, _, err := ledger.BlockCert(basics.Round(round))
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, err, v2.Log)
	}

	encoded, err := encode(handle, b)
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, err, v2.Log)
	}

	return ctx.JSON(http.StatusOK, generated.BlockResponse{
		Block: encoded,
	})
}

// Get the current supply reported by the ledger.
// (GET /v2/ledger/supply)
func (v2 *V2Handlers) GetSupply(ctx echo.Context) error {
	latest := v2.Node.Ledger().Latest()
	totals, err := v2.Node.Ledger().Totals(latest)
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, fmt.Errorf("GetSupply(): round %d, failed: %v", latest, err), v2.Log)
	}

	supply := generated.SupplyResponse{
		CurrentRound: uint64(latest),
		TotalMoney:   totals.Participating().Raw,
		OnlineMoney:  totals.Online.Money.Raw,
	}

	return ctx.JSON(http.StatusOK, supply)
}

// Gets the current node status.
// (GET /v2/status)
func (v2 *V2Handlers) GetStatus(ctx echo.Context) error {
	stat, err := v2.Node.Status()
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, errors.New(errFailedRetrievingNodeStatus), v2.Log)
	}

	response := generated.NodeStatusResponse{
		LastRound:                 uint64(stat.LastRound),
		LastVersion:               string(stat.LastVersion),
		NextVersion:               string(stat.NextVersion),
		NextVersionRound:          uint64(stat.NextVersionRound),
		NextVersionSupported:      stat.NextVersionSupported,
		TimeSinceLastRound:        uint64(stat.TimeSinceLastRound().Nanoseconds()),
		CatchupTime:               uint64(stat.CatchupTime.Nanoseconds()),
		StoppedAtUnsupportedRound: stat.StoppedAtUnsupportedRound,
	}

	return ctx.JSON(http.StatusOK, response)
}

// Gets the node status after waiting for the given round.
// (GET /v2/status/wait-for-block-after/{round}/)
func (v2 *V2Handlers) WaitForBlock(ctx echo.Context, round uint64) error {
	ledger := v2.Node.Ledger()
	latestBlkHdr, err := ledger.BlockHdr(ledger.Latest())
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, err, v2.Log)
	}

	// Check if we're stalled due to an unsupported protocol version
	if latestBlkHdr.NextProtocol != "" {
		if _, nextProtocolSupported := config.Consensus[latestBlkHdr.NextProtocol]; !nextProtocolSupported {
			// see if the desired protocol switch is expect to happen before or after the above point.
			if latestBlkHdr.NextProtocolSwitchOn <= basics.Round(round+1) {
				// we would never reach to this round, since this round would happen after the (unsupported) protocol upgrade.
				return returnError(ctx, http.StatusBadRequest, err, v2.Log)
			}
		}
	}

	// Wait
	select {
	case <-v2.Shutdown:
		return returnError(ctx, http.StatusInternalServerError, err, v2.Log)
	case <-time.After(1 * time.Minute):
	case <-ledger.Wait(basics.Round(round + 1)):
	}

	// Return status after the wait
	return v2.GetStatus(ctx)
}

// Broadcasts a raw transaction to the network.
// (POST /v2/transactions)
func (v2 *V2Handlers) RawTransaction(ctx echo.Context) error {
	var txgroup []transactions.SignedTxn
	dec := protocol.NewDecoder(ctx.Request().Body)
	for {
		var st transactions.SignedTxn
		err := dec.Decode(&st)
		if err == io.EOF {
			break
		}
		if err != nil {
			return returnError(ctx, http.StatusBadRequest, err, v2.Log)
		}
		txgroup = append(txgroup, st)
	}

	if len(txgroup) == 0 {
		err := errors.New("empty txgroup")
		return returnError(ctx, http.StatusBadRequest, err, v2.Log)
	}

	err := v2.Node.BroadcastSignedTxGroup(txgroup)
	if err != nil {
		return returnError(ctx, http.StatusBadRequest, err, v2.Log)
	}

	// For backwards compatibility, return txid of first tx in group
	txid := txgroup[0].ID()
	return ctx.JSON(http.StatusOK, generated.PostTransactionsResponse{TxId:txid.String()})
}

// Get parameters for constructing a new transaction
// (GET /v2/transactions/params)
func (v2 *V2Handlers) TransactionParams(ctx echo.Context) error {
	stat, err := v2.Node.Status()
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, err, v2.Log)
	}

	gh := v2.Node.GenesisHash()

	var params generated.TransactionParams
	params.Fee = v2.Node.SuggestedFee().Raw
	params.GenesisID = v2.Node.GenesisID()
	params.GenesisHash = gh[:]
	params.LastRound = uint64(stat.LastRound)
	params.ConsensusVersion = string(stat.LastVersion)

	proto := config.Consensus[stat.LastVersion]
	params.MinFee = proto.MinTxnFee

	return ctx.JSON(http.StatusOK, params)
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



// Get a specific pending transaction.
// (GET /v2/transactions/pending/{txid})
func (v2 *V2Handlers) PendingTransactionInformation(ctx echo.Context, txid string, params generated.PendingTransactionInformationParams) error {
	txID := transactions.Txid{}
	if txID.UnmarshalText([]byte(txid)) != nil {
		return returnError(ctx, http.StatusBadRequest, fmt.Errorf(errNoTxnSpecified), v2.Log)
	}

	if txn, ok := v2.Node.GetPendingTransaction(txID); ok {
		response := generated.PendingTransactionResponse{
			Txn:             "",
			PoolError:       "",
			ClosingAmount:   nil,
			ConfirmedRound:  nil,
			SenderRewards:   nil,
			ReceiverRewards: nil,
			CloseRewards:    nil,
		}

		handle, err := getCodecHandle(params.Format)
		if err != nil {
			return returnError(ctx, http.StatusBadRequest, err, v2.Log)
		}

		encoded, err := encode(handle, txn.Txn)
		if err != nil {
			return returnError(ctx, http.StatusInternalServerError, err, v2.Log)
		}

		response.Txn = encoded

		if txn.ConfirmedRound != 0 {
			r := uint64(txn.ConfirmedRound)
			response.ConfirmedRound = &r

			response.ClosingAmount = &txn.ApplyData.ClosingAmount.Raw
			response.SenderRewards = &txn.ApplyData.SenderRewards.Raw
			response.ReceiverRewards = &txn.ApplyData.ReceiverRewards.Raw
			response.CloseRewards = &txn.ApplyData.CloseRewards.Raw

			response.AssetIndex = computeAssetIndexFromTxn(txn, v2.Node.Ledger())
		}

		return ctx.JSON(http.StatusOK, response)
	}

	// We didn't find it, return a failure
	return returnError(ctx, http.StatusNotFound, errors.New(errTransactionNotFound), v2.Log)
}

func (v2 *V2Handlers) getPendingTransactions(ctx echo.Context, max *uint64, format *string, addrFilter *string) error {
	var addrPtr *basics.Address

	if addrFilter != nil {
		addr, err := basics.UnmarshalChecksumAddress(*addrFilter)
		if err != nil {
			return returnError(ctx, http.StatusBadRequest, errors.New(errFailedToParseAddress), v2.Log)
		}
		addrPtr = &addr
	}

	txns, err := v2.Node.GetPendingTxnsFromPool()
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, err, v2.Log)
	}

	handle, err := getCodecHandle(format)
	if err != nil {
		return returnError(ctx, http.StatusBadRequest, err, v2.Log)
	}

	// TODO: What should I put in here? MatchAddress uses this to check the FeeSink so I think this is fine.
	spec := transactions.SpecialAddresses{
		FeeSink:     basics.Address{},
		RewardsPool: basics.Address{},
	}

	// Convert transactions to msgp / json strings
	encodedTxns := make([]string, 0)
	for _, txn := range txns {
		// break out if we've reached the max number of transactions
		if max != nil && uint64(len(encodedTxns)) >= *max {
			break;
		}

		// continue if we have an address filter and the address doesn't match the transaction.
		if addrPtr != nil && !txn.Txn.MatchAddress(*addrPtr, spec) {
			continue;
		}

		// Encode the transaction and added to the results
		encodedTxn, err := encode(handle, txn)
		if err != nil {
			return returnError(ctx, http.StatusInternalServerError, err, v2.Log)
		}
		encodedTxns = append(encodedTxns, encodedTxn)
	}

	return ctx.JSON(http.StatusOK, generated.PendingTransactionsResponse{
		TopTransactions:   encodedTxns,
		TotalTransactions: uint64(len(txns)),
	})
}

// Get a list of unconfirmed transactions currently in the transaction pool.
// (GET /v2/transactions/pending)
func (v2 *V2Handlers) GetPendingTransactions(ctx echo.Context, params generated.GetPendingTransactionsParams) error {
	return v2.getPendingTransactions(ctx, params.Max, params.Format, nil)
}

// Get a list of unconfirmed transactions currently in the transaction pool by address.
// (GET /v2/accounts/{addr}/transactions/pending)
func (v2 *V2Handlers) GetPendingTransactionsByAddress(ctx echo.Context, addr string, params generated.GetPendingTransactionsByAddressParams) error {
	return v2.getPendingTransactions(ctx, params.Max, params.Format, &addr)
}

