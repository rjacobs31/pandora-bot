package mock

import (
	"fmt"

	pandora ".."
)

var _ pandora.FactoidService = &FactoidService{}

// FactoidService mocked FactoidService which only returns mocked responses
// and marks them as being called. Supports a range of set response types, each
// of which is handled differently.
type FactoidService struct {
	ResponseVal             interface{}
	PutVal                  interface{}
	CalledGetRandomResponse bool
	CalledPutResponse       bool
}

// GetRandomResponse marks the function as called and returns a mocked response,
// based on the type of the response.
func (s *FactoidService) GetRandomResponse(trigger string) (r string, err error) {
	switch v := s.ResponseVal.(type) {
	case string:
		r, err = v, nil
	case error:
		r, err = "", v
	case map[string]string:
		r, _ = v[trigger]
	case func(string) (string, error):
		r, err = v(trigger)
	case map[string]func() (string, error):
		f, ok := v[trigger]
		if ok {
			r, err = f()
		}
	case fmt.Stringer:
		r = v.String()
	}
	s.CalledGetRandomResponse = true
	return
}

// PutResponse marks the function as called and returns a mocked response,
// based on the type of the response.
func (s *FactoidService) PutResponse(trigger, response string) (err error) {
	switch v := s.ResponseVal.(type) {
	case error:
		err = v
	case func(string, string) error:
		err = v(trigger, response)
	case map[string]func() error:
		f, ok := v[trigger]
		if ok {
			err = f()
		}
	}
	s.CalledPutResponse = true
	return
}
