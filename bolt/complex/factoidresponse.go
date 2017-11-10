package raw

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"

	pandora "github.com/rjacobs31/pandora-bot"
	"github.com/rjacobs31/pandora-bot/bolt/ftypes"
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
func MarshalFactoidResponse(pr *ftypes.FactoidResponse) (buf []byte, err error) {
	return proto.Marshal(pr)
}

// UnmarshalFactoidResponse Unmarshals from protobuf bytes to *pandora.FactoidResponse.
func UnmarshalFactoidResponse(b []byte) (pr *ftypes.FactoidResponse, err error) {
	pr = &ftypes.FactoidResponse{}
	err = proto.Unmarshal(b, pr)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

// FactoidResponse Fetches a FactoidResponse with a given ID in BoltDB.
func (s *FactoidResponseService) FactoidResponse(id uint64) (r *ftypes.FactoidResponse, ok bool) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()
	b := responseBucket(tx)

	v := b.Get(ItoB(id))
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
func (s *FactoidResponseService) Create(r *ftypes.FactoidResponse) (id uint64, err error) {
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

	err = b.Put(ItoB(id), buf)
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

	err = b.Delete(ItoB(id))
	if err == nil {
		return tx.Commit()
	}
	return
}

// DeleteForFactoid Deletes FactoidResponses for a given factoid ID from BoltDB.
func (s *FactoidResponseService) DeleteForFactoid(factoidID uint64) (err error) {
	tx, err := s.DB.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()
	b := responseBucket(tx)
	c := b.Cursor()

	for k, buf := c.First(); k != nil; k, buf = c.Next() {
		var r *ftypes.FactoidResponse
		r, err = UnmarshalFactoidResponse(buf)
		if err != nil {
			return
		} else if r.FactoidID == factoidID {
			b.Delete(k)
		}
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

	v := b.Get(ItoB(id))
	return v != nil && len(v) != 0
}

// Put Puts a FactoidResponse under a given ID in BoltDB. Will replace an
// existing FactoidResponse.
func (s *FactoidResponseService) Put(id uint64, r *ftypes.FactoidResponse) (err error) {
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
	err = b.Put(ItoB(id), buf)

	if err == nil {
		return tx.Commit()
	}
	return
}

// ResponseCount Counts the number of responses associated with the given
// factoid ID in BoltDB.
func (s *FactoidResponseService) ResponseCount(id uint64) (count int, err error) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()

	var (
		b = responseBucket(tx)
		c = b.Cursor()
		r *ftypes.FactoidResponse
	)
	for k, buf := c.First(); k != nil; k, buf = c.Next() {
		if r, err = UnmarshalFactoidResponse(buf); err != nil {
			return
		} else if r.FactoidID == id {
			count++
		}
	}
	return
}

// ResponseByIndex Returns the `n`th response associated with the given
// factoid ID in BoltDB.
func (s *FactoidResponseService) ResponseByIndex(id uint64, n uint64) (r *ftypes.FactoidResponse, err error) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()

	b := responseBucket(tx)
	c := b.Cursor()
	i := uint64(0)
	for k, buf := c.First(); k != nil; k, buf = c.Next() {
		if r, err = UnmarshalFactoidResponse(buf); err != nil {
			return
		} else if r.FactoidID == id {
			if i >= n {
				return
			}
			i++
		}
	}

	return nil, errors.New("FactoidService: Outside index")
}

// ResponseRange Returns the `count` responses associated with the given
// factoid ID in BoltDB.
func (s *FactoidResponseService) ResponseRange(id uint64, startID uint64, count uint64) (responses []*ftypes.FactoidResponse, err error) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()

	var (
		b = responseBucket(tx)
		c = b.Cursor()
		i = uint64(0)
		r *ftypes.FactoidResponse
	)
	responses = make([]*ftypes.FactoidResponse, 0, count)
	for k, buf := c.First(); k != nil; k, buf = c.Next() {
		if r, err = UnmarshalFactoidResponse(buf); err != nil {
			return
		} else if r.FactoidID == id {
			responses = append(responses, r)
			if i >= count {
				return
			}
			i++
		}
	}
	return
}
