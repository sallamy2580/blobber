package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/allocation"
	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/readmarker"
	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/reference"
	"github.com/0chain/blobber/code/go/0chain.net/core/common"
	"gorm.io/gorm"
)

// verifyAuthTicket verifies authTicket and returns authToken and error if any. For any error authToken is nil
func verifyAuthTicket(ctx context.Context, db *gorm.DB, authTokenString string, allocationObj *allocation.Allocation, refRequested *reference.Ref, clientID string) (*readmarker.AuthTicket, error) {
	if authTokenString == "" {
		return nil, common.NewError("invalid_parameters", "Auth ticket is required")
	}

	authToken := &readmarker.AuthTicket{}
	if err := json.Unmarshal([]byte(authTokenString), &authToken); err != nil {
		return nil, common.NewError("invalid_parameters", "Error parsing the auth ticket for download."+err.Error())
	}

	if err := authToken.Verify(allocationObj, clientID); err != nil {
		return nil, err
	}

	if refRequested.LookupHash != authToken.FilePathHash {
		authTokenRef, err := reference.GetLimitedRefFieldsByLookupHashWith(ctx, db, authToken.AllocationID, authToken.FilePathHash, []string{"id", "path"})
		if err != nil {
			return nil, err
		}

		if matched, _ := regexp.MatchString(fmt.Sprintf("^%v", authTokenRef.Path), refRequested.Path); !matched {
			return nil, common.NewError("invalid_parameters", "Auth ticket is not valid for the resource being requested")
		}
	}

	return authToken, nil
}
