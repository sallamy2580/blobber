package allocation

import (
	"context"
	"encoding/json"

	"0chain.net/blobbercore/datastore"
	"0chain.net/core/chain"
	"0chain.net/core/common"
	. "0chain.net/core/logging"
	"0chain.net/core/node"
	"0chain.net/core/transaction"

	"github.com/jinzhu/gorm"
)

func VerifyAllocationTransaction(ctx context.Context, allocationID string, readonly bool) (*Allocation, error) {
	a := &Allocation{}
	db := datastore.GetStore().GetTransaction(ctx)
	err := db.Where(&Allocation{ID: allocationID}).First(a).Error
	if err == nil {
		return a, nil
	}

	if err != nil && gorm.IsRecordNotFoundError(err) {
		t, err := transaction.VerifyTransaction(allocationID, chain.GetServerChain())
		if err != nil {
			return nil, common.NewError("invalid_allocation", "Invalid Allocation id. Allocation not found in blockchain. "+err.Error())
		}
		var storageAllocation transaction.StorageAllocation
		err = json.Unmarshal([]byte(t.TransactionOutput), &storageAllocation)
		if err != nil {
			return nil, common.NewError("transaction_output_decode_error", "Error decoding the allocation transaction output."+err.Error())
		}
		foundBlobber := false
		for _, blobberConnection := range storageAllocation.Blobbers {
			if blobberConnection.ID == node.Self.ID {
				foundBlobber = true
				a.AllocationRoot = ""
				a.BlobberSize = (storageAllocation.Size + int64(len(storageAllocation.Blobbers)-1)) / int64(len(storageAllocation.Blobbers))
				a.BlobberSizeUsed = 0
				break
			}
		}
		if !foundBlobber {
			return nil, common.NewError("invalid_blobber", "Blobber is not part of the open connection transaction")
		}
		a.ID = allocationID
		a.Expiration = storageAllocation.Expiration
		a.OwnerID = storageAllocation.OwnerID
		a.OwnerPublicKey = storageAllocation.OwnerPublicKey
		a.TotalSize = storageAllocation.Size
		a.UsedSize = storageAllocation.UsedSize
		if !readonly {
			Logger.Info("Saving the allocation to DB")
			db.Exec("INSERT INTO allocations (id, size, used_size, expiration_date, owner_id, owner_public_key, blobber_size) VALUES (?,?,?,?,?,?,?) ON CONFLICT (id) DO NOTHING;", a.ID, a.TotalSize, a.UsedSize, a.Expiration, a.OwnerID, a.OwnerPublicKey, a.BlobberSize)
			return a, nil
		}
		return a, nil
	}
	return nil, err
}
