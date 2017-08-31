package raw_test

import (
	"testing"

	raw "."
	pandora "../.."
)

func TestFactoidService(t *testing.T) {
	var (
		ok  bool
		err error
	)
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

	f, ok = s.Factoid(id)
	if !ok || f.ID != id || f.Trigger != "Blah" {
		t.Error("Factoid retrieval failed")
	}

	f.Trigger = "Blarg"
	err = s.Put(id, f)
	if err != nil {
		t.Error("Factoid put failed")
	}
	f, ok = s.Factoid(id)
	if !ok || f.Trigger != "Blarg" {
		t.Error("Factoid put not persisted")
	}

	f.Trigger = "Honk"
	err = s.Put(f.ID, f)
	if err != nil {
		t.Error("Factoid put by trigger failed")
	}
	f, ok = s.FactoidByTrigger("Honk")
	if !ok || f.Trigger != "Honk" {
		t.Errorf("Factoid put by trigger not persisted (got %s)", f.Trigger)
	}

	if err = s.Delete(f.ID); err != nil {
		t.Error("FactoidResponse delete failed")
	}
	if f, ok := s.Factoid(f.ID); f != nil || ok {
		t.Error("FactoidResponse available after delete")
	}

	return
}
