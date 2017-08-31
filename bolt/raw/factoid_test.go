package raw_test

import (
	"testing"

	raw "."
	pandora "../.."
)

func TestFactoidService(t *testing.T) {
	db, close, err := newTestDB()
	if db == nil || close == nil || err != nil {
		panic(err)
	}
	defer close()
	defer db.Close()
	s, err := raw.NewFactoidService(db)
	if err != nil {
		t.Error("FactoidService creation failed")
	}

	f := &pandora.Factoid{
		Trigger: "Blah",
	}
	id, err := s.Create(f)
	if id == 0 || err != nil {
		t.Error("Factoid creation failed")
	}

	if err = s.Delete(f.ID); err != nil {
		t.Error("FactoidResponse delete failed")
	}
	if f, ok := s.Factoid(f.ID); f != nil || ok {
		t.Error("FactoidResponse available after delete")
	}

	return
}
