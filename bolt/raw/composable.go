package raw

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"

	"github.com/rjacobs31/pandora-bot/bolt/ftypes"
)

// Bucket names
const (
	FactoidBucketName      = "Factoids"
	TriggerIndexBucketName = "FactoidTriggerIndex"
)

// Error types
var (
	ErrFactoidExists   = errors.New("factoid exists")
	ErrFactoidNotExist = errors.New("no such factoid")
	ErrTriggerExists   = errors.New("trigger exists")
	ErrTriggerNotExist = errors.New("no such trigger")
)

// InitFactoid initialises necessary factoid buckets in given DB
func InitFactoid(db *bolt.DB) (err error) {
	if db == nil {
		err = errors.New("FactoidService: No DB provided")
		return
	}

	// Initialize top-level buckets.
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(FactoidBucketName))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(TriggerIndexBucketName))
		return err
	})
	return err
}

// FactoidBucket Gets the BoltDB bucket for Factoid objects.
func FactoidBucket(tx *bolt.Tx) (b *bolt.Bucket) {
	return tx.Bucket([]byte(FactoidBucketName))
}

// TriggerIndexBucket Gets the ID for Factoid objects with given ID.
func TriggerIndexBucket(tx *bolt.Tx) (b *bolt.Bucket) {
	return tx.Bucket([]byte(TriggerIndexBucketName))
}

/************
 * Factoids *
 ************/

func upgradeResponseMap(pf *ftypes.Factoid) (err error) {
	if pf.DeprecatedResponses == nil || len(pf.DeprecatedResponses) == 0 {
		return
	}

	responseSet := make(map[string]bool, len(pf.DeprecatedResponses)+len(pf.Responses))
	var highest uint64

	for k, v := range pf.Responses {
		responseSet[v.Response] = true
		if k > highest {
			highest = k
		}
	}

	for _, v := range pf.DeprecatedResponses {
		if _, ok := responseSet[v.Response]; !ok {
			responseSet[v.Response] = true
			highest++
			pf.Responses[highest] = v
		}
	}

	pf.DeprecatedResponses = nil
	return
}

// MarshalFactoid Marshals from *pandora.Factoid to protobuf bytes.
func MarshalFactoid(pf *ftypes.Factoid) ([]byte, error) {
	upgradeResponseMap(pf)
	return proto.Marshal(pf)
}

// UnmarshalFactoid Unmarshals from protobuf bytes to *pandora.Factoid.
func UnmarshalFactoid(b []byte) (*ftypes.Factoid, error) {
	pf := &ftypes.Factoid{}
	err := proto.Unmarshal(b, pf)
	if err != nil {
		return nil, err
	}
	upgradeResponseMap(pf)
	return pf, nil
}

// FactoidExists checks the factoid bucket for existence of a Factoid with
// given ID.
func FactoidExists(tx *bolt.Tx, id uint64) (ok bool) {
	if b := FactoidBucket(tx).Get(ItoB(id)); b != nil && len(b) > 0 {
		ok = true
	}
	return
}

// FactoidExistsByTrigger checks trigger index bucket for Factoid with a given
// trigger.
func FactoidExistsByTrigger(tx *bolt.Tx, trigger string) (ok bool) {
	if v := TriggerIndexBucket(tx).Get([]byte(trigger)); v != nil && len(v) > 0 {
		ok = true
	}
	return
}

// FetchFactoid retreives a Factoid by ID.
func FetchFactoid(tx *bolt.Tx, id uint64) (*ftypes.Factoid, error) {
	f := FactoidBucket(tx).Get(ItoB(id))
	if f == nil || len(f) == 0 {
		return nil, nil
	}
	return UnmarshalFactoid(f)
}

// FetchFactoidByTrigger retreives a Factoid by trigger.
func FetchFactoidByTrigger(tx *bolt.Tx, trigger string) (*ftypes.Factoid, error) {
	id := TriggerIndexBucket(tx).Get([]byte(trigger))
	if id == nil || len(id) == 0 {
		return nil, nil
	}
	f := FactoidBucket(tx).Get(id)
	if f == nil || len(f) == 0 {
		return nil, nil
	}
	return UnmarshalFactoid(f)
}

