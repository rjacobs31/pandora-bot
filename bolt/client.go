package bolt

import (
	"encoding/binary"
	"time"

	"github.com/boltdb/bolt"

	pandora ".."
)

var _ pandora.DataClient = &Client{}

// Client A client for connecting to BoltDB.
type Client struct {
	Path string
	Now  func() time.Time
	DB   *bolt.DB

	pandora.FactoidService
	pandora.RawFactoidService
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
	c.RawFactoidService = &RawFactoidService{DB: db}

	// Initialize top-level buckets.
	tx, err := c.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists([]byte(factBucket)); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists([]byte(factTrigBucket)); err != nil {
		return err
	}

	return tx.Commit()
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
