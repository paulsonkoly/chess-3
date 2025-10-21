// package checksum provides an opaque and non-mutable checksum type.
package checksum

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"hash"
	"io"
)

// Checksum represents a calculated checksum of some data.
type Checksum struct {
	data [sha256.Size]byte
}

var ErrLengthInvalid = errors.New("length invalid")

// FromBytes converts a slice of bytes to checksum. The byte slice can be
// extracted from an existing checksum with Checksum.Bytes.
func FromBytes(bytes []byte) (Checksum, error) {
	if len(bytes) != sha256.Size {
		return Checksum{}, ErrLengthInvalid
	}
	result := Checksum{}

	copy(result.data[:], bytes)

	return result, nil
}

// ReadFrom calculates a checksum based on data read from r.
func ReadFrom(r io.Reader) (Checksum, error) {
	sha := sha256.New()
	result := Checksum{}

	if _, err := io.Copy(sha, r); err != nil {
		return result, err
	}

	copy(result.data[:], sha.Sum(nil))

	return result, nil
}

// Collector creates checksum incrementally.
type Collector struct {
	sha hash.Hash
}

// NewCollector creates a new Collector.
func NewCollector() Collector { return Collector{sha256.New()} }

// Collect feeds a single line data to c.
func (c *Collector) Collect(line string) error {
	c.sha.Write([]byte(line))
	return nil
}

// Checksum returns the accumulated Checksum in c.
func (c Collector) Checksum() Checksum {
	result := Checksum{}
	copy(result.data[:], c.sha.Sum(nil))
	return result
}

// Bytes returns a byte slice representation of c.
func (c Checksum) Bytes() []byte {
	return c.data[:]
}

// String is a base64 url encoded string representation of c.
func (c Checksum) String() string {
	return base64.URLEncoding.EncodeToString(c.data[:])
}

// Matches determines if c and other are the same checksums.
func (c Checksum) Matches(other Checksum) bool {
	return bytes.Equal(c.data[:], other.data[:])
}
