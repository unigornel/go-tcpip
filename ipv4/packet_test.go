package ipv4

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/hverr/go-testutils"
	"github.com/stretchr/testify/assert"
)

var addresses = []struct {
	Address Address
	String  string
	Valid   bool
}{
	{[4]byte{192, 168, 100, 1}, "192.168.100.1", true},
	{[4]byte{}, "not an address", false},
	{[4]byte{}, "::1", false},
}

func TestAddress(t *testing.T) {
	for i, test := range addresses {
		a, ok := NewAddress(test.String)
		if !test.Valid {
			assert.False(t, ok, "Parsed invalid IP %d", i)
		} else {
			assert.True(t, ok, "Could not parse IP %d", i)
			assert.True(t, reflect.DeepEqual(test.Address, a), "Could not parse IP %d", i)
			assert.Equal(t, test.String, a.String(), "Could not convert IP %d to string", i)
		}
	}
}

var headers = []struct {
	Bytes     string
	RawHeader RawHeader
	Header    Header
}{
	{
		"450000349c8340003e0656dbc0a86401c0a86413",
		RawHeader{
			VersionIHL:          0x45,
			ToS:                 0x00,
			TotalLength:         52,
			Identification:      40067,
			FlagsFragmentOffset: 0x4000,
			TTL:                 62,
			Protocol:            6,
			Checksum:            0x56db,
			Source:              [4]byte{192, 168, 100, 1},
			Destination:         [4]byte{192, 168, 100, 19},
		},
		Header{
			Version:        4,
			IHL:            5,
			ToS:            0,
			TotalLength:    52,
			Identification: 40067,
			Flags:          0x02,
			FragmentOffset: 0,
			TTL:            62,
			Protocol:       6,
			Checksum:       0x56db,
			Source:         [4]byte{192, 168, 100, 1},
			Destination:    [4]byte{192, 168, 100, 19},
		},
	},
}

func TestRawHeader(t *testing.T) {
	for i, test := range headers {
		b, err := hex.DecodeString(test.Bytes)
		assert.Nil(t, err)

		h, err := NewRawHeader(bytes.NewReader(b))
		assert.Nil(t, err, "Cannot read raw header %d", i)
		assert.True(t, reflect.DeepEqual(h, test.RawHeader), "Cannot read raw header %d", i)

		o := bytes.NewBuffer(nil)
		err = h.Write(o)
		assert.Nil(t, err, "Cannot write raw header %d", i)
		s := hex.EncodeToString(o.Bytes())
		assert.Equal(t, test.Bytes, s, "Cannot write raw header %d", i)
	}
}

func TestHeader(t *testing.T) {
	for i, test := range headers {
		b, err := hex.DecodeString(test.Bytes)
		assert.Nil(t, err)

		h, err := NewHeader(bytes.NewReader(b))
		assert.Nil(t, err, "Cannot read header %d", i)
		assert.True(t, reflect.DeepEqual(h, test.Header), "Cannot read header %d", i)

		o := bytes.NewBuffer(nil)
		err = h.Write(o)
		assert.Nil(t, err, "Cannot write header %d", i)
		s := hex.EncodeToString(o.Bytes())
		assert.Equal(t, test.Bytes, s, "Cannot write header %d", i)
	}

	// A header with options.
	{
		test := headers[0]
		test.Header.IHL = 6
		test.Header.Options = []byte{0, 1, 2, 3, 4, 5, 6, 7}
		test.Bytes += "0001020304050607"
		test.Bytes = test.Bytes[:1] + "6" + test.Bytes[2:]

		b, err := hex.DecodeString(test.Bytes)
		assert.Nil(t, err)

		h, err := NewHeader(bytes.NewReader(b))
		assert.Nil(t, err, "Cannot read header with options")
		assert.True(
			t,
			reflect.DeepEqual(h, test.Header),
			"Cannot read header with options: %v != %v)",
			test.Header, h,
		)

		o := bytes.NewBuffer(nil)
		err = h.Write(o)
		assert.Nil(t, err, "Cannot write header with options")
		s := hex.EncodeToString(o.Bytes())
		assert.Equal(t, test.Bytes, s, "Cannot write header with options")
	}

	// With an invalid IHL field.
	{
		test := headers[0]
		test.Header.IHL = 4
		test.Bytes = test.Bytes[:1] + "4" + test.Bytes[2:]

		b, err := hex.DecodeString(test.Bytes)
		assert.Nil(t, err)

		_, err = NewHeader(bytes.NewReader(b))
		assert.Equal(t, ErrInvalidIHL, err)
	}

	// With an invalid reader.
	{
		_, err := NewHeader(testutils.NewErrorReader())
		assert.Equal(t, testutils.ErrorReaderDefaultError, err)
	}

	// With an invalid writer.
	{
		w := testutils.NewErrorWriter()
		err := headers[0].Header.Write(w)
		assert.Equal(t, testutils.ErrorWriterDefaultError, err)
	}
}

func TestHeaderConversion(t *testing.T) {
	for i, test := range headers {
		h := test.RawHeader.Header()
		assert.True(t, reflect.DeepEqual(test.Header, h), "Could not convert raw to header %d", i)

		r := test.Header.RawHeader()
		assert.True(t, reflect.DeepEqual(test.RawHeader, r), "COuld not convert header to raw %d", i)
	}
}

func TestHeaderCheck(t *testing.T) {
	for i, test := range headers {
		assert.Nil(t, test.Header.Check(), "Header check %d failed", i)
	}

	validIHL := []Header{
		{IHL: 6, Options: []byte{0x1, 0x2, 0x3, 0x4}},
	}
	for i, test := range validIHL {
		test.Checksum = test.CalculateChecksum()
		assert.Nil(t, test.Check(), "Header check %d failed", i)
	}

	invalidIHL := []Header{
		{IHL: 6, Options: nil},
		{IHL: 7, Options: []byte{0x1, 0x2, 0x3, 0x4}},
		{IHL: 6, Options: []byte{0x1}},
	}
	for i, test := range invalidIHL {
		test.Checksum = test.CalculateChecksum()
		assert.Equal(t, ErrInvalidIHL, test.Check(), "Header check %d failed", i)
	}

	for i, test := range headers {
		test.Header.Checksum = ^test.Header.Checksum
		assert.Equal(t, ErrInvalidChecksum, test.Header.Check(), "Header check %d failed", i)
	}
}
