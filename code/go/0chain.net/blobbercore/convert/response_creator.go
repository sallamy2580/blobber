package convert

import (
	"encoding/json"

	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/allocation"
	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobberhttp"
	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/reference"
	stats2 "github.com/0chain/blobber/code/go/0chain.net/blobbercore/stats"
)

func GetAllocationResponseCreator(resp interface{}) *blobbergrpc.GetAllocationResponse {
	if resp == nil {
		return nil
	}

	alloc, _ := resp.(*allocation.Allocation)
	return &blobbergrpc.GetAllocationResponse{Allocation: AllocationToGRPCAllocation(alloc)}
}

func GetFileMetaDataResponseCreator(httpResp interface{}) *blobbergrpc.GetFileMetaDataResponse {
	if httpResp == nil {
		return nil
	}

	r, _ := httpResp.(map[string]interface{})

	var resp blobbergrpc.GetFileMetaDataResponse
	collaborators, _ := r["collaborators"].([]reference.Collaborator)
	for i := 0; i < len(collaborators); i++ {
		c := collaborators[i]
		resp.Collaborators = append(resp.Collaborators, CollaboratorToGRPCCollaborator(&c))
	}

	resp.MetaData = FileRefToFileRefGRPC(reference.ListingDataToRef(r))
	return &resp
}

func GetFileStatsResponseCreator(r interface{}) *blobbergrpc.GetFileStatsResponse {
	if r == nil {
		return nil
	}

	httpResp, _ := r.(map[string]interface{})

	var resp blobbergrpc.GetFileStatsResponse
	resp.MetaData = FileRefToFileRefGRPC(reference.ListingDataToRef(httpResp))

	respRaw, _ := json.Marshal(httpResp)
	var stats stats2.FileStats
	_ = json.Unmarshal(respRaw, &stats)
	resp.Stats = FileStatsToFileStatsGRPC(&stats)

	return &resp
}

func ListEntitesResponseCreator(r interface{}) *blobbergrpc.ListEntitiesResponse {
	if r == nil {
		return nil
	}

	httpResp, _ := r.(*blobberhttp.ListResult)

	var resp blobbergrpc.ListEntitiesResponse
	for i := range httpResp.Entities {
		resp.Entities = append(resp.Entities, FileRefToFileRefGRPC(reference.ListingDataToRef(httpResp.Entities[i])))
	}

	resp.MetaData = FileRefToFileRefGRPC(reference.ListingDataToRef(httpResp.Meta))
	resp.AllocationRoot = httpResp.AllocationRoot
	return &resp
}

func GetReferencePathResponseCreator(r interface{}) *blobbergrpc.GetReferencePathResponse {
	if r == nil {
		return nil
	}

	httpResp, _ := r.(*blobberhttp.ReferencePathResult)
	var resp blobbergrpc.GetReferencePathResponse

	var recursionCount int
	resp.LatestWm = WriteMarkerToWriteMarkerGRPC(httpResp.LatestWM)
	resp.ReferencePath = ReferencePathToReferencePathGRPC(&recursionCount, httpResp.ReferencePath)
	return &resp
}

func GetObjectTreeResponseCreator(r interface{}) *blobbergrpc.GetObjectTreeResponse {
	if r == nil {
		return nil
	}

	httpResp, _ := r.(*blobberhttp.ReferencePathResult)
	var resp blobbergrpc.GetObjectTreeResponse

	var recursionCount int
	resp.LatestWm = WriteMarkerToWriteMarkerGRPC(httpResp.LatestWM)
	resp.ReferencePath = ReferencePathToReferencePathGRPC(&recursionCount, httpResp.ReferencePath)
	return &resp
}

func GetObjectPathResponseCreator(r interface{}) *blobbergrpc.GetObjectPathResponse {
	if r == nil {
		return nil
	}

	httpResp, _ := r.(*blobberhttp.ObjectPathResult)
	var resp blobbergrpc.GetObjectPathResponse

	var pathList []*blobbergrpc.FileRef
	pl, _ := httpResp.Path["list"].([]map[string]interface{})
	for _, v := range pl {
		pathList = append(pathList, FileRefToFileRefGRPC(reference.ListingDataToRef(v)))
	}

	resp.LatestWriteMarker = WriteMarkerToWriteMarkerGRPC(httpResp.LatestWM)
	resp.ObjectPath = &blobbergrpc.ObjectPath{
		RootHash:     httpResp.RootHash,
		Meta:         FileRefToFileRefGRPC(reference.ListingDataToRef(httpResp.Meta)),
		Path:         FileRefToFileRefGRPC(reference.ListingDataToRef(httpResp.Path)),
		PathList:     pathList,
		FileBlockNum: httpResp.FileBlockNum,
	}

	return &resp
}

