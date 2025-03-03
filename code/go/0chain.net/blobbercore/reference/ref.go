package reference

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/datastore"
	"github.com/0chain/blobber/code/go/0chain.net/core/common"
	"github.com/0chain/blobber/code/go/0chain.net/core/encryption"

	"gorm.io/gorm"
)

const (
	FILE      = "f"
	DIRECTORY = "d"

	CHUNK_SIZE = 64 * 1024

	DIR_LIST_TAG  = "dirlist"
	FILE_LIST_TAG = "filelist"
)

type Ref struct {
	ID                  int64  `gorm:"column:id;primaryKey"`
	Type                string `gorm:"column:type;size:1" dirlist:"type" filelist:"type"`
	AllocationID        string `gorm:"column:allocation_id;size:64;not null;index:idx_path_alloc,priority:1;index:idx_lookup_hash_alloc,priority:1" dirlist:"allocation_id" filelist:"allocation_id"`
	LookupHash          string `gorm:"column:lookup_hash;size:64;not null;index:idx_lookup_hash_alloc,priority:2" dirlist:"lookup_hash" filelist:"lookup_hash"`
	Name                string `gorm:"column:name;size:100;not null" dirlist:"name" filelist:"name"`
	Path                string `gorm:"column:path;size:1000;not null;index:idx_path_alloc,priority:2;index:path_idx" dirlist:"path" filelist:"path"`
	Hash                string `gorm:"column:hash;size:64;not null" dirlist:"hash" filelist:"hash"`
	NumBlocks           int64  `gorm:"column:num_of_blocks;not null;default:0" dirlist:"num_of_blocks" filelist:"num_of_blocks"`
	PathHash            string `gorm:"column:path_hash;size:64;not null" dirlist:"path_hash" filelist:"path_hash"`
	ParentPath          string `gorm:"column:parent_path;size:999"`
	PathLevel           int    `gorm:"column:level;not null;default:0"`
	CustomMeta          string `gorm:"column:custom_meta;not null" filelist:"custom_meta"`
	ContentHash         string `gorm:"column:content_hash;size:64;not null" filelist:"content_hash"`
	Size                int64  `gorm:"column:size;not null;default:0" dirlist:"size" filelist:"size"`
	MerkleRoot          string `gorm:"column:merkle_root;size:64;not null" filelist:"merkle_root"`
	ActualFileSize      int64  `gorm:"column:actual_file_size;not null;default:0" filelist:"actual_file_size"`
	ActualFileHash      string `gorm:"column:actual_file_hash;size:64;not null" filelist:"actual_file_hash"`
	MimeType            string `gorm:"column:mimetype;size:64;not null" filelist:"mimetype"`
	WriteMarker         string `gorm:"column:write_marker;size:64;not null"`
	ThumbnailSize       int64  `gorm:"column:thumbnail_size;not null;default:0" filelist:"thumbnail_size"`
	ThumbnailHash       string `gorm:"column:thumbnail_hash;size:64;not null" filelist:"thumbnail_hash"`
	ActualThumbnailSize int64  `gorm:"column:actual_thumbnail_size;not null;default:0" filelist:"actual_thumbnail_size"`
	ActualThumbnailHash string `gorm:"column:actual_thumbnail_hash;size:64;not null" filelist:"actual_thumbnail_hash"`
	EncryptedKey        string `gorm:"column:encrypted_key;size:64" filelist:"encrypted_key"`
	Children            []*Ref `gorm:"-"`
	childrenLoaded      bool
	OnCloud             bool `gorm:"column:on_cloud;default:false" filelist:"on_cloud"`

	CommitMetaTxns []CommitMetaTxn  `gorm:"foreignkey:ref_id" filelist:"commit_meta_txns"`
	CreatedAt      common.Timestamp `gorm:"column:created_at;index:idx_created_at,sort:desc" dirlist:"created_at" filelist:"created_at"`
	UpdatedAt      common.Timestamp `gorm:"column:updated_at;index:idx_updated_at,sort:desc;" dirlist:"updated_at" filelist:"updated_at"`

	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"` // soft deletion

	ChunkSize        int64 `gorm:"column:chunk_size;not null;default:65536" dirlist:"chunk_size" filelist:"chunk_size"`
	HashToBeComputed bool  `gorm:"-"`
}

