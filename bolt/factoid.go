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

	pandora ".."
	"./internal"
)

var _ pandora.FactoidService = &FactoidService{}
var _ pandora.RawFactoidService = &RawFactoidService{}

// FactoidService BoltDB implementation of FactoidService interface.
type FactoidService struct {
	DB *bolt.DB
}

// RawFactoidService BoltDB implementation of FactoidService interface.
type RawFactoidService struct {
	DB *bolt.DB
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
		DateCreated: dateCreated,
		DateEdited:  dateEdited,
		Protected:   pf.Protected,
		Responses:   responses,
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
		DateCreated: dateCreated,
		DateEdited:  dateEdited,
		Protected:   pf.Protected,
		Responses:   responses,
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

// GetFactoid Fetches factoid with a given trigger from BoltDB.
func (s *RawFactoidService) GetFactoid(trigger string) (*pandora.Factoid, error) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("factoids"))

	buf := b.Get([]byte(trigger))
	if buf == nil || len(buf) == 0 {
		return nil, errors.New("factoid not exist")
	}
	return UnmarshalFactoid(buf)
}

// PutFactoid Inserts factoid with a given trigger into BoltDB.
func (s *RawFactoidService) PutFactoid(trigger string, pf *pandora.Factoid) error {
	tx, err := s.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	sort.Slice(pf.Responses[:], func(i int, j int) bool {
		return pf.Responses[i].Response < pf.Responses[j].Response
	})

	b := tx.Bucket([]byte("factoids"))
	buf, err := MarshallFactoid(pf)
	if err != nil {
		return err
	}
	return b.Put([]byte(trigger), buf)
}

// DeleteFactoid Deletes a factoid with a given trigger from BoltDB.
func (s *RawFactoidService) DeleteFactoid(trigger string) error {
	tx, err := s.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("factoids"))
	return b.Delete([]byte(trigger))
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
	var f internal.Factoid
	b := tx.Bucket([]byte("factoids"))
	now := ptypes.TimestampNow()
	if buf := b.Get([]byte(trigger)); buf == nil || len(buf) == 0 {
		f = internal.Factoid{
			DateCreated: now,
			DateEdited:  now,
			Protected:   false,
			Responses:   []*internal.FactoidResponse{},
		}
	} else if err := proto.Unmarshal(buf, &f); err != nil {
		return err
	}
	f.DateEdited = now

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

	if enc, err := proto.Marshal(&f); err != nil {
		return err
	} else if err := b.Put([]byte(trigger), enc); err != nil {
		return err
	}
	return tx.Commit()
}

// GetRandomResponse fetch random response
func (s *FactoidService) GetRandomResponse(trigger string) (string, error) {
	tx, err := s.DB.Begin(false)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	trigger = CleanTrigger(trigger)
	b := tx.Bucket([]byte("factoids"))
	if buf := b.Get([]byte(trigger)); buf == nil || len(buf) <= 0 {
		return "", errors.New("no factoid exists")
	} else if f, err := UnmarshalFactoid(buf); err != nil {
		return "", err
	} else if len(f.Responses) > 0 {
		r := f.Responses[rand.Intn(len(f.Responses))]
		return r.Response, nil
	}
	return "", nil
}
