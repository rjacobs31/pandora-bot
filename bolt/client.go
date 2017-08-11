package bolt

import (
	"time"

	"github.com/boltdb/bolt"

	pandora ".."
)

var _ pandora.DataClient = &Client{}

// Client A client for connecting to BoltDB.
type Client struct {
	Path string
	Now  func() time.Time

	pandora.FactoidService
	pandora.RawFactoidService

	db *bolt.DB
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
	c.db = db

	// Initialize top-level buckets.
	tx, err := c.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists([]byte("factoids")); err != nil {
		return err
	}

	return tx.Commit()
}

// Close Closes a connection to BoltDB when done.
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