func CommitWriteResponseCreator(r interface{}) *blobbergrpc.CommitResponse {
	if r == nil {
		return nil
	}

	httpResp, _ := r.(*blobberhttp.CommitResult)

	return &blobbergrpc.CommitResponse{
		AllocationRoot: httpResp.AllocationRoot,
		WriteMarker:    WriteMarkerToWriteMarkerGRPC(httpResp.WriteMarker),
		ErrorMessage:   httpResp.ErrorMessage,
		Success:        httpResp.Success,
	}
}

func GetCalculateHashResponseCreator(r interface{}) *blobbergrpc.CalculateHashResponse {
	httpResp, _ := r.(map[string]interface{})
	msg, _ := httpResp["msg"].(string)

	return &blobbergrpc.CalculateHashResponse{Message: msg}
}

func GetCommitMetaTxnResponseCreator(r interface{}) *blobbergrpc.CommitMetaTxnResponse {
	msg, _ := r.(struct {
		Msg string `json:"msg"`
	})

	return &blobbergrpc.CommitMetaTxnResponse{Message: msg.Msg}
}

func CollaboratorResponseCreator(r interface{}) *blobbergrpc.CollaboratorResponse {
	if r == nil {
		return nil
	}

	msg, _ := r.(struct {
		Msg string `json:"msg"`
	})
	var resp blobbergrpc.CollaboratorResponse
	if msg.Msg != "" {
		resp.Message = msg.Msg
		return &resp
	}

	collabs, _ := r.([]reference.Collaborator)
	for i := 0; i < len(collabs); i++ {
		c := collabs[i]
		resp.Collaborators = append(resp.Collaborators, CollaboratorToGRPCCollaborator(&c))
	}

	return &resp
}

func CopyObjectResponseCreator(r interface{}) *blobbergrpc.CopyObjectResponse {
	if r == nil {
		return nil
	}

	httpResp, _ := r.(*blobberhttp.UploadResult)
	return &blobbergrpc.CopyObjectResponse{
		Filename:     httpResp.Filename,
		Size:         httpResp.Size,
		ContentHash:  httpResp.Hash,
		MerkleRoot:   httpResp.MerkleRoot,
		UploadLength: httpResp.UploadLength,
		UploadOffset: httpResp.UploadOffset,
	}
}

func RenameObjectResponseCreator(r interface{}) *blobbergrpc.RenameObjectResponse {
	if r == nil {
		return nil
	}

	httpResp, _ := r.(*blobberhttp.UploadResult)
	return &blobbergrpc.RenameObjectResponse{
		Filename:     httpResp.Filename,
		Size:         httpResp.Size,
		ContentHash:  httpResp.Hash,
		MerkleRoot:   httpResp.MerkleRoot,
		UploadLength: httpResp.UploadLength,
		UploadOffset: httpResp.UploadOffset,
	}
}

func DownloadFileResponseCreator(r interface{}) *blobbergrpc.DownloadFileResponse {
	if r == nil {
		return nil
	}

	switch httpResp := r.(type) {
	case []byte:
		return &blobbergrpc.DownloadFileResponse{
			Data: httpResp,
		}
	case *blobberhttp.DownloadResponse:
		return &blobbergrpc.DownloadFileResponse{
			Success:      httpResp.Success,
			Data:         httpResp.Data,
			AllocationId: httpResp.AllocationID,
			Path:         httpResp.Path,
		}
	}

	return nil
}

func UploadFileResponseCreator(r interface{}) *blobbergrpc.UploadFileResponse {
	if r == nil {
		return nil
	}

	httpResp, _ := r.(*blobberhttp.UploadResult)
	return &blobbergrpc.UploadFileResponse{
		Filename:     httpResp.Filename,
		Size:         httpResp.Size,
		ContentHash:  httpResp.Hash,
		MerkleRoot:   httpResp.MerkleRoot,
		UploadLength: httpResp.UploadLength,
		UploadOffset: httpResp.UploadOffset,
	}
}
