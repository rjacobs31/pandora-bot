package bolt_test

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"

	bolt_internal "."
	"./internal"
)

var Now = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

func NewClient() *bolt_internal.Client {
	// Generate temporary filename.
	f, err := ioutil.TempFile("", "pandora-bolt-client-")
	if err != nil {
		panic(err)
	}
	f.Close()

	// Create client wrapper.
	c := bolt_internal.NewClient(f.Name())
	c.Now = func() time.Time { return Now }

	if err := c.Open(); err != nil {
		panic(err)
	}

	return c
}

func CloseClient(c *bolt_internal.Client) error {
	defer os.Remove(c.Path)
	return c.Close()
}

func TestFactoidService(t *testing.T) {
	c := NewClient()
	defer CloseClient(c)

	if err := c.PutResponse("this", "this is a test"); err != nil {
		t.Error(err)
		return
	}
	err := c.DB.View(func(tx *bolt.Tx) (err error) {
		if b := tx.Bucket([]byte("factoids")); b == nil {
			return errors.New("bucket not exist")
		} else if buf := b.Get([]byte("this")); buf == nil || len(buf) < 1 {
			return errors.New("value not exist")
		} else {
			f := &internal.Factoid{}
			proto.Unmarshal(buf, f)
			if len(f.Responses) != 1 {
				return errors.New("unexpected num responses")
			} else if f.Responses[0].Response != "this is a test" {
				return errors.New("unexpected response")
			}
		}
		return
	})
	if err != nil {
		t.Error(err)
		return
	}
	r, err := c.GetRandomResponse("this")
	if err != nil {
		t.Error(err)
		return
	} else if r != "this is a test" {
		t.Error("unexpected random response")
		return
	}
}
