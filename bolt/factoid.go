package bolt

import (
	"bytes"
	"errors"
	"math/rand"
	"sort"
	"strings"
	"unicode"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	pandora "github.com/rjacobs31/pandora-bot"
	"github.com/rjacobs31/pandora-bot/bolt/internal"
)

var _ pandora.SimpleFactoidService = &FactoidService{}

// FactoidService BoltDB implementation of FactoidService interface.
type FactoidService struct {
	DB  *bolt.DB
	Now func() *timestamp.Timestamp
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
