package bolt

import (
	"errors"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	pandora ".."
	"./internal"
)

var _ pandora.RawFactoidService = &RawFactoidService{}

const (
	factBucket     = "factoids"
	factTrigBucket = "factoid_trigger_index"

	// MaxFactoidFetch The maximum number of factoids that can be fetched together
	MaxFactoidFetch = 100
)

// RawFactoidService BoltDB implementation of FactoidService interface.
type RawFactoidService struct {
	DB  *bolt.DB
	Now func() time.Time
}

// MarshallFactoid Marshals from *pandora.Factoid to protobuf bytes.
func MarshallFactoid(pf *pandora.Factoid) ([]byte, error) {
	responses := make([]*internal.FactoidResponse, 0, len(pf.Responses))
	for _, r := range pf.Responses {
		v, err := packageFactoidResponse(r)
		if err != nil {
			return nil, err
		}
		responses = append(responses, v)
	}

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
		Responses:   responses,
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

	responses := make([]*pandora.FactoidResponse, 0, len(pf.Responses))
	for _, r := range pf.Responses {
		v, errR := unpackageFactoidResponse(r)
		if errR != nil {
			return nil, errR
		}
		responses = append(responses, v)
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
		Responses:   responses,
		Trigger:     pf.Trigger,
	}
	return f, nil
}

func packageFactoidResponse(pf *pandora.FactoidResponse) (*internal.FactoidResponse, error) {
	dateCreated, err := ptypes.TimestampProto(pf.DateCreated)
	if err != nil {
		return nil, err
	}
	dateEdited, err := ptypes.TimestampProto(pf.DateEdited)
	if err != nil {
		return nil, err
	}
	f := &internal.FactoidResponse{
		DateCreated: dateCreated,
		DateEdited:  dateEdited,
		Response:    pf.Response,
	}
	return f, nil
}

func unpackageFactoidResponse(pf *internal.FactoidResponse) (*pandora.FactoidResponse, error) {
	dateCreated, err := ptypes.Timestamp(pf.DateCreated)
	if err != nil {
		return nil, err
	}
	dateEdited, err := ptypes.Timestamp(pf.DateEdited)
	if err != nil {
		return nil, err
	}
	f := &pandora.FactoidResponse{
		DateCreated: dateCreated,
		DateEdited:  dateEdited,
		Response:    pf.Response,
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
func (s *RawFactoidService) Factoid(id uint64) (*pandora.Factoid, error) {
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
func (s *RawFactoidService) FactoidByTrigger(trigger string) (*pandora.Factoid, error) {
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

// FactoidRange Fetches `count` factoids, starting at `fromID`.
func (s *RawFactoidService) FactoidRange(fromID, count uint64) (factoids []*pandora.Factoid, err error) {
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

// InsertFactoid Inserts factoid with a given ID into BoltDB.
func (s *RawFactoidService) InsertFactoid(pf *pandora.Factoid) (id uint64, err error) {
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

	buf, err := MarshallFactoid(pf)
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

// PutFactoid Inserts factoid with a given ID into BoltDB.
func (s *RawFactoidService) PutFactoid(id uint64, pf *pandora.Factoid) error {
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
	buf, err := MarshallFactoid(pf)
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

// PutFactoidByTrigger Inserts factoid with a given trigger into BoltDB.
func (s *RawFactoidService) PutFactoidByTrigger(trigger string, pf *pandora.Factoid) error {
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
	buf, err := MarshallFactoid(pf)
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

// DeleteFactoid Deletes a factoid with a given trigger from BoltDB.
func (s *RawFactoidService) DeleteFactoid(id uint64) error {
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