// BeforeCreate Hook that gets executed to update create and update date
func (ref *Ref) BeforeCreate(tx *gorm.DB) (err error) {
	if !(ref.CreatedAt > 0) {
		return fmt.Errorf("invalid timestamp value while creating for path %s", ref.Path)
	}
	ref.UpdatedAt = ref.CreatedAt
	return nil
}

func (ref *Ref) BeforeSave(tx *gorm.DB) (err error) {
	if !(ref.UpdatedAt > 0) {
		return fmt.Errorf("invalid timestamp value while updating %s", ref.Path)
	}
	return nil
}

func (Ref) TableName() string {
	return TableNameReferenceObjects
}

type PaginatedRef struct { //Gorm smart select fields.
	ID                  int64  `gorm:"column:id" json:"id,omitempty"`
	Type                string `gorm:"column:type" json:"type,omitempty"`
	AllocationID        string `gorm:"column:allocation_id" json:"allocation_id,omitempty"`
	LookupHash          string `gorm:"column:lookup_hash" json:"lookup_hash,omitempty"`
	Name                string `gorm:"column:name" json:"name,omitempty"`
	Path                string `gorm:"column:path" json:"path,omitempty"`
	Hash                string `gorm:"column:hash" json:"hash,omitempty"`
	NumBlocks           int64  `gorm:"column:num_of_blocks" json:"num_blocks,omitempty"`
	PathHash            string `gorm:"column:path_hash" json:"path_hash,omitempty"`
	ParentPath          string `gorm:"column:parent_path" json:"parent_path,omitempty"`
	PathLevel           int    `gorm:"column:level" json:"level,omitempty"`
	CustomMeta          string `gorm:"column:custom_meta" json:"custom_meta,omitempty"`
	ContentHash         string `gorm:"column:content_hash" json:"content_hash,omitempty"`
	Size                int64  `gorm:"column:size" json:"size,omitempty"`
	MerkleRoot          string `gorm:"column:merkle_root" json:"merkle_root,omitempty"`
	ActualFileSize      int64  `gorm:"column:actual_file_size" json:"actual_file_size,omitempty"`
	ActualFileHash      string `gorm:"column:actual_file_hash" json:"actual_file_hash,omitempty"`
	MimeType            string `gorm:"column:mimetype" json:"mimetype,omitempty"`
	WriteMarker         string `gorm:"column:write_marker" json:"write_marker,omitempty"`
	ThumbnailSize       int64  `gorm:"column:thumbnail_size" json:"thumbnail_size,omitempty"`
	ThumbnailHash       string `gorm:"column:thumbnail_hash" json:"thumbnail_hash,omitempty"`
	ActualThumbnailSize int64  `gorm:"column:actual_thumbnail_size" json:"actual_thumbnail_size,omitempty"`
	ActualThumbnailHash string `gorm:"column:actual_thumbnail_hash" json:"actual_thumbnail_hash,omitempty"`
	EncryptedKey        string `gorm:"column:encrypted_key" json:"encrypted_key,omitempty"`

	OnCloud   bool             `gorm:"column:on_cloud" json:"on_cloud,omitempty"`
	CreatedAt common.Timestamp `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt common.Timestamp `gorm:"column:updated_at" json:"updated_at,omitempty"`

	ChunkSize int64 `gorm:"column:chunk_size" json:"chunk_size"`
}

// GetReferenceLookup hash(allocationID + ":" + path)
func GetReferenceLookup(allocationID, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}

func NewDirectoryRef() *Ref {
	return &Ref{Type: DIRECTORY}
}

func NewFileRef() *Ref {
	return &Ref{Type: FILE}
}

// Mkdir create dirs if they don't exits. do nothing if dir exists. last dir will be return without child
func Mkdir(ctx context.Context, allocationID, destpath string) (*Ref, error) {
	var dirRef *Ref
	db := datastore.GetStore().GetTransaction(ctx)
	// cleaning path to avoid edge case issues: append '/' prefix if not added and removing suffix '/' if added
	destpath = strings.TrimSuffix(filepath.Clean("/"+destpath), "/")
	dirs := strings.Split(destpath, "/")

	for i := range dirs {
		currentPath := filepath.Join("/", filepath.Join(dirs[:i+1]...))
		ref, err := GetReference(ctx, allocationID, currentPath)
		if err == nil {
			dirRef = ref
			continue
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// unexpected sql error
			return nil, err
		}

		// dir doesn't exists , create it
		newRef := NewDirectoryRef()
		newRef.AllocationID = allocationID
		newRef.Path = currentPath
		newRef.ParentPath = filepath.Join("/", filepath.Join(dirs[:i]...))
		newRef.Name = dirs[i]
		newRef.Type = DIRECTORY
		newRef.PathLevel = i + 1
		newRef.LookupHash = GetReferenceLookup(allocationID, newRef.Path)
		err = db.Create(newRef).Error
		if err != nil {
			return nil, err
		}

		dirRef = newRef
	}

	return dirRef, nil
}

// GetReference get FileRef with allcationID and path from postgres
func GetReference(ctx context.Context, allocationID, path string) (*Ref, error) {
	ref := &Ref{}
	db := datastore.GetStore().GetTransaction(ctx)
	err := db.Where(&Ref{AllocationID: allocationID, Path: path}).First(ref).Error
	if err != nil {
		return nil, err
	}
	return ref, nil
}

// GetLimitedRefFieldsByPath get FileRef selected fields with allocationID and path from postgres
func GetLimitedRefFieldsByPath(ctx context.Context, allocationID, path string, selectedFields []string) (*Ref, error) {
	ref := &Ref{}
	db := datastore.GetStore().GetTransaction(ctx)
	db = db.Select(selectedFields)
	err := db.Where(&Ref{AllocationID: allocationID, Path: path}).First(ref).Error
	if err != nil {
		return nil, err
	}
	return ref, nil
}

// GetLimitedRefFieldsByLookupHash get FileRef selected fields with allocationID and lookupHash from postgres
func GetLimitedRefFieldsByLookupHashWith(ctx context.Context, db *gorm.DB, allocationID, lookupHash string, selectedFields []string) (*Ref, error) {
	ref := &Ref{}

	err := db.
		Select(selectedFields).
		Where(&Ref{AllocationID: allocationID, LookupHash: lookupHash}).
		First(ref).Error

	if err != nil {
		return nil, err
	}
	return ref, nil
}

// GetLimitedRefFieldsByLookupHash get FileRef selected fields with allocationID and lookupHash from postgres
func GetLimitedRefFieldsByLookupHash(ctx context.Context, allocationID, lookupHash string, selectedFields []string) (*Ref, error) {
	ref := &Ref{}
	db := datastore.GetStore().GetTransaction(ctx)
	db = db.Select(selectedFields)
	err := db.Where(&Ref{AllocationID: allocationID, LookupHash: lookupHash}).First(ref).Error
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func GetReferenceByLookupHash(ctx context.Context, allocationID, pathHash string) (*Ref, error) {
	ref := &Ref{}
	db := datastore.GetStore().GetTransaction(ctx)
	err := db.Where(&Ref{AllocationID: allocationID, LookupHash: pathHash}).First(ref).Error
	if err != nil {
		return nil, err
	}
	return ref, nil
}

// IsRefExist checks if ref with given path exists and returns error other than gorm.ErrRecordNotFound
func IsRefExist(ctx context.Context, allocationID, path string) (bool, error) {
	db := datastore.GetStore().GetTransaction(ctx)
	var count int64
	if err := db.Model(&Ref{}).Where("allocation_id=? AND path=?", allocationID, path).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetRefsTypeFromPaths Give list of paths it will return refs of respective path with only Type and Path selected in sql query
func GetRefsTypeFromPaths(ctx context.Context, allocationID string, paths []string) (refs []*Ref, err error) {
	if len(paths) == 0 {
		return
	}

	db := datastore.GetStore().GetTransaction(ctx)
	db = db.Select("path", "type")
	for _, p := range paths {
		db = db.Or(Ref{AllocationID: allocationID, Path: p})
	}

	err = db.Find(&refs).Error
	return
}

func GetSubDirsFromPath(p string) []string {
	path := p
	parent, cur := filepath.Split(path)
	parent = filepath.Clean(parent)
	subDirs := make([]string, 0)
	for len(cur) > 0 {
		if cur == "." {
			break
		}
		subDirs = append([]string{cur}, subDirs...)
		parent, cur = filepath.Split(parent)
		parent = filepath.Clean(parent)
	}
	return subDirs
}

func GetRefWithChildren(ctx context.Context, allocationID, path string) (*Ref, error) {
	var refs []Ref
	db := datastore.GetStore().GetTransaction(ctx)
	db = db.Where(Ref{ParentPath: path, AllocationID: allocationID}).Or(Ref{Type: DIRECTORY, Path: path, AllocationID: allocationID})
	err := db.Order("path").Find(&refs).Error
	if err != nil {
		return nil, err
	}
	if len(refs) == 0 {
		return &Ref{Type: DIRECTORY, Path: path, AllocationID: allocationID}, nil
	}
	curRef := &refs[0]
	if curRef.Path != path {
		return nil, common.NewError("invalid_dir_tree", "DB has invalid tree. Root not found in DB")
	}
	for i := 1; i < len(refs); i++ {
		if refs[i].ParentPath == curRef.Path {
			curRef.Children = append(curRef.Children, &refs[i])
		} else {
			return nil, common.NewError("invalid_dir_tree", "DB has invalid tree.")
		}
	}
	return &refs[0], nil
}

func GetRefWithSortedChildren(ctx context.Context, allocationID, path string) (*Ref, error) {
	var refs []*Ref
	db := datastore.GetStore().GetTransaction(ctx)
	db = db.Where(
		Ref{ParentPath: path, AllocationID: allocationID}).
		Or(Ref{Type: DIRECTORY, Path: path, AllocationID: allocationID})

	err := db.Order("path").Find(&refs).Error
	if err != nil {
		return nil, err
	}

	if len(refs) == 0 {
		return &Ref{Type: DIRECTORY, Path: path, AllocationID: allocationID}, nil
	}

	curRef := refs[0]
	if curRef.Path != path {
		return nil, common.NewError("invalid_dir_tree", "DB has invalid tree. Root not found in DB")
	}

	for i := 1; i < len(refs); i++ {
		if refs[i].ParentPath == curRef.Path {
			curRef.Children = append(curRef.Children, refs[i])
		} else {
			return nil, common.NewError("invalid_dir_tree", "DB has invalid tree.")
		}
	}
	return refs[0], nil
}

func (fr *Ref) GetFileHashData() string {
	hashArray := make([]string, 0, 10)
	hashArray = append(hashArray,
		fr.AllocationID,
		fr.Type, // don't need to add it as well
		fr.Name, // don't see any utility as fr.Path below has name in it
		fr.Path,
		strconv.FormatInt(fr.Size, 10),
		fr.ContentHash,
		fr.MerkleRoot,
		strconv.FormatInt(fr.ActualFileSize, 10),
		fr.ActualFileHash,
		strconv.FormatInt(fr.ChunkSize, 10),
	)

	return strings.Join(hashArray, ":")
}

func (fr *Ref) CalculateFileHash(ctx context.Context, saveToDB bool) (string, error) {

	fr.Hash = encryption.Hash(fr.GetFileHashData())
	fr.NumBlocks = int64(math.Ceil(float64(fr.Size*1.0) / float64(fr.ChunkSize)))
	fr.PathLevel = len(strings.Split(strings.TrimRight(fr.Path, "/"), "/"))
	fr.LookupHash = GetReferenceLookup(fr.AllocationID, fr.Path)
	fr.PathHash = fr.LookupHash

	var err error
	if saveToDB && fr.HashToBeComputed {
		err = fr.SaveFileRef(ctx)
	}
	return fr.Hash, err
}

func (r *Ref) CalculateDirHash(ctx context.Context, saveToDB bool) (h string, err error) {
	if !r.HashToBeComputed {
		h = r.Hash
		return
	}

	l := len(r.Children)

	// if l == 0 && !r.childrenLoaded {
	// 	h = r.Hash
	// 	return
	// }

	defer func() {
		if err == nil && saveToDB {
			err = r.SaveDirRef(ctx)

		}
	}()

	childHashes := make([]string, l)
	childPathHashes := make([]string, l)
	var refNumBlocks, size int64

	for i, childRef := range r.Children {
		if childRef.HashToBeComputed {
			_, err := childRef.CalculateHash(ctx, saveToDB)
			if err != nil {
				return "", err
			}
		}

		childHashes[i] = childRef.Hash
		childPathHashes[i] = childRef.PathHash
		refNumBlocks += childRef.NumBlocks
		size += childRef.Size
	}

	r.Hash = encryption.Hash(strings.Join(childHashes, ":"))
	r.PathHash = encryption.Hash(strings.Join(childPathHashes, ":"))
	r.NumBlocks = refNumBlocks
	r.Size = size
	r.PathLevel = len(GetSubDirsFromPath(r.Path)) + 1
	r.LookupHash = GetReferenceLookup(r.AllocationID, r.Path)

	return r.Hash, err
}

func (r *Ref) CalculateHash(ctx context.Context, saveToDB bool) (string, error) {
	if r.Type == DIRECTORY {
		return r.CalculateDirHash(ctx, saveToDB)
	}
	return r.CalculateFileHash(ctx, saveToDB)
}

func (r *Ref) AddChild(child *Ref) {
	if r.Children == nil {
		r.Children = make([]*Ref, 0)
	}
	var index int
	var ltFound bool
	// Add child in sorted fashion
	for i, ref := range r.Children {
		if strings.Compare(child.Path, ref.Path) == -1 {
			index = i
			ltFound = true
			break
		}
	}
	if ltFound {
		r.Children = append(r.Children[:index+1], r.Children[index:]...)
		r.Children[index] = child
	} else {
		r.Children = append(r.Children, child)
	}
	r.childrenLoaded = true
}

func (r *Ref) RemoveChild(idx int) {
	if idx < 0 {
		return
	}
	r.Children = append(r.Children[:idx], r.Children[idx+1:]...)
	r.childrenLoaded = true
}

func (r *Ref) UpdatePath(newPath, parentPath string) {
	r.Path = newPath
	r.ParentPath = parentPath
	r.PathLevel = len(GetSubDirsFromPath(r.Path)) + 1
	r.LookupHash = GetReferenceLookup(r.AllocationID, r.Path)
}

func DeleteReference(ctx context.Context, refID int64, pathHash string) error {
	if refID <= 0 {
		return common.NewError("invalid_ref_id", "Invalid reference ID to delete")
	}
	db := datastore.GetStore().GetTransaction(ctx)
	return db.Where("path_hash = ?", pathHash).Delete(&Ref{ID: refID}).Error
}

func (r *Ref) SaveFileRef(ctx context.Context) error {
	db := datastore.GetStore().GetTransaction(ctx)
	db = db.Model(r).Where("id = ?", r.ID).Updates(map[string]interface{}{
		"allocation_id":         r.AllocationID,
		"lookup_hash":           r.LookupHash,
		"name":                  r.Name,
		"path":                  r.Path,
		"hash":                  r.Hash,
		"num_of_blocks":         r.NumBlocks,
		"path_hash":             r.PathHash,
		"parent_path":           r.ParentPath,
		"level":                 r.PathLevel,
		"write_marker":          r.WriteMarker,
		"mimetype":              r.MimeType,
		"custom_meta":           r.CustomMeta,
		"thumbnail_hash":        r.ThumbnailHash,
		"thumbnail_size":        r.ThumbnailSize,
		"actual_thumbnail_hash": r.ActualThumbnailHash,
		"actual_thumbnail_size": r.ActualThumbnailSize,
		"encrypted_key":         r.EncryptedKey,
		"content_hash":          r.ContentHash,
		"size":                  r.Size,
		"merkle_root":           r.MerkleRoot,
		"actual_file_size":      r.ActualFileSize,
		"actual_file_hash":      r.ActualFileHash,
		"chunk_size":            r.ChunkSize,
	})
	if errors.Is(db.Error, gorm.ErrRecordNotFound) || db.RowsAffected == 0 {
		err := db.Save(r).Error
		return err
	} else {
		return db.Error
	}
}

func (r *Ref) SaveDirRef(ctx context.Context) error {
	db := datastore.GetStore().GetTransaction(ctx)
	db = db.Model(r).Where("id = ?", r.ID).Updates(map[string]interface{}{
		"allocation_id": r.AllocationID,
		"lookup_hash":   r.LookupHash,
		"name":          r.Name,
		"path":          r.Path,
		"hash":          r.Hash,
		"num_of_blocks": r.NumBlocks,
		"path_hash":     r.PathHash,
		"parent_path":   r.ParentPath,
		"level":         r.PathLevel,
		"write_marker":  r.WriteMarker,
		"content_hash":  r.ContentHash,
		"size":          r.Size,
		"merkle_root":   r.MerkleRoot,
		"chunk_size":    r.ChunkSize,
	})
	if errors.Is(db.Error, gorm.ErrRecordNotFound) || db.RowsAffected == 0 {
		err := db.Save(r).Error
		return err
	} else {
		return db.Error
	}
}

func (r *Ref) Save(ctx context.Context) error {
	db := datastore.GetStore().GetTransaction(ctx)
	return db.Save(r).Error
}

// GetListingData reflect and convert all fields into map[string]interface{}
func (r *Ref) GetListingData(ctx context.Context) map[string]interface{} {
	if r == nil {
		return make(map[string]interface{})
	}

	if r.Type == FILE {
		return GetListingFieldsMap(*r, FILE_LIST_TAG)
	}
	return GetListingFieldsMap(*r, DIR_LIST_TAG)
}

func ListingDataToRef(refMap map[string]interface{}) *Ref {
	if len(refMap) < 1 {
		return nil
	}

	ref := &Ref{}

	refType, _ := refMap["type"].(string)
	var tagName string
	if refType == FILE {
		tagName = FILE_LIST_TAG
	} else {
		tagName = DIR_LIST_TAG
	}

	t := reflect.TypeOf(ref).Elem()
	v := reflect.ValueOf(ref).Elem()

	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get the field tag value
		tag := field.Tag.Get(tagName)
		// Skip if tag is not defined or ignored
		if tag == "" || tag == "-" {
			continue
		}

		val := refMap[tag]
		if val != nil {
			v.FieldByName(field.Name).Set(reflect.ValueOf(val))
		}
	}

	return ref
}

func GetListingFieldsMap(refEntity interface{}, tagName string) map[string]interface{} {
	result := make(map[string]interface{})
	t := reflect.TypeOf(refEntity)
	v := reflect.ValueOf(refEntity)
	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get the field tag value
		tag := field.Tag.Get(tagName)
		// Skip if tag is not defined or ignored
		if !field.Anonymous && (tag == "" || tag == "-") {
			continue
		}

		if field.Anonymous {
			listMap := GetListingFieldsMap(v.FieldByName(field.Name).Interface(), tagName)
			if len(listMap) > 0 {
				for k, v := range listMap {
					result[k] = v
				}
			}
		} else {
			fieldValue := v.FieldByName(field.Name).Interface()
			if fieldValue == nil {
				continue
			}
			result[tag] = fieldValue
		}
	}
	return result
}
