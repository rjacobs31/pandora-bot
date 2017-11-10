package bolt

import (
	"bytes"
	"math/rand"
	"strings"
	"unicode"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	pandora "github.com/rjacobs31/pandora-bot"
	"github.com/rjacobs31/pandora-bot/bolt/ftypes"
	"github.com/rjacobs31/pandora-bot/bolt/raw"
)

var _ pandora.FactoidService = &FactoidService{}

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
		f      *ftypes.Factoid
		uintID uint64
	)

	if s.Now == nil {
		s.Now = ptypes.TimestampNow
	}
	now := s.Now()
	f, err = raw.FetchFactoidByTrigger(tx, trigger)
	if err != nil {
		return err
	}
	if f == nil {
		f = &ftypes.Factoid{
			ID:          uintID,
			DateCreated: now,
			DateEdited:  now,
			Protected:   false,
			Responses:   map[uint64]*ftypes.FactoidResponse{},
			Trigger:     trigger,
		}
	}
	f.DateEdited = now

	r := &ftypes.FactoidResponse{
		DateCreated: now,
		DateEdited:  now,
		Response:    *proto.String(response),
	}
	var highest uint64
	var replaced bool
	for k, v := range f.Responses {
		if k > highest {
			highest = k
		}
		if v.Response == r.Response {
			replaced = true
			break
		}
	}
	if !replaced {
		f.Responses[highest+1] = r
	}
	raw.PutFactoidByTrigger(tx, f)
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
	if f, err := raw.FetchFactoidByTrigger(tx, trigger); err != nil {
		return "", err
	} else if f != nil && len(f.Responses) > 0 {
		i := rand.Intn(len(f.Responses))
		for _, v := range f.Responses {
			if i <= 0 {
				return v.Response, nil
			}
			i--
		}
	}
	return "", nil
}
