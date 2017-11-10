package raw_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"

	pandora "github.com/rjacobs31/pandora-bot"
	"github.com/rjacobs31/pandora-bot/bolt/raw"
)

func newTestDB() (db *bolt.DB, close func() error, err error) {
	// Generate temporary filename.
	f, err := ioutil.TempFile("", "pandora-bolt-client-")
	if err != nil {
		panic(err)
	}
	f.Close()

	db, err = bolt.Open(f.Name(), 0600, &bolt.Options{Timeout: 1 * time.Second})
	close = func() error {
		return os.Remove(f.Name())
	}
	return
}

func TestFactoidResponse(t *testing.T) {
	db, close, err := newTestDB()
	if db == nil || close == nil || err != nil {
		panic(err)
	}
	defer close()
	defer db.Close()
	s, err := raw.NewFactoidResponseService(db)
	if err != nil {
		t.Error("FactoidResponseService creation failed")
	}

	r := &pandora.FactoidResponse{
		Response: "Blah",
	}
	id, err := s.Create(r)
	if id != 0 || err.Error() != "FactoidResponseService: Put without FactoidID" {
		t.Error("FactoidResponse creation successful without factoid ID")
	}
	r.FactoidID = 1
	id, err = s.Create(r)
	if id == 0 || err != nil {
		t.Error("FactoidResponse creation failed")
	}

	r, ok := s.FactoidResponse(id)
	if !ok || r == nil {
		t.Error("FactoidResponse get failed")
	} else if r.Response != "Blah" {
		t.Error("FactoidResponse get returned wrong response")
	}

	r2 := &pandora.FactoidResponse{
		FactoidID: 2,
		Response:  "Honk",
	}
	id, err = s.Create(r2)
	if id == 0 || err != nil {
		t.Error("FactoidResponse second creation failed")
	}

	r.Response = "Blarg"
	err = s.Put(r.ID, r)
	if err != nil {
		t.Error("FactoidResponse put failed")
	}

	if r, ok = s.FactoidResponse(r.ID); r == nil || !ok {
		t.Error("FactoidResponse put->get failed")
	} else if r.Response != "Blarg" {
		t.Error("FactoidResponse put->get wrong response")
	}

	if err = s.Delete(r.ID); err != nil {
		t.Error("FactoidResponse delete failed")
	}
	if s.Exist(r.ID) {
		t.Error("FactoidResponse exists after delete")
	}
	if r, ok = s.FactoidResponse(r.ID); r != nil || ok {
		t.Error("FactoidResponse available after delete")
	}

	return
}
