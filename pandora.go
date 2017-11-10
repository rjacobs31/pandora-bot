package pandora

import (
	"github.com/rjacobs31/pandora-bot/bolt/ftypes"
)

type DataClient interface {
	FactoidService
}

type RawFactoidService interface {
	Factoid(id uint64) (*ftypes.Factoid, error)
	FactoidByTrigger(trigger string) (*ftypes.Factoid, error)
	FactoidRange(fromID, count uint64) (factoids []*ftypes.Factoid, err error)
	InsertFactoid(f *ftypes.Factoid) (id uint64, err error)
	PutFactoid(id uint64, f *ftypes.Factoid) error
	PutFactoidByTrigger(trigger string, f *ftypes.Factoid) error
	DeleteFactoid(id uint64) error
}

type FactoidService interface {
	PutResponse(trigger, response string) error
	RandomResponse(trigger string) (string, error)
}

// Interpolator replaces values in a string, based on a map.
type Interpolator interface {
	Interpolate(template string, lookup map[string]interface{}) (string, error)
}
