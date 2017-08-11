package pandora

import "time"

type FactoidResponse struct {
	DateCreated time.Time
	DateEdited  time.Time
	Response    string
}

type Factoid struct {
	DateCreated time.Time
	DateEdited  time.Time
	Protected   bool
	Responses   []*FactoidResponse
}

type DataClient interface {
	FactoidService
	RawFactoidService
}

type RawFactoidService interface {
	GetFactoid(trigger string) (*Factoid, error)
	PutFactoid(trigger string, f *Factoid) error
	DeleteFactoid(trigger string) error
}

type FactoidService interface {
	PutResponse(trigger, response string) error
	GetRandomResponse(trigger string) (string, error)
}

// Interpolator replaces values in a string, based on a map.
type Interpolator interface {
	Interpolate(template string, lookup map[string]interface{}) (string, error)
}
