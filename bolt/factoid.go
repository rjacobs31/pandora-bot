package bolt

import (
	"bytes"
	"errors"
	"math/rand"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	pandora ".."
	"./internal"
)

var _ pandora.FactoidService = &FactoidService{}
var _ pandora.RawFactoidService = &RawFactoidService{}

const (
	factBucket     = "factoids"
	factTrigBucket = "factoid_trigger_index"

	// MaxFactoidFetch The maximum number of factoids that can be fetched together
	MaxFactoidFetch = 100
)

// FactoidService BoltDB implementation of FactoidService interface.
type FactoidService struct {
	DB  *bolt.DB
	Now func() *timestamp.Timestamp
}

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

// CleanTrigger Clean a trigger for use as DB key
func CleanTrigger(trigger string) string {
	trigger = strings.TrimSpace(strings.ToLower(trigger))
	result := bytes.Buffer{}
	var prev rune

	for _, c := range trigger {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			result.WriteRune(c)
		} else if unicode.IsSpace(c) || unicode.IsSymbol(c) {
			c = ' '
			if prev != ' ' {
				result.WriteRune(' ')
			}
		}
		prev = c
	}

	return result.String()
}

// PutResponse insert given response under trigger
func (s *FactoidService) PutResponse(trigger, response string) error {
	tx, err := s.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	trigger = CleanTrigger(trigger)
	var (
		f      internal.Factoid
		id     []byte
		uintID uint64
	)
	b := tx.Bucket([]byte(factBucket))
	bt := tx.Bucket([]byte(factTrigBucket))

	// get existing factoid ID or assign one if not present
	if id = bt.Get([]byte(trigger)); id == nil || len(id) == 0 {
		uintID, _ = b.NextSequence()
		id = itob(uintID)
		bt.Put([]byte(trigger), id)
	}

	if s.Now == nil {
		s.Now = ptypes.TimestampNow
	}
	now := s.Now()
	if buf := b.Get(id); buf == nil || len(buf) == 0 {
		f = internal.Factoid{
			ID:          uintID,
			DateCreated: now,
			DateEdited:  now,
			Protected:   false,
			Responses:   []*internal.FactoidResponse{},
			Trigger:     trigger,
		}
	} else if err = proto.Unmarshal(buf, &f); err != nil {
		return err
	}
	f.DateEdited = now

	// sort responses and find insertion index
	sort.Slice(f.Responses[:], func(i int, j int) bool {
		return f.Responses[i].Response < f.Responses[j].Response
	})
	i := sort.Search(len(f.Responses), func(i int) bool {
		return f.Responses[i].Response >= response
	})

	r := &internal.FactoidResponse{
		DateCreated: now,
		DateEdited:  now,
		Response:    *proto.String(response),
	}
	if i < len(f.Responses) {
		if f.Responses[i].Response == response {
			return FactoidAlreadyExistsError(response)
		}
		f.Responses = append(f.Responses, nil)
		copy(f.Responses[i+1:], f.Responses[i:])
		f.Responses[i] = r
	} else {
		f.Responses = append(f.Responses, r)
	}

	enc, err := proto.Marshal(&f)
	if err != nil {
		return err
	}
	if err := bt.Put([]byte(trigger), id); err != nil {
		return err
	}
	if err := b.Put(id, enc); err != nil {
		return err
	}
	return tx.Commit()
}

// RandomResponse fetch random response
func (s *FactoidService) RandomResponse(trigger string) (string, error) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	trigger = CleanTrigger(trigger)
	b := tx.Bucket([]byte(factBucket))
	bt := tx.Bucket([]byte(factTrigBucket))
	if id := bt.Get([]byte(trigger)); id == nil || len(id) <= 0 {
		return "", errors.New("no factoid exists")
	} else if buf := b.Get(id); buf == nil || len(buf) <= 0 {
		return "", errors.New("no factoid exists")
	} else if f, err := UnmarshalFactoid(buf); err != nil {
		return "", err
	} else if len(f.Responses) > 0 {
		r := f.Responses[rand.Intn(len(f.Responses))]
		return r.Response, nil
	}
	return "", nil
}
