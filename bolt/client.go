package bolt

import (
	"encoding/binary"
	"time"

	"github.com/boltdb/bolt"

	pandora "github.com/rjacobs31/pandora-bot"
	"github.com/rjacobs31/pandora-bot/bolt/raw"
)

var _ pandora.DataClient = &Client{}

// Client A client for connecting to BoltDB.
type Client struct {
	Path string
	Now  func() time.Time
	DB   *bolt.DB

	pandora.FactoidService
}

// NewClient creates a new BoltDB client.
func NewClient(path string) *Client {
	return &Client{
		Path: path,
		Now:  time.Now,
	}
}

// Open Opens a connection to BoltDB.
func (c *Client) Open() error {
	db, err := bolt.Open(c.Path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	c.DB = db
	c.FactoidService = &FactoidService{DB: db}

	// Initialize top-level buckets.
	if raw.InitFactoid(db); err != nil {
		return err
	}
	return nil
}

// Close Closes a connection to BoltDB when done.
func (c *Client) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
