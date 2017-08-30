package raw

import (
	"errors"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	pandora "../.."
	"../internal"
)

var _ pandora.FactoidService = &FactoidService{}

const (
	factBucket     = "factoids"
	factTrigBucket = "factoid_trigger_index"

	// MaxFactoidFetch The maximum number of factoids that can be fetched together
	MaxFactoidFetch = 100
)

// FactoidService BoltDB implementation of FactoidService interface.
type FactoidService struct {
	DB  *bolt.DB
	Now func() time.Time
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
	b := tx.Bucket([]byte(factBucket))
	return b.Get(itob(id))
}

func fetchFactoidByTrigger(tx *bolt.Tx, trigger string) []byte {
	b := tx.Bucket([]byte(factBucket))
	bt := tx.Bucket([]byte(factTrigBucket))
	id := bt.Get([]byte(trigger))
	return b.Get(id)
}

// Factoid Fetches factoid with a given ID from BoltDB.
func (s *FactoidService) Factoid(id uint64) (*pandora.Factoid, error) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte(factBucket))
	buf := b.Get(itob(id))
	if buf == nil || len(buf) == 0 {
		return nil, errors.New("factoid not exist")
	}
	return UnmarshalFactoid(buf)
}

// FactoidByTrigger Fetches factoid with a given trigger from BoltDB.
func (s *FactoidService) FactoidByTrigger(trigger string) (*pandora.Factoid, error) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte(factBucket))
	bt := tx.Bucket([]byte(factTrigBucket))

	id := bt.Get([]byte(trigger))
	buf := b.Get(id)
	if buf == nil || len(buf) == 0 {
		return nil, errors.New("factoid not exist")
	}
	return UnmarshalFactoid(buf)
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

	b := tx.Bucket([]byte(factBucket))
	cur := b.Cursor()

	k, v := cur.Seek(itob(fromID))
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
	tx, err := s.DB.Begin(true)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	sort.Slice(pf.Responses[:], func(i int, j int) bool {
		return pf.Responses[i].Response < pf.Responses[j].Response
	})

	b := tx.Bucket([]byte(factBucket))
	bt := tx.Bucket([]byte(factTrigBucket))

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

	err = b.Put(itob(id), buf)
	if err != nil {
		return
	}
	err = bt.Put([]byte(trigger), itob(id))
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
	sort.Slice(pf.Responses[:], func(i int, j int) bool {
		return pf.Responses[i].Response < pf.Responses[j].Response
	})

	if s.Now == nil {
		s.Now = time.Now
	}
	now := s.Now()
	pf.DateEdited = now

	b := tx.Bucket([]byte(factBucket))
	bt := tx.Bucket([]byte(factTrigBucket))
	buf, err := MarshalFactoid(pf)
	if err != nil {
		return err
	}

	err = b.Put(itob(id), buf)
	if err != nil {
		return err
	}
	err = bt.Put([]byte(trigger), itob(id))
	if err != nil {
		return err
	}

	return tx.Commit()
}

// PutByTrigger Inserts factoid with a given trigger into BoltDB.
func (s *FactoidService) PutByTrigger(trigger string, pf *pandora.Factoid) error {
	var id []byte
	tx, err := s.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if s.Now == nil {
		s.Now = time.Now
	}
	now := s.Now()
	pf.DateEdited = now

	sort.Slice(pf.Responses[:], func(i int, j int) bool {
		return pf.Responses[i].Response < pf.Responses[j].Response
	})

	b := tx.Bucket([]byte(factBucket))
	bt := tx.Bucket([]byte(factTrigBucket))
	if id = bt.Get([]byte(trigger)); id == nil || len(id) < 1 {
		uintID, _ := b.NextSequence()
		pf.ID = uintID
		id = itob(uintID)
	}
	buf, err := MarshalFactoid(pf)
	if err != nil {
		return err
	}

	err = b.Put(id, buf)
	if err != nil {
		return err
	}
	err = bt.Put([]byte(trigger), id)
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

	b := tx.Bucket([]byte(factBucket))
	bt := tx.Bucket([]byte(factTrigBucket))
	buf := b.Get(itob(id))
	if buf == nil || len(buf) < 1 {
		return errors.New("factoid not exist")
	}
	f, err := UnmarshalFactoid(buf)
	if err != nil {
		return errors.New("unmarshal failed")
	}

	bt.Delete([]byte(f.Trigger))
	b.Delete(itob(id))
	return tx.Commit()
}