// FactoidIDByTrigger retreives a Factoid by trigger.
func FactoidIDByTrigger(tx *bolt.Tx, trigger string) []byte {
	return TriggerIndexBucket(tx).Get([]byte(trigger))
}

// DeleteFactoid deletes a factoid by ID.
func DeleteFactoid(tx *bolt.Tx, id uint64) (f *ftypes.Factoid, err error) {
	f, err = FetchFactoid(tx, id)
	if err != nil || f == nil {
		return nil, ErrFactoidNotExist
	}
	TriggerIndexBucket(tx).Delete([]byte(f.Trigger))
	FactoidBucket(tx).Delete(ItoB(id))
	return
}

// CreateFactoid creates a Factoid in DB and returns ID.
func CreateFactoid(tx *bolt.Tx, f *ftypes.Factoid) (id uint64, err error) {
	factoidBucket := FactoidBucket(tx)
	triggerBucket := TriggerIndexBucket(tx)
	// if factoid doesn't have an ID, generate one
	if existingID := triggerBucket.Get([]byte(f.Trigger)); existingID != nil && len(existingID) > 0 {
		return BtoI(existingID), ErrFactoidExists
	}
	id, _ = factoidBucket.NextSequence()
	f.ID = id
	b, err := MarshalFactoid(f)
	if err != nil {
		return
	}
	err = factoidBucket.Put(ItoB(id), b)
	if err != nil {
		return
	}
	err = triggerBucket.Put([]byte(f.Trigger), ItoB(id))
	return
}

// PutFactoid creates a factoid if it doesn't exist or updates it by ID if it
// does exist.
func PutFactoid(tx *bolt.Tx, f *ftypes.Factoid) (id uint64, err error) {
	if FactoidExists(tx, f.ID) {
		id, err = UpdateFactoid(tx, f)
	} else {
		id, err = CreateFactoid(tx, f)
	}
	return
}

// PutFactoidByTrigger creates a factoid if it doesn't exist or updates it by trigger if
// it does exist.
func PutFactoidByTrigger(tx *bolt.Tx, f *ftypes.Factoid) (id uint64, err error) {
	if FactoidExistsByTrigger(tx, f.Trigger) {
		id, err = UpdateFactoidByTrigger(tx, f)
	} else {
		id, err = CreateFactoid(tx, f)
	}
	return
}

// UpdateFactoid updates a factoid by ID.
func UpdateFactoid(tx *bolt.Tx, f *ftypes.Factoid) (id uint64, err error) {
	factoidBucket := FactoidBucket(tx)
	triggerBucket := TriggerIndexBucket(tx)
	id = f.ID
	if id == 0 {
		return 0, ErrFactoidNotExist
	}
	if existingID := triggerBucket.Get([]byte(f.Trigger)); BtoI(existingID) != f.ID {
		return 0, ErrTriggerExists
	}
	b, err := MarshalFactoid(f)
	if err != nil {
		return
	}
	err = factoidBucket.Put(ItoB(id), b)
	if err != nil {
		return
	}
	err = triggerBucket.Put([]byte(f.Trigger), ItoB(id))
	return
}

// UpdateFactoidByTrigger updates a factoid by trigger.
func UpdateFactoidByTrigger(tx *bolt.Tx, f *ftypes.Factoid) (id uint64, err error) {
	factoidBucket := FactoidBucket(tx)
	triggerBucket := TriggerIndexBucket(tx)
	if existingID := triggerBucket.Get([]byte(f.Trigger)); BtoI(existingID) == 0 {
		return 0, ErrTriggerNotExist
	}
	b, err := MarshalFactoid(f)
	if err != nil {
		return
	}
	err = factoidBucket.Put(ItoB(id), b)
	if err != nil {
		return
	}
	err = triggerBucket.Put([]byte(f.Trigger), ItoB(id))
	return
}
