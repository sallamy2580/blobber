package storage

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/0chain/blobber/code/go/0chain.net/core/cache"
	"github.com/0chain/blobber/code/go/0chain.net/core/common"
	"github.com/0chain/blobber/code/go/0chain.net/core/logging"
	"github.com/0chain/blobber/code/go/0chain.net/core/node"

	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"
)

var lru = cache.NewLRUCache(10000)

func ChallengeHandler(ctx context.Context, r *http.Request) (interface{}, error) {
	if r.Method == "GET" {
		return nil, common.NewError("invalid_method", "Invalid method used for the upload URL. Use multi-part form POST instead")
	}
	logging.Logger.Info("Got validation request. Decoding the input")
	requestHash := r.Header.Get("X-App-Request-Hash")
	h := sha3.New256()
	tReader := io.TeeReader(r.Body, h)
	var challengeRequest ChallengeRequest
	decoder := json.NewDecoder(tReader)
	err := decoder.Decode(&challengeRequest)
	if err != nil {
		logging.Logger.Error("Error decoding the input to validator")
		return nil, common.NewError("input_decode_error", "Error in decoding the input."+err.Error())
	}
	challengeHash := hex.EncodeToString(h.Sum(nil))

	if requestHash != challengeHash {
		logging.Logger.Error("Header hash and request hash do not match")
		return nil, common.NewError("invalid_parameters", "Header hash and request hash do not match")
	}

	if challengeRequest.ObjPath == nil {
		logging.Logger.Error("Not object path found in the input")
		return nil, common.NewError("invalid_parameters", "Empty object path or merkle path")
	}

	logging.Logger.Info("Processing validation.", zap.Any("challenge_id", challengeRequest.ChallengeID))
	vt, err := lru.Get(challengeHash)
	retVT, ok := vt.(*ValidationTicket)
	if vt != nil && err == nil && ok {
		return retVT, nil
	}

	var validationTicket ValidationTicket
	challengeObj, err := GetProtocolImpl().VerifyChallengeTransaction(ctx, &challengeRequest)
	if err != nil {
		logging.Logger.Error("Error verifying the challenge from BC",
			zap.Any("challenge_id", challengeRequest.ChallengeID),
			zap.Error(err))
		return nil, common.NewError("invalid_parameters", "Challenge could not be verified. "+err.Error())
	}

	time.Sleep(1 * time.Second)

	allocationObj, err := GetProtocolImpl().VerifyAllocationTransaction(ctx, challengeObj.AllocationID)
	if err != nil {
		logging.Logger.Error("Error verifying the allocation from BC", zap.Any("allocation_id", challengeObj.AllocationID), zap.Error(err))
		return nil, common.NewError("invalid_parameters", "Allocation could not be verified. "+err.Error())
	}

	err = challengeRequest.VerifyChallenge(challengeObj, allocationObj)
	if err != nil {
		errCode := err.Error()
		commError, ok := err.(*common.Error)
		if ok {
			errCode = commError.Code
		}

		logging.Logger.Error("Validation Failed - Error verifying the challenge", zap.Any("challenge_id", challengeObj.ID), zap.Error(err))
		validationTicket.BlobberID = challengeObj.BlobberID
		validationTicket.ChallengeID = challengeObj.ID
		validationTicket.Result = false
		validationTicket.MessageCode = errCode
		validationTicket.Message = err.Error()
		validationTicket.ValidatorID = node.Self.ID
		validationTicket.ValidatorKey = node.Self.PublicKey
		validationTicket.Timestamp = common.Now()

		if err := validationTicket.Sign(); err != nil {
			return nil, common.NewError("invalid_parameters", err.Error())
		}
		return &validationTicket, nil
	}

	validationTicket.BlobberID = challengeObj.BlobberID
	validationTicket.ChallengeID = challengeObj.ID
	validationTicket.Result = true
	validationTicket.MessageCode = "success"
	validationTicket.Message = "Challenge passed"
	validationTicket.ValidatorID = node.Self.ID
	validationTicket.ValidatorKey = node.Self.PublicKey
	validationTicket.Timestamp = common.Now()
	if err := validationTicket.Sign(); err != nil {
		return nil, common.NewError("invalid_parameters", err.Error())
	}
	logging.Logger.Info("Validation passed.", zap.Any("challenge_id", challengeRequest.ChallengeID))

	lru.Add(challengeHash, &validationTicket) //nolint:errcheck // never returns an error anyway
	return &validationTicket, nil
}
