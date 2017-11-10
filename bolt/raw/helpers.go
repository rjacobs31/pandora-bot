package raw

import "encoding/binary"

// ItoB returns an 8-byte big endian representation of v.
func ItoB(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// BtoI returns a uint64 representation of v.
func BtoI(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
