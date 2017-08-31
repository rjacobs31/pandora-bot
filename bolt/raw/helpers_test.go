package raw_test

import (
	"bytes"
	"testing"

	"github.com/rjacobs31/pandora-bot/bolt/raw"
)

func TestItoB(t *testing.T) {
	testCases := []struct {
		num uint64
		b   []byte
	}{
		{num: 0, b: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{num: 1, b: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{num: 256, b: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00}},
		{num: 512, b: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00}},
		{num: 1023, b: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xff}},
	}

	for k, v := range testCases {
		if result := raw.ItoB(v.num); bytes.Compare(result, v.b) != 0 {
			t.Errorf("[%d] ItoB(%d): expected %v, got %v", k, v.num, v.b, result)
		}
	}
}
