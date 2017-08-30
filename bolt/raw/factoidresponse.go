package raw

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	pandora "../.."
	internal "../internal"
)

var _ pandora.FactoidResponseService = &FactoidResponseService{}

const responseBucketName = "FactoidResponse"

// responseBucket Gets the BoltDB bucket for FactoidResponse objects.
func responseBucket(tx *bolt.Tx) (b *bolt.Bucket) {
	return tx.Bucket([]byte(responseBucketName))
}

// FactoidResponseService BoltDB implementation of a raw factoid response
// service.
type FactoidResponseService struct {
	DB *bolt.DB
}

// NewFactoidResponseService instantiates a new FactoidResponseService.
func NewFactoidResponseService(db *bolt.DB) (s *FactoidResponseService, err error) {
	if db == nil {
		err = errors.New("FactoidResponseService: No DB provided")
		return
	}

	// Initialize top-level buckets.
	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	_, err = tx.CreateBucketIfNotExists([]byte(responseBucketName))
	if err != nil {
		return
	}
	return &FactoidResponseService{DB: db}, tx.Commit()
}

// MarshalFactoidResponse Marshals from *pandora.FactoidResponse to protobuf bytes.
func MarshalFactoidResponse(pf *pandora.FactoidResponse) (buf []byte, err error) {
	dateCreated, err := ptypes.TimestampProto(pf.DateCreated)
	if err != nil {
		return
	}
	dateEdited, err := ptypes.TimestampProto(pf.DateEdited)
	if err != nil {
		return
	}

	r := &internal.FactoidResponse{
		ID:          pf.ID,
		FactoidID:   pf.FactoidID,
		DateCreated: dateCreated,
		DateEdited:  dateEdited,
		Response:    pf.Response,
	}
	return proto.Marshal(r)
}

// UnmarshalFactoidResponse Unmarshals from protobuf bytes to *pandora.FactoidResponse.
func UnmarshalFactoidResponse(b []byte) (r *pandora.FactoidResponse, err error) {
	pf := &internal.FactoidResponse{}
	err = proto.Unmarshal(b, pf)
	if err != nil {
		return
	}

	dateCreated, err := ptypes.Timestamp(pf.DateCreated)
	if err != nil {
		return
	}
	dateEdited, err := ptypes.Timestamp(pf.DateEdited)
	if err != nil {
		return
	}

	r = &pandora.FactoidResponse{
		ID:          pf.ID,
		FactoidID:   pf.FactoidID,
		DateCreated: dateCreated,
		DateEdited:  dateEdited,
		Response:    pf.Response,
	}
	return
}

// FactoidResponse Fetches a FactoidResponse with a given ID in BoltDB.
func (s *FactoidResponseService) FactoidResponse(id uint64) (r *pandora.FactoidResponse, ok bool) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()
	b := responseBucket(tx)

	v := b.Get(itob(id))
	if v == nil || len(v) < 1 {
		return
	}

	r, err = UnmarshalFactoidResponse(v)
	if err == nil {
		ok = true
	}
	return
}

// Create Creates a new FactoidResponse in BoltDB.
func (s *FactoidResponseService) Create(r *pandora.FactoidResponse) (id uint64, err error) {
	if r.FactoidID == 0 {
		return 0, errors.New("FactoidResponseService: Put without FactoidID")
	}

	tx, err := s.DB.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()
	b := responseBucket(tx)

	id, _ = b.NextSequence()
	r.ID = id
	buf, err := MarshalFactoidResponse(r)
	if err != nil {
		return
	}

	err = b.Put(itob(id), buf)
	if err == nil {
		err = tx.Commit()
	}
	return
}

// Delete Deletes a FactoidResponse with a given ID from BoltDB.
func (s *FactoidResponseService) Delete(id uint64) (err error) {
	tx, err := s.DB.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()
	b := responseBucket(tx)

	err = b.Delete(itob(id))
	if err == nil {
		return tx.Commit()
	}
	return
}

// Exist Checks existence of FactoidResponse with a given ID from BoltDB.
func (s *FactoidResponseService) Exist(id uint64) (exists bool) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()
	b := responseBucket(tx)

	v := b.Get(itob(id))
	return v != nil && len(v) != 0
}

// Put Puts a FactoidResponse under a given ID in BoltDB. Will replace an
// existing FactoidResponse.
func (s *FactoidResponseService) Put(id uint64, r *pandora.FactoidResponse) (err error) {
	if id == 0 {
		return errors.New("FactoidResponseService: Put without ID")
	} else if r == nil {
		return errors.New("FactoidResponseService: Put without value")
	} else if r.FactoidID == 0 {
		return errors.New("FactoidResponseService: Put without FactoidID")
	}

	tx, err := s.DB.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()
	b := responseBucket(tx)

	buf, err := MarshalFactoidResponse(r)
	if err != nil {
		return
	}
	err = b.Put(itob(id), buf)

	if err == nil {
		return tx.Commit()
	}
	return
}
