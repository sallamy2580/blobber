package allocation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/datastore"
	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/filestore"
	"github.com/0chain/blobber/code/go/0chain.net/core/chain"
	"github.com/0chain/blobber/code/go/0chain.net/core/common"
	. "github.com/0chain/blobber/code/go/0chain.net/core/logging"
	"github.com/0chain/blobber/code/go/0chain.net/core/node"
	"github.com/0chain/blobber/code/go/0chain.net/core/transaction"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetAllocationByID from DB. This function doesn't load related terms.
func GetAllocationByID(ctx context.Context, allocID string) (a *Allocation, err error) {
	var tx = datastore.GetStore().GetTransaction(ctx)

	a = new(Allocation)
	err = tx.Model(&Allocation{}).
		Where("id=?", allocID).
		First(a).Error
	return
}

// LoadTerms loads corresponding terms from DB. Since, the GetAllocationByID
// doesn't loads up related Terms (isn't needed in most cases) this method
// loads the Terms for an allocation.
func (a *Allocation) LoadTerms(ctx context.Context) (err error) {
	// get transaction
	var tx = datastore.GetStore().GetTransaction(ctx)
	// load related terms
	var terms []*Terms
	err = tx.Model(terms).
		Where("allocation_id = ?", a.ID).
		Find(&terms).Error
	if err != nil {
		// unexpected DB error, including a RecordNotFoundError, since
		// an allocation can't be without its terms (the terms must exist)
		return
	}
	a.Terms = terms // set field
	return          // found in DB
}

// VerifyAllocationTransaction try to get allocation from postgres.if it doesn't exists, get it from sharders, and insert it into postgres.
func VerifyAllocationTransaction(ctx context.Context, allocationTx string, readonly bool) (a *Allocation, err error) {
	var tx = datastore.GetStore().GetTransaction(ctx)

	a = new(Allocation)
	err = tx.Model(&Allocation{}).
		Where(&Allocation{Tx: allocationTx}).
		First(a).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewError("bad_db_operation", err.Error()) // unexpected DB error
	}

	if err == nil {
		// load related terms
		var terms []*Terms
		err = tx.Model(terms).
			Where("allocation_id = ?", a.ID).
			Find(&terms).Error
		if err != nil {
			return nil, common.NewError("bad_db_operation", err.Error()) // unexpected DB error
		}
		a.Terms = terms // set field
		return          // found in DB
	}

	t, err := transaction.VerifyTransaction(allocationTx, chain.GetServerChain())
	if err != nil {
		return nil, common.NewError("invalid_allocation",
			"Invalid Allocation id. Allocation not found in blockchain. "+err.Error())
	}
	var sa transaction.StorageAllocation
	err = json.Unmarshal([]byte(t.TransactionOutput), &sa)
	if err != nil {
		return nil, common.NewError("transaction_output_decode_error",
			"Error decoding the allocation transaction output."+err.Error())
	}

	var isExist bool
	err = tx.Model(&Allocation{}).
		Where("id = ?", sa.ID).
		First(a).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewError("bad_db_operation", err.Error()) // unexpected
	}

	isExist = (a.ID != "")

	Logger.Info("VerifyAllocationTransaction",
		zap.Bool("isExist", isExist),
		zap.Any("allocation", a),
		zap.Any("storageAllocation", sa),
		zap.String("node.Self.ID", node.Self.ID))

	if !isExist {
		foundBlobber := false
		for _, blobberConnection := range sa.BlobberDetails {
			if blobberConnection.BlobberID != node.Self.ID {
				continue
			}
			foundBlobber = true
			a.AllocationRoot = ""
			a.BlobberSize = (sa.Size + int64(len(sa.BlobberDetails)-1)) /
				int64(len(sa.BlobberDetails))
			a.BlobberSizeUsed = 0
			break
		}
		if !foundBlobber {
			return nil, common.NewError("invalid_blobber",
				"Blobber is not part of the open connection transaction")
		}
	}

	// set/update fields
	a.ID = sa.ID
	a.Tx = sa.Tx
	a.Expiration = sa.Expiration
	a.OwnerID = sa.OwnerID
	a.OwnerPublicKey = sa.OwnerPublicKey
	a.RepairerID = t.ClientID // blobber node id
	a.TotalSize = sa.Size
	a.UsedSize = sa.UsedSize
	a.Finalized = sa.Finalized
	a.TimeUnit = sa.TimeUnit
	a.IsImmutable = sa.IsImmutable

	m := map[string]interface{}{
		"allocation_id":  a.ID,
		"allocated_size": (sa.Size + int64(len(sa.BlobberDetails)-1)) / int64(len(sa.BlobberDetails)),
	}

	err = filestore.GetFileStore().UpdateAllocationMetaData(m)
	if err != nil {
		return nil, common.NewError("meta_data_update_error", err.Error())
	}
	// go update allocation data in file store map
	// related terms
	a.Terms = make([]*Terms, 0, len(sa.BlobberDetails))
	for _, d := range sa.BlobberDetails {
		a.Terms = append(a.Terms, &Terms{
			BlobberID:    d.BlobberID,
			AllocationID: a.ID,
			ReadPrice:    d.Terms.ReadPrice,
			WritePrice:   d.Terms.WritePrice,
		})
	}

	if readonly {
		return a, nil
	}

	Logger.Info("Saving the allocation to DB")

	if isExist {
		err = tx.Save(a).Error
	} else {
		err = tx.Create(a).Error
	}

	if err != nil {
		return nil, err
	}

	// save/update related terms
	for _, t := range a.Terms {
		if isExist {
			err = tx.Save(t).Error
		} else {
			err = tx.Create(t).Error
		}
		if err != nil {
			return nil, err
		}
	}

	return a, nil
}

func RequestReadPoolStat(clientID string) (*ReadPool, error) {
	Logger.Info("request read pools")

	params := map[string]string{
		"client_id": clientID,
	}
	resp, err := transaction.MakeSCRestAPICall(transaction.STORAGE_CONTRACT_ADDRESS, "/getReadPoolStat", params, chain.GetServerChain())
	if err != nil {
		return nil, fmt.Errorf("requesting read pools stat: %v", err)
	}

	var readPool ReadPool
	if err = json.Unmarshal(resp, &readPool); err != nil {
		return nil, fmt.Errorf("decoding read pools stat response: %v, \n==> resp: %s", err, string(resp))
	}

	return &readPool, nil
}

func RequestWritePool(allocationID string) (wps *WritePool, err error) {
	Logger.Info("request write pools")

	var (
		resp []byte
	)

	params := map[string]string{
		"allocation": allocationID,
	}
	resp, err = transaction.MakeSCRestAPICall(transaction.STORAGE_CONTRACT_ADDRESS, "/allocation", params, chain.GetServerChain())
	if err != nil {
		return nil, fmt.Errorf("requesting write pools stat: %v", err)
	}

	var allocation = struct {
		ID        string `json:"id"`
		WritePool uint64 `json:"write_pool"`
	}{}
	if err = json.Unmarshal(resp, &allocation); err != nil {
		return nil, fmt.Errorf("decoding write pools stat response: %v", err)
	}

	return &WritePool{
		AllocationID: allocationID,
		Balance:      allocation.WritePool,
	}, nil
}
