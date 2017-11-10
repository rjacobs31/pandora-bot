package raw

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/ptypes"

	pandora "github.com/rjacobs31/pandora-bot"
	"github.com/rjacobs31/pandora-bot/bolt/ftypes"
)

var _ pandora.FactoidService = &FactoidService{}

const (
	factoidBucketName      = "Factoids"
	triggerIndexBucketName = "FactoidTriggerIndex"

	// MaxFactoidFetch The maximum number of factoids that can be fetched together
	MaxFactoidFetch = 100
)

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
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(factoidBucketName))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(triggerIndexBucketName))
		return err
	})

	if err == nil {
		s = &FactoidService{DB: db}
	}
	return
}

func fetchFactoid(tx *bolt.Tx, id uint64) []byte {
	b := FactoidBucket(tx)
	return b.Get(ItoB(id))
}

func fetchFactoidByTrigger(tx *bolt.Tx, trigger string) []byte {
	b := FactoidBucket(tx)
	bt := TriggerIndexBucket(tx)
	id := bt.Get([]byte(trigger))
	return b.Get(id)
}

// Factoid Fetches factoid with a given ID from BoltDB.
func (s *FactoidService) Factoid(id uint64) (f *ftypes.Factoid, ok bool) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()

	b := FactoidBucket(tx)
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
func (s *FactoidService) FactoidByTrigger(trigger string) (f *ftypes.Factoid, ok bool) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()

	b := FactoidBucket(tx)
	bt := TriggerIndexBucket(tx)

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
func (s *FactoidService) Range(fromID, count uint64) (factoids []*ftypes.Factoid, err error) {
	var f *ftypes.Factoid

	if count > MaxFactoidFetch {
		count = MaxFactoidFetch
	}

	tx, err := s.DB.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := FactoidBucket(tx)
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
func (s *FactoidService) Create(pf *ftypes.Factoid) (id uint64, err error) {
	if pf.Trigger == "" {
		return 0, errors.New("FactoidService: Empty trigger")
	}

	tx, err := s.DB.Begin(true)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	b := FactoidBucket(tx)
	bt := TriggerIndexBucket(tx)

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
	now := ptypes.TimestampNow()
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
func (s *FactoidService) Put(id uint64, pf *ftypes.Factoid) error {
	tx, err := s.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	trigger := pf.Trigger

	if s.Now == nil {
		s.Now = time.Now
	}
	now := ptypes.TimestampNow()
	pf.DateEdited = now

	b := FactoidBucket(tx)
	bt := TriggerIndexBucket(tx)

	// Clear old trigger index if it exists and differs
	oldBytes := b.Get(ItoB(id))
	if oldBytes != nil && len(oldBytes) != 0 {
		var oldF *ftypes.Factoid
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

	b := FactoidBucket(tx)
	bt := TriggerIndexBucket(tx)
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
