package raw

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	pandora "../.."
	"../internal"
)

var _ pandora.FactoidService = &FactoidService{}

const (
	factoidBucketName      = "Factoids"
	triggerIndexBucketName = "FactoidTriggerIndex"

	// MaxFactoidFetch The maximum number of factoids that can be fetched together
	MaxFactoidFetch = 100
)

// responseBucket Gets the BoltDB bucket for FactoidResponse objects.
func factoidBucket(tx *bolt.Tx) (b *bolt.Bucket) {
	return tx.Bucket([]byte(factoidBucketName))
}

// responseBucket Gets the BoltDB bucket for FactoidResponse objects.
func triggerIndexBucket(tx *bolt.Tx) (b *bolt.Bucket) {
	return tx.Bucket([]byte(triggerIndexBucketName))
}

// FactoidService BoltDB implementation of FactoidService interface.
type FactoidService struct {
	DB  *bolt.DB
	Now func() time.Time
}

// NewFactoidService instantiates a new FactoidService.
func NewFactoidService(db *bolt.DB) (s *FactoidService, err error) {
	if db == nil {
		err = errors.New("FactoidService: No DB provided")
		return
	}

	// Initialize top-level buckets.
	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	_, err = tx.CreateBucketIfNotExists([]byte(factoidBucketName))
	if err != nil {
		return
	}
	_, err = tx.CreateBucketIfNotExists([]byte(triggerIndexBucketName))
	if err != nil {
		return
	}
	return &FactoidService{DB: db}, tx.Commit()
}

// MarshalFactoid Marshals from *pandora.Factoid to protobuf bytes.
func MarshalFactoid(pf *pandora.Factoid) ([]byte, error) {
	dateCreated, err := ptypes.TimestampProto(pf.DateCreated)
	if err != nil {
		return nil, err
	}
	dateEdited, err := ptypes.TimestampProto(pf.DateEdited)
	if err != nil {
		return nil, err
	}

	f := &internal.Factoid{
		ID:          pf.ID,
		DateCreated: dateCreated,
		DateEdited:  dateEdited,
		Protected:   pf.Protected,
		Trigger:     pf.Trigger,
	}
	return proto.Marshal(f)
}

// UnmarshalFactoid Unmarshals from protobuf bytes to *pandora.Factoid.
func UnmarshalFactoid(b []byte) (*pandora.Factoid, error) {
	pf := &internal.Factoid{}
	err := proto.Unmarshal(b, pf)
	if err != nil {
		return nil, err
	}

	dateCreated, err := ptypes.Timestamp(pf.DateCreated)
	if err != nil {
		return nil, err
	}
	dateEdited, err := ptypes.Timestamp(pf.DateEdited)
	if err != nil {
		return nil, err
	}

	f := &pandora.Factoid{
		ID:          pf.ID,
		DateCreated: dateCreated,
		DateEdited:  dateEdited,
		Protected:   pf.Protected,
		Trigger:     pf.Trigger,
	}
	return f, nil
}

func fetchFactoid(tx *bolt.Tx, id uint64) []byte {
	b := factoidBucket(tx)
	return b.Get(ItoB(id))
}

func fetchFactoidByTrigger(tx *bolt.Tx, trigger string) []byte {
	b := factoidBucket(tx)
	bt := triggerIndexBucket(tx)
	id := bt.Get([]byte(trigger))
	return b.Get(id)
}

// Factoid Fetches factoid with a given ID from BoltDB.
func (s *FactoidService) Factoid(id uint64) (f *pandora.Factoid, ok bool) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()

	b := factoidBucket(tx)
	buf := b.Get(ItoB(id))
	if buf == nil || len(buf) == 0 {
		return
	}
	if f, err = UnmarshalFactoid(buf); err == nil {
		ok = true
	}
	return
}

