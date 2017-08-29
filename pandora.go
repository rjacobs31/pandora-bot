package pandora

import "time"

type FactoidResponse struct {
	ID          uint64
	FactoidID   uint64
	DateCreated time.Time
	DateEdited  time.Time
	Response    string
}

type Factoid struct {
	ID          uint64
	DateCreated time.Time
	DateEdited  time.Time
	Protected   bool
	Responses   []*FactoidResponse
	Trigger     string
}

type DataClient interface {
	FactoidService
	RawFactoidService
}

type RawFactoidService interface {
	Factoid(id uint64) (*Factoid, error)
	FactoidByTrigger(trigger string) (*Factoid, error)
	FactoidRange(fromID, count uint64) (factoids []*Factoid, err error)
	InsertFactoid(f *Factoid) (id uint64, err error)
	PutFactoid(id uint64, f *Factoid) error
	PutFactoidByTrigger(trigger string, f *Factoid) error
	DeleteFactoid(id uint64) error
}

type FactoidResponseService interface {
	FactoidResponse(id uint64) (r *FactoidResponse, ok bool)
	Create(r *FactoidResponse) (id uint64, err error)
	Put(id uint64, r *FactoidResponse) (err error)
	Delete(id uint64) (err error)
}

type FactoidService interface {
	PutResponse(trigger, response string) error
	RandomResponse(trigger string) (string, error)
}

// Interpolator replaces values in a string, based on a map.
type Interpolator interface {
	Interpolate(template string, lookup map[string]interface{}) (string, error)
}