// FactoidByTrigger Fetches factoid with a given trigger from BoltDB.
func (s *FactoidService) FactoidByTrigger(trigger string) (f *pandora.Factoid, ok bool) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()

	b := factoidBucket(tx)
	bt := triggerIndexBucket(tx)

	id := bt.Get([]byte(trigger))
	if id == nil || len(id) == 0 {
		return
	}
	buf := b.Get(id)
	if buf == nil || len(buf) == 0 {
		return
	}
	if f, err = UnmarshalFactoid(buf); err == nil {
		ok = true
	}
	return
}

// Range Fetches `count` factoids, starting at `fromID`.
func (s *FactoidService) Range(fromID, count uint64) (factoids []*pandora.Factoid, err error) {
	var f *pandora.Factoid

	if count > MaxFactoidFetch {
		count = MaxFactoidFetch
	}

	tx, err := s.DB.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := factoidBucket(tx)
	cur := b.Cursor()

	k, v := cur.Seek(ItoB(fromID))
	for i := uint64(0); i < count-1; i++ {
		if k == nil || v == nil {
			break
		}
		f, err = UnmarshalFactoid(v)
		if err != nil {
			return
		}
		factoids = append(factoids, f)
		k, v = cur.Next()
	}
	return
}

// Create Inserts factoid with a given ID into BoltDB.
func (s *FactoidService) Create(pf *pandora.Factoid) (id uint64, err error) {
	if pf.Trigger == "" {
		return 0, errors.New("FactoidService: Empty trigger")
	}

	tx, err := s.DB.Begin(true)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	b := factoidBucket(tx)
	bt := triggerIndexBucket(tx)

	// if ID already present then can't insert, else get new ID
	trigger := pf.Trigger
	if bid := bt.Get([]byte(trigger)); bid != nil || len(bid) > 0 {
		return id, errors.New("factoid already exists")
	}
	id, _ = bt.NextSequence()
	pf.ID = id

	if s.Now == nil {
		s.Now = time.Now
	}
	now := s.Now()
	pf.DateCreated = now
	pf.DateEdited = now

	buf, err := MarshalFactoid(pf)
	if err != nil {
		return
	}

	err = b.Put(ItoB(id), buf)
	if err != nil {
		return
	}
	err = bt.Put([]byte(trigger), ItoB(id))
	if err != nil {
		return
	}

	return id, tx.Commit()
}

// Put Inserts factoid with a given ID into BoltDB.
func (s *FactoidService) Put(id uint64, pf *pandora.Factoid) error {
	tx, err := s.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	trigger := pf.Trigger

	if s.Now == nil {
		s.Now = time.Now
	}
	now := s.Now()
	pf.DateEdited = now

	b := factoidBucket(tx)
	bt := triggerIndexBucket(tx)

	// Clear old trigger index if it exists and differs
	oldBytes := b.Get(ItoB(id))
	if oldBytes != nil && len(oldBytes) != 0 {
		var oldF *pandora.Factoid
		oldF, err = UnmarshalFactoid(oldBytes)
		if err == nil && oldF.Trigger != pf.Trigger {
			bt.Delete([]byte(oldF.Trigger))
		}
	}

	buf, err := MarshalFactoid(pf)
	if err != nil {
		return err
	}

	err = b.Put(ItoB(id), buf)
	if err != nil {
		return err
	}
	err = bt.Put([]byte(trigger), ItoB(id))
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete Deletes a factoid with a given trigger from BoltDB.
func (s *FactoidService) Delete(id uint64) error {
	tx, err := s.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b := factoidBucket(tx)
	bt := triggerIndexBucket(tx)
	buf := b.Get(ItoB(id))
	if buf == nil || len(buf) < 1 {
		return errors.New("factoid not exist")
	}
	f, err := UnmarshalFactoid(buf)
	if err != nil {
		return errors.New("unmarshal failed")
	}

	bt.Delete([]byte(f.Trigger))
	b.Delete(ItoB(id))
	return tx.Commit()
}
